package direct

import (
	"errors"
	"fmt"
	"github.com/madokast/direct/memory"
	"github.com/madokast/direct/memory/trace_type"
	"github.com/madokast/direct/utils"
	"reflect"
	"strings"
	"sync"
)

type Map[Key comparable, Value any] memory.Pointer

var userDefinedHashEqualFuncSetMu sync.Mutex
var userDefinedHashEqualFuncSet = map[reflect.Value]int{} // funcVal->objRefCnt prevents GC

type mapHeader[Key comparable, Value any] struct {
	table             Slice[entry[Key, Value]] // hashTable + list
	tableLength       SizeType
	tableBasePtr      memory.Pointer
	count             SizeType
	mask              SizeType
	free              SizeType
	hash              func(Key) SizeType
	equal             func(Key, Key) bool
	headerPageHandler memory.PageHandler // for free
}

type entry[Key comparable, Value any] struct {
	key   Key
	value Value
	next  SizeType
}

const nilMap = 0
const emptyTableFlag = 0 // a slot in table is empty is entry.next = emptyTableFlag
const listTailFlag = 1   // a slot is the tail of list if entry.next = listTailFlag

func MakeMap[Key comparable, Value any](capacity SizeType) (Map[Key, Value], error) {
	if isSimpleType[Key]() {
		return makeCustomMap0[Key, Value](capacity, simpleHash[Key], simpleEqual[Key], 3)
	} else if isString[Key]() {
		return makeCustomMap0[Key, Value](capacity, hashString[Key], equalString[Key], 3)
	} else {
		var k Key
		str := fmt.Sprintf("%T is not simple type. Use MakeCustomMap", k)
		return nilMap, errors.New(str)
	}
}

func MakeMapFromGoMap[Key comparable, Value any](gm map[Key]Value) (m Map[Key, Value], err error) {
	if isSimpleType[Key]() {
		m, err = makeCustomMap0[Key, Value](SizeType(len(gm)), simpleHash[Key], simpleEqual[Key], 3)
	} else if isString[Key]() {
		m, err = makeCustomMap0[Key, Value](SizeType(len(gm)), hashString[Key], equalString[Key], 3)
	} else {
		var k Key
		str := fmt.Sprintf("%T is not simple type. Use MakeCustomMap", k)
		return nilMap, errors.New(str)
	}
	if err != nil {
		return nilMap, err
	}
	for k, v := range gm {
		m.directPutNoGrow(k, v)
	}
	return m, nil
}

func MakeCustomMap[Key comparable, Value any](capacity SizeType, hash func(Key) SizeType, equal func(key1 Key, key2 Key) bool) (Map[Key, Value], error) {
	return makeCustomMap0[Key, Value](capacity, hash, equal, 3)
}

func makeCustomMap0[Key comparable, Value any](capacity SizeType, hash func(Key) SizeType, equal func(Key, Key) bool, traceSkip int) (Map[Key, Value], error) {
	if utils.Asserted {
		if hash == nil {
			panic("hash function is nil")
		}
		if equal == nil {
			panic("equal function is nil")
		}
	}
	var normalizeCap SizeType = 8
	for normalizeCap < capacity {
		normalizeCap <<= 1
	}
	normalizeCap <<= 1 // enlarge once again because a test encounters capacity expansion.

	table, err := makeSliceWithLength0[entry[Key, Value]](normalizeCap, trace_type.MapTable, traceSkip+2) // <<= 1 for link space
	if err != nil {
		return nilMap, err
	}
	pageMapHeader, err := Global.allocPage(1, trace_type.MapHeader, traceSkip)
	if err != nil {
		table.Free()
		return nilMap, err
	}
	var theMap = Map[Key, Value](Global.pagePointerOf(pageMapHeader))
	header := theMap.header()

	header.table = table
	header.tableLength = normalizeCap
	header.tableBasePtr = table.header().elementBasePointer
	header.count = 0
	header.mask = (normalizeCap >> 1) - 1
	header.free = normalizeCap >> 1 // do (>>1) for link
	header.hash = hash
	header.equal = equal
	header.headerPageHandler = pageMapHeader

	header.hashEqualRefCtrl(1)
	return theMap, nil
}

func (m Map[Key, Value]) Put(k Key, v Value) error {
	err := m.checkCapacity(1)
	if err != nil {
		return err
	}
	header := m.header()
	loc := header.hash(k) & header.mask
	slot := header.dataAt(loc)
	next := slot.next
	if next == emptyTableFlag {
		*slot = entry[Key, Value]{
			key:   k,
			value: v,
			next:  listTailFlag, // no next
		}
		header.count++
		return nil
	} else if next == listTailFlag { // only one, no link
		if header.equal(slot.key, k) { // replace
			slot.value = v
		} else { // add link
			slot.next = header.free
			*header.dataAt(header.free) = entry[Key, Value]{
				key:   k,
				value: v,
				next:  listTailFlag,
			}
			header.free++
			header.count++
		}
		return nil
	} else { // has next
		if header.equal(slot.key, k) {
			slot.value = v // replace
			return nil
		} else {
			for {
				slot = header.dataAt(next)
				if header.equal(slot.key, k) { // replace
					slot.value = v
					return nil
				}
				next = slot.next
				if next == listTailFlag { // end
					slot.next = header.free
					*header.dataAt(header.free) = entry[Key, Value]{
						key:   k,
						value: v,
						next:  listTailFlag,
					}
					header.free++
					header.count++
					return nil
				}
			}
		}
	}
}

// DirectPut asserts no replacement
func (m Map[Key, Value]) DirectPut(k Key, v Value) error {
	err := m.checkCapacity(1)
	if err != nil {
		return err
	}
	header := m.header()
	loc := header.hash(k) & header.mask
	slot := header.dataAt(loc)
	next := slot.next
	if next == emptyTableFlag { // empty
		*slot = entry[Key, Value]{
			key:   k,
			value: v,
			next:  listTailFlag, // no next
		}
		header.count++
		return nil
	} else if next == listTailFlag { // only one, no link
		if utils.Asserted {
			if header.equal(slot.key, k) {
				panic(fmt.Sprintf("DirectPut faces duplicated key %v value %v, %v", slot.key, slot.value, v))
			}
		}
		slot.next = header.free
		*header.dataAt(header.free) = entry[Key, Value]{
			key:   k,
			value: v,
			next:  listTailFlag,
		}
		header.free++
		header.count++
		return nil
	} else { // has next
		if utils.Asserted {
			if header.equal(slot.key, k) {
				panic(fmt.Sprintf("DirectPut faces duplicated key %v value %v, %v", slot.key, slot.value, v))
			}
		}
		for {
			slot = header.dataAt(next)
			if utils.Asserted {
				if header.equal(slot.key, k) {
					panic(fmt.Sprintf("DirectPut faces duplicated key %v value %v, %v", slot.key, slot.value, v))
				}
			}
			next = slot.next
			if next == listTailFlag { // end
				slot.next = header.free
				*header.dataAt(header.free) = entry[Key, Value]{
					key:   k,
					value: v,
					next:  listTailFlag,
				}
				header.free++
				header.count++
				return nil
			}
		}
	}
}

// directPutNoGrow asserts no replacement and no grow-capacity
func (m Map[Key, Value]) directPutNoGrow(k Key, v Value) {
	if utils.Asserted {
		if m == nilMap {
			panic("use a moved or freed or null map")
		}
	}
	header := m.header()
	if utils.Asserted {
		if header.table == nullSlice {
			panic("use a moved or freed or null map")
		}
		if header.free == header.tableLength {
			panic("full map calls directPutNoGrow")
		}
	}
	loc := header.hash(k) & header.mask
	slot := header.dataAt(loc)
	next := slot.next
	if next == emptyTableFlag { // empty
		*slot = entry[Key, Value]{
			key:   k,
			value: v,
			next:  listTailFlag, // no next
		}
		header.count++
		return
	} else if next == listTailFlag { // only one, no link
		if utils.Asserted {
			if header.equal(slot.key, k) {
				panic(fmt.Sprintf("DirectPut faces duplicated key %v value %v, %v", slot.key, slot.value, v))
			}
		}
		slot.next = header.free
		*header.dataAt(header.free) = entry[Key, Value]{
			key:   k,
			value: v,
			next:  listTailFlag,
		}
		header.free++
		header.count++
		return
	} else { // has next
		if utils.Asserted {
			if header.equal(slot.key, k) {
				panic(fmt.Sprintf("DirectPut faces duplicated key %v value %v, %v", slot.key, slot.value, v))
			}
		}
		for {
			slot = header.dataAt(next)
			if utils.Asserted {
				if header.equal(slot.key, k) {
					panic(fmt.Sprintf("DirectPut faces duplicated key %v value %v, %v", slot.key, slot.value, v))
				}
			}
			next = slot.next
			if next == listTailFlag { // end
				slot.next = header.free
				*header.dataAt(header.free) = entry[Key, Value]{
					key:   k,
					value: v,
					next:  listTailFlag,
				}
				header.free++
				header.count++
				return
			}
		}
	}
}

func (m Map[Key, Value]) Get2(k Key) (val Value, ok bool) {
	if utils.Asserted {
		if m == nilMap {
			panic("use a moved or freed or null map")
		}
	}
	header := m.header()
	if utils.Asserted {
		if header.table == nullSlice {
			panic("use a moved or freed or null map")
		}
	}
	loc := header.hash(k) & header.mask
	slot := header.dataAt(loc)
	next := slot.next
	if next == emptyTableFlag {
		return val, false
	} else if next == listTailFlag {
		if header.equal(slot.key, k) {
			return slot.value, true
		} else {
			return val, false
		}
	} else {
		if header.equal(slot.key, k) {
			return slot.value, true
		} else {
			for {
				slot = header.dataAt(next)
				if header.equal(slot.key, k) {
					return slot.value, true
				}
				next = slot.next
				if next == listTailFlag {
					return val, false
				}
			}
		}
	}
}

func (m Map[Key, Value]) Get(k Key) (val Value) {
	if utils.Asserted {
		if m == nilMap {
			panic("use a moved or freed or null map")
		}
	}
	header := m.header()
	if utils.Asserted {
		if header.table == nullSlice {
			panic("use a moved or freed or null map")
		}
	}
	loc := header.hash(k) & header.mask
	slot := header.dataAt(loc)
	next := slot.next
	if next == emptyTableFlag {
		return val
	} else if next == listTailFlag {
		if header.equal(slot.key, k) {
			return slot.value
		} else {
			return val
		}
	} else {
		if header.equal(slot.key, k) {
			return slot.value
		} else {
			for {
				slot = header.dataAt(next)
				if header.equal(slot.key, k) {
					return slot.value
				}
				next = slot.next
				if next == listTailFlag {
					return val
				}
			}
		}
	}
}

func (m Map[Key, Value]) Delete(k Key) {
	if utils.Asserted {
		if m == nilMap {
			panic("use a moved or freed or null map")
		}
	}
	header := m.header()
	if utils.Asserted {
		if header.table == nullSlice {
			panic("use a moved or freed or null map")
		}
	}
	loc := header.hash(k) & header.mask
	slot := header.dataAt(loc)
	next := slot.next
	if next == emptyTableFlag {
		return
	} else if next == listTailFlag {
		if header.equal(slot.key, k) {
			slot.next = emptyTableFlag
			header.count--
			return
		}
	} else {
		if header.equal(slot.key, k) {
			*slot = *header.dataAt(next)
			header.count--
			return
		} else {
			var targetSlot *entry[Key, Value] = nil
			var last2Slot *entry[Key, Value] = nil
			for {
				last2Slot = slot
				slot = header.dataAt(next)
				if targetSlot == nil && header.equal(slot.key, k) {
					targetSlot = slot
				}
				next = slot.next
				if next == listTailFlag {
					break
				}
			}
			if targetSlot != nil {
				if last2Slot.next+1 == header.free {
					header.free--
				}
				if targetSlot == slot {
					last2Slot.next = listTailFlag
				} else {
					targetSlot.key = slot.key
					targetSlot.value = slot.value
					last2Slot.next = listTailFlag
				}
				header.count--
			}
			return
		}
	}
}

func (m Map[Key, Value]) Length() int {
	if utils.Asserted {
		if m == nilMap {
			panic("use a moved or freed or null map")
		}
	}
	header := m.header()
	return int(header.count)
}

func (mh *mapHeader[Key, Value]) dataAt(index SizeType) *entry[Key, Value] {
	if utils.Asserted {
		if index >= mh.tableLength {
			panic(fmt.Sprintf("Map table out of bound(%d, %d)", index, mh.tableLength))
		}
	}
	return memory.PointerAs[entry[Key, Value]](mh.tableBasePtr + memory.Pointer(index*memory.Sizeof[entry[Key, Value]]()))
}

func (mh *mapHeader[Key, Value]) hashEqualRefCtrl(cnt int) {
	hashVal := reflect.ValueOf(mh.hash)
	equalVal := reflect.ValueOf(mh.equal)

	userDefinedHashEqualFuncSetMu.Lock()
	{
		hashCnt := userDefinedHashEqualFuncSet[hashVal] + cnt
		if hashCnt <= 0 {
			if utils.Asserted {
				if hashCnt < 0 {
					panic("bad code: hashCnt < 0")
				}
			}
			delete(userDefinedHashEqualFuncSet, hashVal)
		} else {
			userDefinedHashEqualFuncSet[hashVal] = hashCnt
		}
	}
	{
		equalCnt := userDefinedHashEqualFuncSet[equalVal] + cnt
		if equalCnt <= 0 {
			if utils.Asserted {
				if equalCnt < 0 {
					panic("bad code: equalCnt < 0")
				}
			}
			delete(userDefinedHashEqualFuncSet, equalVal)
		} else {
			userDefinedHashEqualFuncSet[equalVal] = equalCnt
		}
	}
	userDefinedHashEqualFuncSetMu.Unlock()
}

func (m Map[Key, Value]) GoMap() map[Key]Value {
	if utils.Asserted {
		if m == nilMap {
			panic("use a moved or freed or null map")
		}
	}
	header := m.header()
	gm := make(map[Key]Value, header.count)
	m.Iterate(func(k Key, v Value) {
		gm[k] = v
	})
	return gm
}

func (m *Map[Key, Value]) Move() (moved Map[Key, Value]) {
	if utils.Asserted {
		if *m == nilMap {
			panic("use a moved or freed or null map")
		}
	}
	moved = *m
	*m = nilMap
	return moved
}

func (m Map[Key, Value]) Moved() bool {
	return m == nilMap
}

// Free in defer use
// :: defer func() {m.Free(&m)}()
// instead of
// :: m.Free(&m)
func (m Map[Key, Value]) Free() {
	if m.pointer().IsNotNull() {
		header := m.header()
		if header.headerPageHandler.IsNull() {
			panic("double free?")
		}

		header.hashEqualRefCtrl(-1)

		header.table.Free()
		Global.freePage(header.headerPageHandler)
	}
}

func (m Map[Key, Value]) String() string {
	if m.IsNull() {
		return "NilMap"
	}
	return fmt.Sprintf("%v", m.GoMap())
}

func (m Map[Key, Value]) debugString() string {
	if m.IsNull() {
		return "NilMap"
	}
	header := m.header()
	var sb strings.Builder
	header.table.IterateRefIndexBreakable(func(index SizeType, slot *entry[Key, Value]) bool {
		if index > header.mask {
			return false
		}
		next := slot.next
		if next != emptyTableFlag {
			sb.WriteString(fmt.Sprintf("(%d)[%v:%v", index, slot.key, slot.value))
			for next != listTailFlag {
				slot = header.dataAt(next)
				sb.WriteString(fmt.Sprintf("->%v:%v", slot.key, slot.value))
				next = slot.next
			}
			sb.WriteString("]\n")
		}
		return true
	})
	s := sb.String()
	if len(s) > 0 {
		if utils.Asserted {
			if s[len(s)-1] != '\n' {
				panic("wrong string " + s)
			}
		}
		s = s[:len(s)-1]
	}
	return s
}

func (m Map[Key, Value]) checkCapacity(appendSize SizeType) error {
	if utils.Asserted {
		if m == nilMap {
			panic("use a moved or freed or null map")
		}
	}
	header := m.header()
	if header.free+appendSize > header.tableLength {
		if utils.Debug {
			fmt.Println("map capacity expense")
		}
		// new table
		var newTable Slice[entry[Key, Value]]
		var newNormalizeCap SizeType = 8
		{
			var err error
			targetCapacity := header.count + appendSize
			for newNormalizeCap < targetCapacity {
				newNormalizeCap <<= 1
			}
			newNormalizeCap <<= 1 // enlarge once again because a test encounters capacity expansion.

			newTable, err = makeSliceWithLength0[entry[Key, Value]](newNormalizeCap, trace_type.MapTable, 5)
			if err != nil {
				return err
			}
		}

		// put
		{
			oldTable := header.table
			oldMask := header.mask
			oldTableBasePtr := header.tableBasePtr

			header.table = newTable
			header.tableLength = newNormalizeCap
			header.tableBasePtr = newTable.header().elementBasePointer
			header.count = 0
			header.mask = (newNormalizeCap >> 1) - 1
			header.free = newNormalizeCap >> 1

			oldTable.IterateRefIndexBreakable(func(index SizeType, slot *entry[Key, Value]) bool {
				if index > oldMask {
					return false
				}
				next := slot.next
				if next != emptyTableFlag {
					m.directPutNoGrow(slot.key, slot.value)
					for next != listTailFlag {
						slot = memory.PointerAs[entry[Key, Value]](oldTableBasePtr + memory.Pointer(next*memory.Sizeof[entry[Key, Value]]()))
						m.directPutNoGrow(slot.key, slot.value)
						next = slot.next
					}
				}
				return true
			})
			oldTable.Free()
		}
	}
	return nil
}

func (m Map[Key, Value]) pointer() memory.Pointer {
	return memory.Pointer(m)
}

func (m Map[Key, Value]) IsNull() bool {
	return m.pointer().IsNull()
}

func (m Map[Key, Value]) header() *mapHeader[Key, Value] {
	if utils.Asserted {
		if m.pointer().IsNull() {
			panic("header of null")
		}
	}
	return memory.PointerAs[mapHeader[Key, Value]](m.pointer())
}
