package direct

import (
	"fmt"
	"github.com/madokast/direct/memory"
	"github.com/madokast/direct/utils"
	"math/rand"
	"runtime"
	d2 "runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func Test_sizeof(t *testing.T) {
	t.Log(memory.Sizeof[int8]())
	t.Log(memory.Sizeof[int16]())
	t.Log(memory.Sizeof[int32]())
	t.Log(memory.Sizeof[int64]())
	t.Log(memory.Sizeof[string]())
	t.Log(memory.Sizeof[[]byte]())
}

func Test_sizeof1(t *testing.T) {
	t.Log(memory.Sizeof[Slice[string]]())
	t.Log(memory.Sizeof[Slice[int]]())
}

func Test_newSlice(t *testing.T) {
	Global.Init(4 * 1024)
	defer Global.Free()

	slice, err := MakeSlice[int32](10)
	utils.PanicErr(err)
	defer slice.Free()

	t.Log(memory.Sizeof[sliceHeader]())
	t.Log(slice.header())
}

func TestSlice_Get(t *testing.T) {
	Global.Init(4 * 1024)
	defer Global.Free()

	defer func() {}()

	var s Slice[int32]
	defer func() { s.Free() }()

	utils.PanicErr(s.Append(5))
	utils.PanicErr(s.Append(2))
	utils.PanicErr(s.Append(0))

	utils.Assert(s.Get(0) == 5)
	utils.Assert(s.Get(1) == 2)
	utils.Assert(s.Get(2) == 0)
}

func TestSlice_Iterate(t *testing.T) {
	Global.Init(4 * 1024)
	defer Global.Free()

	var s Slice[int32]
	defer func() {
		s.Free()
	}()

	utils.PanicErr(s.Append(5))
	utils.PanicErr(s.Append(2))
	utils.PanicErr(s.Append(0))

	s.Iterate(func(i int32) {
		t.Log(i)
	})

	t.Log(s.Length())
	utils.Assert(s.Length() == 3)
	t.Log(s.Capacity())
}

func TestSlice_new2(t *testing.T) {
	Global.Init(4 * 1024)
	defer Global.Free()

	var s Slice[int32]
	utils.PanicErr(s.Append(5))
	s.Free()

	s = 0
	utils.PanicErr(s.Append(5))
	s.Free()
}

func Test_BenchmarkMySlice_Append(t *testing.T) {
	Global.Init(32 * 1024 * 1024)

	for i := 0; i < 10000; i++ {
		var s Slice[int]
		for j := 0; j < 100; j++ {
			utils.PanicErr(s.Append(j))
		}
		s.Free()
	}

	Global.Free()
}

func Test_BenchmarkMySlice_make(t *testing.T) {
	Global.Init(32 * 1024 * 1024)

	for i := 0; i < 10000; i++ {
		ss, _ := MakeSlice[string](1024)
		_ = ss.Append("")
		ss.Free()
	}

	Global.Free()
}

func BenchmarkSlice_Append(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var s []int
		for j := 0; j < 100; j++ {
			s = append(s, j)
		}
	}
}

func BenchmarkMySlice_Append(b *testing.B) {
	Global.Init(32 * 1024 * 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s Slice[int]
		for j := 0; j < 100; j++ {
			_ = s.Append(j)
		}
		s.Free()
	}
	b.StopTimer()

	Global.Free()
}

func BenchmarkSlice_make(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ss := make([]string, 0, 1024)
		ss = append(ss, "")
	}
}

func BenchmarkMySlice_make(b *testing.B) {
	Global.Init(32 * 1024 * 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ss, _ := MakeSlice[string](1024)
		_ = ss.Append("")
		ss.Free()
	}
	b.StopTimer()

	Global.Free()
}

func BenchmarkMySlice_make2(b *testing.B) {
	Global.Init(32 * 1024 * 1024)

	b.ResetTimer()
	ss, _ := MakeSlice[string](1024)
	_ = ss.Append("")
	for i := 0; i < b.N; i++ {
		ss.Set(0, "")
	}
	ss.Free()
	b.StopTimer()

	Global.Free()
}

func BenchmarkSlice_make_parallel(b *testing.B) {
	wg := sync.WaitGroup{}
	for k := 0; k < 32; k++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < b.N; i++ {
				ss := make([]string, 0, 1024)
				ss = append(ss, "")
			}
		}()
	}
	wg.Wait()
}

func BenchmarkMySlice_make_parallel(b *testing.B) {
	Global.Init(256 * 1024 * 1024)
	b.ResetTimer()
	wg := sync.WaitGroup{}
	for k := 0; k < 32; k++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for i := 0; i < b.N; i++ {
				ss, _ := MakeSlice[string](1024)
				_ = ss.Append("")
				ss.Free()
			}

		}()
	}
	wg.Wait()
	b.StopTimer()
	Global.Free()
}

func Test_BenchmarkMySlice_make_parallel(t *testing.T) {
	Global.Init(256 * 1024 * 1024)
	memories := locals
	wg := sync.WaitGroup{}
	for k := 0; k < 32; k++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				ss, _ := MakeSlice[string](1024)
				_ = ss.Append("")
				ss.Free()
			}
		}()
	}
	t.Log(len(memories))
	wg.Wait()
	Global.Free()
}

func TestMakeSliceWithLength(t *testing.T) {
	Global.Init(4 * 1024)
	defer Global.Free()

	s, err := MakeSliceWithLength[int32](10)
	utils.PanicErr(err)
	defer func() { s.Free() }()

	utils.Assert(s.Length() == 10, s.Length())
	utils.Assert(s.Capacity() >= 10, s.Capacity())

	for i := SizeType(0); i < s.Length(); i++ {
		utils.Assert(s.Get(i) == 0, s.Get(i))
	}
}

func TestSlice_String(t *testing.T) {
	Global.Init(4 * 1024)
	defer Global.Free()

	s, err := MakeSliceWithLength[int32](10)
	utils.PanicErr(err)
	defer func() { s.Free() }()

	s.Set(0, 10)
	s.Set(5, 520)
	s.Set(9, 111)

	t.Log(s)
}

func Test2DSlice_String(t *testing.T) {
	Global.Init(4 * 1024)
	defer Global.Free()

	var ss Slice[Slice[int32]]
	defer func() { ss.Free() }()
	{
		s, err := MakeSliceWithLength[int32](10)
		utils.PanicErr(err)
		defer func() { s.Free() }()

		s.Set(0, 10)
		s.Set(5, 520)
		s.Set(9, 111)

		t.Log(s)

		utils.PanicErr(ss.Append(s))
	}

	{
		s, err := MakeSliceWithLength[int32](5)
		utils.PanicErr(err)
		defer func() { s.Free() }()

		s.Set(1, 4)
		s.Set(3, 8)

		t.Log(s)

		utils.PanicErr(ss.Append(s))
	}

	t.Log(ss)
}

func TestIntersection(t *testing.T) {
	Global.Init(1 * memory.MB)

	var intersection Slice[Slice[Slice[int32]]]

	var err error
	for i := 0; i < 3; i++ {
		t.Log("before", intersection)
		//intersection, err = calIntersection(intersection)
		intersection, err = calIntersectionParallel(intersection)
		utils.PanicErr(err)
		t.Log("after", intersection)
	}

	freeIntersection(intersection)

	Global.Free()
}

func freeIntersection(intersection Slice[Slice[Slice[int32]]]) {
	intersection.Iterate(func(idPairs Slice[Slice[int32]]) {
		idPairs.Iterate(func(ids Slice[int32]) {
			ids.Free()
		})
		idPairs.Free()
	})
	intersection.Free()
}

// calIntersection should free intersection
func calIntersection(intersection Slice[Slice[Slice[int32]]]) (Slice[Slice[Slice[int32]]], error) {
	if intersection.Length() < 1 {
		return generatePredicateIntersection()
	}
	var resultIntersection Slice[Slice[Slice[int32]]]
	var err error
	intersection.Iterate(func(idPairs Slice[Slice[int32]]) {
		var tempIdPairs Slice[Slice[Slice[int32]]]
		value2Index := make(map[int32]SizeType)
		idPairs.IterateIndex(func(tableIndex SizeType, ids Slice[int32]) {
			ids.Iterate(func(rowId int32) {
				value := rowId % 4
				index, ok := value2Index[value]
				if !ok {
					value2Index[value] = tempIdPairs.Length()
					index = tempIdPairs.Length()
					var temp Slice[Slice[int32]]
					temp, err = MakeSliceWithLength[Slice[int32]](2)
					if err != nil {
						return
					}
					err = tempIdPairs.Append(temp)
					if err != nil {
						return
					}
				}
				temp := tempIdPairs.Get(index).Get(tableIndex)
				err = temp.Append(rowId)
				if err != nil {
					return
				}
				tempIdPairs.Get(index).Set(tableIndex, temp)
			})
		})
		tempIdPairs.Iterate(func(temp Slice[Slice[int32]]) {
			err = resultIntersection.Append(temp)
			if err != nil {
				return
			}
		})
		tempIdPairs.Free()
	})
	freeIntersection(intersection)
	return resultIntersection, err
}

func calIntersectionParallel(intersection Slice[Slice[Slice[int32]]]) (Slice[Slice[Slice[int32]]], error) {
	if intersection.Length() < 1 {
		return generatePredicateIntersection()
	}
	var resultIntersection Slice[Slice[Slice[int32]]]
	var theErr atomic.Value
	var wg sync.WaitGroup
	var lock sync.Mutex
	intersection.Iterate(func(idPairs Slice[Slice[int32]]) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var tempIdPairs Slice[Slice[Slice[int32]]]
			value2Index := make(map[int32]SizeType)
			idPairs.IterateIndexBreakable(func(tableIndex SizeType, ids Slice[int32]) bool {
				if theErr.Load() != nil {
					return false
				}
				ids.IterateBreakable(func(rowId int32) bool {
					value := rowId % 4
					index, ok := value2Index[value]
					if !ok {
						value2Index[value] = tempIdPairs.Length()
						index = tempIdPairs.Length()
						temp, err := MakeSliceWithLength[Slice[int32]](2)
						if err != nil {
							theErr.Store(err)
							return false
						}
						err = tempIdPairs.Append(temp)
						if err != nil {
							theErr.Store(err)
							return false
						}
					}
					temp := tempIdPairs.Get(index).Get(tableIndex)
					err := temp.Append(rowId)
					if err != nil {
						theErr.Store(err)
						return false
					}
					tempIdPairs.Get(index).Set(tableIndex, temp)

					time.Sleep(10 * time.Millisecond)
					return true
				})
				return true
			})
			if theErr.Load() != nil {
				return
			}
			tempIdPairs.IterateBreakable(func(temp Slice[Slice[int32]]) bool {
				lock.Lock()
				err := resultIntersection.Append(temp)
				lock.Unlock()
				if err != nil {
					theErr.Store(err)
					return false
				}
				return true
			})
			if theErr.Load() != nil {
				return
			}
			tempIdPairs.Free()
		}()
	})
	wg.Wait()
	err := theErr.Load()
	if err != nil {
		return 0, err.(error)
	}
	freeIntersection(intersection)
	return resultIntersection, nil
}

func generatePredicateIntersection() (Slice[Slice[Slice[int32]]], error) {
	var intersection Slice[Slice[Slice[int32]]]
	var leftPli = map[int32][]int32{
		0: {0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		1: {11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
		3: {100, 200},
		4: {55, 66, 77, 88, 99},
	}
	for _, int32s := range leftPli {
		idPairs, err := MakeSliceWithLength[Slice[int32]](2)
		if err != nil {
			return 0, err
		}
		s1, err := MakeSliceFromGoSlice(int32s)
		if err != nil {
			return 0, err
		}
		s2, err := MakeSliceFromGoSlice(int32s)
		if err != nil {
			return 0, err
		}
		idPairs.Set(0, s1)
		idPairs.Set(1, s2)

		err = intersection.Append(idPairs)
		if err != nil {
			return 0, err
		}
	}

	return intersection, nil
}

func TestMakeSliceFromGoSlice(t *testing.T) {
	Global.Init(4 * 1024)
	defer Global.Free()

	var ss = utils.PanicErr1(MakeSliceFromGoSlice([]string{"hello", ", ", "world", "!"}))
	defer func() { ss.Free() }()

	t.Log(ss)
}

func TestMakeSliceFromGoSlice1000(t *testing.T) {
	Global.Init(1024 * 1024)
	defer Global.Free()

	var temp []int
	for i := 0; i < 1000; i++ {
		temp = append(temp, rand.Int())
	}

	var ss = utils.PanicErr1(MakeSliceFromGoSlice(temp))
	defer func() { ss.Free() }()

	if len(temp) != int(ss.Length()) {
		panic(fmt.Sprintf("%d != %d", len(temp), int(ss.Length())))
	}

	for i, e := range temp {
		if e != ss.Get(SizeType(i)) {
			panic(fmt.Sprintf("%d != %d", e, ss.Get(SizeType(i))))
		}
	}
}

func TestMakeSliceFromGoSliceEmpty(t *testing.T) {
	Global.Init(4 * 1024)
	defer Global.Free()

	var temp []string
	var ss = utils.PanicErr1(MakeSliceFromGoSlice(temp))
	defer func() { ss.Free() }()

	t.Log(ss)
}

func TestSlice_Copy(t *testing.T) {
	Global.Init(4 * 1024)
	defer Global.Free()

	var ss = utils.PanicErr1(MakeSliceFromGoSlice([]string{"hello", ", ", "world", "!"}))
	defer func() { ss.Free() }()

	t.Log(ss)

	ss2, err := ss.Copy()
	utils.PanicErr(err)
	defer func() { ss2.Free() }()
	t.Log(ss2)
}

func TestMakeSliceCopy1000(t *testing.T) {
	Global.Init(1024 * 1024)
	defer Global.Free()

	var temp []int
	for i := 0; i < 1000; i++ {
		temp = append(temp, rand.Int())
	}

	var ss = utils.PanicErr1(MakeSliceFromGoSlice(temp))
	defer func() { ss.Free() }()

	if len(temp) != int(ss.Length()) {
		panic(fmt.Sprintf("%d != %d", len(temp), int(ss.Length())))
	}

	ss2, err := ss.Copy()
	utils.PanicErr(err)
	defer func() { ss2.Free() }()

	if len(temp) != int(ss2.Length()) {
		panic(fmt.Sprintf("%d != %d", len(temp), int(ss2.Length())))
	}

	t.Log(len(temp), ss.Length(), ss2.Length())

	for i, e := range temp {
		if e != ss.Get(SizeType(i)) {
			panic(fmt.Sprintf("%d != %d", e, ss.Get(SizeType(i))))
		}
		if e != ss2.Get(SizeType(i)) {
			panic(fmt.Sprintf("%d != %d", e, ss2.Get(SizeType(i))))
		}
	}
}

func TestSlice_AppendBatch(t *testing.T) {
	Global.Init(1024 * 1024)
	defer Global.Free()

	var ss = utils.PanicErr1(MakeSliceFromGoSlice([]int{1, 2, 3}))
	defer func() { ss.Free() }()

	var s2 Slice[int]
	defer func() { s2.Free() }()
	err := s2.AppendBatch(ss)
	utils.PanicErr(err)

	t.Log(ss)
	t.Log(s2)
}

func TestSlice_AppendBatch2(t *testing.T) {
	Global.Init(1024 * 1024)
	defer Global.Free()

	var ss = utils.PanicErr1(MakeSliceFromGoSlice([]int{1, 2, 3}))
	defer func() { ss.Free() }()

	var s2 Slice[int]
	defer func() { s2.Free() }()
	err := s2.AppendBatch(ss)
	utils.PanicErr(err)
	err = s2.AppendBatch(ss)
	utils.PanicErr(err)

	t.Log(ss)
	t.Log(s2)
}

func TestSlice_AppendBatchSelf(t *testing.T) {
	Global.Init(1024 * 1024)
	defer Global.Free()

	var ss = utils.PanicErr1(MakeSliceFromGoSlice([]int{1, 2, 3}))
	defer func() { ss.Free() }()

	err := ss.AppendBatch(ss)
	utils.PanicErr(err)

	t.Log(ss)
}

func TestSlice_AppendBatch1000(t *testing.T) {
	Global.Init(1024 * 1024)
	defer Global.Free()

	gs := make([]int, 500)
	for i := range gs {
		gs[i] = 22
	}

	var ss = utils.PanicErr1(MakeSliceFromGoSlice(gs))
	defer func() { ss.Free() }()

	err := ss.AppendBatch(ss)
	utils.PanicErr(err)
	t.Log(ss.Length())
	utils.Assert(ss.Length() == 1000)
	ss.Iterate(func(i int) {
		if i != 22 {
			panic(i)
		}
	})
}

func TestSlice_Iterator(t *testing.T) {
	Global.Init(1 * memory.KB)
	defer Global.Free()
	slice := utils.PanicErr1(MakeSliceFromGoSlice([]int{1, 2, 5, 10}))
	iter := slice.Iterator()
	for iter.Next() {
		t.Log(iter.Index(), iter.Value())
	}
	slice.Free()
}

func TestSlice_Move(t *testing.T) {
	Global.Init(1 * memory.MB)
	defer Global.Free()

	var ss Slice[Slice[int]]
	defer func() {
		ss.Iterate(func(s Slice[int]) {
			s.Free()
		})
		ss.Free()
	}()
	for i := 0; i < 10; i++ {
		var s Slice[int]
		defer func() { s.Free() }()
		for j := 0; j < i; j++ {
			_ = s.Append(j)
		}
		if i%2 == 0 {
			_ = ss.Append(s.Move())
		}
	}

	// [[0 1 2 3 4 5 6 7 8 9] [0 1 2 3 4 5 6 7 8 9] [0 1 2 3 4 5 6 7 8 9] [0 1 2 3 4 5 6 7 8 9] [0 1 2 3 4 5 6 7 8 9]]
	t.Log(ss)
}

func TestSlice_DoubleFreeCheck(t *testing.T) {
	Global.Init(1 * memory.KB)
	defer Global.Free()

	slice := utils.PanicErr1(MakeSliceFromGoSlice([]int{1, 2, 5, 10}))
	t.Log(slice)
	slice.Free()

	if utils.Asserted {
		t.Log("========== THE FOLLOWING ERROR IS OK ============")
		var err any
		func() {
			defer func() {
				err = recover()
			}()
			slice.Free()
		}()

		utils.Assert(err != nil, err)
	}
}

func Benchmark_MultiSliceAppendGo(b *testing.B) {
	rand.Seed(1)
	size := 100 * 100 * 100
	var values = make([]int, size)
	for i := 0; i < size; i++ {
		values[i] = rand.Int()
	}

	for i := 0; i < b.N; i++ {
		m := map[int]int{}
		var ss [][]int
		for _, value := range values {
			key := value % 128
			index, ok := m[key]
			if !ok {
				index = len(ss)
				ss = append(ss, nil)
				m[key] = index
			}
			ss[index] = append(ss[index], value)
		}
		m = map[int]int{}
		ss = nil
	}
}

func Benchmark_MultiSliceAppendRock(b *testing.B) {
	rand.Seed(1)
	size := 100 * 100 * 100
	var values = make([]int, size)
	for i := 0; i < size; i++ {
		values[i] = rand.Int()
	}
	Global.Init(1 * memory.GB)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := utils.PanicErr1(MakeMap[int, SizeType](0))
		var ss Slice[Slice[int]] = nullSlice

		for _, value := range values {
			key := value % 128
			index, ok := m.Get2(key)
			if !ok {
				index = ss.Length()
				utils.PanicErr(ss.Append(nullSlice))
				utils.PanicErr(m.DirectPut(key, index))
			}
			utils.PanicErr(ss.RefAt(index).Append(value))
		}
		ssIter := ss.Iterator()
		for ssIter.Next() {
			ssIter.Ref().Free()
		}
		ss.Free()
		m.Free()
	}

	b.StopTimer()
	Global.Free()
}

func TestGcFail(t *testing.T) {
	Global.Init(1024)

	var s Slice[string]
	sb := strings.Builder{}
	sb.WriteString("xxxxxxxxxxxxxxx" + string(make([]byte, 1024*1024*1024)))
	s.Append(sb.String())
	sb.Reset()

	runtime.GC()
	d2.FreeOSMemory()
	//fmt.Println(s.Get(0)[:20])

	s.Free()
	Global.Free()
}
