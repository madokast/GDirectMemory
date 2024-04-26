package direct

import (
	"fmt"
	"github.com/madokast/direct/memory"
	"github.com/madokast/direct/memory/trace_type"
	"github.com/madokast/direct/utils"
)

// Stack represents a managed stack
// the zero value is a zero-cap stack
// elements in the stack will be stored in linked nodes
// stackHeader -> nodeHeader -> nodeHeader ...
type Stack[T any] memory.Pointer

func (s *Stack[T]) Push(val T) (err error) {
	err = s.checkCapacity()
	if err != nil {
		return err
	}
	header := s.header()
	if utils.Asserted {
		if header.length >= header.capacity {
			panic(fmt.Sprintf("header.length(%d) >= header.capacity(%d)", header.length, header.capacity))
		}
	}
	header.length++
	*memory.PointerAs[T](header.nextElementPtr) = val
	header.nextElementPtr += memory.Pointer(memory.Sizeof[T]()) // do not check bound. check it in checkCapacity
	return nil
}

func (s Stack[T]) Length() SizeType {
	if s.pointer().IsNull() {
		return 0
	}
	return s.header().length
}

func (s Stack[T]) Top() T {
	return *s.TopRef()
}

func (s Stack[T]) TopRef() *T {
	if utils.Asserted {
		if s.Length() == 0 {
			panic("top of empty stack")
		}
	}
	return memory.PointerAs[T](s.header().nextElementPtr - memory.Pointer(memory.Sizeof[T]()))
}

func (s *Stack[T]) checkCapacity() error {
	if s.pointer().IsNull() {
		// init
		pageSize := stackNodePageSize[T]()
		page, err := Global.allocPage(pageSize, trace_type.StackHeader, 3)
		if err != nil {
			return err
		}
		ptr := Global.pagePointerOf(page)
		header := memory.PointerAs[stackHeader](ptr)
		header.length = 0
		header.capacity = ((pageSize << memory.BasePageSizeShiftNumber) - stackHeaderSize) / memory.Sizeof[T]()
		header.next = memory.NullPointer
		header.last = memory.NullPointer
		header.nextElementPtr = ptr + memory.Pointer(stackHeaderSize)
		if utils.Asserted {
			if header.capacity <= 0 {
				panic(fmt.Sprintf("header.capacity(%d) <= 0", header.capacity))
			}
		}
		if utils.Debug {
			fmt.Println("init stack header", ptr.String())
		}
		*s = Stack[T](ptr)
		return nil
	}

	header := s.header()
	if utils.Asserted {
		if header.length > header.capacity {
			panic(fmt.Sprintf("header.length(%d) > header.capacity(%d)", header.length, header.capacity))
		}
	}
	if header.length == header.capacity {
		// next node
		pageSize := stackNodePageSize[T]()
		page, err := Global.allocPage(pageSize, trace_type.StackNode, 3)
		if err != nil {
			return err
		}
		ptr := Global.pagePointerOf(page)
		nodeHeader := memory.PointerAs[stackNodeHeader](ptr)
		nodeHeader.next = memory.NullPointer

		header.capacity += ((pageSize << memory.BasePageSizeShiftNumber) - stackNodeHeaderSize) / memory.Sizeof[T]()
		if utils.Asserted {
			if header.length >= header.capacity {
				panic(fmt.Sprintf("after append new node. header.length(%d) >= header.capacity(%d)", header.length, header.capacity))
			}
		}
		if header.next == memory.NullPointer {
			// stack header only
			header.next = ptr
			header.last = ptr
		} else {
			// link to last
			if utils.Asserted {
				if header.last.IsNull() {
					panic("last is null")
				}
				if memory.PointerAs[stackNodeHeader](header.last).next.IsNotNull() {
					panic("next node of last node is not null")
				}
			}
			memory.PointerAs[stackNodeHeader](header.last).next = ptr
			header.last = ptr
		}
		header.nextElementPtr = ptr + memory.Pointer(stackNodeHeaderSize)
		if utils.Debug {
			fmt.Println("alloc stack node", ptr.String())
		}
	}
	return nil
}

func (s *Stack[T]) Move() (moved Stack[T]) {
	moved = *s
	*s = nullStack
	return moved
}

func (s Stack[T]) Moved() bool {
	return s == nullSlice
}

func (s Stack[T]) ToGoSlice() []T {
	gs := make([]T, 0, int(s.Length()))
	s.Iterate(func(e T) {
		gs = append(gs, e)
	})
	return gs
}

func (s Stack[T]) String() string {
	return fmt.Sprintf("%+v", s.ToGoSlice())
}

func (s Stack[T]) Free() {
	if s.pointer().IsNotNull() {
		if utils.Asserted {
			if s.header().capacity == 0 {
				panic("double free?")
			}
		}
		pageSize := stackNodePageSize[T]()
		next := s.header().next
		for next.IsNotNull() {
			nextNext := memory.PointerAs[stackNodeHeader](next).next
			Global.freePointer(next, pageSize)
			if utils.Debug {
				fmt.Println("free stack node", next.String())
			}
			next = nextNext
		}
		Global.freePointer(s.pointer(), pageSize)
		if utils.Debug {
			fmt.Println("free stack header", s.pointer().String())
		}
	}
}

func (s Stack[T]) pointer() memory.Pointer {
	return memory.Pointer(s)
}

func (s Stack[T]) header() *stackHeader {
	if utils.Asserted {
		if s.pointer().IsNull() {
			panic("header of null")
		}
	}
	return memory.PointerAs[stackHeader](s.pointer())
}

// stackHeader is the first node
type stackHeader struct {
	length         SizeType       // total element number in the stack
	capacity       SizeType       // full check
	next           memory.Pointer // null for tail
	last           memory.Pointer // point to last node for quick enstack
	nextElementPtr memory.Pointer // insert enstackd element
}

// nodeHeader is the following node
type stackNodeHeader struct {
	next memory.Pointer // null for tail
}

var stackHeaderSize = memory.Sizeof[stackHeader]()
var stackNodeHeaderSize = memory.Sizeof[stackNodeHeader]()

const nullStack = 0

func stackNodePageSize[T any]() SizeType {
	headerSize := memory.Sizeof[T]() + stackHeaderSize
	return (headerSize + memory.BasePageSize - 1) >> memory.BasePageSizeShiftNumber
}
