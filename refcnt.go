package direct

import (
	"fmt"
	"github.com/madokast/direct/memory"
	"github.com/madokast/direct/memory/trace_type"
	"github.com/madokast/direct/utils"
	"sync/atomic"
)

// Shared represent a thread-safe shared obj.
// It should be created by SharedFactory
// un-init index == 0 && holder == 0
// freed   index == 0 && holder != 0
// in-use  index != 0 && holder != 0
type Shared[T object] struct {
	index  SizeType // 0 for holder-element-count
	holder Slice[objRefCnt[T]]
}

type objRefCnt[T object] struct {
	obj    T
	refCnt int64
}

type object interface {
	Free()
	Moved() bool
	String() string
}

func (s *Shared[T]) NotInit() bool {
	if utils.Asserted {
		if s.index == 0 && s.holder != nullSlice {
			panic("use a freed shared obj")
		}
	}
	return s.index == 0
}

func (s *Shared[T]) Value() T {
	return *s.Ref()
}

func (s *Shared[T]) Ref() *T {
	if utils.Asserted {
		if s.index == 0 {
			if s.holder == nullSlice {
				panic("use an un-init shared obj")
			} else {
				panic("use a freed shared obj")
			}
		}
		if atomic.LoadInt64(&s.holder.RefAt(s.index).refCnt) == 0 {
			panic("use a freed shared obj")
		}
	}
	return &s.holder.RefAt(s.index).obj
}

func (s Shared[T]) String() string {
	return (*s.Ref()).String()
}

func (s *Shared[T]) Share() Shared[T] {
	if utils.Asserted {
		if s.index == 0 {
			if s.holder == nullSlice {
				panic("use an un-init shared obj")
			} else {
				panic("use a freed shared obj")
			}
		}
		if atomic.LoadInt64(&s.holder.RefAt(s.index).refCnt) == 0 {
			panic("share a freed obj")
		}
	}
	cnt := atomic.AddInt64(&s.holder.RefAt(s.index).refCnt, 1)
	if utils.Debug {
		fmt.Println("share obj cnt", cnt)
	}
	return *s
}

func (s *Shared[T]) Free() {
	if s.index == 0 {
		if s.holder == nullSlice {
			return // free an un-init shared is ok
		} else {
			panic("double free shared obj!")
		}
	}
	objCnt := s.holder.RefAt(s.index)
	if utils.Asserted {
		if atomic.LoadInt64(&objCnt.refCnt) == 0 {
			panic("double free shared obj!")
		}
	}
	cnt := atomic.AddInt64(&objCnt.refCnt, -1)
	if utils.Debug {
		fmt.Println("free shared obj cnt", cnt)
	}
	if utils.Asserted {
		if cnt < 0 {
			panic(fmt.Sprint("double free shared obj. Cnt =", cnt))
		}
	}
	if cnt == 0 {
		objCnt.obj.Free()
		header := s.holder.header()
		holderRefCnt := atomic.AddInt32(memory.PointerAs[int32](header.elementBasePointer), -1)
		if utils.Debug {
			fmt.Println("do free shared obj, holder cnt", holderRefCnt)
		}
		if holderRefCnt == 0 {
			s.holder.Free()
		}
	}
	s.index = 0
}

/*======================== FACTORY ==========================*/

// SharedFactory creates sharedObj. Thread unsafe
type SharedFactory[T object] struct {
	holder Slice[objRefCnt[T]] // index 0 for holder-element-count
	noCopy utils.NoCopy
}

func CreateSharedFactory[T object]() SharedFactory[T] {
	return SharedFactory[T]{
		holder: nullSlice,
	}
}

func (sf *SharedFactory[T]) MakeShared(obj T) (s Shared[T], err error) {
	var sfHolder = sf.holder                                  // register
	var sfHolderHeader *sliceHeader = nil                     // register
	var sfHolderHeaderElementBasePointer = memory.NullPointer // register
	if sfHolder == nullSlice {
		// index 0 for count
		sfHolder, err = makeSlice0[objRefCnt[T]](2, trace_type.Shared, 3)
		if err != nil {
			return s, err
		}
		sf.holder = sfHolder
		sfHolderHeader = sfHolder.header()
		sfHolderHeader.length = 1 // index 0 for holder-element-count
		sfHolderHeaderElementBasePointer = sfHolderHeader.elementBasePointer
		// init ref count. One means the factory holding it
		*memory.PointerAs[int32](sfHolderHeaderElementBasePointer) = 1
	} else {
		sfHolderHeader = sfHolder.header()
		sfHolderHeaderElementBasePointer = sfHolderHeader.elementBasePointer
	}

	if utils.Asserted {
		if sf.holder != sfHolder {
			panic("bad code: f.holder != sfHolder")
		}
		if sfHolderHeader == nil {
			panic("bad code: sfHolderHeader == nil")
		}
		if sf.holder.Length()+1 > sf.holder.Capacity() {
			panic(fmt.Sprintf("slice capacity is not enough slice length(%d) capacity(%d)",
				sf.holder.Length(), sf.holder.Capacity()))
		}
		if sfHolderHeaderElementBasePointer.IsNull() {
			panic("bad code: sfHolderHeaderElementBasePointer.IsNull()")
		}
	}

	s.holder = sfHolder
	s.index = sfHolderHeader.length

	var objCnt = objRefCnt[T]{
		obj:    obj,
		refCnt: 1, // cnt for the element
	}
	objTargetPtr := sfHolderHeaderElementBasePointer + memory.Pointer(s.index*memory.Sizeof[objRefCnt[T]]())
	*memory.PointerAs[objRefCnt[T]](objTargetPtr) = objCnt
	sfHolderHeader.length += 1

	// add holder count
	holderRefCnt := atomic.AddInt32(memory.PointerAs[int32](sfHolderHeaderElementBasePointer), 1)
	if utils.Asserted {
		if holderRefCnt <= 0 {
			panic(fmt.Sprintf("holder count(%d) <= 0", holderRefCnt))
		}
	}
	if utils.Debug {
		fmt.Printf("make shared obj in holder count %d\n", holderRefCnt)
	}

	// detach the full holder
	if sfHolderHeader.length == sfHolderHeader.capacity {
		holderRefCnt = atomic.AddInt32(memory.PointerAs[int32](sfHolderHeaderElementBasePointer), -1)
		if utils.Asserted {
			if holderRefCnt <= 0 {
				panic(fmt.Sprintf("holderRefCnt(%d) <= 0", holderRefCnt))
			}
		}
		sf.holder = nullSlice
	}

	return s, nil
}

func (sf *SharedFactory[T]) Destroy() {
	holder := sf.holder      // register
	if holder != nullSlice { // maybe null
		holderRefCnt := atomic.AddInt32(memory.PointerAs[int32](holder.header().elementBasePointer), -1)
		if utils.Asserted {
			if holderRefCnt < 0 {
				panic(fmt.Sprintf("holderRefCnt(%d) <= 0", holderRefCnt))
			}
		}
		if holderRefCnt == 0 {
			holder.Free()
		}
		sf.holder = nullSlice // detach the holder anyway
	}
}
