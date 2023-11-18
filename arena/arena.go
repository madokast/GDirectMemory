package arena

import (
	"github.com/madokast/direct/stdlib"
)

// Arena offers direct memory
// It's useful to create many gc-free object.
// The zero value of Arena is ready.
type Arena struct {
	allBlocks    []stdlib.Memory
	currentBlock stdlib.Memory
	blockSize    uint
	free         uintptr // free pointer of last block
}

func New(blockSize uint) *Arena {
	blk := stdlib.Malloc(blockSize)
	blocks := make([]stdlib.Memory, 1, 16)
	blocks[0] = blk
	return &Arena{
		allBlocks:    blocks,
		currentBlock: blk,
		blockSize:    blockSize,
		free:         0,
	}
}

// Allocate direct memory of size with align
func (a *Arena) Allocate(size uint, align uint) (ptr uintptr) {
	padSize := alignPadSize(a.free+a.currentBlock.Address, align)
	if a.free+padSize+uintptr(size) > uintptr(a.blockSize) {
		a.currentBlock = stdlib.Malloc(a.blockSize)
		a.allBlocks = append(a.allBlocks, a.currentBlock)
		a.free = 0
		return a.Allocate(size, align)
	}

	ptr = a.currentBlock.Address + a.free + padSize
	a.free += padSize + uintptr(size)
	return ptr
}

func (a *Arena) Free() {
	for _, blk := range a.allBlocks {
		blk.Free()
	}
}

func alignPadSize(ptr uintptr, align uint) uintptr {
	var alignPtr = ptr
	if alignPtr%uintptr(align) != 0 {
		alignPtr &= ^(uintptr(align) - 1)
		alignPtr += uintptr(align)
	}
	return alignPtr - ptr
}
