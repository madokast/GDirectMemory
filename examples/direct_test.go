package examples

import "github.com/madokast/direct"

func ExampleMemory() {
	m := direct.New(1024)
	m.Free()
}
