package managed_memory

import (
	"gitlab.grandhoo.com/rock/storage/internal/logger"
	"gitlab.grandhoo.com/rock/storage/storage2/utils/test"
	"testing"
)

func TestMemory_memoryHeaderSize(t *testing.T) {
	logger.Info(memoryHeaderSize)
}

func TestMemory_newEmpty(t *testing.T) {
	memory := New(4 * 1024)
	defer memory.Free()
	logger.Info(memory)
}

func TestMemory_alloc1(t *testing.T) {
	memory := New(4 * 1024)
	defer memory.Free()
	page, err := memory.allocPage(1)
	test.PanicErr(err)
	logger.Info(page)
	logger.Info(memory)
	memory.freePage(page)
}

func TestMemory_alloc1_free(t *testing.T) {
	memory := New(4 * 1024)
	defer memory.Free()
	page, err := memory.allocPage(1)
	memory.freePage(page)
	test.PanicErr(err)
	logger.Info(page)
	logger.Info(memory)
}

func TestMemory_alloc11_free(t *testing.T) {
	memory := New(4 * 1024)
	defer memory.Free()
	page, err := memory.allocPage(1)
	memory.freePage(page)
	page, err = memory.allocPage(1)
	memory.freePage(page)
	test.PanicErr(err)
	logger.Info(page)
	logger.Info(memory)
}

func TestMemory_alloc2_free(t *testing.T) {
	memory := New(4 * 1024)
	defer memory.Free()
	page, err := memory.allocPage(2)
	memory.freePage(page)
	test.PanicErr(err)
	logger.Info(page)
	logger.Info(memory)
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
	test.PanicErr(err)
	logger.Info(memory)
}

func TestMemory_alloc_2433(t *testing.T) {
	memory := New(4 * 1024)
	defer memory.Free()
	page, err := memory.allocPage(2)
	logger.Info(memory.allocatedMemorySize())
	memory.freePage(page)
	page, err = memory.allocPage(4)
	logger.Info(memory.allocatedMemorySize())
	memory.freePage(page)
	page2, err := memory.allocPage(3)
	logger.Info(memory.allocatedMemorySize())
	page, err = memory.allocPage(3)
	logger.Info(memory.allocatedMemorySize())
	test.PanicErr(err)
	memory.freePage(page)
	logger.Info(memory)
	memory.freePage(page2)
	logger.Info(memory)
}
