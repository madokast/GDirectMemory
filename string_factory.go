package direct

import (
	"fmt"
	"reflect"
	"sync/atomic"
	"unsafe"
)

// StringFactory creates String. Thread-unsafe
type StringFactory struct {
	holder Slice[byte] // str-count + data
	noCopy noCopy
}

func NewStringFactory() StringFactory {
	return StringFactory{
		holder: nullSlice,
	}
}

func (sf *StringFactory) CreateFromGoString(gs string, m *LocalMemory) (s String, err error) {
	var gsLength = SizeType(len(gs))
	if gsLength == 0 {
		return emptyString, nil
	}

	var sfHolder = sf.holder                           // register
	var sfHolderHeader *sliceHeader = nil              // register
	var sfHolderHeaderElementBasePointer = nullPointer // register
	if sfHolder != nullSlice {
		sfHolderHeader = sfHolder.header()
		sfHolderHeaderElementBasePointer = sfHolderHeader.elementBasePointer
		if sfHolderHeader.length+gsLength > sfHolderHeader.capacity {
			if debug {
				fmt.Printf("holder(len=%d, cap=%d) cannot add new string(len=%d)\n", sfHolderHeader.length, sfHolderHeader.capacity, gsLength)
			}
			// detach
			cnt := atomic.AddInt32(pointerAs[int32](sfHolderHeaderElementBasePointer), -1)
			if asserted {
				if cnt < 0 {
					panic(fmt.Sprintf("string holder cnt(%d) < 0", cnt))
				}
			}
			if cnt == 0 {
				sfHolder.Free(m)
			}
			sfHolder = nullSlice
		}
	}
	if sfHolder == nullSlice {
		// int32 for count
		sfHolder, err = makeSlice0[byte](m, gsLength+Sizeof[int32](), "String", 3)
		if err != nil {
			return emptyString, err
		}
		sf.holder = sfHolder
		sfHolderHeader = sfHolder.header()
		sfHolderHeader.length = Sizeof[int32]()
		sfHolderHeaderElementBasePointer = sfHolderHeader.elementBasePointer
		// ref cnt 1 for the factory
		*pointerAs[int32](sfHolderHeaderElementBasePointer) = 1
	}

	if asserted {
		if sf.holder != sfHolder {
			panic("bad code: f.holder != sfHolder")
		}
		if sfHolderHeader == nil {
			panic("bad code: sfHolderHeader == nil")
		}
		if sf.holder.Length()+gsLength > sf.holder.Capacity() {
			panic(fmt.Sprintf("slice capacity is not enough slice length(%d) go string length(%d) slice capacity(%d)",
				sf.holder.Length(), gsLength, sf.holder.Capacity()))
		}
		if sfHolderHeaderElementBasePointer.IsNull() {
			panic("bad code: sfHolderHeaderElementBasePointer.IsNull()")
		}
	}

	s.holder = sfHolder
	s.length = gsLength
	s.ptr = sfHolderHeaderElementBasePointer + pointer(sfHolderHeader.length)

	libMemMove(s.ptr, pointer((*reflect.StringHeader)(unsafe.Pointer(&gs)).Data), gsLength)
	sfHolderHeader.length += s.length

	// add count
	cnt := atomic.AddInt32(pointerAs[int32](sfHolderHeaderElementBasePointer), 1)
	if debug {
		fmt.Printf("alloc string %s in holder count %d\n", gs, cnt)
	}
	return s, nil
}

func (sf *StringFactory) Destroy(m *LocalMemory) {
	holder := sf.holder
	if holder != nullSlice {
		cnt := atomic.AddInt32(pointerAs[int32](holder.header().elementBasePointer), -1)
		if asserted {
			if cnt < 0 {
				panic(fmt.Sprintf("string holder cnt(%d) < 0", cnt))
			}
		}
		if cnt == 0 {
			holder.Free(m)
		}
		sf.holder = nullSlice
	}
}

func init() {
	if Sizeof[byte]() != 1 {
		panic(fmt.Sprint("size of byte is not 1. ", Sizeof[byte]()))
	}
}
