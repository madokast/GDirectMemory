package managed_memory

import (
	"gitlab.grandhoo.com/rock/storage/internal/logger"
	"sync"
	"unsafe"
)

// Memory is thread-unsafe. Make a LocalMemory in thread
type Memory pointer

type memoryTrait interface {
	allocPage(pageNumber SizeType) (PageHandler, error)
	freePage(pageHandler PageHandler)
	pagePointerOf(pageHandler PageHandler) pointer
}

type memoryHeader struct {
	pageBasePointer         cacheShareWord[pointer] // pointer to the per first page
	freedBasePageHeader     PageHandler             // a list header of freed page of basePageSize. Nullable
	freedCombinedPageHeader PageHandler             // a list header of freed combined page. Nullable
	maxPageIndex            SizeType                // OOM when emptyPageIndex > maxPageIndex and no free
	emptyPageIndex          SizeType                // next page when no proper freed page
	libPointer              pointer                 // used for free
	allocatedPageNumber     SizeType                // statistical
}

func New(size SizeType) Memory {
	ptr := libMalloc(((size + 7) & (sizeTypeMax - 7)) + 8) // align
	m := Memory((ptr + 7) & pointer(sizeTypeMax-7))
	header := m.header()
	header.pageBasePointer = cacheShareWord[pointer]{
		value: m.pointer() + pointer(memoryHeaderSize) - pointer(basePageSize), // subtract as from one
	}
	header.freedBasePageHeader = nullPageHandle
	header.freedCombinedPageHeader = nullPageHandle
	header.maxPageIndex = (size - memoryHeaderSize) >> basePageSizeShiftNumber
	header.emptyPageIndex = 1 // from one pass the null
	header.libPointer = ptr
	header.allocatedPageNumber = 0
	memoryLockerMap[m] = new(sync.Mutex)
	return m
}

func (m Memory) Free() {
	if m.allocatedMemorySize() != 0 {
		logger.Panic("memory leak!", m.String())
	}
	libFree(m.header().libPointer)
}

func (m Memory) pointer() pointer {
	return pointer(m)
}

func (m Memory) maxPageIndex() SizeType {
	return m.header().maxPageIndex
}

func (m Memory) emptyPageIndex() SizeType {
	return m.header().emptyPageIndex
}

func (m Memory) freedBasePageHeader() PageHandler {
	return m.header().freedBasePageHeader
}

func (m Memory) freedCombinedPageHeader() PageHandler {
	return m.header().freedCombinedPageHeader
}

var memoryHeaderSize = SizeType(unsafe.Sizeof(memoryHeader{}))

func (m Memory) header() *memoryHeader {
	return (*memoryHeader)(unsafe.Pointer(uintptr(m)))
}
