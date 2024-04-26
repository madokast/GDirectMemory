package memory

import (
	"fmt"
	"reflect"
	"unsafe"
)

func LibMalloc(size SizeType) Pointer {
	pointer := Pointer(uintptr(sysAllocOS(size.UIntPtr())))
	if pointer.IsNull() {
		panic(fmt.Sprintf("cannot allocate memory %s", HumanFriendlyMemorySize(size)))
	}
	return pointer
}

func LibCalloc(size SizeType) Pointer {
	ptr := LibMalloc(size)
	LibZero(ptr, size)
	return ptr
}

func LibFree(ptr Pointer) {
	sysFreeOS(ptr.UnsafePointer(), 0) // 0 is ok, just for log
}

func LibZero(ptr Pointer, size SizeType) {
	memclrNoHeapPointers(unsafe.Pointer(ptr.UIntPtr()), size.UIntPtr())
}

func LibMemMove(to, from Pointer, size SizeType) {
	memmove(unsafe.Pointer(to.UIntPtr()), unsafe.Pointer(from.UIntPtr()), size.UIntPtr())
}

func LibMemEqual(p1, p2 Pointer, size SizeType) bool {
	return memequal(unsafe.Pointer(p1.UIntPtr()), unsafe.Pointer(p2.UIntPtr()), size.UIntPtr())
}

func LibGoSliceHeaderPointer[E any](slice []E) Pointer {
	return Pointer((*reflect.SliceHeader)(unsafe.Pointer(&slice)).Data)
}

const markFreeMemoryValue = 0b1010

func libMarkFreedMemory(ptr Pointer, size SizeType) {
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

//go:linkname sysAllocOS runtime.sysAllocOS
//go:noescape GoUnusedParameter
//goland:noinspection GoUnusedParameter
func sysAllocOS(n uintptr) unsafe.Pointer

//go:linkname sysFreeOS runtime.sysFreeOS
//go:noescape GoUnusedParameter
//goland:noinspection GoUnusedParameter
func sysFreeOS(v unsafe.Pointer, n uintptr)

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
