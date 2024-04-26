package example

import (
	"fmt"
	"github.com/madokast/direct"
	"testing"
)

func TestSlice(t *testing.T) {
	direct.Global.Init(10 * direct.MB)
	defer direct.Global.Free()

	var s direct.Slice[int]
	_ = s.Append(1)                  // [1]
	_ = s.AppendGoSlice([]int{2, 3}) // [1, 2, 3]
	_ = s.AppendBatch(s)             // [1, 2, 3, 1, 2, 3]

	fmt.Println(s)

	s.Set(1, 100) // [1, 100, 3, 1, 2, 3]

	iter := s.Iterator()
	for iter.Next() {
		fmt.Print(iter.Index(), ":", iter.Value(), ", ")
	}
	fmt.Println()

	s.Free()
}

func TestSliceLeak(t *testing.T) {
	direct.Global.Init(10 * direct.MB)
	defer direct.Global.Free()

	var s, _ = direct.MakeSliceFromGoSlice([]int{1, 2, 3}) // line:35

	// [1 2 3]
	fmt.Println(s)

	//s.Free()

	// Addr:0x1FCC64A0340 index:4 type:Slice size:256B allocated at D:/learn/repo/go_repo/direct/example/slice_example_test.go:35
	fmt.Println(direct.Global.MemoryLeakInfo())
}

func TestSliceOOM(t *testing.T) {
	direct.Global.Init(10 * direct.KB)
	defer direct.Global.Free()

	var s, err = direct.MakeSlice[int](1024 * 1024)
	defer s.Free()

	if _, ok := err.(*direct.OOMError); ok {
		fmt.Println("OOM", err) // OOM out of memory when alloc 32769 pages. The memory details is ...
	}
}
