package memory

import (
	"github.com/madokast/direct/utils"
	"testing"
)

func TestLocalMemory(t *testing.T) {
	memory := New(4096)
	defer memory.Free()

	localMemory := memory.NewLocalMemory()
	defer func() {
		t.Log(localMemory.localPages)
		localMemory.Destroy()
		t.Log(memory)
	}()

	page, err := localMemory.AllocPage(5)
	utils.PanicErr(err)
	t.Log(page)
	localMemory.FreePage(page)

	page, err = localMemory.AllocPage(2)
	utils.PanicErr(err)
	t.Log(page)
	localMemory.FreePage(page)
}

func TestLocalMemory2(t *testing.T) {
	memory := New(4096)
	defer memory.Free()

	localMemory := memory.NewLocalMemory()
	defer func() {
		t.Log(localMemory.localPages)
		localMemory.Destroy()
		t.Log(memory)
	}()

	page, err := localMemory.AllocPage(5)
	utils.PanicErr(err)
	t.Log(page)
	ptr := localMemory.PagePointerOf(page)
	utils.Assert(memory.PointerToPageIndex(ptr) == page.PageIndex(), page, memory.PointerToPageIndex(ptr))
	localMemory.FreePointer(ptr, 5)

	page, err = localMemory.AllocPage(2)
	utils.PanicErr(err)
	t.Log(page)
	ptr = localMemory.PagePointerOf(page)
	utils.Assert(memory.PointerToPageIndex(ptr) == page.PageIndex(), page, memory.PointerToPageIndex(ptr))
	localMemory.FreePointer(ptr, 5)

}
