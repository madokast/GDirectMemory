package memory

import (
	"fmt"
	"github.com/madokast/direct/utils"
)

type LocalMemory struct {
	localPages   []PageHandler
	globalMemory Memory
	noCopy       utils.NoCopy
}

const localPoolCapacity = 64         // max capacity of localPages
const maxAllocTimes = 4              // alloc pageNumber*maxAllocTimes page into local
const maxAllocOncePageNumber = 40960 // max alloc a-40960 page once from globalMemory

func (m Memory) NewLocalMemory() LocalMemory {
	return LocalMemory{
		localPages:   make([]PageHandler, 0, localPoolCapacity),
		globalMemory: m,
	}
}

func (m *LocalMemory) NewLocalMemory() LocalMemory {
	return LocalMemory{
		localPages:   make([]PageHandler, 0, localPoolCapacity),
		globalMemory: m.globalMemory,
	}
}

func (m *LocalMemory) Destroy() {
	if utils.Asserted {
		if len(m.localPages) == 1 && m.localPages[0] == nullPageHandle {
			panic("double free local-memory")
		}
	}
	mu := &m.globalMemory.header().mu
	mu.Lock()
	for _, page := range m.localPages {
		m.globalMemory.freePage(page)
	}
	mu.Unlock()
	if utils.Asserted {
		m.localPages = m.localPages[:0]
		m.localPages = append(m.localPages, nullPageHandle) // for check only
	}
}

func (m *LocalMemory) AllocPage(pageNumber SizeType) (PageHandler, error) {
	if utils.Asserted {
		if len(m.localPages) == 1 && m.localPages[0] == nullPageHandle {
			panic("use a destroyed memory")
		}
	}
	if utils.Asserted {
		if pageNumber == 0 {
			panic("AllocPage 0 pageNumber")
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
		//	m.localPages[i] = MakePageHandler(curPageNumber-pageNumber, curPageIndex+pageNumber)
		//	return MakePageHandler(pageNumber, curPageIndex), nil
		//}
	}

	// if localPages is full but all is unuseful, free all
	if localPageLength >= localPoolCapacity {
		if utils.Asserted {
			if localPageLength > localPoolCapacity {
				panic(fmt.Sprintf("AllocPage bad code %d %d", localPageLength, localPoolCapacity))
			}
		}
		for _, page := range m.localPages {
			m.globalMemory.freePage(page)
		}
		m.localPages = m.localPages[:0]
	}

	/* ------------------------------ alloc from globalMemory ------------------------------ */
	mu := &m.globalMemory.header().mu

	if pageNumber >= maxAllocOncePageNumber {
		// too big alloc
		mu.Lock()
		page, err := m.globalMemory.allocPage(pageNumber)
		mu.Unlock()
		return page, err
	}

	var totalAllocPageNumber SizeType = 0
	var allocTimes = 0
	var err error
	var page PageHandler
	mu.Lock()
	for allocTimes < maxAllocTimes && totalAllocPageNumber < maxAllocOncePageNumber && len(m.localPages) < localPoolCapacity {
		page, err = m.globalMemory.allocPage(pageNumber)
		if err != nil {
			break
		}
		m.localPages = append(m.localPages, page)
		allocTimes++
		totalAllocPageNumber += pageNumber
	}
	mu.Unlock()

	last := len(m.localPages) - 1
	if page.IsNull() {
		if utils.Asserted {
			if err == nil {
				panic("bad code page.IsNull() and err == nil")
			}
		}
		if allocTimes > 0 {
			if utils.Asserted {
				if last < 0 {
					panic(fmt.Sprintf("bad code allocTimes(%d) > 0 but last(%d) < 0", allocTimes, last))
				}
			}
			lastPage := m.localPages[last]
			if utils.Asserted {
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

	if utils.Asserted {
		if err != nil {
			panic(fmt.Sprint("bad code err is not nil", err, " ", m.localPages))
		}
		if last < 0 {
			panic(fmt.Sprint("last < 0", last, m.localPages))
		}
	}

	lastPage := m.localPages[last]
	m.localPages = m.localPages[:last]
	if utils.Asserted {
		if lastPage.PageNumber() < pageNumber {
			panic(fmt.Sprintf("AllocPage bad code lastPage.PageNumber(%d) < pageNumber(%d)", lastPage.PageNumber(), pageNumber))
		}
	}

	return lastPage, nil
}

func (m *LocalMemory) FreePage(pageHandler PageHandler) {
	if utils.Asserted {
		if len(m.localPages) == 1 && m.localPages[0] == nullPageHandle {
			panic("use a destroyed memory")
		}
	}
	if utils.Asserted {
		if pageHandler.IsNull() {
			panic("free null page")
		}
		LibZero(m.PagePointerOf(pageHandler), pageHandler.Size())
	}

	// free too big
	if pageHandler.PageNumber() > maxAllocOncePageNumber {
		mu := &m.globalMemory.header().mu
		mu.Lock()
		m.globalMemory.freePage(pageHandler)
		mu.Unlock()
		return
	}

	localPageLength := len(m.localPages)
	if localPageLength >= localPoolCapacity {
		mu := &m.globalMemory.header().mu
		if utils.Asserted {
			if localPageLength > localPoolCapacity {
				panic(fmt.Sprintf("bad code localPageLength(%d) > localPoolCapacity(%d)", localPageLength, localPoolCapacity))
			}
		}
		last := localPageLength - 1
		lastPage := m.localPages[last]

		// free the smaller one
		mu.Lock()
		if pageHandler.PageNumber() < lastPage.PageNumber() {
			m.globalMemory.freePage(pageHandler)
		} else {
			m.globalMemory.freePage(lastPage)
			m.localPages[last] = pageHandler
		}
		mu.Unlock()
		return
	}

	// usually cache it
	m.localPages = append(m.localPages, pageHandler)
}

func (m *LocalMemory) PagePointerOf(pageHandler PageHandler) Pointer {
	return m.globalMemory.PagePointerOf(pageHandler)
}

func (m *LocalMemory) FreePointer(ptr Pointer, pageNumber SizeType) {
	index := m.globalMemory.PointerToPageIndex(ptr)
	pageHandler := MakePageHandler(pageNumber, index)
	m.FreePage(pageHandler)
}

func (m *LocalMemory) String() string {
	return fmt.Sprintf("{localPages:%s}", utils.Jsonify(m.localPages))
}
