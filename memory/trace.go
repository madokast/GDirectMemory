package memory

import (
	"fmt"
	"github.com/madokast/direct/memory/trace_type"
	"github.com/madokast/direct/utils"
	"strings"
	"sync"
)

/**
A Tracer records every alloc/free point.
Used only for leak detection.
Bad performance.
*/

const Trace = true

type traceRecord struct {
	pageIndex SizeType
	file      string
	lineNo    int
	size      SizeType
	_type     trace_type.Type
}

type tracer struct {
	traceMu      sync.Mutex
	traceRecords map[Pointer]traceRecord
	memory       Memory
}

func newTrace(m Memory) *tracer {
	return &tracer{
		traceMu:      sync.Mutex{},
		traceRecords: map[Pointer]traceRecord{},
		memory:       m,
	}
}

func (t *tracer) TraceAlloc(ptr Pointer, _type trace_type.Type, size SizeType, file string, lineNo int) {
	if !utils.Asserted {
		if trace_type.SkipTrace(_type) {
			return
		}
	}

	t.traceMu.Lock()
	if utils.Asserted {
		tr, ok := t.traceRecords[ptr]
		if ok {
			panic(fmt.Sprintf("pointer %s has been traced as %s", ptr.String(), tr.String()))
		}
	}

	t.traceRecords[ptr] = traceRecord{
		pageIndex: t.memory.PointerToPageIndex(ptr),
		file:      file,
		lineNo:    lineNo,
		size:      size,
		_type:     _type,
	}
	if utils.Debug {
		record := t.traceRecords[ptr]
		fmt.Println("trace", ptr, record.String())
	}
	t.traceMu.Unlock()
}

func (t *tracer) DeTraceAlloc(ptr Pointer) {
	t.traceMu.Lock()
	if utils.Asserted {
		_, ok := t.traceRecords[ptr]
		if !ok {
			panic(fmt.Sprintf("remove a un-traced pointer %s", ptr.String()))
		}
	}
	if utils.Debug {
		record := t.traceRecords[ptr]
		fmt.Println("de-trace", ptr, record.String())
	}
	delete(t.traceRecords, ptr)
	t.traceMu.Unlock()
}

func (t *tracer) cleanTrace() {
	t.traceMu.Lock()
	defer t.traceMu.Unlock()
	t.traceRecords = make(map[Pointer]traceRecord)
}

func (t *tracer) hasLeak() bool {
	t.traceMu.Lock()
	defer t.traceMu.Unlock()
	return len(t.traceRecords) > 0

}

func (t *tracer) leakReport() string {
	t.traceMu.Lock()
	defer t.traceMu.Unlock()
	var sb strings.Builder
	for ptr, record := range t.traceRecords {
		sb.WriteString(fmt.Sprintf("Addr:%s %s\n", ptr.String(), record.String()))
	}
	return sb.String()
}

func (tr *traceRecord) String() string {
	return fmt.Sprintf("index:%d type:%s size:%s allocated at %s:%d", tr.pageIndex, tr._type, HumanFriendlyMemorySize(tr.size), tr.file, tr.lineNo)
}

/*--------------------- trace user -------------------*/

var tracerMapMu sync.Mutex
var tracerMap = map[Memory]*tracer{}

func (m Memory) Tracer() *tracer {
	if !Trace {
		panic("call Tracer() after testing trace flag")
	}
	tracerMapMu.Lock()
	t := tracerMap[m]
	tracerMapMu.Unlock()
	if utils.Asserted {
		if t == nil {
			panic("Tracer is not init")
		}
	}
	return t
}

func (m Memory) startTrace() {
	if Trace {
		if utils.Asserted {
			tracerMapMu.Lock()
			_, ok := tracerMap[m]
			tracerMapMu.Unlock()
			if ok {
				panic("a Tracer has been attached to the memory")
			}
		}

		tracerMapMu.Lock()
		tracerMap[m] = newTrace(m)
		tracerMapMu.Unlock()
	}
}

func (m Memory) deleteTracer() {
	if Trace {
		if utils.Asserted {
			tracerMapMu.Lock()
			_, ok := tracerMap[m]
			tracerMapMu.Unlock()
			if !ok {
				panic("no Tracer attaches to the memory")
			}
		}

		tracerMapMu.Lock()
		delete(tracerMap, m)
		tracerMapMu.Unlock()
	}
}
