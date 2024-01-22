package direct

import (
	"errors"
	"fmt"
	"reflect"
	"unsafe"
)

func (m Memory) allocPage(pageNumber SizeType) (PageHandler, error) {
	header := m.header()
	// alloc from freed
	if pageNumber == 1 {
		freedBasePageHeader := header.freedBasePageHeader
		if freedBasePageHeader.IsNotNull() {
			nextFreedPage := m.nextFreedPage(freedBasePageHeader)
			header.freedBasePageHeader = nextFreedPage
			if debug {
				fmt.Println("alloc page from freedBasePage", freedBasePageHeader)
			}
			header.allocatedPageNumber++
			return freedBasePageHeader, nil
		}
	} else {
		previous := nullPageHandle
		combinedPageHeader := header.freedCombinedPageHeader
		for combinedPageHeader.IsNotNull() {
			linked := pointerAs[linkedFreePageHeader](m.pagePointerOf(combinedPageHeader))
			if linked.pageNumber >= pageNumber {
				if previous.IsNull() {
					header.freedCombinedPageHeader = linked.next
				} else {
					pointerAs[linkedFreePageHeader](m.pagePointerOf(previous)).next = linked.next
				}
				if debug {
					fmt.Println("alloc", pageNumber, "pages from combinedPage", combinedPageHeader)
				}
				header.allocatedPageNumber += linked.pageNumber
				return combinedPageHeader, nil
			}
			previous = combinedPageHeader
			combinedPageHeader = linked.next
		}
		if debug {
			if header.freedCombinedPageHeader.IsNotNull() {
				fmt.Println("failed to reuse combined pages when alloc", pageNumber, "pages")
			}
		}
	}
	// from new
	newEmpty := header.emptyPageIndex + pageNumber
	if newEmpty > header.maxPageIndex {
		e := fmt.Sprintf("out of memory when alloc %d pages. The memory details is \n%s", pageNumber, m.String())
		return nullPageHandle, errors.New(e)
	}
	handler := makePageHandler(pageNumber, header.emptyPageIndex)
	header.emptyPageIndex = newEmpty
	if debug {
		fmt.Println("alloc page from empty", handler)
	}
	header.allocatedPageNumber += pageNumber
	return handler, nil
}

func (m Memory) freePage(pageHandler PageHandler) {
	if asserted {
		if pageHandler.IsNull() {
			panic("free null page")
		}
		libZero(m.pagePointerOf(pageHandler), pageHandler.Size())
	}
	pageNumber := pageHandler.PageNumber()
	if asserted {
		if pageNumber == 0 {
			panic("pageNumber == 0")
		}
	}
	header := m.header()

	pagePtr := m.pagePointerOf(pageHandler)
	// return to linklist
	if pageNumber == 1 {
		if header.freedBasePageHeader.IsNull() {
			header.freedBasePageHeader = pageHandler
			pointerAs[linkedFreePageHeader](pagePtr).next = nullPageHandle
		} else {
			pointerAs[linkedFreePageHeader](pagePtr).next = header.freedBasePageHeader
			header.freedBasePageHeader = pageHandler
		}
	} else {
		pointerAs[linkedFreePageHeader](pagePtr).pageNumber = pageNumber
		if header.freedCombinedPageHeader.IsNull() {
			header.freedCombinedPageHeader = pageHandler
			pointerAs[linkedFreePageHeader](pagePtr).next = nullPageHandle
		} else {
			pointerAs[linkedFreePageHeader](pagePtr).next = header.freedCombinedPageHeader
			header.freedCombinedPageHeader = pageHandler
		}
	}
	header.allocatedPageNumber -= pageNumber
}

func (m Memory) numberOfFreedBasePages() SizeType {
	var number SizeType = 0
	header := m.header().freedBasePageHeader
	for header.IsNotNull() {
		number++
		header = m.nextFreedPage(header)
	}
	return number
}

// numberOfFreedCombinedPages number is number of FreedCombinedPages and pageNumber is the sum of FreedCombinedPages size
// if FreedCombinedPages are 3,3,4 then number is 3 and pageNumber is 10
func (m Memory) numberOfFreedCombinedPages() (number SizeType, pageNumber SizeType) {
	header := m.header().freedCombinedPageHeader
	for header.IsNotNull() {
		linked := pointerAs[linkedFreePageHeader](m.pagePointerOf(header))
		pageNumber += linked.pageNumber
		number++
		header = linked.next
	}
	return
}

func (m Memory) nextFreedPage(freedPageHandler PageHandler) PageHandler {
	if asserted {
		if freedPageHandler.IsNull() {
			panic("nextFreedPage of null")
		}
	}
	return pointerAs[linkedFreePageHeader](m.pagePointerOf(freedPageHandler)).next
}

type linkedFreePageHeader struct {
	pageNumber SizeType    // 空页大小
	next       PageHandler // 链表
}

// pagePointerOf thread-safe
func (m Memory) pagePointerOf(pageHandler PageHandler) pointer {
	if asserted {
		if pageHandler.IsNull() {
			panic("pagePointerOf of null")
		}
	}
	offset := pointer(pageHandler.PageIndex() << basePageSizeShiftNumber)
	return m.header().pageBasePointer.value + offset
}

func (m Memory) pageZero(pageHandler PageHandler) {
	ptr := m.pagePointerOf(pageHandler)
	size := pageHandler.PageNumber() << basePageSizeShiftNumber
	libZero(ptr, size)
}

func (m Memory) pageAsBytes(pageHandler PageHandler) []byte {
	var bs []byte
	ptr := m.pagePointerOf(pageHandler)
	size := pageHandler.PageNumber() << basePageSizeShiftNumber
	*((*reflect.SliceHeader)(unsafe.Pointer(&bs))) = reflect.SliceHeader{
		Data: ptr.UIntPtr(),
		Len:  size.Int(),
		Cap:  size.Int(),
	}
	return bs
}

func (m Memory) allocatedMemorySize() SizeType {
	header := m.header()
	return header.allocatedPageNumber << basePageSizeShiftNumber
}

func (m Memory) Json() map[string]interface{} {
	info := map[string]interface{}{}
	header := m.header()
	info["pointer"] = m.pointer().String()
	info["libPointer"] = header.libPointer.String()
	info["pageBasePointer"] = header.pageBasePointer.value.String()
	info["maxPageIndex"] = header.maxPageIndex
	info["emptyPageIndex"] = header.emptyPageIndex
	totalPageNumber := header.maxPageIndex
	info["totalPageNumber"] = totalPageNumber
	info["totalMemory"] = totalPageNumber << basePageSizeShiftNumber
	info["totalMemory_h"] = humanFriendlyMemorySize(totalPageNumber << basePageSizeShiftNumber)
	allocatedPageNumber := header.allocatedPageNumber
	info["allocatedPageNumber"] = allocatedPageNumber
	info["allocatedMemory"] = allocatedPageNumber << basePageSizeShiftNumber
	info["allocatedMemory_h"] = humanFriendlyMemorySize(allocatedPageNumber << basePageSizeShiftNumber)
	return info
}

func (m Memory) String() string {
	return Jsonify(m.Json())
}
