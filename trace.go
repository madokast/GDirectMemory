package direct

import (
	"fmt"
	"strings"
	"sync"
)

/**
A tracer records every alloc/free point.
Used only for leak detection.
Bad performance.
*/

const trace = true

type traceRecord struct {
	file   string
	lineNo int
	size   SizeType
	_type  string
}

type tracer struct {
	traceMu      sync.Mutex
	traceRecords map[pointer]traceRecord
}

func newTrace() *tracer {
	return &tracer{
		traceMu:      sync.Mutex{},
		traceRecords: map[pointer]traceRecord{},
	}
}

func (t *tracer) traceAlloc(ptr pointer, _type string, size SizeType, file string, lineNo int) {
	t.traceMu.Lock()
	if asserted {
		tr, ok := t.traceRecords[ptr]
		if ok {
			panic(fmt.Sprintf("pointer %s has been traced as %s", ptr.String(), tr.String()))
		}
	}
	t.traceRecords[ptr] = traceRecord{
		file:   file,
		lineNo: lineNo,
		size:   size,
		_type:  _type,
	}
	t.traceMu.Unlock()
}

func (t *tracer) deTraceAlloc(ptr pointer) {
	t.traceMu.Lock()
	if asserted {
		_, ok := t.traceRecords[ptr]
		if !ok {
			panic(fmt.Sprintf("remove a un-traced pointer %s", ptr.String()))
		}
	}
	delete(t.traceRecords, ptr)
	t.traceMu.Unlock()
}

func (t *tracer) cleanTrace() {
	t.traceMu.Lock()
	defer t.traceMu.Unlock()
	t.traceRecords = make(map[pointer]traceRecord)
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
	return fmt.Sprintf("type:%s size:%s allocated at %s:%d", tr._type, humanFriendlyMemorySize(tr.size), tr.file, tr.lineNo)
}

/*--------------------- trace user -------------------*/

var tracerMapMu sync.Mutex
var tracerMap = map[Memory]*tracer{}

func (m Memory) tracer() *tracer {
	if !trace {
		panic("call tracer() after testing trace flag")
	}
	tracerMapMu.Lock()
	t := tracerMap[m]
	tracerMapMu.Unlock()
	if asserted {
		if t == nil {
			panic("tracer is not init")
		}
	}
	return t
}

func (m Memory) startTrace() {
	if trace {
		if asserted {
			tracerMapMu.Lock()
			_, ok := tracerMap[m]
			tracerMapMu.Unlock()
			if ok {
				panic("a tracer has been attached to the memory")
			}
		}

		tracerMapMu.Lock()
		tracerMap[m] = newTrace()
		tracerMapMu.Unlock()
	}
}

func (m Memory) deleteTracer() {
	if trace {
		if asserted {
			tracerMapMu.Lock()
			_, ok := tracerMap[m]
			tracerMapMu.Unlock()
			if !ok {
				panic("no tracer attaches to the memory")
			}
		}

		tracerMapMu.Lock()
		delete(tracerMap, m)
		tracerMapMu.Unlock()
	}
}
