package direct

import (
	"testing"
)

func TestlocalMemory(t *testing.T) {
	memory := New(4096)
	defer memory.Free()

	localMemory := memory.NewLocalMemory()
	defer func() {
		t.Log(localMemory.localPages)
		localMemory.Destroy()
		t.Log(memory)
	}()

	page, err := localMemory.allocPage(5, "test", 1)
	PanicErr(err)
	t.Log(page)
	localMemory.freePage(page)

	page, err = localMemory.allocPage(2, "test", 1)
	PanicErr(err)
	t.Log(page)
	localMemory.freePage(page)
}
