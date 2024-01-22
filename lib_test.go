package direct

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"
)

func Test_libCMalloc(t *testing.T) {
	const size = 16
	ptr := libMalloc(size)
	libZero(ptr, size)
	var bs []byte
	((*reflect.SliceHeader)(unsafe.Pointer(&bs))).Data = ptr.UIntPtr()
	((*reflect.SliceHeader)(unsafe.Pointer(&bs))).Cap = size
	((*reflect.SliceHeader)(unsafe.Pointer(&bs))).Len = size
	for i, b := range bs {
		t.Log(i, b)
	}
	libFree(ptr)
}

func Test_libCMalloc2(t *testing.T) {
	const size = 1024 * 1024
	ptr := libMalloc(size)
	libZero(ptr, size)
	var bs []byte
	((*reflect.SliceHeader)(unsafe.Pointer(&bs))).Data = ptr.UIntPtr()
	((*reflect.SliceHeader)(unsafe.Pointer(&bs))).Cap = size
	((*reflect.SliceHeader)(unsafe.Pointer(&bs))).Len = size
	for i, b := range bs {
		if b != 0 {
			panic(fmt.Sprint(i, ", ", b))
		}
	}
	libFree(ptr)
}

func Test_libCCalloc(t *testing.T) {
	const size = 1024 * 1024
	ptr := libCalloc(size)
	var bs []byte
	((*reflect.SliceHeader)(unsafe.Pointer(&bs))).Data = ptr.UIntPtr()
	((*reflect.SliceHeader)(unsafe.Pointer(&bs))).Cap = size
	((*reflect.SliceHeader)(unsafe.Pointer(&bs))).Len = size
	for i, b := range bs {
		if b != 0 {
			panic(fmt.Sprint(i, ", ", b))
		}
	}
	libFree(ptr)
}

func Benchmark_malloc(b *testing.B) {
	const size = 128 * 1024 * 1024
	for i := 0; i < b.N; i++ {
		ptr := libMalloc(size)
		libFree(ptr)
	}
}

func Benchmark_go_alloc(b *testing.B) {
	const size = 128 * 1024 * 1024
	for i := 0; i < b.N; i++ {
		_ = make([]byte, size)
	}
}

func Test_libMarkFreedMemory(t *testing.T) {
	const size = 1024
	ptr := libMalloc(size)
	defer libFree(ptr)

	var bs []byte
	((*reflect.SliceHeader)(unsafe.Pointer(&bs))).Data = ptr.UIntPtr()
	((*reflect.SliceHeader)(unsafe.Pointer(&bs))).Cap = size
	((*reflect.SliceHeader)(unsafe.Pointer(&bs))).Len = size

	libZero(ptr, size)
	for i, b := range bs {
		if b != 0 {
			panic(fmt.Sprint(i, ", ", b))
		}
	}

	libMarkFreedMemory(ptr, size)
	for i, b := range bs {
		if b != markFreeMemoryValue {
			panic(fmt.Sprint(i, ", ", b))
		}
	}
}
