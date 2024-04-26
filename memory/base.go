package memory

import (
	"fmt"
	"github.com/madokast/direct/utils"
	"math"
	"unsafe"
)

type Word = uint64    // link:wordSize
type Pointer Word     // OS pointer
type PageHandler Word // point to a page
type SizeType Word    // size_t

const wordSize = 8
const BasePageSizeShiftNumber = 8
const BasePageSize SizeType = 1 << BasePageSizeShiftNumber

const pageNumberShift = 32
const pageIndexMask PageHandler = (1 << pageNumberShift) - 1
const pageNumberMask = pageIndexMask << pageNumberShift

const nullPageHandle = PageHandler(0)
const NullPointer = Pointer(0)
const SizeTypeMax SizeType = math.MaxUint64

var pageHandlerSize = unsafe.Sizeof(nullPageHandle)
var sizeTypeSize = unsafe.Sizeof(SizeType(0))

func MakePageHandler(pageNumber SizeType, pageIndex SizeType) PageHandler {
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
	return SizeType(p&pageNumberMask) >> (pageNumberShift - BasePageSizeShiftNumber)
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
	return utils.Jsonify(p.Json())
}

func (p Pointer) UIntPtr() uintptr {
	return uintptr(p)
}

func (p Pointer) UnsafePointer() unsafe.Pointer {
	return unsafe.Pointer(uintptr(p))
}

func (p Pointer) IsNull() bool {
	return p == NullPointer
}

func (p Pointer) IsNotNull() bool {
	return p != NullPointer
}

func (p Pointer) String() string {
	if p.IsNull() {
		return "null"
	}
	return fmt.Sprintf("0x%X", p.UIntPtr())
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

func PointerAs[T any](p Pointer) *T {
	return (*T)(unsafe.Pointer(uintptr(p)))
}

const KB = 1024
const MB = 1024 * KB
const GB = 1024 * MB

var memorySizeUints = []string{"B", "KB", "MB", "GB"}

func HumanFriendlyMemorySize(memorySize SizeType) string {
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
	if Sizeof[Word]() < Sizeof[uintptr]() {
		panic(fmt.Sprintf("word cannot hold a pointer %d %d", Sizeof[Word](), Sizeof[uintptr]()))
	}
	if Sizeof[Word]() != wordSize {
		panic(fmt.Sprintf("wordSize is not correct %d %d", Sizeof[Word](), wordSize))
	}
}
