package managed_memory

import (
	"gitlab.grandhoo.com/rock/storage/internal/logger"
	"gitlab.grandhoo.com/rock/storage/storage2/utils/test"
	"testing"
)

func TestMemory_pageAsBytes(t *testing.T) {
	memory := New(4*1024 + 3)
	logger.Info(memory)
	defer memory.Free()
	page, err := memory.allocPage(4)
	test.PanicErr(err)
	defer memory.freePage(page)
	bytes := memory.pageAsBytes(page)
	memory.pageZero(page)
	bytes = memory.pageAsBytes(page)
	logger.Info("page size", page.Size())
	logger.Info("bytes size", len(bytes))
	for _, b := range bytes {
		if b != 0 {
			panic(b)
		}
	}
}
