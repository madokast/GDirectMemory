package direct

import (
	"testing"
)

func TestConcurrentMemory(t *testing.T) {
	memory := New(4096)
	defer memory.Free()

	concurrentMemory := memory.NewConcurrentMemory()
	defer func() {
		t.Log(concurrentMemory.localPages)
		concurrentMemory.Destroy()
		t.Log(memory)
	}()

	page, err := concurrentMemory.allocPage(5, "test", 1)
	PanicErr(err)
	t.Log(page)
	concurrentMemory.freePage(page)

	page, err = concurrentMemory.allocPage(2, "test", 1)
	PanicErr(err)
	t.Log(page)
	concurrentMemory.freePage(page)
}
