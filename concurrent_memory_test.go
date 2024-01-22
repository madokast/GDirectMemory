package managed_memory

import (
	"gitlab.grandhoo.com/rock/storage/internal/logger"
	"gitlab.grandhoo.com/rock/storage/storage2/utils/test"
	"testing"
)

func TestConcurrentMemory(t *testing.T) {
	memory := New(4096)
	defer memory.Free()

	concurrentMemory := memory.NewConcurrentMemory()
	defer func() {
		logger.Info(concurrentMemory.localPages)
		concurrentMemory.Destroy()
		logger.Info(memory)
	}()

	page, err := concurrentMemory.allocPage(5)
	test.PanicErr(err)
	logger.Info(page)
	concurrentMemory.freePage(page)

	page, err = concurrentMemory.allocPage(2)
	test.PanicErr(err)
	logger.Info(page)
	concurrentMemory.freePage(page)
}
