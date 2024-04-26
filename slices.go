package direct

import (
	"fmt"
	"github.com/madokast/direct/memory"
	"github.com/madokast/direct/memory/trace_type"
	"github.com/madokast/direct/utils"
	"reflect"
	"unsafe"
)

// Slice represents a managed slice
// the zero value is a zero-cap slice
type Slice[T any] memory.Pointer

// MakeSlice == make([]T, 0, elementCapacity)
func MakeSlice[T any](elementCapacity SizeType) (Slice[T], error) {
	return makeSlice0[T](elementCapacity, trace_type.Slice, 3)
}

// makeSlice0 is the root make func doing alloc
func makeSlice0[T any](elementCapacity SizeType, _type trace_type.Type, traceSkip int) (Slice[T], error) {
	sliceByteSize := memory.Sizeof[sliceHeader]() + elementCapacity*memory.Sizeof[T]()
	pageNumber := (sliceByteSize + memory.BasePageSize - 1) >> memory.BasePageSizeShiftNumber

	pageHandler, err := Global.allocPage(pageNumber, _type, traceSkip)
	if err != nil {
		return nullSlice, err
	}
	pagePointer := Global.pagePointerOf(pageHandler)

	header := memory.PointerAs[sliceHeader](pagePointer)
	header.length = 0
	header.capacity = (pageHandler.Size() - memory.Sizeof[sliceHeader]()) / memory.Sizeof[T]()
	header.pageHandler = pageHandler
	header.elementBasePointer = pagePointer + memory.Pointer(memory.Sizeof[sliceHeader]())

	return Slice[T](pagePointer), nil
}

// MakeSliceWithLength == make([]T, elementLength)
func MakeSliceWithLength[T any](elementLength SizeType) (Slice[T], error) {
	return makeSliceWithLength0[T](elementLength, trace_type.Slice, 4)
}

func makeSliceWithLength0[T any](elementLength SizeType, _type trace_type.Type, traceSkip int) (Slice[T], error) {
	s, err := makeSlice0[T](elementLength, _type, traceSkip)
	if err != nil {
		return nullSlice, err
	}

	header := s.header()
	if utils.Asserted {
		if header.capacity < elementLength {
			panic(fmt.Sprintf("bad code header.capacity(%d) < elementLength(%d)", header.capacity, elementLength))
		}
	}
	header.length = elementLength
	memory.LibZero(header.elementBasePointer, elementLength*memory.Sizeof[T]())
	return s, nil
}

func MakeSliceFromGoSlice[T any](gs []T) (Slice[T], error) {
	elementLength := SizeType(len(gs))
	if elementLength == 0 {
		return nullSlice, nil
	}
	s, err := makeSlice0[T](elementLength, trace_type.Slice, 3)
	if err != nil {
		return nullSlice, err
	}
	//for i, e := range gs {
	//	s.Set(SizeType(i), e)
	//}
	header := s.header()
	if utils.Asserted {
		if header.length != 0 {
			panic(fmt.Sprintf("MakeSlice header.length is %d not 0", header.length))
		}
	}
	header.length = elementLength
	memory.LibMemMove(header.elementBasePointer, memory.LibGoSliceHeaderPointer(gs), elementLength*memory.Sizeof[T]())
	return s, nil
}

func (s *Slice[T]) Append(val T) (err error) {
	err = s.checkCapacity(1)
	if err != nil {
		return err
	}
	header := s.header()
	last := header.length
	header.length++
	//*s.RefAt(last) = val
	ptr := header.elementBasePointer + memory.Pointer(last*memory.Sizeof[T]())
	*memory.PointerAs[T](ptr) = val
	return nil
}

func (s *Slice[T]) AppendBatch(values Slice[T]) (err error) {
	appendNumber := values.Length()
	if appendNumber == 0 {
		return nil
	}
	s0 := *s
	err = s.checkCapacity(appendNumber)
	if err != nil {
		return err
	}
	// appending to self!
	if s0 == values {
		values = *s
	}
	header := s.header()
	start := header.length
	header.length += appendNumber
	if utils.Asserted {
		if header.length > header.capacity {
			panic(fmt.Sprintf("AppendBatch header.length(%d) > header.capacity(%d)", header.length, header.capacity))
		}
	}
	memory.LibMemMove(header.elementBasePointer+memory.Pointer(start*memory.Sizeof[T]()), values.header().elementBasePointer, appendNumber*memory.Sizeof[T]())
	return nil
}

func (s *Slice[T]) AppendGoSlice(values []T) (err error) {
	appendNumber := SizeType(len(values))
	if appendNumber == 0 {
		return nil
	}
	err = s.checkCapacity(appendNumber)
	if err != nil {
		return err
	}
	header := s.header()
	start := header.length
	header.length += appendNumber
	if utils.Asserted {
		if header.length > header.capacity {
			panic(fmt.Sprintf("AppendBatch header.length(%d) > header.capacity(%d)", header.length, header.capacity))
		}
	}
	memory.LibMemMove(header.elementBasePointer+memory.Pointer(start*memory.Sizeof[T]()),
		memory.Pointer(((*reflect.SliceHeader)(unsafe.Pointer(&values))).Data),
		appendNumber*memory.Sizeof[T]())
	return nil
}

func (s Slice[T]) Length() SizeType {
	if s.pointer().IsNull() {
		return 0
	}
	return s.header().length
}

func (s Slice[T]) Capacity() SizeType {
	if s.pointer().IsNull() {
		return 0
	}
	return s.header().capacity
}

func (s Slice[T]) Get(index SizeType) T {
	return *s.RefAt(index)
}

func (s Slice[T]) Set(index SizeType, val T) {
	*s.RefAt(index) = val
}

func (s Slice[T]) RefAt(index SizeType) *T {
	if utils.Asserted {
		if index >= s.Length() {
			panic(fmt.Sprintf("slice out of bound %d %d", index, s.header().length))
		}
	}
	ptr := s.header().elementBasePointer + memory.Pointer(index*memory.Sizeof[T]())
	return memory.PointerAs[T](ptr)
}

func (s *Slice[T]) checkCapacity(appendNumber SizeType) (err error) {
	if utils.Asserted {
		if appendNumber == 0 {
			panic("bad code appendNumber is 0")
		}
	}
	// handle null
	if s.pointer().IsNull() {
		*s, err = makeSlice0[T](appendNumber|8, trace_type.Slice, 4)
		return err
	}

	header := memory.PointerAs[sliceHeader](s.pointer())
	originLength := header.length
	targetCapacity := originLength + appendNumber
	if targetCapacity > header.capacity {
		if utils.Asserted {
			if originLength > header.capacity {
				panic(fmt.Sprintf("bad code originLength(%d) > header.capacity(%d)", originLength, header.capacity))
			}
		}
		var s2 Slice[T]
		minTarget := originLength + (originLength >> 2)
		if targetCapacity < minTarget {
			targetCapacity = minTarget
		}
		s2, err = makeSlice0[T](targetCapacity, trace_type.Slice, 4)
		if err != nil {
			return err
		}
		s2Header := s2.header()
		s2Header.length = header.length
		memory.LibMemMove(s2Header.elementBasePointer, header.elementBasePointer, originLength*memory.Sizeof[T]())
		s.Free()
		*s = s2
	}
	return nil
}

func (s Slice[T]) GoSlice() []T {
	gs := make([]T, int(s.Length()))
	s.IterateIndex(func(index SizeType, element T) {
		gs[index] = element
	})
	return gs
}

func (s Slice[T]) String() string {
	return fmt.Sprintf("%+v", s.GoSlice())
}

func (s Slice[T]) Copy() (Slice[T], error) {
	if s == nullSlice {
		return nullSlice, nil
	}
	srcHeader := s.header()
	srcLength := srcHeader.length
	if srcLength == 0 {
		return nullSlice, nil
	}
	cp, err := makeSlice0[T](srcLength, trace_type.Slice, 3)
	if err != nil {
		return nullSlice, err
	}
	cpHeader := cp.header()
	cpHeader.length = srcLength
	memory.LibMemMove(cpHeader.elementBasePointer, srcHeader.elementBasePointer, srcLength*memory.Sizeof[T]())
	return cp, nil
}

func (s *Slice[T]) Move() (moved Slice[T]) {
	moved = *s
	*s = nullSlice
	return moved
}

func (s Slice[T]) Moved() bool {
	return s == nullSlice
}

func (s Slice[T]) Free() {
	if s.pointer().IsNotNull() {
		if utils.Asserted {
			if s.header().pageHandler.IsNull() {
				panic("double free?")
			}
		}
		Global.freePage(s.header().pageHandler)
	}
}

func (s Slice[T]) pointer() memory.Pointer {
	return memory.Pointer(s)
}

func (s Slice[T]) header() *sliceHeader {
	if utils.Asserted {
		if s.pointer().IsNull() {
			panic("header of null")
		}
	}
	return memory.PointerAs[sliceHeader](s.pointer())
}

func SliceCopy[T any](src, dst Slice[T], elementNumber SizeType) {
	if utils.Asserted {
		if src.Length() < elementNumber {
			panic(fmt.Sprintf("SliceCopy src.Length(%d) < elementNumber(%d)", src.Length(), elementNumber))
		}
		if dst.Length() < elementNumber {
			panic(fmt.Sprintf("SliceCopy dst.Length(%d) < elementNumber(%d)", dst.Length(), elementNumber))
		}
	}
	memory.LibMemMove(dst.header().elementBasePointer, src.header().elementBasePointer, elementNumber*memory.Sizeof[T]())
}

const nullSlice = 0

type sliceHeader struct {
	length             SizeType
	capacity           SizeType
	pageHandler        memory.PageHandler // for free
	elementBasePointer memory.Pointer     // pointer to first element
}
