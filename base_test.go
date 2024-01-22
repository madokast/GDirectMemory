package managed_memory

import (
	"fmt"
	"gitlab.grandhoo.com/rock/storage/internal/logger"
	"testing"
)

func Test_pageIndexMask(t *testing.T) {
	logger.Info(fmt.Sprintf("%x", pageIndexMask))
}

func Test_pageNumberMask(t *testing.T) {
	logger.Info(fmt.Sprintf("%x", pageNumberMask))
}

func Test_pageIdOf(t *testing.T) {
	page := makePageHandler(123, 321)
	logger.Info(fmt.Sprintf("%x", page))
	logger.Info(page.PageIndex())
	logger.Info(page.Size())
	logger.Info(page.Size() >> basePageSizeShiftNumber)
}

func Test_pageHandlerSize(t *testing.T) {
	logger.Info(pageHandlerSize)
}

func Test_sizeTypeSize(t *testing.T) {
	logger.Info(sizeTypeSize)
}

func Test_humanFriendlyMemorySize(t *testing.T) {
	logger.Info(humanFriendlyMemorySize(1))
	logger.Info(humanFriendlyMemorySize(300))
	logger.Info(humanFriendlyMemorySize(1024))
	logger.Info(humanFriendlyMemorySize(1025))
	logger.Info(humanFriendlyMemorySize(1024 * 1024))
	logger.Info(humanFriendlyMemorySize(5*1024*1024 + 300*1024))
	logger.Info(humanFriendlyMemorySize(2 * 1024 * 1024 * 1024))
	logger.Info(humanFriendlyMemorySize(2*1024*1024*1024 + 500*1024*1024 + 300))
}
