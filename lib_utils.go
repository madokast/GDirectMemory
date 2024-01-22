package managed_memory

import (
	"reflect"
	"unsafe"
)

func libZero(ptr pointer, size SizeType) {
	memclrNoHeapPointers(unsafe.Pointer(ptr.UIntPtr()), size.UIntPtr())
}

func libMemMove(to, from pointer, size SizeType) {
	memmove(unsafe.Pointer(to.UIntPtr()), unsafe.Pointer(from.UIntPtr()), size.UIntPtr())
}

func libMemEqual(p1, p2 pointer, size SizeType) bool {
	return memequal(unsafe.Pointer(p1.UIntPtr()), unsafe.Pointer(p2.UIntPtr()), size.UIntPtr())
}

func libGoSliceHeaderPointer[E any](slice []E) pointer {
	return pointer((*reflect.SliceHeader)(unsafe.Pointer(&slice)).Data)
}

const markFreeMemoryValue = 0b1010

func libMarkFreedMemory(ptr pointer, size SizeType) {
	var bs []byte
	*((*reflect.SliceHeader)(unsafe.Pointer(&bs))) = reflect.SliceHeader{
		Data: ptr.UIntPtr(),
		Len:  size.Int(),
		Cap:  size.Int(),
	}
	for i := range bs {
		bs[i] = markFreeMemoryValue
	}
}

//go:noescape
//go:linkname memclrNoHeapPointers runtime.memclrNoHeapPointers
//goland:noinspection GoUnusedParameter
func memclrNoHeapPointers(ptr unsafe.Pointer, n uintptr)

//go:linkname memmove runtime.memmove
//go:noescape GoUnusedParameter
//goland:noinspection GoUnusedParameter
func memmove(to, from unsafe.Pointer, n uintptr)

//go:linkname memequal runtime.memequal
//go:noescape GoUnusedParameter
//goland:noinspection GoUnusedParameter
func memequal(a, b unsafe.Pointer, size uintptr) bool
