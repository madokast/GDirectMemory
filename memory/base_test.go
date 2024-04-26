package memory

import (
	"fmt"
	"testing"
)

func Test_pageIndexMask(t *testing.T) {
	t.Log(fmt.Sprintf("%x", pageIndexMask))
}

func Test_pageNumberMask(t *testing.T) {
	t.Log(fmt.Sprintf("%x", pageNumberMask))
}

func Test_pageIdOf(t *testing.T) {
	page := MakePageHandler(123, 321)
	t.Log(fmt.Sprintf("%x", page))
	t.Log(page.PageIndex())
	t.Log(page.Size())
	t.Log(page.Size() >> BasePageSizeShiftNumber)
}

func Test_pageHandlerSize(t *testing.T) {
	t.Log(pageHandlerSize)
}

func Test_sizeTypeSize(t *testing.T) {
	t.Log(sizeTypeSize)
}

func Test_humanFriendlyMemorySize(t *testing.T) {
	t.Log(HumanFriendlyMemorySize(1))
	t.Log(HumanFriendlyMemorySize(300))
	t.Log(HumanFriendlyMemorySize(1024))
	t.Log(HumanFriendlyMemorySize(1025))
	t.Log(HumanFriendlyMemorySize(1024 * 1024))
	t.Log(HumanFriendlyMemorySize(5*1024*1024 + 300*1024))
	t.Log(HumanFriendlyMemorySize(2 * 1024 * 1024 * 1024))
	t.Log(HumanFriendlyMemorySize(2*1024*1024*1024 + 500*1024*1024 + 300))
}
