package directstring

import (
	"strings"
	"unsafe"
)

// String represents a string in direct memory
type String struct {
	address uintptr
	length  int
}

// CastFromString casts a go string as direct string.
// warning: do not hold this direct string when the
// go string are not referenced.
func CastFromString(s string) String {
	return *((*String)(unsafe.Pointer(&s)))
}

func (s String) Len() int {
	return s.length
}

// ToHeapString returns a go string in heap.
// The go string can be safely used after the memory
// of direct String has been released.
func (s String) ToHeapString() string {
	return strings.Clone(s.CastAsString())
}

// CastAsString casts a direct String as go string
// warning: do not hold this go string when the memory
// of direct String has been released.
// It'is useful when communicating with go string's function
// such as strings.Index(s.CastAsString(), "xx")
func (s String) CastAsString() string {
	return *((*string)(unsafe.Pointer(&s)))
}

func (s String) ToString() string {
	return s.ToHeapString()
}

func (s String) HashCode() uint64 {
	hash := uint64(2166136261)
	const prime64 = uint64(16777619)
	var i uintptr = 0
	for i+8 <= uintptr(s.length) {
		hash *= prime64
		hash ^= *((*uint64)(unsafe.Pointer(s.address + i)))
		i += 8
	}
	for i+2 <= uintptr(s.length) {
		hash *= prime64
		hash ^= uint64(*((*uint16)(unsafe.Pointer(s.address + i))))
		i += 2
	}
	for i < uintptr(s.length) {
		hash *= prime64
		hash ^= uint64(*((*uint8)(unsafe.Pointer(s.address + i))))
		i += 1
	}
	return hash
}
