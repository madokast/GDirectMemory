package direct

import "testing"

func Test_Trace_MakeSlice(t *testing.T) {
	memory := New(4 * KB)
	concurrentMemory := memory.NewLocalMemory()

	_, err := MakeSlice[int](&concurrentMemory, 10)
	PanicErr(err)

	concurrentMemory.Destroy()
	Assert(memory.IsMemoryLeak())
	t.Log(memory.MemoryLeakInfo())
	memory.Free()
}

func Test_Trace_MakeSliceWithLength(t *testing.T) {
	memory := New(4 * KB)
	concurrentMemory := memory.NewLocalMemory()

	_, err := MakeSliceWithLength[int](&concurrentMemory, 10)
	PanicErr(err)

	concurrentMemory.Destroy()
	Assert(memory.IsMemoryLeak())
	t.Log(memory.MemoryLeakInfo())
	memory.Free()
}

func Test_Trace_Append(t *testing.T) {
	memory := New(4 * KB)
	concurrentMemory := memory.NewLocalMemory()

	var s Slice[int]
	err := s.Append(1, &concurrentMemory)
	PanicErr(err)

	concurrentMemory.Destroy()
	Assert(memory.IsMemoryLeak())
	t.Log(memory.MemoryLeakInfo())
	memory.Free()
}

func Test_Trace_Append2(t *testing.T) {
	memory := New(4 * KB)
	concurrentMemory := memory.NewLocalMemory()

	var s Slice[int]
	err := s.Append(1, &concurrentMemory)
	PanicErr(err)
	err = s.AppendGoSlice([]int{1}, &concurrentMemory)
	PanicErr(err)

	concurrentMemory.Destroy()
	Assert(memory.IsMemoryLeak())
	t.Log(memory.MemoryLeakInfo())
	memory.Free()
}

func Test_Trace_Append3(t *testing.T) {
	memory := New(4 * KB)
	concurrentMemory := memory.NewLocalMemory()

	var s Slice[int]
	err := s.Append(1, &concurrentMemory)
	PanicErr(err)
	err = s.AppendGoSlice(make([]int, 100), &concurrentMemory)
	PanicErr(err)

	concurrentMemory.Destroy()
	Assert(memory.IsMemoryLeak())
	t.Log(memory.MemoryLeakInfo())
	memory.Free()
}

func Test_Trace_Copy(t *testing.T) {
	memory := New(4 * KB)
	concurrentMemory := memory.NewLocalMemory()

	s, err := MakeSliceWithLength[int](&concurrentMemory, 10)
	PanicErr(err)

	_, err = s.Copy(&concurrentMemory)
	PanicErr(err)

	concurrentMemory.Destroy()
	Assert(memory.IsMemoryLeak())
	t.Log(memory.MemoryLeakInfo())
	memory.Free()
}

func Test_MakeSliceFromGoSlice(t *testing.T) {
	memory := New(4 * KB)
	concurrentMemory := memory.NewLocalMemory()

	_, err := MakeSliceFromGoSlice(&concurrentMemory, []int{3, 2, 1})
	PanicErr(err)

	concurrentMemory.Destroy()
	Assert(memory.IsMemoryLeak())
	t.Log(memory.MemoryLeakInfo())
	memory.Free()
}

func Test_CreateFromGoString(t *testing.T) {
	memory := New(4 * KB)
	concurrentMemory := memory.NewLocalMemory()

	factory := NewStringFactory()
	_, err := factory.CreateFromGoString("hello", &concurrentMemory)
	PanicErr(err)
	_, err = factory.CreateFromGoString("world", &concurrentMemory)
	PanicErr(err)

	concurrentMemory.Destroy()
	Assert(memory.IsMemoryLeak())
	t.Log(memory.MemoryLeakInfo())
	memory.Free()
}

func Test_MakeCustomMap(t *testing.T) {
	memory := New(4 * KB)
	concurrentMemory := memory.NewLocalMemory()

	_, err := MakeCustomMap[int, int](0, func(key int) SizeType {
		return SizeType(key)
	}, func(key int, key2 int) bool {
		return key == key2
	}, &concurrentMemory)
	PanicErr(err)

	concurrentMemory.Destroy()
	Assert(memory.IsMemoryLeak())
	t.Log(memory.MemoryLeakInfo())
	memory.Free()
}

func Test_MakeMap(t *testing.T) {
	memory := New(4 * KB)
	concurrentMemory := memory.NewLocalMemory()

	_, err := MakeMap[int, int](0, &concurrentMemory)
	PanicErr(err)

	concurrentMemory.Destroy()
	Assert(memory.IsMemoryLeak())
	t.Log(memory.MemoryLeakInfo())
	memory.Free()
}

func Test_MakeMapFromGoMap(t *testing.T) {
	memory := New(4 * KB)
	concurrentMemory := memory.NewLocalMemory()

	_, err := MakeMapFromGoMap(map[int]int{1: 1}, &concurrentMemory)
	PanicErr(err)

	concurrentMemory.Destroy()
	Assert(memory.IsMemoryLeak())
	t.Log(memory.MemoryLeakInfo())
	memory.Free()
}

func Test_MakeMapFromGoMap2(t *testing.T) {
	memory := New(4 * KB)
	concurrentMemory := memory.NewLocalMemory()

	m, err := MakeMapFromGoMap(map[int]int{1: 1}, &concurrentMemory)
	PanicErr(err)
	m.Free(&concurrentMemory)

	concurrentMemory.Destroy()
	Assert(!memory.IsMemoryLeak())
	t.Log(memory.MemoryLeakInfo())
	memory.Free()
}
