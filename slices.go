package managed_memory

import (
	"fmt"
	"gitlab.grandhoo.com/rock/storage/internal/logger"
	"reflect"
	"unsafe"
)

// Slice represents a managed slice
// the zero value is a zero-cap slice
// the slice cannot be moved because of the exposed pointer
type Slice[T any] pointer

func (s *Slice[T]) Append(val T, m *LocalMemory) (err error) {
	err = s.checkCapacity(1, m)
	if err != nil {
		return err
	}
	header := s.header()
	last := header.length
	header.length++
	//*s.RefAt(last) = val
	ptr := header.elementBasePointer + pointer(last*Sizeof[T]())
	*pointerAs[T](ptr) = val
	return nil
}

func (s *Slice[T]) AppendBatch(values Slice[T], m *LocalMemory) (err error) {
	appendNumber := values.Length()
	if appendNumber == 0 {
		return nil
	}
	s0 := *s
	err = s.checkCapacity(appendNumber, m)
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
	if asserted {
		if header.length > header.capacity {
			logger.Panic(fmt.Sprintf("AppendBatch header.length(%d) > header.capacity(%d)", header.length, header.capacity))
		}
	}
	libMemMove(header.elementBasePointer+pointer(start*Sizeof[T]()), values.header().elementBasePointer, appendNumber*Sizeof[T]())
	return nil
}

func (s *Slice[T]) AppendGoSlice(values []T, m *LocalMemory) (err error) {
	appendNumber := SizeType(len(values))
	if appendNumber == 0 {
		return nil
	}
	err = s.checkCapacity(appendNumber, m)
	if err != nil {
		return err
	}
	header := s.header()
	start := header.length
	header.length += appendNumber
	if asserted {
		if header.length > header.capacity {
			logger.Panic(fmt.Sprintf("AppendBatch header.length(%d) > header.capacity(%d)", header.length, header.capacity))
		}
	}
	libMemMove(header.elementBasePointer+pointer(start*Sizeof[T]()),
		pointer(((*reflect.SliceHeader)(unsafe.Pointer(&values))).Data),
		appendNumber*Sizeof[T]())
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

func (s Slice[T]) Iterate(iter func(T)) {
	if s.pointer().IsNotNull() {
		header := s.header()
		ptr := header.elementBasePointer
		length := header.length
		for i := SizeType(0); i < length; i++ {
			iter(*pointerAs[T](ptr))
			ptr += pointer(Sizeof[T]())
		}
	}
}

func (s Slice[T]) IterateRef(iter func(*T)) {
	if s.pointer().IsNotNull() {
		header := s.header()
		ptr := header.elementBasePointer
		length := header.length
		for i := SizeType(0); i < length; i++ {
			iter(pointerAs[T](ptr))
			ptr += pointer(Sizeof[T]())
		}
	}
}

func (s Slice[T]) IterateBreakable(iter func(T) (_continue_ bool)) {
	if s.pointer().IsNotNull() {
		header := s.header()
		ptr := header.elementBasePointer
		length := header.length
		for i := SizeType(0); i < length; i++ {
			if !iter(*pointerAs[T](ptr)) {
				break
			}
			ptr += pointer(Sizeof[T]())
		}
	}
}

func (s Slice[T]) IterateIndex(iter func(index SizeType, element T)) {
	if s.pointer().IsNotNull() {
		header := s.header()
		ptr := header.elementBasePointer
		length := header.length
		for i := SizeType(0); i < length; i++ {
			iter(i, *pointerAs[T](ptr))
			ptr += pointer(Sizeof[T]())
		}
	}
}

func (s Slice[T]) IterateRefIndex(iter func(index SizeType, ref *T)) {
	if s.pointer().IsNotNull() {
		header := s.header()
		ptr := header.elementBasePointer
		length := header.length
		for i := SizeType(0); i < length; i++ {
			iter(i, pointerAs[T](ptr))
			ptr += pointer(Sizeof[T]())
		}
	}
}

func (s Slice[T]) IterateIndexBreakable(iter func(index SizeType, element T) (_continue_ bool)) {
	if s.pointer().IsNotNull() {
		header := s.header()
		ptr := header.elementBasePointer
		length := header.length
		for i := SizeType(0); i < length; i++ {
			if !iter(i, *pointerAs[T](ptr)) {
				break
			}
			ptr += pointer(Sizeof[T]())
		}
	}
}

func (s Slice[T]) IterateRefIndexBreakable(iter func(index SizeType, element *T) (_continue_ bool)) {
	if s.pointer().IsNotNull() {
		header := s.header()
		ptr := header.elementBasePointer
		length := header.length
		for i := SizeType(0); i < length; i++ {
			if !iter(i, pointerAs[T](ptr)) {
				break
			}
			ptr += pointer(Sizeof[T]())
		}
	}
}

func (s Slice[T]) RefAt(index SizeType) *T {
	if asserted {
		if index >= s.Length() {
			logger.Panic(fmt.Sprintf("slice out of bound %d %d", index, s.header().length))
		}
	}
	ptr := s.header().elementBasePointer + pointer(index*Sizeof[T]())
	return pointerAs[T](ptr)
}

func (s *Slice[T]) checkCapacity(appendNumber SizeType, m *LocalMemory) (err error) {
	if asserted {
		if appendNumber == 0 {
			logger.Panic("bad code appendNumber is 0")
		}
	}
	// handle null
	if s.pointer().IsNull() {
		*s, err = MakeSlice[T](m, appendNumber|8)
		return err
	}

	header := pointerAs[sliceHeader](s.pointer())
	originLength := header.length
	targetCapacity := originLength + appendNumber
	if targetCapacity > header.capacity {
		if asserted {
			if originLength > header.capacity {
				logger.Panic(fmt.Sprintf("bad code originLength(%d) > header.capacity(%d)", originLength, header.capacity))
			}
		}
		var s2 Slice[T]
		minTarget := originLength + (originLength >> 2)
		if targetCapacity < minTarget {
			targetCapacity = minTarget
		}
		s2, err = MakeSlice[T](m, targetCapacity)
		if err != nil {
			return err
		}
		s2Header := s2.header()
		s2Header.length = header.length
		libMemMove(s2Header.elementBasePointer, header.elementBasePointer, originLength*Sizeof[T]())
		s.Free(m)
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

func SliceCopy[T any](src, dst Slice[T], elementNumber SizeType) {
	if asserted {
		if src.Length() < elementNumber {
			logger.Panic(fmt.Sprintf("SliceCopy src.Length(%d) < elementNumber(%d)", src.Length(), elementNumber))
		}
		if dst.Length() < elementNumber {
			logger.Panic(fmt.Sprintf("SliceCopy dst.Length(%d) < elementNumber(%d)", dst.Length(), elementNumber))
		}
	}
	libMemMove(dst.header().elementBasePointer, src.header().elementBasePointer, elementNumber*Sizeof[T]())
}

// MakeSlice == make([]T, 0, elementCapacity)
func MakeSlice[T any](m *LocalMemory, elementCapacity SizeType) (Slice[T], error) {
	sliceByteSize := Sizeof[sliceHeader]() + elementCapacity*Sizeof[T]()
	pageNumber := (sliceByteSize + basePageSize - 1) >> basePageSizeShiftNumber

	pageHandler, err := m.allocPage(pageNumber)
	if err != nil {
		return nullSlice, err
	}
	pagePointer := m.pagePointerOf(pageHandler)

	header := pointerAs[sliceHeader](pagePointer)
	header.length = 0
	header.capacity = (pageHandler.Size() - Sizeof[sliceHeader]()) / Sizeof[T]()
	header.pageHandler = pageHandler
	header.elementBasePointer = pagePointer + pointer(Sizeof[sliceHeader]())

	return Slice[T](pagePointer), nil
}

// MakeSliceWithLength == make([]T, elementLength)
func MakeSliceWithLength[T any](m *LocalMemory, elementLength SizeType) (Slice[T], error) {
	s, err := MakeSlice[T](m, elementLength)
	if err != nil {
		return nullSlice, err
	}

	header := s.header()
	if asserted {
		if header.capacity < elementLength {
			logger.Panic(fmt.Sprintf("bad code header.capacity(%d) < elementLength(%d)", header.capacity, elementLength))
		}
	}
	header.length = elementLength
	libZero(header.elementBasePointer, elementLength*Sizeof[T]())
	return s, nil
}

func MakeSliceFromGoSlice[T any](m *LocalMemory, gs []T) (Slice[T], error) {
	elementLength := SizeType(len(gs))
	if elementLength == 0 {
		return nullSlice, nil
	}
	s, err := MakeSlice[T](m, elementLength)
	if err != nil {
		return nullSlice, err
	}
	//for i, e := range gs {
	//	s.Set(SizeType(i), e)
	//}
	header := s.header()
	if asserted {
		if header.length != 0 {
			logger.Panic(fmt.Sprintf("MakeSlice header.length is %d not 0", header.length))
		}
	}
	header.length = elementLength
	libMemMove(header.elementBasePointer, libGoSliceHeaderPointer(gs), elementLength*Sizeof[T]())
	return s, nil
}

func (s Slice[T]) Copy(m *LocalMemory) (Slice[T], error) {
	if s == nullSlice {
		return nullSlice, nil
	}
	srcHeader := s.header()
	srcLength := srcHeader.length
	if srcLength == 0 {
		return nullSlice, nil
	}
	cp, err := MakeSlice[T](m, srcLength)
	if err != nil {
		return nullSlice, err
	}
	cpHeader := cp.header()
	cpHeader.length = srcLength
	libMemMove(cpHeader.elementBasePointer, srcHeader.elementBasePointer, srcLength*Sizeof[T]())
	return cp, nil
}

func (s *Slice[T]) Move() (moved Slice[T]) {
	moved = *s
	*s = nullSlice
	return moved
}

func (s Slice[T]) Free(m *LocalMemory) {
	if s.pointer().IsNotNull() {
		if asserted {
			if s.header().pageHandler == 0 {
				logger.Panic("double free?")
			}
		}
		m.freePage(s.header().pageHandler)
	}
}

func (s Slice[T]) pointer() pointer {
	return pointer(s)
}

func (s Slice[T]) header() *sliceHeader {
	if asserted {
		if s.pointer().IsNull() {
			logger.Panic("header of null")
		}
	}
	return pointerAs[sliceHeader](s.pointer())
}

const nullSlice = 0

type sliceHeader struct {
	length             SizeType
	capacity           SizeType
	pageHandler        PageHandler // for free
	elementBasePointer pointer     // pointer to first element
}

type SliceIterator[T any] struct {
	cur    pointer
	end    pointer
	index  SizeType
	noCopy noCopy
}

func (s Slice[T]) Iterator() (iter SliceIterator[T]) {
	sp := s.pointer()
	if sp.IsNull() {
		// do nothing
	} else {
		header := pointerAs[sliceHeader](sp)
		iter.cur = header.elementBasePointer - pointer(Sizeof[T]()) // -1
		iter.end = header.elementBasePointer + pointer(header.length*Sizeof[T]())
		iter.index = sizeTypeMax // -1
	}
	return
}

func (it *SliceIterator[T]) Next() bool {
	it.cur += pointer(Sizeof[T]())
	it.index += 1
	return it.cur < it.end
}

func (it SliceIterator[T]) Value() T {
	return *it.Ref()
}

func (it SliceIterator[T]) Ref() *T {
	if asserted {
		if it.cur >= it.end {
			logger.Panic("iterator index out of bound")
		}
		if it.end == nullPointer {
			logger.Panic("iterator accesses an empty slice")
		}
		if it.index == sizeTypeMax {
			logger.Panic("iterator accesses before call Next()")
		}
	}
	return pointerAs[T](it.cur)
}

func (it SliceIterator[T]) Index() SizeType {
	if asserted {
		if it.cur >= it.end {
			logger.Panic("iterator index out of bound")
		}
		if it.end == nullPointer {
			logger.Panic("iterator accesses an empty slice")
		}
		if it.index == sizeTypeMax {
			logger.Panic("iterator accesses before call Next()")
		}
	}
	return it.index
}
