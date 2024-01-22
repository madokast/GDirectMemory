package managed_memory

import (
	"fmt"
	"gitlab.grandhoo.com/rock/storage/internal/logger"
	"reflect"
	"strings"
	"unsafe"
)

// String represents an un-modifiable string in managed_memory.
type String struct {
	ptr    pointer
	length SizeType
	holder Slice[byte]
}

func (s *String) Move() (moved String) {
	moved = *s
	*s = emptyString
	return moved
}

func (s String) Length() SizeType {
	return s.length
}

func (s String) AsGoString() string {
	return *((*string)(unsafe.Pointer(&s)))
}

func (s String) CopyToGoString() string {
	var sb strings.Builder
	sb.WriteString(s.AsGoString())
	return sb.String()
}

func (s String) Equal(s2 String) bool {
	return s.AsGoString() == s2.AsGoString()
}

func isString[T any]() bool {
	var v T
	return reflect.TypeOf(v) == stringType
}

func equalString[str any](s1, s2 str) bool {
	if asserted {
		if !isString[str]() {
			logger.Panic("call equalString using non-string type", s1, s2, fmt.Sprintf("%T", s1))
		}
	}
	return *((*string)(unsafe.Pointer(&s1))) == *((*string)(unsafe.Pointer(&s2)))
}

func (s String) Hashcode() (hash SizeType) {
	hash = 2166136261
	const prime32 = 16777619
	i := SizeType(0)
	for ; i < s.length; i++ {
		hash *= prime32
		//logger.Debug("at", fmt.Sprintf("%c", *pointerAs[byte](s.ptr)))
		hash ^= SizeType(*pointerAs[byte](s.ptr))
		s.ptr++
	}
	return hash
}

func hashString[str any](s str) SizeType {
	if asserted {
		if !isString[str]() {
			logger.Panic("call hashString using non-string type", s, fmt.Sprintf("%T", s))
		}
	}
	return ((*String)(unsafe.Pointer(&s))).Hashcode()
}

func (s String) String() string {
	return s.CopyToGoString()
}

func (s String) Free(m *LocalMemory) {
	s.holder.Free(m)
}

var emptyString = String{}
var stringType = reflect.TypeOf(emptyString)

func init() {
	if Sizeof[String]()-Sizeof[Slice[byte]]() != Sizeof[string]() {
		logger.Panic(fmt.Sprintf("string size is not correct %d %d", Sizeof[String](), Sizeof[word]()))
	}
}
