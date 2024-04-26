package memory

import (
	"fmt"
	"github.com/madokast/direct/utils"
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
			if utils.Debug {
				fmt.Println("alloc page from freedBasePage", freedBasePageHeader)
			}
			header.allocatedPageNumber++
			return freedBasePageHeader, nil
		}
	} else {
		previous := nullPageHandle
		combinedPageHeader := header.freedCombinedPageHeader
		for combinedPageHeader.IsNotNull() {
			linked := PointerAs[linkedFreePageHeader](m.PagePointerOf(combinedPageHeader))
			if linked.pageNumber >= pageNumber {
				if previous.IsNull() {
					header.freedCombinedPageHeader = linked.next
				} else {
					PointerAs[linkedFreePageHeader](m.PagePointerOf(previous)).next = linked.next
				}
				if utils.Debug {
					fmt.Println("alloc", pageNumber, "pages from combinedPage", combinedPageHeader)
				}
				header.allocatedPageNumber += linked.pageNumber
				return combinedPageHeader, nil
			}
			previous = combinedPageHeader
			combinedPageHeader = linked.next
		}
		if utils.Debug {
			if header.freedCombinedPageHeader.IsNotNull() {
				fmt.Println("failed to reuse combined pages when alloc", pageNumber, "pages")
			}
		}
	}
	// from new
	newEmpty := header.emptyPageIndex + pageNumber
	if newEmpty > header.maxPageIndex {
		e := OOMError{
			pageNumber: pageNumber,
			details:    m.String(),
		}
		return nullPageHandle, &e
	}
	handler := MakePageHandler(pageNumber, header.emptyPageIndex)
	header.emptyPageIndex = newEmpty
	if utils.Debug {
		fmt.Println("alloc page from empty", handler)
	}
	header.allocatedPageNumber += pageNumber
	return handler, nil
}

func (m Memory) freePage(pageHandler PageHandler) {
	if utils.Asserted {
		if pageHandler.IsNull() {
			panic("free null page")
		}
		LibZero(m.PagePointerOf(pageHandler), pageHandler.Size())
	}
	pageNumber := pageHandler.PageNumber()
	if utils.Asserted {
		if pageNumber == 0 {
			panic("pageNumber == 0")
		}
	}
	header := m.header()

	pagePtr := m.PagePointerOf(pageHandler)
	// return to linklist
	if pageNumber == 1 {
		if header.freedBasePageHeader.IsNull() {
			header.freedBasePageHeader = pageHandler
			PointerAs[linkedFreePageHeader](pagePtr).next = nullPageHandle
		} else {
			PointerAs[linkedFreePageHeader](pagePtr).next = header.freedBasePageHeader
			header.freedBasePageHeader = pageHandler
		}
	} else {
		PointerAs[linkedFreePageHeader](pagePtr).pageNumber = pageNumber
		if header.freedCombinedPageHeader.IsNull() {
			header.freedCombinedPageHeader = pageHandler
			PointerAs[linkedFreePageHeader](pagePtr).next = nullPageHandle
		} else {
			PointerAs[linkedFreePageHeader](pagePtr).next = header.freedCombinedPageHeader
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
		linked := PointerAs[linkedFreePageHeader](m.PagePointerOf(header))
		pageNumber += linked.pageNumber
		number++
		header = linked.next
	}
	return
}

func (m Memory) nextFreedPage(freedPageHandler PageHandler) PageHandler {
	if utils.Asserted {
		if freedPageHandler.IsNull() {
			panic("nextFreedPage of null")
		}
	}
	return PointerAs[linkedFreePageHeader](m.PagePointerOf(freedPageHandler)).next
}

type linkedFreePageHeader struct {
	pageNumber SizeType    // 空页大小
	next       PageHandler // 链表
}

// PagePointerOf thread-safe
func (m Memory) PagePointerOf(pageHandler PageHandler) Pointer {
	if utils.Asserted {
		if pageHandler.IsNull() {
			panic("PagePointerOf of null")
		}
	}
	offset := Pointer(pageHandler.PageIndex() << BasePageSizeShiftNumber)
	return m.header().pageBasePointer + offset
}

func (m Memory) PointerToPageIndex(ptr Pointer) SizeType {
	if utils.Asserted {
		if ptr.IsNull() {
			panic("call PointerToPageIndex by null pointer")
		}
	}
	offset := ptr - m.header().pageBasePointer
	if utils.Asserted {
		if SizeType(offset)%BasePageSize != 0 {
			panic(fmt.Sprintf("%s is not a point to a page", ptr.String()))
		}
	}
	return SizeType(offset >> BasePageSizeShiftNumber)
}

func (m Memory) pageZero(pageHandler PageHandler) {
	ptr := m.PagePointerOf(pageHandler)
	size := pageHandler.PageNumber() << BasePageSizeShiftNumber
	LibZero(ptr, size)
}

func (m Memory) pageAsBytes(pageHandler PageHandler) []byte {
	var bs []byte
	ptr := m.PagePointerOf(pageHandler)
	size := pageHandler.PageNumber() << BasePageSizeShiftNumber
	*((*reflect.SliceHeader)(unsafe.Pointer(&bs))) = reflect.SliceHeader{
		Data: ptr.UIntPtr(),
		Len:  size.Int(),
		Cap:  size.Int(),
	}
	return bs
}

func (m Memory) allocatedMemorySize() SizeType {
	header := m.header()
	return header.allocatedPageNumber << BasePageSizeShiftNumber
}

func (m Memory) Json() map[string]interface{} {
	info := map[string]interface{}{}
	header := m.header()
	info["pointer"] = m.pointer().String()
	info["libPointer"] = header.libPointer.String()
	info["pageBasePointer"] = header.pageBasePointer.String()
	info["maxPageIndex"] = header.maxPageIndex
	info["emptyPageIndex"] = header.emptyPageIndex
	totalPageNumber := header.maxPageIndex
	info["totalPageNumber"] = totalPageNumber
	info["totalMemory"] = totalPageNumber << BasePageSizeShiftNumber
	info["totalMemory_h"] = HumanFriendlyMemorySize(totalPageNumber << BasePageSizeShiftNumber)
	allocatedPageNumber := header.allocatedPageNumber
	info["allocatedPageNumber"] = allocatedPageNumber
	info["allocatedMemory"] = allocatedPageNumber << BasePageSizeShiftNumber
	info["allocatedMemory_h"] = HumanFriendlyMemorySize(allocatedPageNumber << BasePageSizeShiftNumber)
	return info
}

func (m Memory) String() string {
	return utils.Jsonify(m.Json())
}

type OOMError struct {
	pageNumber SizeType
	details    string
}

func (o *OOMError) Error() string {
	return fmt.Sprintf("out of memory when alloc %d pages. The memory details is \n%s", o.pageNumber, o.details)
}
