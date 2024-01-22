package direct

import (
	"testing"
)

func TestSharedFactory_MakeShared(t *testing.T) {
	memory := New(1 * MB)
	defer memory.Free()
	concurrentMemory := memory.NewLocalMemory()
	defer concurrentMemory.Destroy()

	m, err := MakeMapFromGoMap(map[int]int{123: 546}, &concurrentMemory)
	PanicErr(err)
	defer func() { m.Free(&concurrentMemory) }()
	t.Log(m)

	factory := CreateSharedFactory[Map[int, int]]()
	defer factory.Destroy(&concurrentMemory)

	ms, err := factory.MakeShared(m.Move(), &concurrentMemory)
	PanicErr(err)
	defer ms.Free(&concurrentMemory)
	t.Log(ms.String())

	ms2 := ms.Share()
	defer ms2.Free(&concurrentMemory)
	t.Log(ms2.String())
}

func TestSharedFactory_MakeShared2(t *testing.T) {
	memory := New(1 * MB)
	defer memory.Free()
	concurrentMemory := memory.NewLocalMemory()
	defer concurrentMemory.Destroy()

	m, err := MakeMapFromGoMap(map[int]int{123: 546}, &concurrentMemory)
	PanicErr(err)
	defer func() { m.Free(&concurrentMemory) }()
	t.Log(m)

	factory := CreateSharedFactory[Map[int, int]]()
	defer factory.Destroy(&concurrentMemory)

	ms, err := factory.MakeShared(m.Move(), &concurrentMemory)
	PanicErr(err)
	defer ms.Free(&concurrentMemory)
	t.Log(ms.String())

	ms2 := ms.Share()
	defer ms2.Free(&concurrentMemory)
	t.Log(ms2.String())
}
