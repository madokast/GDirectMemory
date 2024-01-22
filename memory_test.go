package direct

import (
	"testing"
)

func TestMemory_memoryHeaderSize(t *testing.T) {
	t.Log(memoryHeaderSize)
}

func TestMemory_newEmpty(t *testing.T) {
	memory := New(4 * 1024)
	defer memory.Free()
	t.Log(memory)
}

func TestMemory_alloc1(t *testing.T) {
	memory := New(4 * 1024)
	defer memory.Free()
	page, err := memory.allocPage(1)
	PanicErr(err)
	t.Log(page)
	t.Log(memory)
	memory.freePage(page)
}

func TestMemory_alloc1_free(t *testing.T) {
	memory := New(4 * 1024)
	defer memory.Free()
	page, err := memory.allocPage(1)
	memory.freePage(page)
	PanicErr(err)
	t.Log(page)
	t.Log(memory)
}

func TestMemory_alloc11_free(t *testing.T) {
	memory := New(4 * 1024)
	defer memory.Free()
	page, err := memory.allocPage(1)
	memory.freePage(page)
	page, err = memory.allocPage(1)
	memory.freePage(page)
	PanicErr(err)
	t.Log(page)
	t.Log(memory)
}

func TestMemory_alloc2_free(t *testing.T) {
	memory := New(4 * 1024)
	defer memory.Free()
	page, err := memory.allocPage(2)
	memory.freePage(page)
	PanicErr(err)
	t.Log(page)
	t.Log(memory)
}

func TestMemory_alloc2121_free(t *testing.T) {
	memory := New(4 * 1024)
	defer memory.Free()
	page, err := memory.allocPage(2)
	memory.freePage(page)
	page, err = memory.allocPage(1)
	memory.freePage(page)
	page, err = memory.allocPage(2)
	memory.freePage(page)
	page, err = memory.allocPage(1)
	memory.freePage(page)
	PanicErr(err)
	t.Log(memory)
}

func TestMemory_alloc_2433(t *testing.T) {
	memory := New(4 * 1024)
	defer memory.Free()
	page, err := memory.allocPage(2)
	t.Log(memory.allocatedMemorySize())
	memory.freePage(page)
	page, err = memory.allocPage(4)
	t.Log(memory.allocatedMemorySize())
	memory.freePage(page)
	page2, err := memory.allocPage(3)
	t.Log(memory.allocatedMemorySize())
	page, err = memory.allocPage(3)
	t.Log(memory.allocatedMemorySize())
	PanicErr(err)
	memory.freePage(page)
	t.Log(memory)
	memory.freePage(page2)
	t.Log(memory)
}
