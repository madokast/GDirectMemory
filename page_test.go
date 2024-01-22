package direct

import (
	"testing"
)

func TestMemory_pageAsBytes(t *testing.T) {
	memory := New(4*1024 + 3)
	t.Log(memory)
	defer memory.Free()
	page, err := memory.allocPage(4)
	PanicErr(err)
	defer memory.freePage(page)
	bytes := memory.pageAsBytes(page)
	memory.pageZero(page)
	bytes = memory.pageAsBytes(page)
	t.Log("page size", page.Size())
	t.Log("bytes size", len(bytes))
	for _, b := range bytes {
		if b != 0 {
			panic(b)
		}
	}
}
