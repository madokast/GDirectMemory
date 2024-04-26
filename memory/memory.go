package memory

import (
	"fmt"
	"github.com/madokast/direct/utils/spin"
	"os"
	"unsafe"
)

// Memory is thread-unsafe. Make a LocalMemory in thread
type Memory Pointer

type memoryHeader struct {
	pageBasePointer         Pointer     // pointer to the zero-th page
	freedBasePageHeader     PageHandler // a list header of freed page of BasePageSize. Nullable
	freedCombinedPageHeader PageHandler // a list header of freed combined page. Nullable
	maxPageIndex            SizeType    // OOM when emptyPageIndex > maxPageIndex and no free
	emptyPageIndex          SizeType    // next page when no proper freed page
	libPointer              Pointer     // used for free
	allocatedPageNumber     SizeType    // statistical
	mu                      spin.Mutex
}

const NullMemory = Memory(NullPointer)

func New(size SizeType) Memory {
	ptr := LibMalloc(((size + 7) & (SizeTypeMax - 7)) + 8) // align
	m := Memory((ptr + 7) & Pointer(SizeTypeMax-7))
	header := m.header()
	header.pageBasePointer = m.pointer() + Pointer(memoryHeaderSize) - Pointer(BasePageSize)
	header.freedBasePageHeader = nullPageHandle
	header.freedCombinedPageHeader = nullPageHandle
	header.maxPageIndex = (size - memoryHeaderSize) >> BasePageSizeShiftNumber
	header.emptyPageIndex = 1 // from one pass the null
	header.libPointer = ptr
	header.allocatedPageNumber = 0
	header.mu = spin.Mutex{} // zero
	if Trace {
		m.startTrace()
	}
	return m
}

func (m Memory) Free() {
	LibFree(m.header().libPointer)
	if Trace {
		m.deleteTracer()
	}
}

// IsMemoryLeak call it before Free
func (m Memory) IsMemoryLeak() bool {
	if Trace {
		return m.Tracer().hasLeak()
	} else {
		_, _ = fmt.Fprintln(os.Stderr, "Memory trace is off. Turn on for leak check")
		return false
	}
}

// MemoryLeakInfo call it before Free
func (m Memory) MemoryLeakInfo() string {
	if Trace {
		return fmt.Sprint("Statistic info: \n", m.String(), "\n leaking objects: \n", m.Tracer().leakReport())
	} else {
		_, _ = fmt.Fprintln(os.Stderr, "Memory trace is off. Turn on for more info")
		return fmt.Sprint("Statistic info: ", m.String())
	}
}

func (m Memory) IsNull() bool {
	return m.pointer().IsNull()
}

func (m Memory) pointer() Pointer {
	return Pointer(m)
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

func (m Memory) AllocatedPageNumber() SizeType {
	return m.header().allocatedPageNumber
}

var memoryHeaderSize = SizeType(unsafe.Sizeof(memoryHeader{}))

func (m Memory) header() *memoryHeader {
	return (*memoryHeader)(unsafe.Pointer(uintptr(m)))
}
