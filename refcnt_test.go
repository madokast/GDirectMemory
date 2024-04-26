package direct

import (
	memory2 "github.com/madokast/direct/memory"
	"github.com/madokast/direct/utils"
	"github.com/madokast/direct/utils/gpm"
	"sync"
	"testing"
)

func TestSharedFactory_MakeShared(t *testing.T) {
	Global.Init(1 * memory2.MB)
	defer Global.Free()

	m, err := MakeMapFromGoMap(map[int]int{123: 546})
	utils.PanicErr(err)
	defer func() { m.Free() }()
	t.Log(m)

	factory := CreateSharedFactory[Map[int, int]]()
	defer factory.Destroy()

	ms, err := factory.MakeShared(m.Move())
	utils.PanicErr(err)
	defer ms.Free()
	t.Log(ms.String())

	ms2 := ms.Share()
	defer ms2.Free()
	t.Log(ms2.String())
}

func TestSharedFactory_MakeShared2(t *testing.T) {
	Global.Init(1 * memory2.MB)
	defer Global.Free()

	m, err := MakeMapFromGoMap(map[int]int{123: 546})
	utils.PanicErr(err)
	defer func() { m.Free() }()
	t.Log(m)

	factory := CreateSharedFactory[Map[int, int]]()
	defer factory.Destroy()

	ms, err := factory.MakeShared(m.Move())
	utils.PanicErr(err)
	defer ms.Free()
	t.Log(ms.String())

	ms2 := ms.Share()
	defer ms2.Free()
	t.Log(ms2.String())
}

func BenchmarkAdd(b *testing.B) {
	var s int
	for i := 0; i < b.N; i++ {
		s += i // 0.2205
	}
}

func BenchmarkAddLock(b *testing.B) {
	var s int
	var mu sync.Mutex
	for i := 0; i < b.N; i++ {
		mu.Lock()
		s += i // 9.438 ns/op
		mu.Unlock()
	}
}

func BenchmarkAddPreempt(b *testing.B) {
	var s int
	for i := 0; i < b.N; i++ {
		mp := gpm.DisablePreempt()
		s += i // 3.701 ns/op
		gpm.EnablePreempt(mp)
	}
}
