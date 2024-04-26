package direct

import (
	"fmt"
	"github.com/madokast/direct/memory"
	"github.com/madokast/direct/utils"
	"runtime"
	"strings"
	"testing"
)

func Test_Trace_MakeSlice(t *testing.T) {
	Global.Init(4 * memory.KB)

	_, file, line, _ := runtime.Caller(0)
	_, err := MakeSlice[int](10)
	utils.PanicErr(err)

	if memory.Trace {
		utils.Assert(Global.IsMemoryLeak())
		info := Global.MemoryLeakInfo()
		t.Log(info)
		t.Log("should be", file, line+1)
		utils.Assert(strings.Contains(info, fmt.Sprintf("allocated at %s:%d", file, line+1)))
	}
	Global.Free()
}

func Test_Trace_MakeSliceWithLength(t *testing.T) {
	Global.Init(4 * memory.KB)
	_, file, line, _ := runtime.Caller(0)
	_, err := MakeSliceWithLength[int](10)
	utils.PanicErr(err)

	if memory.Trace {
		utils.Assert(Global.IsMemoryLeak())
		info := Global.MemoryLeakInfo()
		t.Log(info)
		t.Log("should be", file, line+1)
		utils.Assert(strings.Contains(info, fmt.Sprintf("allocated at %s:%d", file, line+1)))
	}
	Global.Free()
}

func Test_MakeSliceFromGoSlice(t *testing.T) {
	Global.Init(4 * memory.KB)

	_, file, line, _ := runtime.Caller(0)
	_, err := MakeSliceFromGoSlice([]int{3, 2, 1})
	utils.PanicErr(err)

	if memory.Trace {
		utils.Assert(Global.IsMemoryLeak())
		info := Global.MemoryLeakInfo()
		t.Log(info)
		t.Log("should be", file, line+1)
		utils.Assert(strings.Contains(info, fmt.Sprintf("allocated at %s:%d", file, line+1)))
	}
	Global.Free()
}

func Test_Trace_Append(t *testing.T) {
	Global.Init(4 * memory.KB)

	var s Slice[int]
	_, file, line, _ := runtime.Caller(0)
	err := s.Append(1)
	utils.PanicErr(err)

	if memory.Trace {
		utils.Assert(Global.IsMemoryLeak())
		info := Global.MemoryLeakInfo()
		t.Log(info)
		t.Log("should be", file, line+1)
		utils.Assert(strings.Contains(info, fmt.Sprintf("allocated at %s:%d", file, line+1)))
	}
	Global.Free()
}

func Test_Trace_Append2(t *testing.T) {
	Global.Init(4 * memory.KB)

	var s Slice[int]
	_, file, line, _ := runtime.Caller(0)
	err := s.Append(1)
	utils.PanicErr(err)
	err = s.AppendGoSlice([]int{2})
	utils.PanicErr(err)

	if memory.Trace {
		utils.Assert(Global.IsMemoryLeak())
		info := Global.MemoryLeakInfo()
		t.Log(info)
		t.Log("should be", file, line+1)
		utils.Assert(strings.Contains(info, fmt.Sprintf("allocated at %s:%d", file, line+1)))
	}
	Global.Free()
}

func Test_Trace_Append3(t *testing.T) {
	Global.Init(4 * memory.KB)

	var s Slice[int]
	err := s.Append(1)
	utils.PanicErr(err)
	_, file, line, _ := runtime.Caller(0)
	err = s.AppendGoSlice(make([]int, 100))
	utils.PanicErr(err)

	if memory.Trace {
		utils.Assert(Global.IsMemoryLeak())
		info := Global.MemoryLeakInfo()
		t.Log(info)
		t.Log("should be", file, line+1)
		utils.Assert(strings.Contains(info, fmt.Sprintf("allocated at %s:%d", file, line+1)))
	}
	Global.Free()
}

func Test_Trace_Copy(t *testing.T) {
	Global.Init(4 * memory.KB)

	_, file, line, _ := runtime.Caller(0)
	s, err := MakeSliceWithLength[int](10)
	utils.PanicErr(err)

	_, file2, line2, _ := runtime.Caller(0)
	_, err = s.Copy()
	utils.PanicErr(err)

	if memory.Trace {
		utils.Assert(Global.IsMemoryLeak())
		info := Global.MemoryLeakInfo()
		t.Log(info)
		t.Log("should be", file, line+1)
		utils.Assert(strings.Contains(info, fmt.Sprintf("allocated at %s:%d", file, line+1)))
		t.Log("should be", file2, line2+1)
		utils.Assert(strings.Contains(info, fmt.Sprintf("allocated at %s:%d", file2, line2+1)))
	}
	Global.Free()
}

func Test_CreateFromGoString(t *testing.T) {
	Global.Init(4 * memory.KB)

	factory := NewStringFactory()
	_, file, line, _ := runtime.Caller(0)
	_, err := factory.CreateFromGoString("hello")
	utils.PanicErr(err)
	_, err = factory.CreateFromGoString("world")
	utils.PanicErr(err)

	if memory.Trace {
		utils.Assert(Global.IsMemoryLeak())
		info := Global.MemoryLeakInfo()
		t.Log(info)
		t.Log("should be", file, line+1)
		utils.Assert(strings.Contains(info, fmt.Sprintf("allocated at %s:%d", file, line+1)))
	}
	Global.Free()
}

func Test_MakeCustomMap(t *testing.T) {
	Global.Init(4 * memory.KB)

	_, file, line, _ := runtime.Caller(0)
	_, err := MakeCustomMap[int, int](0, func(key int) SizeType {
		return SizeType(key)
	}, func(key int, key2 int) bool {
		return key == key2
	})
	utils.PanicErr(err)

	if memory.Trace {
		utils.Assert(Global.IsMemoryLeak())
		info := Global.MemoryLeakInfo()
		t.Log(info)
		t.Log("should be", file, line+1)
		utils.Assert(strings.Contains(info, fmt.Sprintf("allocated at %s:%d", file, line+1)))
	}
	Global.Free()
}

func Test_MakeMap(t *testing.T) {
	Global.Init(4 * memory.KB)

	_, file, line, _ := runtime.Caller(0)
	_, err := MakeMap[int, int](0)
	utils.PanicErr(err)

	if memory.Trace {
		utils.Assert(Global.IsMemoryLeak())
		info := Global.MemoryLeakInfo()
		t.Log(info)
		t.Log("should be", file, line+1)
		utils.Assert(strings.Contains(info, fmt.Sprintf("allocated at %s:%d", file, line+1)))
	}
	t.Log(Global.MemoryLeakInfo())
	Global.Free()
}

func Test_MakeMapFromGoMap(t *testing.T) {
	Global.Init(4 * memory.KB)

	_, file, line, _ := runtime.Caller(0)
	_, err := MakeMapFromGoMap(map[int]int{1: 1})
	utils.PanicErr(err)

	if memory.Trace {
		utils.Assert(Global.IsMemoryLeak())
		info := Global.MemoryLeakInfo()
		t.Log(info)
		t.Log("should be", file, line+1)
		utils.Assert(strings.Contains(info, fmt.Sprintf("allocated at %s:%d", file, line+1)))
	}
	Global.Free()
}

func Test_MakeMapFromGoMap2(t *testing.T) {
	Global.Init(4 * memory.KB)

	m, err := MakeMapFromGoMap(map[int]int{1: 1})
	utils.PanicErr(err)
	m.Free()

	if memory.Trace {
		utils.Assert(!Global.IsMemoryLeak())
	}
	t.Log(Global.MemoryLeakInfo())
	Global.Free()
}

func Test_MapAppend(t *testing.T) {
	Global.Init(4 * memory.MB)

	_, file2, line2, _ := runtime.Caller(0)
	m, err := MakeMap[int, int](0)
	utils.PanicErr(err)

	var file string
	var line int
	for i := 0; i < 100; i++ {
		_, file, line, _ = runtime.Caller(0)
		err = m.Put(i, i)
		utils.PanicErr(err)
	}

	if memory.Trace {
		utils.Assert(Global.IsMemoryLeak())
		info := Global.MemoryLeakInfo()
		t.Log(info)
		t.Log("should be", file, line+1)
		utils.Assert(strings.Contains(info, fmt.Sprintf("allocated at %s:%d", file, line+1)))
		t.Log("should be", file2, line2+1)
		utils.Assert(strings.Contains(info, fmt.Sprintf("allocated at %s:%d", file2, line2+1)))
	}
	Global.Free()
}

func Test_Stack(t *testing.T) {
	Global.Init(4 * memory.KB)

	var s Stack[int]
	_, file, line, _ := runtime.Caller(0)
	err := s.Push(1)
	utils.PanicErr(err)

	if memory.Trace {
		utils.Assert(Global.IsMemoryLeak())
		info := Global.MemoryLeakInfo()
		t.Log(info)
		t.Log("should be", file, line+1)
		utils.Assert(strings.Contains(info, fmt.Sprintf("allocated at %s:%d", file, line+1)))
	}
	Global.Free()
}

func Test_StackPushMore(t *testing.T) {
	Global.Init(4 * memory.KB)

	var s Stack[int]
	_, file, line, _ := runtime.Caller(0)
	err := s.Push(1)
	utils.PanicErr(err)

	var file2 string
	var line2 int
	for i := 0; i < 100; i++ {
		_, file2, line2, _ = runtime.Caller(0)
		err = s.Push(1)
		utils.PanicErr(err)
	}

	if memory.Trace {
		utils.Assert(Global.IsMemoryLeak())
		info := Global.MemoryLeakInfo()
		t.Log(info)
		t.Log("should be", file, line+1)
		utils.Assert(strings.Contains(info, fmt.Sprintf("allocated at %s:%d", file, line+1)))
		t.Log("should be", file2, line2+1)
		utils.Assert(strings.Contains(info, fmt.Sprintf("allocated at %s:%d", file2, line2+1)))
	}
	Global.Free()
}
