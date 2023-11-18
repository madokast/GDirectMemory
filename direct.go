package direct

import "github.com/madokast/direct/arena"

type Memory struct {
	*arena.Arena
}

func New(blockSize uint) Memory {
	return Memory{Arena: arena.New(blockSize)}
}
