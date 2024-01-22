package managed_memory

import (
	"errors"
	"fmt"
	"gitlab.grandhoo.com/rock/storage/internal/logger"
	"reflect"
	"strings"
	"unsafe"
)

type Map[Key comparable, Value any] struct {
	table        Slice[entry[Key, Value]] // Table + list
	tableLength  SizeType
	tableBasePtr pointer
	count        SizeType
	mask         SizeType
	free         SizeType
	hash         func(Key) SizeType
	equal        func(Key, Key) bool
	memory       *LocalMemory
}

type entry[Key comparable, Value any] struct {
	key   Key
	value Value
	next  SizeType
}

const emptyTableFlag = 0 // a slot in table is empty is entry.next = emptyTableFlag
const listTailFlag = 1   // a slot is the tail of list if entry.next = listTailFlag

func MakeMap[Key comparable, Value any](capacity SizeType, m *LocalMemory) (*Map[Key, Value], error) {
	if isSimpleType[Key]() {
		return MakeCustomMap[Key, Value](capacity, simpleHash[Key], simpleEqual[Key], m)
	} else if isString[Key]() {
		return MakeCustomMap[Key, Value](capacity, hashString[Key], equalString[Key], m)
	} else {
		var k Key
		str := fmt.Sprintf("%T is not simle type. Use MakeCustomMap", k)
		logger.Error(str)
		return nil, errors.New(str)
	}
}

func MakeCustomMap[Key comparable, Value any](capacity SizeType, hash func(Key) SizeType, equal func(Key, Key) bool, m *LocalMemory) (*Map[Key, Value], error) {
	if asserted {
		if hash == nil {
			logger.Panic("hash function is nil")
		}
		if equal == nil {
			logger.Panic("equal function is nil")
		}
	}
	var normalizeCap SizeType = 8
	for normalizeCap < capacity {
		normalizeCap <<= 1
	}
	normalizeCap <<= 1 // enlarge once again because a test encounters capacity expansion.

	table, err := MakeSliceWithLength[entry[Key, Value]](m, normalizeCap) // <<= 1 for link space
	if err != nil {
		return nil, err
	}
	return &Map[Key, Value]{
		table:        table,
		tableLength:  normalizeCap,
		tableBasePtr: table.header().elementBasePointer,
		count:        0,
		mask:         (normalizeCap >> 1) - 1,
		free:         normalizeCap >> 1, // do (>>1) for link
		hash:         hash,
		equal:        equal,
		memory:       m,
	}, nil
}

func (m *Map[Key, Value]) Put(k Key, v Value) error {
	err := m.checkCapacity(1)
	if err != nil {
		logger.Error(err)
		return err
	}
	loc := m.hash(k) & m.mask
	slot := m.dataAt(loc)
	next := slot.next
	if next == emptyTableFlag {
		*slot = entry[Key, Value]{
			key:   k,
			value: v,
			next:  listTailFlag, // no next
		}
		m.count++
		return nil
	} else if next == listTailFlag { // only one, no link
		if m.equal(slot.key, k) { // replace
			slot.value = v
		} else { // add link
			slot.next = m.free
			*m.dataAt(m.free) = entry[Key, Value]{
				key:   k,
				value: v,
				next:  listTailFlag,
			}
			m.free++
			m.count++
		}
		return nil
	} else { // has next
		if m.equal(slot.key, k) {
			slot.value = v // replace
			return nil
		} else {
			for {
				slot = m.dataAt(next)
				if m.equal(slot.key, k) { // replace
					slot.value = v
					return nil
				}
				next = slot.next
				if next == listTailFlag { // end
					slot.next = m.free
					*m.dataAt(m.free) = entry[Key, Value]{
						key:   k,
						value: v,
						next:  listTailFlag,
					}
					m.free++
					m.count++
					return nil
				}
			}
		}
	}
}

func (m *Map[Key, Value]) DirectPut(k Key, v Value) error {
	err := m.checkCapacity(1)
	if err != nil {
		logger.Error(err)
		return err
	}
	loc := m.hash(k) & m.mask
	slot := m.dataAt(loc)
	next := slot.next
	if next == emptyTableFlag { // empty
		*slot = entry[Key, Value]{
			key:   k,
			value: v,
			next:  listTailFlag, // no next
		}
		m.count++
		return nil
	} else if next == listTailFlag { // only one, no link
		if asserted {
			if m.equal(slot.key, k) {
				logger.Panic(fmt.Sprintf("DirectPut faces duplicated key %v value %v, %v", slot.key, slot.value, v))
			}
		}
		slot.next = m.free
		*m.dataAt(m.free) = entry[Key, Value]{
			key:   k,
			value: v,
			next:  listTailFlag,
		}
		m.free++
		m.count++
		return nil
	} else { // has next
		if asserted {
			if m.equal(slot.key, k) {
				logger.Panic(fmt.Sprintf("DirectPut faces duplicated key %v value %v, %v", slot.key, slot.value, v))
			}
		}
		for {
			slot = m.dataAt(next)
			if asserted {
				if m.equal(slot.key, k) {
					logger.Panic(fmt.Sprintf("DirectPut faces duplicated key %v value %v, %v", slot.key, slot.value, v))
				}
			}
			next = slot.next
			if next == listTailFlag { // end
				slot.next = m.free
				*m.dataAt(m.free) = entry[Key, Value]{
					key:   k,
					value: v,
					next:  listTailFlag,
				}
				m.free++
				m.count++
				return nil
			}
		}
	}
}

func (m *Map[Key, Value]) directPutNoCheck(k Key, v Value) {
	if asserted {
		if m.table == nullSlice {
			logger.Panic("use a moved or freed or null map")
		}
	}
	if asserted {
		if m.free == m.tableLength {
			logger.Panic("full map calls directPutNoCheck")
		}
	}
	loc := m.hash(k) & m.mask
	slot := m.dataAt(loc)
	next := slot.next
	if next == emptyTableFlag { // empty
		*slot = entry[Key, Value]{
			key:   k,
			value: v,
			next:  listTailFlag, // no next
		}
		m.count++
		return
	} else if next == listTailFlag { // only one, no link
		if asserted {
			if m.equal(slot.key, k) {
				logger.Panic(fmt.Sprintf("DirectPut faces duplicated key %v value %v, %v", slot.key, slot.value, v))
			}
		}
		slot.next = m.free
		*m.dataAt(m.free) = entry[Key, Value]{
			key:   k,
			value: v,
			next:  listTailFlag,
		}
		m.free++
		m.count++
		return
	} else { // has next
		if asserted {
			if m.equal(slot.key, k) {
				logger.Panic(fmt.Sprintf("DirectPut faces duplicated key %v value %v, %v", slot.key, slot.value, v))
			}
		}
		for {
			slot = m.dataAt(next)
			if asserted {
				if m.equal(slot.key, k) {
					logger.Panic(fmt.Sprintf("DirectPut faces duplicated key %v value %v, %v", slot.key, slot.value, v))
				}
			}
			next = slot.next
			if next == listTailFlag { // end
				slot.next = m.free
				*m.dataAt(m.free) = entry[Key, Value]{
					key:   k,
					value: v,
					next:  listTailFlag,
				}
				m.free++
				m.count++
				return
			}
		}
	}
}

func (m *Map[Key, Value]) Get2(k Key) (val Value, ok bool) {
	if asserted {
		if m.table == nullSlice {
			logger.Panic("use a moved or freed or null map")
		}
	}
	loc := m.hash(k) & m.mask
	slot := m.dataAt(loc)
	next := slot.next
	if next == emptyTableFlag {
		return val, false
	} else if next == listTailFlag {
		if m.equal(slot.key, k) {
			return slot.value, true
		} else {
			return val, false
		}
	} else {
		if m.equal(slot.key, k) {
			return slot.value, true
		} else {
			for {
				slot = m.dataAt(next)
				if m.equal(slot.key, k) {
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

func (m *Map[Key, Value]) Get(k Key) (val Value) {
	if asserted {
		if m.table == nullSlice {
			logger.Panic("use a moved or freed or null map")
		}
	}
	loc := m.hash(k) & m.mask
	slot := m.dataAt(loc)
	next := slot.next
	//switch next {
	//case emptyTableFlag:
	//	return val
	//case listTailFlag:
	//	if m.equal(slot.key, k) {
	//		return slot.value
	//	} else {
	//		return val
	//	}
	//default:
	//	if m.equal(slot.key, k) {
	//		return slot.value
	//	} else {
	//		for {
	//			slot = m.dataAt(next)
	//			if m.equal(slot.key, k) {
	//				return slot.value
	//			}
	//			next = slot.next
	//			if next == listTailFlag {
	//				return val
	//			}
	//		}
	//	}
	//}

	if next == emptyTableFlag {
		return val
	} else if next == listTailFlag {
		if m.equal(slot.key, k) {
			return slot.value
		} else {
			return val
		}
	} else {
		if m.equal(slot.key, k) {
			return slot.value
		} else {
			for {
				slot = m.dataAt(next)
				if m.equal(slot.key, k) {
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

func (m *Map[Key, Value]) Delete(k Key) {
	if asserted {
		if m.table == nullSlice {
			logger.Panic("use a moved or freed or null map")
		}
	}
	loc := m.hash(k) & m.mask
	slot := m.dataAt(loc)
	next := slot.next
	if next == emptyTableFlag {
		return
	} else if next == listTailFlag {
		if m.equal(slot.key, k) {
			slot.next = emptyTableFlag
			m.count--
			return
		}
	} else {
		if m.equal(slot.key, k) {
			*slot = *m.dataAt(next)
			m.count--
			return
		} else {
			var targetSlot *entry[Key, Value] = nil
			var last2Slot *entry[Key, Value] = nil
			for {
				last2Slot = slot
				slot = m.dataAt(next)
				if targetSlot == nil && m.equal(slot.key, k) {
					targetSlot = slot
				}
				next = slot.next
				if next == listTailFlag {
					break
				}
			}
			if targetSlot != nil {
				if last2Slot.next+1 == m.free {
					m.free--
				}
				if targetSlot == slot {
					last2Slot.next = listTailFlag
				} else {
					targetSlot.key = slot.key
					targetSlot.value = slot.value
					last2Slot.next = listTailFlag
				}
				m.count--
			}
			return
		}
	}
}

func (m *Map[Key, Value]) Length() int {
	if asserted {
		if m.table == nullSlice {
			logger.Panic("use a moved or freed or null map")
		}
	}
	return int(m.count)
}

func (m *Map[Key, Value]) dataAt(index SizeType) *entry[Key, Value] {
	if asserted {
		if m.table == nullSlice {
			logger.Panic("use a moved or freed or null map")
		}
		if index >= m.tableLength {
			logger.Panic(fmt.Sprintf("Map table out of bound(%d, %d)", index, m.tableLength))
		}
	}
	return pointerAs[entry[Key, Value]](m.tableBasePtr + pointer(index*Sizeof[entry[Key, Value]]()))
}

func (m *Map[Key, Value]) Iterate(iter func(Key, Value)) {
	if asserted {
		if m.table == nullSlice {
			logger.Panic("use a moved or freed or null map")
		}
	}
	m.table.IterateRefIndexBreakable(func(index SizeType, slot *entry[Key, Value]) bool {
		if index > m.mask {
			return false
		}
		next := slot.next
		if next != emptyTableFlag {
			iter(slot.key, slot.value)
			for next != listTailFlag {
				slot = m.dataAt(next)
				iter(slot.key, slot.value)
				next = slot.next
			}
		}
		return true
	})
}

func (m *Map[Key, Value]) IterateBreakable(iter func(Key, Value) (_continue_ bool)) {
	if asserted {
		if m.table == nullSlice {
			logger.Panic("use a moved or freed or null map")
		}
	}
	m.table.IterateRefIndexBreakable(func(index SizeType, slot *entry[Key, Value]) bool {
		if index > m.mask {
			return false
		}
		next := slot.next
		if next != emptyTableFlag {
			if !iter(slot.key, slot.value) {
				return false
			}
			for next != listTailFlag {
				slot = m.dataAt(next)
				if !iter(slot.key, slot.value) {
					return false
				}
				next = slot.next
			}
		}
		return true
	})
}

func (m *Map[Key, Value]) GoMap() map[Key]Value {
	if asserted {
		if m.table == nullSlice {
			logger.Panic("use a moved or freed or null map")
		}
	}
	gm := make(map[Key]Value, m.count)
	m.Iterate(func(k Key, v Value) {
		gm[k] = v
	})
	return gm
}

func (m *Map[Key, Value]) Move() (moved *Map[Key, Value]) {
	if asserted {
		if m.memory == nil {
			logger.Panic("move a null map")
		}
	}
	moved = new(Map[Key, Value])
	*moved = *m
	m.memory = nil
	if asserted {
		m.table = nullSlice
		m.tableBasePtr = nullPointer
	}
	return moved
}

func (m *Map[Key, Value]) Moved() bool {
	return m.memory == nil
}

func (m *Map[Key, Value]) Free() {
	if m.memory != nil {
		if asserted {
			if m.table == nullSlice {
				logger.Panic("double free map")
			}
			if m.tableBasePtr == nullSlice {
				logger.Panic("double free map")
			}
		}
		m.table.Free(m.memory)
		if asserted {
			m.table = nullSlice
			m.tableBasePtr = nullPointer
		}
	}
}

func (m *Map[Key, Value]) String() string {
	return fmt.Sprintf("%v", m.GoMap())
}

func (m *Map[Key, Value]) debugString() string {
	if asserted {
		if m.table == nullSlice {
			logger.Panic(fmt.Sprintf("use a moved or freed or null map"))
		}
	}
	var sb strings.Builder
	m.table.IterateRefIndexBreakable(func(index SizeType, slot *entry[Key, Value]) bool {
		if index > m.mask {
			return false
		}
		next := slot.next
		if next != emptyTableFlag {
			sb.WriteString(fmt.Sprintf("(%d)[%v:%v", index, slot.key, slot.value))
			for next != listTailFlag {
				slot = m.dataAt(next)
				sb.WriteString(fmt.Sprintf("->%v:%v", slot.key, slot.value))
				next = slot.next
			}
			sb.WriteString("]\n")
		}
		return true
	})
	s := sb.String()
	if len(s) > 0 {
		if asserted {
			if s[len(s)-1] != '\n' {
				logger.Panic("wrong string", s)
			}
		}
		s = s[:len(s)-1]
	}
	return s
}

func (m *Map[Key, Value]) checkCapacity(appendSize SizeType) error {
	if asserted {
		if m.table == nullSlice {
			logger.Panic("use a moved or freed or null map")
		}
	}
	if m.free+appendSize > m.tableLength {
		if debug {
			logger.Debug("map capacity expense")
		}
		m2, err := MakeCustomMap[Key, Value](m.count+appendSize, m.hash, m.equal, m.memory)
		if err != nil {
			return err
		}
		m.Iterate(m2.directPutNoCheck)
		m.Free()
		*m = *m2
	}
	return nil
}

type MapIterator[Key comparable, Value any] struct {
	tableIter    SliceIterator[entry[Key, Value]]
	currentSlot  *entry[Key, Value]
	tableBasePtr pointer
	noCopy       noCopy
}

func (m *Map[Key, Value]) Iterator() (iter MapIterator[Key, Value]) {
	if asserted {
		if m.table == nullSlice {
			logger.Panic(fmt.Sprintf("use a moved or freed or null map"))
		}
	}
	return MapIterator[Key, Value]{
		tableIter: SliceIterator[entry[Key, Value]]{
			cur:   m.tableBasePtr - pointer(Sizeof[entry[Key, Value]]()),
			end:   m.tableBasePtr + pointer((m.mask+1)*Sizeof[entry[Key, Value]]()),
			index: sizeTypeMax,
		},
		currentSlot:  nil,
		tableBasePtr: m.tableBasePtr,
	}
}

func (mi *MapIterator[Key, Value]) Next() bool {
	currentSlot := mi.currentSlot
	if currentSlot != nil { // next slot in list
		next := currentSlot.next
		if next != listTailFlag {
			mi.currentSlot = pointerAs[entry[Key, Value]](mi.tableBasePtr + pointer(next*Sizeof[entry[Key, Value]]()))
			return true
		}
		// else find in table
	}

	for mi.tableIter.Next() {
		currentSlot = mi.tableIter.Ref()
		if currentSlot.next != emptyTableFlag {
			mi.currentSlot = currentSlot
			return true
		}
	}
	if asserted {
		mi.currentSlot = nil
	}
	return false
}

func (mi *MapIterator[Key, Value]) Key() Key {
	return *mi.KeyRef()
}

func (mi *MapIterator[Key, Value]) Value() Value {
	return *mi.ValueRef()
}

func (mi *MapIterator[Key, Value]) KeyRef() *Key {
	if asserted {
		if mi.currentSlot == nil {
			logger.Panic("access iter-map before check Next()")
		}
	}
	return &mi.currentSlot.key
}

func (mi *MapIterator[Key, Value]) ValueRef() *Value {
	if asserted {
		if mi.currentSlot == nil {
			logger.Panic("access iter-map before check Next()")
		}
	}
	return &mi.currentSlot.value
}

func isSimpleType[T any]() bool {
	var value T
	return isSimpleType0(reflect.TypeOf(value))
}

func isSimpleType0(tp reflect.Type) bool {
	if tp == nil {
		return false
	}
	if tp == stringType {
		return false
	}
	if !tp.Comparable() {
		return false
	}
	switch tp.Kind() {
	case reflect.Float64:
		logger.Warn("default hash func has poor dispersion for float64")
		return true
	case reflect.Float32:
		logger.Warn("default hash func has poor dispersion for float32")
		return true
	case reflect.Invalid:
		return false
	case reflect.Array:
		return false
	case reflect.Chan:
		return false
	case reflect.Func:
		return false
	case reflect.Interface:
		return false
	case reflect.Map:
		return false
	case reflect.Pointer:
		return false
	case reflect.Slice:
		return false
	case reflect.String:
		return false
	case reflect.UnsafePointer:
		return false
	case reflect.Struct:
		for i := 0; i < tp.NumField(); i++ {
			field := tp.Field(i)
			if !isSimpleType0(field.Type) {
				return false
			}
		}
		return true
	default:
		return true
	}
}

func simpleHash[T any](value T) SizeType {
	switch Sizeof[T]() {
	case 0:
		return 0
	case 1:
		return SizeType(*pointerAs[uint8](pointer(uintptr(unsafe.Pointer(&value)))))
	case 2:
		return SizeType(*pointerAs[uint16](pointer(uintptr(unsafe.Pointer(&value)))))
	case 4:
		return SizeType(*pointerAs[uint32](pointer(uintptr(unsafe.Pointer(&value)))))
	case 8:
		return SizeType(*pointerAs[uint64](pointer(uintptr(unsafe.Pointer(&value)))))
	case 16:
		return SizeType(*pointerAs[uint64](pointer(uintptr(unsafe.Pointer(&value))))) +
			SizeType(*pointerAs[uint64](pointer(uintptr(unsafe.Pointer(&value))))+8)
	default:
		helper := simpleHashHelper[T]{
			value: value,
		}
		ptr := pointer(uintptr(unsafe.Pointer(&helper)))
		return *pointerAs[SizeType](ptr)
	}
}

type simpleHashHelper[T any] struct {
	value T
	_     SizeType
}

func simpleEqual[T comparable](e1, e2 T) bool {
	return e1 == e2
}
