package managed_memory

import (
	"gitlab.grandhoo.com/rock/storage/internal/logger"
	"reflect"
	"unsafe"
)

// StringFactory creates String. Thread-unsafe
type StringFactory struct {
	m      *LocalMemory
	holder Slice[byte]
	noCopy noCopy
}

func NewStringFactory(m *LocalMemory) StringFactory {
	return StringFactory{
		m:      m,
		holder: nullSlice,
	}
}

func (sf *StringFactory) CreateFromGoString(gs string) (s String, err error) {
	if asserted {
		if sf.m == nil {
			logger.Panic("use a freed string-factory")
		}
	}

	if len(gs) == 0 {
		return emptyString, nil
	}
	var sfHolder = sf.holder              // register
	var sfHolderHeader *sliceHeader = nil // register
	if sfHolder != nullSlice {
		sfHolderHeader = sfHolder.header()
		if sfHolderHeader.length+SizeType(len(gs)) > sfHolderHeader.capacity {
			sfHolder = nullSlice
		}
	}
	if sfHolder == nullSlice {
		sfHolder, err = MakeSlice[byte](sf.m, SizeType(len(gs)))
		if err != nil {
			return emptyString, err
		}
		sf.holder = sfHolder
		s.holder = sfHolder
		sfHolderHeader = sfHolder.header()
	}

	if asserted {
		if sf.holder != sfHolder {
			logger.Panic("bad code: f.holder != sfHolder")
		}
		if sfHolderHeader == nil {
			logger.Panic("bad code: sfHolderHeader == nil")
		}
		if sf.holder.Length()+SizeType(len(gs)) > sf.holder.Capacity() {
			logger.Panic("slice capacity is not enough", sf.holder.Length(), len(gs), sf.holder.Capacity())
		}
	}

	s.length = SizeType(len(gs))
	s.ptr = sfHolderHeader.elementBasePointer + pointer(sfHolderHeader.length)

	libMemMove(s.ptr, pointer((*reflect.StringHeader)(unsafe.Pointer(&gs)).Data), SizeType(len(gs)))
	sfHolderHeader.length += s.length

	return s, nil
}

func (sf *StringFactory) Destroy() {
	// do nothing
	if asserted {
		if sf.m == nil {
			logger.Panic("double free")
		}
		sf.m = nil
	}
}

func init() {
	if Sizeof[byte]() != 1 {
		logger.Panic("size of byte is not 1. ", Sizeof[byte]())
	}
}
