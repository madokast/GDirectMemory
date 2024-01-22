//go:build !cgo

package managed_memory

import (
	"sync"
	"unsafe"
)

var pointers = map[pointer]SizeType{}
var pointersMu sync.Mutex

func libMalloc(size SizeType) pointer {
	ptr := pointer(uintptr(sysAllocOS(size.UIntPtr())))
	pointersMu.Lock()
	pointers[ptr] = size
	pointersMu.Unlock()
	return ptr
}

func libCalloc(size SizeType) pointer {
	ptr := libMalloc(size)
	libZero(ptr, size)
	return ptr
}

func libFree(ptr pointer) {
	pointersMu.Lock()
	sysFreeOS(ptr.UnsafePointer(), pointers[ptr].UIntPtr())
	pointersMu.Unlock()
}

//go:linkname sysAllocOS runtime.sysAllocOS
//go:noescape GoUnusedParameter
//goland:noinspection GoUnusedParameter
func sysAllocOS(n uintptr) unsafe.Pointer

//go:linkname sysFreeOS runtime.sysFreeOS
//go:noescape GoUnusedParameter
//goland:noinspection GoUnusedParameter
func sysFreeOS(v unsafe.Pointer, n uintptr)
