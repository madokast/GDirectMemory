package direct

import (
	"github.com/madokast/direct/memory"
	"github.com/madokast/direct/utils"
	"testing"
)

func Test_globalAllocPage(t *testing.T) {
	Global.Init(1024)
	defer Global.Free()

	_, err := Global.allocPage(1, "test", 1)
	utils.PanicErr(err)

	t.Log(Global.MemoryLeakInfo())
	if memory.Trace {
		utils.Assert(Global.IsMemoryLeak())
	}
}

func Test_globalAllocPage2(t *testing.T) {
	Global.Init(1024)
	defer Global.Free()

	page, err := Global.allocPage(1, "test", 1)
	utils.PanicErr(err)
	Global.freePage(page)

	t.Log(Global.MemoryLeakInfo())
	utils.Assert(!Global.IsMemoryLeak())
}

var bs []byte
var page memory.PageHandler

func BenchmarkGoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bs = make([]byte, memory.BasePageSize) // 40.16 ns/op
		if bs == nil {
			panic("null")
		}
	}
}

func BenchmarkLocalMemory(b *testing.B) {
	global := memory.New(1024 * 1024)
	local := global.NewLocalMemory()
	var err error
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		page, err = local.AllocPage(1) // 4.178 ns/op trace off / assert off
		utils.PanicErr(err)
		local.FreePage(page)
	}
	b.StopTimer()
	local.Destroy()
	global.Free()
}

func BenchmarkGlobalMemory(b *testing.B) {
	Global.Init(1024 * 1024)
	var err error
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		page, err = Global.allocPage(1, "test", 1) // 12.52 ns/op trace off / assert off
		utils.PanicErr(err)
		Global.freePage(page)
	}
	b.StopTimer()
	Global.Free()
}
