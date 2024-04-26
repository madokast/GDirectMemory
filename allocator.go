package direct

import (
	"fmt"
	"github.com/madokast/direct/memory"
	"github.com/madokast/direct/memory/trace_type"
	"github.com/madokast/direct/utils/gpm"
	"github.com/madokast/direct/utils/spin"
	"runtime"
)

type SizeType = memory.SizeType
type OOMError = memory.OOMError

const SizeTypeMax = memory.SizeTypeMax
const KB = memory.KB
const MB = memory.MB
const GB = memory.GB

type globalMemoryNameSpace struct{} // like a namespace

var (
	Global        globalMemoryNameSpace
	global        memory.Memory
	locals        []memory.LocalMemory          // mid -> local
	extraLocals   map[int64]*memory.LocalMemory // mid -> local
	extraLocalsMu spin.Mutex
)

var localsMaxSize = int64(runtime.NumCPU())

func (g globalMemoryNameSpace) allocPage(pageNumber SizeType, _type trace_type.Type, callerSkip int) (page memory.PageHandler, err error) {
	var local *memory.LocalMemory = nil
retry:
	mp := gpm.DisablePreempt()
	mid := mp.MID()
	if mid < localsMaxSize {
		local = &locals[mid]
	} else {
		extraLocalsMu.Lock()
		local = extraLocals[mid]
		if local == nil {
			gpm.EnablePreempt(mp)
			newLocal := global.NewLocalMemory()
			extraLocals[mid] = &newLocal
			extraLocalsMu.Unlock()
			goto retry
		}
		extraLocalsMu.Unlock()
	}

	page, err = local.AllocPage(pageNumber)
	gpm.EnablePreempt(mp)

	if memory.Trace && err == nil {
		_, file, line, _ := runtime.Caller(callerSkip)
		global.Tracer().TraceAlloc(global.PagePointerOf(page), _type, pageNumber*memory.BasePageSize, file, line)
	}
	return page, err
}

func (g globalMemoryNameSpace) freePage(pageHandler memory.PageHandler) {
	if memory.Trace {
		global.Tracer().DeTraceAlloc(global.PagePointerOf(pageHandler))
	}

	var local *memory.LocalMemory
retry:
	mp := gpm.DisablePreempt()
	mid := mp.MID()
	if mid < localsMaxSize {
		local = &locals[mid]
	} else {
		extraLocalsMu.Lock()
		local = extraLocals[mid]
		if local == nil {
			gpm.EnablePreempt(mp)
			newLocal := global.NewLocalMemory()
			extraLocals[mid] = &newLocal
			extraLocalsMu.Unlock()
			goto retry
		}
		extraLocalsMu.Unlock()
	}
	local.FreePage(pageHandler)
	gpm.EnablePreempt(mp)
}

func (g globalMemoryNameSpace) pagePointerOf(pageHandler memory.PageHandler) memory.Pointer {
	return global.PagePointerOf(pageHandler)
}

func (g globalMemoryNameSpace) freePointer(ptr memory.Pointer, pageNumber SizeType) {
	index := global.PointerToPageIndex(ptr)
	pageHandler := memory.MakePageHandler(pageNumber, index)
	g.freePage(pageHandler)
}

func (g globalMemoryNameSpace) Init(totalSize SizeType) {
	if !global.IsNull() {
		panic("global memory has been initialized")
	}
	global = memory.New(totalSize)
	locals = make([]memory.LocalMemory, localsMaxSize)
	extraLocals = map[int64]*memory.LocalMemory{}

	for i := int64(0); i < localsMaxSize; i++ {
		locals[i] = global.NewLocalMemory()
	}
}

func (g globalMemoryNameSpace) Free() {
	if global.IsNull() {
		panic("free an un-init global memory")
	}
	for i := range locals {
		locals[i].Destroy()
	}
	for mid := range extraLocals {
		extraLocals[mid].Destroy()
	}
	locals = nil
	extraLocals = nil

	allocatedPageNumber := global.AllocatedPageNumber()
	if allocatedPageNumber > 0 {
		fmt.Printf("memory leak %s\n", memory.HumanFriendlyMemorySize(allocatedPageNumber<<memory.BasePageSizeShiftNumber))
		if memory.Trace {
			fmt.Println(global.MemoryLeakInfo())
		}
	}
	global.Free()
	global = memory.NullMemory
}

func (g globalMemoryNameSpace) IsMemoryLeak() bool {
	return global.IsMemoryLeak()
}

func (g globalMemoryNameSpace) MemoryLeakInfo() string {
	return global.MemoryLeakInfo()
}
