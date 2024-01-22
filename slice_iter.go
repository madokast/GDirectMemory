package direct

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

func (it *SliceIterator[T]) Value() T {
	return *it.Ref()
}

func (it *SliceIterator[T]) Ref() *T {
	if asserted {
		if it.cur >= it.end {
			panic("iterator index out of bound")
		}
		if it.end == nullPointer {
			panic("iterator accesses an empty slice")
		}
		if it.index == sizeTypeMax {
			panic("iterator accesses before call Next()")
		}
	}
	return pointerAs[T](it.cur)
}

func (it *SliceIterator[T]) Index() SizeType {
	if asserted {
		if it.cur >= it.end {
			panic("iterator index out of bound")
		}
		if it.end == nullPointer {
			panic("iterator accesses an empty slice")
		}
		if it.index == sizeTypeMax {
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
