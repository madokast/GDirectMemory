package direct

import "github.com/madokast/direct/stdlib"

type Arena struct {
	blocks []block
}

type block struct {
	stdlib.Memory
	free uintptr //
}
