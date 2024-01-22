package direct

import (
	"testing"
)

func TestSharedFactory_MakeShared(t *testing.T) {
	memory := New(1 * MB)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	m, err := MakeMapFromGoMap(map[int]int{123: 546}, &localMemory)
	PanicErr(err)
	defer func() { m.Free(&localMemory) }()
	t.Log(m)

	factory := CreateSharedFactory[Map[int, int]]()
	defer factory.Destroy(&localMemory)

	ms, err := factory.MakeShared(m.Move(), &localMemory)
	PanicErr(err)
	defer ms.Free(&localMemory)
	t.Log(ms.String())

	ms2 := ms.Share()
	defer ms2.Free(&localMemory)
	t.Log(ms2.String())
}

func TestSharedFactory_MakeShared2(t *testing.T) {
	memory := New(1 * MB)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	m, err := MakeMapFromGoMap(map[int]int{123: 546}, &localMemory)
	PanicErr(err)
	defer func() { m.Free(&localMemory) }()
	t.Log(m)

	factory := CreateSharedFactory[Map[int, int]]()
	defer factory.Destroy(&localMemory)

	ms, err := factory.MakeShared(m.Move(), &localMemory)
	PanicErr(err)
	defer ms.Free(&localMemory)
	t.Log(ms.String())

	ms2 := ms.Share()
	defer ms2.Free(&localMemory)
	t.Log(ms2.String())
}
