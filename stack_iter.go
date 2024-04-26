package direct

import (
	"fmt"
	"github.com/madokast/direct/memory"
	"github.com/madokast/direct/utils"
)

type StackIterator[T any] struct {
	cur        memory.Pointer
	nextNode   memory.Pointer
	index      SizeType
	nodeLength SizeType
	length     SizeType
	noCopy     utils.NoCopy
}

func (s Stack[T]) Iterator() (iter StackIterator[T]) {
	ptr := s.pointer()
	if ptr.IsNull() {
		// do nothing
	} else {
		header := memory.PointerAs[stackHeader](ptr)
		iter.cur = ptr + memory.Pointer(stackHeaderSize) - memory.Pointer(memory.Sizeof[T]()) // -1
		iter.nextNode = header.next
		iter.index = SizeTypeMax // -1
		iter.nodeLength = ((stackNodePageSize[T]() << memory.BasePageSizeShiftNumber) - stackHeaderSize) / memory.Sizeof[T]()
		iter.length = header.length
	}
	return
}

func (it *StackIterator[T]) Next() bool {
	if it.nodeLength == 0 {
		if it.nextNode.IsNull() {
			return false
		} else {
			nextNodeHeader := memory.PointerAs[stackNodeHeader](it.nextNode)
			it.cur = it.nextNode + memory.Pointer(stackNodeHeaderSize)
			it.nextNode = nextNodeHeader.next
			it.index++
			it.nodeLength = ((stackNodePageSize[T]()<<memory.BasePageSizeShiftNumber)-stackNodeHeaderSize)/memory.Sizeof[T]() - 1
			if utils.Asserted {
				if it.index >= it.length {
					panic(fmt.Sprintf("it.index(%d) >= it.length(%d)", it.index, it.length))
				}
			}
			return true
		}
	} else {
		it.cur += memory.Pointer(memory.Sizeof[T]())
		it.nodeLength--
		it.index++
		return it.index < it.length
	}
}

func (it *StackIterator[T]) Value() T {
	return *it.Ref()
}

func (it *StackIterator[T]) Ref() *T {
	if utils.Asserted {
		if it.index >= it.length {
			panic("check Next() before access")
		}
	}
	return memory.PointerAs[T](it.cur)
}

func (it *StackIterator[T]) Index() SizeType {
	if utils.Asserted {
		if it.index >= it.length {
			panic("check Next() before access")
		}
	}
	return it.index
}

func (s Stack[T]) Iterate(iter func(T)) {
	if s.pointer().IsNotNull() {
		pageSize := stackNodePageSize[T]()
		header := s.header()
		length := header.length

		var index SizeType = 0
		cursor := s.pointer() + memory.Pointer(stackHeaderSize)
		nodeCap := ((pageSize << memory.BasePageSizeShiftNumber) - stackHeaderSize) / memory.Sizeof[T]()

		for index < length {
			iter(*memory.PointerAs[T](cursor))
			cursor += memory.Pointer(memory.Sizeof[T]())
			index++
			nodeCap--
			if nodeCap == 0 {
				break
			}
		}

		next := header.next
		for next.IsNotNull() {
			cursor = next + memory.Pointer(stackNodeHeaderSize)
			nodeCap = ((pageSize << memory.BasePageSizeShiftNumber) - stackNodeHeaderSize) / memory.Sizeof[T]()
			for index < length {
				iter(*memory.PointerAs[T](cursor))
				cursor += memory.Pointer(memory.Sizeof[T]())
				index++
				nodeCap--
				if nodeCap == 0 {
					break
				}
			}
			next = memory.PointerAs[stackNodeHeader](next).next
		}
	}
}
