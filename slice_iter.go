package direct

import (
	"github.com/madokast/direct/memory"
	"github.com/madokast/direct/utils"
)

type SliceIterator[T any] struct {
	cur    memory.Pointer
	end    memory.Pointer
	index  SizeType
	noCopy utils.NoCopy
}

func (s Slice[T]) Iterator() (iter SliceIterator[T]) {
	sp := s.pointer()
	if sp.IsNull() {
		// do nothing
	} else {
		header := memory.PointerAs[sliceHeader](sp)
		iter.cur = header.elementBasePointer - memory.Pointer(memory.Sizeof[T]()) // -1
		iter.end = header.elementBasePointer + memory.Pointer(header.length*memory.Sizeof[T]())
		iter.index = SizeTypeMax // -1
	}
	return
}

func (it *SliceIterator[T]) Next() bool {
	it.cur += memory.Pointer(memory.Sizeof[T]())
	it.index += 1
	return it.cur < it.end
}

func (it *SliceIterator[T]) Value() T {
	return *it.Ref()
}

func (it *SliceIterator[T]) Ref() *T {
	if utils.Asserted {
		if it.cur >= it.end {
			panic("iterator index out of bound")
		}
		if it.end == memory.NullPointer {
			panic("iterator accesses an empty slice")
		}
		if it.index == SizeTypeMax {
			panic("iterator accesses before call Next()")
		}
	}
	return memory.PointerAs[T](it.cur)
}

func (it *SliceIterator[T]) Index() SizeType {
	if utils.Asserted {
		if it.cur >= it.end {
			panic("iterator index out of bound")
		}
		if it.end == memory.NullPointer {
			panic("iterator accesses an empty slice")
		}
		if it.index == SizeTypeMax {
			panic("iterator accesses before call Next()")
		}
	}
	return it.index
}

func (s Slice[T]) Iterate(iter func(T)) {
	if s.pointer().IsNotNull() {
		header := s.header()
		ptr := header.elementBasePointer
		length := header.length
		for i := SizeType(0); i < length; i++ {
			iter(*memory.PointerAs[T](ptr))
			ptr += memory.Pointer(memory.Sizeof[T]())
		}
	}
}

func (s Slice[T]) IterateRef(iter func(*T)) {
	if s.pointer().IsNotNull() {
		header := s.header()
		ptr := header.elementBasePointer
		length := header.length
		for i := SizeType(0); i < length; i++ {
			iter(memory.PointerAs[T](ptr))
			ptr += memory.Pointer(memory.Sizeof[T]())
		}
	}
}

func (s Slice[T]) IterateBreakable(iter func(T) (_continue_ bool)) {
	if s.pointer().IsNotNull() {
		header := s.header()
		ptr := header.elementBasePointer
		length := header.length
		for i := SizeType(0); i < length; i++ {
			if !iter(*memory.PointerAs[T](ptr)) {
				break
			}
			ptr += memory.Pointer(memory.Sizeof[T]())
		}
	}
}

func (s Slice[T]) IterateIndex(iter func(index SizeType, element T)) {
	if s.pointer().IsNotNull() {
		header := s.header()
		ptr := header.elementBasePointer
		length := header.length
		for i := SizeType(0); i < length; i++ {
			iter(i, *memory.PointerAs[T](ptr))
			ptr += memory.Pointer(memory.Sizeof[T]())
		}
	}
}

func (s Slice[T]) IterateRefIndex(iter func(index SizeType, ref *T)) {
	if s.pointer().IsNotNull() {
		header := s.header()
		ptr := header.elementBasePointer
		length := header.length
		for i := SizeType(0); i < length; i++ {
			iter(i, memory.PointerAs[T](ptr))
			ptr += memory.Pointer(memory.Sizeof[T]())
		}
	}
}

func (s Slice[T]) IterateIndexBreakable(iter func(index SizeType, element T) (_continue_ bool)) {
	if s.pointer().IsNotNull() {
		header := s.header()
		ptr := header.elementBasePointer
		length := header.length
		for i := SizeType(0); i < length; i++ {
			if !iter(i, *memory.PointerAs[T](ptr)) {
				break
			}
			ptr += memory.Pointer(memory.Sizeof[T]())
		}
	}
}

func (s Slice[T]) IterateRefIndexBreakable(iter func(index SizeType, element *T) (_continue_ bool)) {
	if s.pointer().IsNotNull() {
		header := s.header()
		ptr := header.elementBasePointer
		length := header.length
		for i := SizeType(0); i < length; i++ {
			if !iter(i, memory.PointerAs[T](ptr)) {
				break
			}
			ptr += memory.Pointer(memory.Sizeof[T]())
		}
	}
}
