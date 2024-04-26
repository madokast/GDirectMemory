package direct

import (
	"fmt"
	memory2 "github.com/madokast/direct/memory"
	"github.com/madokast/direct/utils"
	"golang.org/x/exp/slices"
	"testing"
)

func Test_nodePageSize(t *testing.T) {
	utils.Assert(stackNodePageSize[int]() == 1, stackNodePageSize[int]())
	utils.Assert(stackNodePageSize[[memory2.BasePageSize]byte]() == 2, stackNodePageSize[[memory2.BasePageSize]byte]())
}

func TestStack_Push(t *testing.T) {
	Global.Init(4096)

	var stack Stack[[8]int]
	for i := 0; i < 10; i++ {
		utils.PanicErr(stack.Push([8]int{i}))
		header := stack.header()
		t.Log(header.length, header.capacity, header.next, header.last, header.nextElementPtr)
	}
	stack.Free()

	utils.Assert(!Global.IsMemoryLeak())
	fmt.Println(Global.MemoryLeakInfo())
	Global.Free()
}

func TestStack_Iter(t *testing.T) {
	Global.Init(4096)

	var stack Stack[int]
	for i := 0; i < 100; i++ {
		utils.PanicErr(stack.Push(i))
	}
	var is []int
	stack.Iterate(func(i int) {
		is = append(is, i)
	})
	utils.Assert(len(is) == 100, len(is))
	for i, e := range is {
		utils.Assert(i == e, i, e)
	}
	stack.Free()

	utils.Assert(!Global.IsMemoryLeak())
	fmt.Println(Global.MemoryLeakInfo())
	Global.Free()
}

func TestStack_Iter2(t *testing.T) {
	Global.Init(4096)

	var stack Stack[int]
	for i := 0; i < 100; i++ {
		utils.PanicErr(stack.Push(i))
		//header := stack.header()
		//t.Log(header.length, header.capacity, header.next, header.last, header.nextElementPtr)
	}
	utils.Assert(stack.Length() == 100, stack.Length())
	iter := stack.Iterator()
	for iter.Next() {
		utils.Assert(iter.Index().Int() == iter.Value(), iter.Index(), iter.Value())
	}
	utils.Assert(!iter.Next())
	utils.Assert(!iter.Next())
	utils.Assert(!iter.Next())
	stack.Free()

	utils.Assert(!Global.IsMemoryLeak())
	fmt.Println(Global.MemoryLeakInfo())
	Global.Free()
}

func TestLarge(t *testing.T) {
	Global.Init(1 * memory2.GB)

	var stack Stack[int]
	for i := 0; i < 100*100*100; i++ {
		utils.PanicErr(stack.Push(i))
	}
	utils.Assert(stack.Length() == 100*100*100, stack.Length())
	iter := stack.Iterator()
	for iter.Next() {
		utils.Assert(iter.Index().Int() == iter.Value(), iter.Index(), iter.Value())
	}
	stack.Free()

	utils.Assert(!Global.IsMemoryLeak())
	fmt.Println(Global.MemoryLeakInfo())
	Global.Free()
}

func TestStack_ToGoSlice(t *testing.T) {
	Global.Init(1 * memory2.MB)

	var stack Stack[int]
	var s []int
	for i := 0; i < 10; i++ {
		s = append(s, i)
		utils.PanicErr(stack.Push(i))
		utils.Assert(slices.Equal(s, stack.ToGoSlice()), s, stack)
		t.Log(s, stack)
	}
	stack.Free()
	utils.Assert(!Global.IsMemoryLeak())
	fmt.Println(Global.MemoryLeakInfo())
	Global.Free()
}

func TestStack_ToGoSlice2(t *testing.T) {
	Global.Init(1 * memory2.MB)

	var stack Stack[int]
	var s []int
	for i := 0; i < 1000; i++ {
		s = append(s, i)
		utils.PanicErr(stack.Push(i))
		utils.Assert(slices.Equal(s, stack.ToGoSlice()), s, stack)
	}
	stack.Free()
	utils.Assert(!Global.IsMemoryLeak())
	fmt.Println(Global.MemoryLeakInfo())
	Global.Free()
}

func TestStack_Top(t *testing.T) {
	Global.Init(1 * memory2.MB)

	var stack Stack[int]
	for i := 0; i < 1000; i++ {
		utils.PanicErr(stack.Push(i))
		utils.Assert(stack.Top() == i, i, stack.Top())
	}
	//stack.Free()
	//utils.Assert(!Global.IsMemoryLeak())
	fmt.Println(Global.MemoryLeakInfo())
	Global.Free()
}

func BenchmarkGoAppend(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var ss []int
		for k := 0; k < 100*100*100; k++ {
			ss = append(ss, k) // 5331791 ns/op
		}
	}
}

func BenchmarkSliceAppend(b *testing.B) {
	Global.Init(1 * memory2.GB)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var ss Slice[int]
		for k := 0; k < 100*100*100; k++ {
			utils.PanicErr(ss.Append(k)) // 4252713 ns/op
		}
		ss.Free()
	}
	b.StopTimer()

	Global.Free()
}

func BenchmarkStackAppend(b *testing.B) {
	Global.Init(1 * memory2.GB)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s Stack[int] = nullStack
		for k := 0; k < 100*100*100; k++ {
			utils.PanicErr(s.Push(k)) // 5489071 ns/op
		}
		s.Free()
	}
	b.StopTimer()

	Global.Free()
}
