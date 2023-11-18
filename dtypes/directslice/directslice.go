package directslice

import (
	"github.com/madokast/direct/arena"
)

// Slice represents a gc-free slice in direct memory
// The structure consists of reflect.SliceHeader amd arena
type Slice[E any] struct {
	address  uintptr
	length   int
	capacity int
	arena    *arena.Arena
}
