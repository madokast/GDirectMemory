package direct

import (
	"fmt"
	"runtime"
	"sync"
)

type LocalMemory struct {
	localPages     []PageHandler
	globalMemory   Memory
	globalMemoryMu *sync.Mutex
	noCopy         noCopy
}

var memoryLockerMap = make(map[Memory]*sync.Mutex)

const localPoolCapacity = 64         // max capacity of localPages
const maxAllocTimes = 4              // alloc pageNumber*maxAllocTimes page into local
const maxAllocOncePageNumber = 40960 // max alloc a-40960 page once from globalMemory

func (m Memory) NewConcurrentMemory() LocalMemory {
	if asserted {
		if _, ok := memoryLockerMap[m]; !ok {
			panic(fmt.Sprintf("no mutex register to memory %s", m.String()))
		}
	}
	return LocalMemory{
		localPages:     make([]PageHandler, 0, localPoolCapacity),
		globalMemory:   m,
		globalMemoryMu: memoryLockerMap[m],
	}
}

func (m *LocalMemory) NewConcurrentMemory() LocalMemory {
	return LocalMemory{
		localPages:     make([]PageHandler, 0, localPoolCapacity),
		globalMemory:   m.globalMemory,
		globalMemoryMu: m.globalMemoryMu,
	}
}

func (m *LocalMemory) Destroy() {
	if asserted {
		if m.globalMemoryMu == nil {
			panic("double free local-memory")
		}
	}
	m.globalMemoryMu.Lock()
	for _, page := range m.localPages {
		m.globalMemory.freePage(page)
	}
	m.globalMemoryMu.Unlock()
	if asserted {
		m.localPages = m.localPages[:0]
		m.globalMemoryMu = nil
	}
}

func (m *LocalMemory) allocPage(pageNumber SizeType, _type string, callerSkip int) (PageHandler, error) {
	if trace {
		pageHandler, err := m.allocPage0(pageNumber)
		if err != nil {
			return pageHandler, err
		} else {
			_, file, line, _ := runtime.Caller(callerSkip)
			m.globalMemory.tracer().traceAlloc(m.pagePointerOf(pageHandler), _type, pageNumber*basePageSize, file, line)
			return pageHandler, err
		}
	} else {
		return m.allocPage0(pageNumber)
	}
}

func (m *LocalMemory) freePage(pageHandler PageHandler) {
	if trace {
		m.globalMemory.tracer().deTraceAlloc(m.pagePointerOf(pageHandler))
	}
	m.freePage0(pageHandler)
}

func (m *LocalMemory) allocPage0(pageNumber SizeType) (PageHandler, error) {
	if asserted {
		if m.globalMemoryMu == nil {
			panic("use a destroyed memory")
		}
	}
	if asserted {
		if pageNumber == 0 {
			panic("allocPage 0 pageNumber")
		}
	}

	// alloc from local
	localPageLength := len(m.localPages)
	for i := 0; i < localPageLength; i++ {
		curPage := m.localPages[i]
		curPageNumber := curPage.PageNumber()
		if curPageNumber >= pageNumber {
			last := localPageLength - 1
			m.localPages[i] = m.localPages[last]
			m.localPages = m.localPages[:last]
			return curPage, nil
		}
		//if curPageNumber > pageNumber {
		//	// break curPage
		//	curPageIndex := curPage.PageIndex()
		//	m.localPages[i] = makePageHandler(curPageNumber-pageNumber, curPageIndex+pageNumber)
		//	return makePageHandler(pageNumber, curPageIndex), nil
		//}
	}

	// if localPages is full but all is unuseful, free all
	if localPageLength >= localPoolCapacity {
		if asserted {
			if localPageLength > localPoolCapacity {
				panic(fmt.Sprintf("allocPage bad code %d %d", localPageLength, localPoolCapacity))
			}
		}
		for _, page := range m.localPages {
			m.globalMemory.freePage(page)
		}
		m.localPages = m.localPages[:0]
	}

	/* ------------------------------ alloc from globalMemory ------------------------------ */

	if pageNumber >= maxAllocOncePageNumber {
		// too big alloc
		m.globalMemoryMu.Lock()
		page, err := m.globalMemory.allocPage(pageNumber)
		m.globalMemoryMu.Unlock()
		return page, err
	}

	var totalAllocPageNumber SizeType = 0
	var allocTimes = 0
	var err error
	var page PageHandler
	m.globalMemoryMu.Lock()
	for allocTimes < maxAllocTimes && totalAllocPageNumber < maxAllocOncePageNumber && len(m.localPages) < localPoolCapacity {
		page, err = m.globalMemory.allocPage(pageNumber)
		if err != nil {
			break
		}
		m.localPages = append(m.localPages, page)
		allocTimes++
		totalAllocPageNumber += pageNumber
	}
	m.globalMemoryMu.Unlock()

	last := len(m.localPages) - 1
	if page.IsNull() {
		if asserted {
			if err == nil {
				panic("bad code page.IsNull() and err == nil")
			}
		}
		if allocTimes > 0 {
			if asserted {
				if last < 0 {
					panic(fmt.Sprintf("bad code allocTimes(%d) > 0 but last(%d) < 0", allocTimes, last))
				}
			}
			lastPage := m.localPages[last]
			if asserted {
				if lastPage.PageNumber() < pageNumber {
					panic(fmt.Sprintf("bad code allocTimes(%d) > 0 but lastPage.PageNumber()(%d) < pageNumber(%d)",
						allocTimes, lastPage.PageNumber(), pageNumber))
				}
			}
			m.localPages = m.localPages[:last] // pop
			return lastPage, nil
		}
		return nullPageHandle, err
	}

	if asserted {
		if err != nil {
			panic(fmt.Sprint("bad code err is not nil", err, " ", m.localPages))
		}
		if last < 0 {
			panic(fmt.Sprint("last < 0", last, m.localPages))
		}
	}

	lastPage := m.localPages[last]
	m.localPages = m.localPages[:last]
	if asserted {
		if lastPage.PageNumber() < pageNumber {
			panic(fmt.Sprintf("allocPage bad code lastPage.PageNumber(%d) < pageNumber(%d)", lastPage.PageNumber(), pageNumber))
		}
	}

	return lastPage, nil

	// alloc pageNumber * maxAllocTimes
	//allocPageFromProxy := pageNumber * maxAllocTimes
	//if allocPageFromProxy > maxAllocOncePageNumber {
	//	allocPageFromProxy = maxAllocOncePageNumber
	//}
	//m.globalMemoryMu.Lock()
	//page, err := m.globalMemory.allocPage(allocPageFromProxy)
	//m.globalMemoryMu.Unlock()
	//if err != nil {
	//	// try itself
	//	m.globalMemoryMu.Lock()
	//	page, err = m.globalMemory.allocPage(pageNumber)
	//	m.globalMemoryMu.Unlock()
	//	return page, err
	//}

	// break page
	//curPageNumber := page.PageNumber() // curPageNumber maybe greater then allocPageFromProxy
	//curPageIndex := page.PageIndex()
	//retPage := makePageHandler(pageNumber, curPageIndex)
	//remainPage := makePageHandler(curPageNumber-pageNumber, curPageIndex+pageNumber)
	//if localPageLength >= localPoolCapacity {
	//	if asserted {
	//		if localPageLength > localPoolCapacity {
	//			panic(pretty.Sprintf("bad code %d %d", localPageLength, localPoolCapacity))
	//		}
	//	}
	//	last := localPageLength - 1
	//	m.globalMemoryMu.Lock()
	//	m.globalMemory.freePage(m.localPages[last])
	//	m.globalMemoryMu.Unlock()
	//	m.localPages[last] = remainPage
	//} else {
	//	m.localPages = append(m.localPages, remainPage)
	//}
	//return retPage, nil
}

func (m *LocalMemory) freePage0(pageHandler PageHandler) {
	if asserted {
		if m.globalMemoryMu == nil {
			panic("use a destroyed memory")
		}
	}
	if asserted {
		if pageHandler.IsNull() {
			panic("free null page")
		}
		libZero(m.pagePointerOf(pageHandler), pageHandler.Size())
	}

	// free too big
	if pageHandler.PageNumber() > maxAllocOncePageNumber {
		m.globalMemoryMu.Lock()
		m.globalMemory.freePage(pageHandler)
		m.globalMemoryMu.Unlock()
		return
	}

	localPageLength := len(m.localPages)
	if localPageLength >= localPoolCapacity {
		if asserted {
			if localPageLength > localPoolCapacity {
				panic(fmt.Sprintf("bad code localPageLength(%d) > localPoolCapacity(%d)", localPageLength, localPoolCapacity))
			}
		}
		last := localPageLength - 1
		lastPage := m.localPages[last]

		// free the smaller one
		m.globalMemoryMu.Lock()
		if pageHandler.PageNumber() < lastPage.PageNumber() {
			m.globalMemory.freePage(pageHandler)
		} else {
			m.globalMemory.freePage(lastPage)
			m.localPages[last] = pageHandler
		}
		m.globalMemoryMu.Unlock()
		return
	}

	// usually cache it
	m.localPages = append(m.localPages, pageHandler)
}

func (m *LocalMemory) pagePointerOf(pageHandler PageHandler) pointer {
	return m.globalMemory.pagePointerOf(pageHandler)
}

func (m *LocalMemory) String() string {
	return fmt.Sprintf("{localPages:%s}", Jsonify(m.localPages))
}

type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}
