package managed_memory

import (
	"fmt"
	"gitlab.grandhoo.com/rock/storage/internal/logger"
	"gitlab.grandhoo.com/rock/storage/storage2/utils/rpc/serializer"
	"math"
	"unsafe"
)

type word = uint64    // link:wordSize
type pointer word     // OS pointer
type PageHandler word // point to a page
type SizeType word    // size_t

const wordSize = 8
const basePageSizeShiftNumber = 8
const basePageSize SizeType = 1 << basePageSizeShiftNumber

const pageNumberShift = 32
const pageIndexMask PageHandler = (1 << pageNumberShift) - 1
const pageNumberMask = pageIndexMask << pageNumberShift

const nullPageHandle = PageHandler(0)
const nullPointer = pointer(0)
const sizeTypeMax SizeType = math.MaxUint64

var pageHandlerSize = unsafe.Sizeof(nullPageHandle)
var sizeTypeSize = unsafe.Sizeof(SizeType(0))

func makePageHandler(pageNumber SizeType, pageIndex SizeType) PageHandler {
	return PageHandler((pageNumber << pageNumberShift) + pageIndex)
}

func (p PageHandler) IsNull() bool {
	return p == nullPageHandle
}

func (p PageHandler) IsNotNull() bool {
	return p != nullPageHandle
}

// PageIndex counts by base page
func (p PageHandler) PageIndex() SizeType {
	return SizeType(p & pageIndexMask)
}

func (p PageHandler) PageNumber() SizeType {
	return SizeType(p&pageNumberMask) >> pageNumberShift
}

func (p PageHandler) Size() SizeType {
	return SizeType(p&pageNumberMask) >> (pageNumberShift - basePageSizeShiftNumber)
}

func (p PageHandler) Json() interface{} {
	if p.IsNull() {
		return "null"
	}
	return map[string]interface{}{
		"id":     p.PageIndex(),
		"number": p.PageNumber(),
	}
}

func (p PageHandler) String() string {
	return serializer.Jsonify(p.Json())
}

func (p pointer) UIntPtr() uintptr {
	return uintptr(p)
}

func (p pointer) UnsafePointer() unsafe.Pointer {
	return unsafe.Pointer(uintptr(p))
}

func (p pointer) IsNull() bool {
	return p == nullPointer
}

func (p pointer) IsNotNull() bool {
	return p != nullPointer
}

func (p pointer) String() string {
	if p.IsNull() {
		return "null"
	}
	return fmt.Sprintf("%x", p.UIntPtr())
}

func (s SizeType) UIntPtr() uintptr {
	return uintptr(s)
}

func (s SizeType) Int() int {
	return int(s)
}

func (s SizeType) BitString() string {
	return fmt.Sprintf("%b", s)
}

func pointerAs[T any](p pointer) *T {
	return (*T)(unsafe.Pointer(uintptr(p)))
}

const cacheLineSize = 64

type cacheShareWord[T ~word] struct {
	value T
	_     [cacheLineSize - wordSize]byte
}

const KB = 1024
const MB = 1024 * KB
const GB = 1024 * MB

var memorySizeUints = []string{"B", "KB", "MB", "GB"}

func humanFriendlyMemorySize(memorySize SizeType) string {
	size := float64(memorySize)
	cnt := 0
	for size >= 1024 {
		cnt++
		size /= 1024
		if cnt == len(memorySizeUints)-1 {
			break
		}
	}
	if cnt == 0 {
		return fmt.Sprintf("%d%s", memorySize, memorySizeUints[0])
	} else {
		return fmt.Sprintf("%.2f%s", size, memorySizeUints[cnt])
	}

}

func Sizeof[T any]() SizeType {
	return SizeType(unsafe.Sizeof(*((*T)(unsafe.Pointer(uintptr(0))))))
}

func init() {
	if Sizeof[word]() < Sizeof[uintptr]() {
		logger.Panic(fmt.Sprintf("word cannot hold a pointer %d %d", Sizeof[word](), Sizeof[uintptr]()))
	}
	if Sizeof[word]() != wordSize {
		logger.Panic(fmt.Sprintf("wordSize is not correct %d %d", Sizeof[word](), wordSize))
	}
}
