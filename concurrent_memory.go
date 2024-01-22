package managed_memory

import (
	"fmt"
	"gitlab.grandhoo.com/rock/storage/internal/logger"
	"gitlab.grandhoo.com/rock/storage/storage2/utils/rpc/serializer"
	"sync"
)

type LocalMemory struct {
	localPages []PageHandler
	proxy      Memory
	mu         *sync.Mutex
	noCopy     noCopy
}

var memoryLockerMap = make(map[Memory]*sync.Mutex)

const localPoolCapacity = 64         // max capacity of localPages
const maxAllocTimes = 4              // alloc pageNumber*maxAllocTimes page into local
const maxAllocOncePageNumber = 40960 // max alloc a-40960 page once from proxy

func (m Memory) NewConcurrentMemory() LocalMemory {
	return LocalMemory{
		localPages: make([]PageHandler, 0, localPoolCapacity),
		proxy:      m,
		mu:         memoryLockerMap[m],
	}
}

func (m *LocalMemory) NewConcurrentMemory() LocalMemory {
	return LocalMemory{
		localPages: make([]PageHandler, 0, localPoolCapacity),
		proxy:      m.proxy,
		mu:         memoryLockerMap[m.proxy],
	}
}

func (m *LocalMemory) Destroy() {
	if asserted {
		if m.mu == nil {
			logger.Panic("double free local-memory")
		}
	}
	m.mu.Lock()
	for _, page := range m.localPages {
		m.proxy.freePage(page)
	}
	m.mu.Unlock()
	m.localPages = m.localPages[:0]
	if asserted {
		m.mu = nil
	}
}

func (m *LocalMemory) allocPage(pageNumber SizeType) (PageHandler, error) {
	if asserted {
		if m.mu == nil {
			logger.Panic("use a destroyed memory")
		}
	}
	if asserted {
		if pageNumber == 0 {
			logger.Panic("allocPage 0 pageNumber")
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
				logger.Panic(fmt.Sprintf("allocPage bad code %d %d", localPageLength, localPoolCapacity))
			}
		}
		m.Destroy()
	}

	/* ------------------------------ alloc from proxy ------------------------------ */

	if pageNumber >= maxAllocOncePageNumber {
		// too big alloc
		m.mu.Lock()
		page, err := m.proxy.allocPage(pageNumber)
		m.mu.Unlock()
		if err != nil {
			logger.Info("pageNumber", pageNumber, "localPages", m.localPages)
		}
		return page, err
	}

	var totalAllocPageNumber SizeType = 0
	var allocTimes = 0
	var err error
	var page PageHandler
	m.mu.Lock()
	for allocTimes < maxAllocTimes && totalAllocPageNumber < maxAllocOncePageNumber && len(m.localPages) < localPoolCapacity {
		page, err = m.proxy.allocPage(pageNumber)
		if err != nil {
			break
		}
		m.localPages = append(m.localPages, page)
		allocTimes++
		totalAllocPageNumber += pageNumber
	}
	m.mu.Unlock()

	last := len(m.localPages) - 1
	if page.IsNull() {
		if asserted {
			if err == nil {
				logger.Panic("bad code page.IsNull() and err == nil")
			}
		}
		if last == -1 {
			m.mu.Lock()
			logger.Error(err, m.proxy.String())
			m.mu.Unlock()
			return nullPageHandle, err
		}
		if last >= 0 {
			lastPage := m.localPages[last]
			if lastPage.PageNumber() == pageNumber {
				m.localPages = m.localPages[:last]
				return lastPage, nil
			} else {
				logger.Panic("allocPage bad code lastPage.PageNumber(%d) < pageNumber(%d)", lastPage.PageNumber(), pageNumber)
				return nullPageHandle, err
			}
		}
	}

	if asserted {
		if err != nil {
			logger.Panic("bad code err is not nil", err, m.localPages)
		}
		if last < 0 {
			logger.Panic("last < 0", last, m.localPages)
		}
	}

	lastPage := m.localPages[last]
	m.localPages = m.localPages[:last]
	if asserted {
		if lastPage.PageNumber() < pageNumber {
			logger.Panic("allocPage bad code lastPage.PageNumber(%d) < pageNumber(%d)", lastPage.PageNumber(), pageNumber)
		}
	}

	return lastPage, nil

	// alloc pageNumber * maxAllocTimes
	//allocPageFromProxy := pageNumber * maxAllocTimes
	//if allocPageFromProxy > maxAllocOncePageNumber {
	//	allocPageFromProxy = maxAllocOncePageNumber
	//}
	//m.mu.Lock()
	//page, err := m.proxy.allocPage(allocPageFromProxy)
	//m.mu.Unlock()
	//if err != nil {
	//	// try itself
	//	m.mu.Lock()
	//	page, err = m.proxy.allocPage(pageNumber)
	//	m.mu.Unlock()
	//	if err != nil {
	//		logger.Info(m.localPages)
	//	}
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
	//			logger.Panic(pretty.Sprintf("bad code %d %d", localPageLength, localPoolCapacity))
	//		}
	//	}
	//	last := localPageLength - 1
	//	m.mu.Lock()
	//	m.proxy.freePage(m.localPages[last])
	//	m.mu.Unlock()
	//	m.localPages[last] = remainPage
	//} else {
	//	m.localPages = append(m.localPages, remainPage)
	//}
	//return retPage, nil
}

func (m *LocalMemory) freePage(pageHandler PageHandler) {
	if asserted {
		if m.mu == nil {
			logger.Panic("use a destroyed memory")
		}
	}
	if asserted {
		if pageHandler.IsNull() {
			logger.Panic("free null page")
		}
		libZero(m.pagePointerOf(pageHandler), pageHandler.Size())
	}

	// free too big
	if pageHandler.PageNumber() > maxAllocOncePageNumber {
		m.mu.Lock()
		m.proxy.freePage(pageHandler)
		m.mu.Unlock()
		return
	}

	localPageLength := len(m.localPages)
	if localPageLength >= localPoolCapacity {
		if asserted {
			if localPageLength > localPoolCapacity {
				logger.Panic(fmt.Sprintf("bad code localPageLength(%d) > localPoolCapacity(%d)", localPageLength, localPoolCapacity))
			}
		}
		last := localPageLength - 1
		lastPage := m.localPages[last]

		// free the smaller one
		m.mu.Lock()
		if pageHandler.PageNumber() < lastPage.PageNumber() {
			m.proxy.freePage(pageHandler)
		} else {
			m.proxy.freePage(lastPage)
			m.localPages[last] = pageHandler
		}
		m.mu.Unlock()
		return
	}

	// usually cache it
	m.localPages = append(m.localPages, pageHandler)
}

func (m *LocalMemory) pagePointerOf(pageHandler PageHandler) pointer {
	return m.proxy.pagePointerOf(pageHandler)
}

func (m *LocalMemory) String() string {
	return fmt.Sprintf("{localPages:%s}", serializer.Jsonify(m.localPages))
}

type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}
