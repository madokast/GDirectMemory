package direct

import (
	"fmt"
	"reflect"
	"unsafe"
)

type MapIterator[Key comparable, Value any] struct {
	tableIter    SliceIterator[entry[Key, Value]]
	currentSlot  *entry[Key, Value]
	tableBasePtr pointer
	noCopy       noCopy
}

func (m Map[Key, Value]) Iterator() (iter MapIterator[Key, Value]) {
	if asserted {
		if m.IsNull() {
			panic(fmt.Sprintf("use a moved or freed or null map"))
		}
	}
	header := m.header()
	return MapIterator[Key, Value]{
		tableIter: SliceIterator[entry[Key, Value]]{
			cur:   header.tableBasePtr - pointer(Sizeof[entry[Key, Value]]()),
			end:   header.tableBasePtr + pointer((header.mask+1)*Sizeof[entry[Key, Value]]()),
			index: sizeTypeMax,
		},
		currentSlot:  nil,
		tableBasePtr: header.tableBasePtr,
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
			panic("access iter-map before check Next()")
		}
	}
	return &mi.currentSlot.key
}

func (mi *MapIterator[Key, Value]) ValueRef() *Value {
	if asserted {
		if mi.currentSlot == nil {
			panic("access iter-map before check Next()")
		}
	}
	return &mi.currentSlot.value
}

/* ==================== Iterate ============================*/

func (m Map[Key, Value]) Iterate(iter func(Key, Value)) {
	if asserted {
		if m == nilMap {
			panic("use a moved or freed or null map")
		}
	}
	header := m.header()
	header.table.IterateRefIndexBreakable(func(index SizeType, slot *entry[Key, Value]) bool {
		if index > header.mask {
			return false
		}
		next := slot.next
		if next != emptyTableFlag {
			iter(slot.key, slot.value)
			for next != listTailFlag {
				slot = header.dataAt(next)
				iter(slot.key, slot.value)
				next = slot.next
			}
		}
		return true
	})
}

func (m Map[Key, Value]) IterateBreakable(iter func(Key, Value) (_continue_ bool)) {
	if asserted {
		if m == nilMap {
			panic("use a moved or freed or null map")
		}
	}
	header := m.header()
	header.table.IterateRefIndexBreakable(func(index SizeType, slot *entry[Key, Value]) bool {
		if index > header.mask {
			return false
		}
		next := slot.next
		if next != emptyTableFlag {
			if !iter(slot.key, slot.value) {
				return false
			}
			for next != listTailFlag {
				slot = header.dataAt(next)
				if !iter(slot.key, slot.value) {
					return false
				}
				next = slot.next
			}
		}
		return true
	})
}

/* ==================== Hash Equal ============================*/

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
		return true
	case reflect.Float32:
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

/* ==================== isMap ============================*/

type isMapHelper interface {
	isMapHelper20240102()
}

var isMapHelperType = reflect.TypeOf((*isMapHelper)(nil)).Elem()
var mapKind = reflect.TypeOf(Map[int, int](nilMap)).Kind()

func (m Map[Key, Value]) isMapHelper20240102() {
	panic("should not call")
}

func isMap[T any]() bool {
	tp := reflect.TypeOf((*T)(nil)).Elem()
	return tp.Kind() == mapKind && tp.Implements(isMapHelperType)
}

func init() {
	if Sizeof[mapHeader[int, int]]() > basePageSize {
		panic(fmt.Sprint("size of mapHeader ", Sizeof[mapHeader[int, int]](), " > 1 page ", basePageSize))
	}
}
