package managed_memory

import (
	"fmt"
	"gitlab.grandhoo.com/rock/storage/config"
	"gitlab.grandhoo.com/rock/storage/internal/logger"
	"gitlab.grandhoo.com/rock/storage/storage2/utils/test"
	"math/rand"
	"strconv"
	"testing"
	"unsafe"
)

func TestMakeMap(t *testing.T) {
	memory := New(1024 * 1024)
	defer memory.Free()
	concurrentMemory := memory.NewConcurrentMemory()
	defer concurrentMemory.Destroy()

	m, err := MakeMap[int, string](8, &concurrentMemory)
	test.PanicErr(err)
	defer func() {
		m.Free()
	}()

	test.PanicErr(m.DirectPut(0, "a"))
	test.PanicErr(m.DirectPut(1, "a"))
	test.PanicErr(m.DirectPut(2, "a"))
	test.PanicErr(m.DirectPut(3, "a"))
	test.PanicErr(m.DirectPut(4, "a"))
	test.PanicErr(m.DirectPut(5, "a"))
	test.PanicErr(m.DirectPut(8, "a"))

	test.PanicErr(m.Put(3, "q"))
	test.PanicErr(m.Put(5, "b"))
	test.PanicErr(m.Put(4, "c"))
	test.PanicErr(m.Put(8, "d"))
	test.PanicErr(m.Put(12, "k"))

	for i, e := range m.table.GoSlice() {
		logger.Info(i, e)
	}

	logger.Info(m.GoMap())
	logger.Info(m.String())
}

func TestMapCorrect(t *testing.T) {
	memory := New(256 * 1024 * 1024)
	defer memory.Free()
	concurrentMemory := memory.NewConcurrentMemory()
	defer concurrentMemory.Destroy()

	const sz = 100_0000
	m := test.PanicErr1(MakeMap[int, int](0, &concurrentMemory))
	gm := map[int]int{}
	for i := 0; i < sz; i++ {
		r := rand.Int()
		test.PanicErr(m.Put(r, 2*r))
		gm[r] = 2 * r

		test.Assert(m.Length() == len(gm), m.Length(), len(gm))
	}
	keys := make([]int, sz)
	for i := 0; i < sz; i++ {
		keys[i] = rand.Int()
	}
	for _, k := range keys {
		v1, ok1 := m.Get2(k)
		v2, ok2 := gm[k]
		test.Assert(v1 == v2, v1, v2)
		test.Assert(ok1 == ok2, ok1, ok2)
	}
	m.Free()
}

func BenchmarkMapGet(b *testing.B) {
	const sz = 100_0000
	rand.Seed(1)
	m := make(map[int]int, sz)
	for i := 0; i < sz; i++ {
		m[rand.Int()] = 0
	}
	keys := make([]int, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = rand.Int()
	}
	b.ResetTimer()
	for _, k := range keys {
		_ = m[k]
	}
}

func BenchmarkMyMapGet(b *testing.B) {
	memory := New(128 * 1024 * 1024)
	defer memory.Free()
	concurrentMemory := memory.NewConcurrentMemory()
	defer concurrentMemory.Destroy()

	const sz = 100_0000
	rand.Seed(1)
	m := test.PanicErr1(MakeMap[int, int](sz, &concurrentMemory))
	for i := 0; i < sz; i++ {
		test.PanicErr(m.Put(rand.Int(), 0))
	}
	keys := make([]int, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = rand.Int()
	}
	b.ResetTimer()
	for _, k := range keys {
		_ = m.Get(k)
	}
	b.StopTimer()
	m.Free()
}

func BenchmarkMyCustomMapGet(b *testing.B) {
	memory := New(128 * 1024 * 1024)
	defer memory.Free()
	concurrentMemory := memory.NewConcurrentMemory()
	defer concurrentMemory.Destroy()

	const sz = 100_0000
	rand.Seed(1)
	m := test.PanicErr1(MakeCustomMap[int, int](sz, func(k int) SizeType {
		return SizeType(k)
	}, func(k1 int, k2 int) bool {
		return k1 == k2
	}, &concurrentMemory))
	for i := 0; i < sz; i++ {
		test.PanicErr(m.Put(rand.Int(), 0))
	}
	keys := make([]int, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = rand.Int()
	}
	b.ResetTimer()
	for _, k := range keys {
		_ = m.Get(k)
	}
	b.StopTimer()
	m.Free()
}

func BenchmarkStringMapGet(b *testing.B) {
	const sz = 100_0000
	rand.Seed(1)
	m := make(map[string]int, sz)
	for i := 0; i < sz; i++ {
		m[strconv.Itoa(rand.Int())] = 0
	}
	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = strconv.Itoa(rand.Int())
	}
	b.ResetTimer()
	for _, k := range keys {
		_ = m[k]
	}
}

func BenchmarkMyStringMapGet(b *testing.B) {
	memory := New(1 * config.GB)
	defer memory.Free()
	concurrentMemory := memory.NewConcurrentMemory()
	defer concurrentMemory.Destroy()
	factory := NewStringFactory(&concurrentMemory)
	defer factory.Destroy()

	const sz = 100_0000
	rand.Seed(1)
	m := test.PanicErr1(MakeMap[String, int](sz, &concurrentMemory))
	keys := make([]String, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = test.PanicErr1(factory.CreateFromGoString(strconv.Itoa(rand.Int())))
	}
	b.ResetTimer()
	for _, k := range keys {
		_ = m.Get(k)
	}
	b.StopTimer()
	for _, key := range keys {
		key.Free(&concurrentMemory)
	}
	m.Free()
}

func BenchmarkMapPutGetExpense(b *testing.B) {
	for ii := 0; ii < b.N; ii++ {
		var sz = 100_0000
		gm := map[int]int{}
		for i := 0; i < sz; i++ {
			r := rand.Int()
			gm[r] = 2 * r
		}
		keys := make([]int, sz)
		for i := 0; i < sz; i++ {
			keys[i] = rand.Int()
		}
		for _, k := range keys {
			_ = gm[k]
		}
	}
}

func BenchmarkMyMapPutGetExpense(b *testing.B) {
	memory := New(256 * 1024 * 1024)
	defer memory.Free()
	concurrentMemory := memory.NewConcurrentMemory()
	defer concurrentMemory.Destroy()

	b.ResetTimer()
	for ii := 0; ii < b.N; ii++ {
		const sz = 100_0000
		m := test.PanicErr1(MakeMap[int, int](0, &concurrentMemory))
		for i := 0; i < sz; i++ {
			r := rand.Int()
			test.PanicErr(m.Put(r, 2*r))
		}
		keys := make([]int, sz)
		for i := 0; i < sz; i++ {
			keys[i] = rand.Int()
		}
		for _, k := range keys {
			_ = m.Get(k)
		}
		m.Free()
	}
}

func Test_simpleHash(t *testing.T) {
	logger.Info(Sizeof[simpleHashHelper[byte]]())
	logger.Info(Sizeof[simpleHashHelper[int32]]())

	logger.Info(simpleHash[byte](1).BitString())
	logger.Info(simpleHash[byte](10).BitString())
	logger.Info(simpleHash[byte](100).BitString())

	logger.Info(simpleHash[int64](1).BitString())
	logger.Info(simpleHash[int64](10).BitString())
	logger.Info(simpleHash[int64](100).BitString())

	logger.Info(simpleHash[int32](1).BitString())
	logger.Info(simpleHash[int32](10).BitString())
	logger.Info(simpleHash[int32](100).BitString())

	logger.Info(simpleHash[float64](1).BitString())
	logger.Info(simpleHash[float64](10).BitString())
	logger.Info(simpleHash[float64](100).BitString())
}

func BenchmarkHashInt(b *testing.B) {
	var h SizeType
	for i := 0; i < b.N; i++ {
		h += SizeType(i)
	}
	_, _ = fmt.Fprint(logger.NullLog, h)
}

func BenchmarkSimpleHashInt(b *testing.B) {
	var h SizeType
	for i := 0; i < b.N; i++ {
		h += simpleHash(i)
	}
	_, _ = fmt.Fprint(logger.NullLog, h)
}

func TestMap_Link(t *testing.T) {
	memory := New(256 * 1024 * 1024)
	defer memory.Free()
	concurrentMemory := memory.NewConcurrentMemory()
	defer concurrentMemory.Destroy()

	m, err := MakeCustomMap[int, int](0, func(key int) SizeType {
		return 3
	}, func(k1, k2 int) bool {
		return k1 == k2
	}, &concurrentMemory)
	test.PanicErr(err)
	defer func() { m.Free() }()

	test.Assert(m.Length() == 0, m.Length())

	test.PanicErr(m.DirectPut(1, 1))
	test.PanicErr(m.DirectPut(2, 2))
	test.PanicErr(m.DirectPut(3, 3))
	test.PanicErr(m.DirectPut(4, 4))
	test.PanicErr(m.DirectPut(5, 5))

	logger.Info(m.String())
	logger.Info(m.debugString())
}

func TestMap_Delete(t *testing.T) {
	memory := New(256 * 1024 * 1024)
	defer memory.Free()
	concurrentMemory := memory.NewConcurrentMemory()
	defer concurrentMemory.Destroy()

	for i := 1; i <= 5; i++ {
		m, err := MakeCustomMap[int, int](0, func(key int) SizeType {
			return 3
		}, func(k1, k2 int) bool {
			return k1 == k2
		}, &concurrentMemory)
		test.PanicErr(err)

		test.Assert(m.Length() == 0, m.Length())

		test.PanicErr(m.DirectPut(1, 1))
		test.PanicErr(m.DirectPut(2, 2))
		test.PanicErr(m.DirectPut(3, 3))
		test.PanicErr(m.DirectPut(4, 4))
		test.PanicErr(m.DirectPut(5, 5))

		test.Assert(m.Length() == 5, m.Length())

		logger.Info(m.debugString())
		m.Delete(i)
		logger.Info("del", i, m.debugString())
		test.Assert(m.Length() == 4, m.Length())

		for k := 1; k <= 5; k++ {
			val, ok := m.Get2(k)
			if k == i {
				test.Assert(!ok, k, val)
				test.Assert(val == 0, k, val)
			} else {
				test.Assert(ok, k, val)
				test.Assert(k == val, k, val)
			}
		}

		m.Free()

	}
}

func TestMap_Delete2(t *testing.T) {
	memory := New(256 * 1024 * 1024)
	defer memory.Free()
	concurrentMemory := memory.NewConcurrentMemory()
	defer concurrentMemory.Destroy()

	for i := 1; i <= 5; i++ {
		for j := 1; j <= 5; j++ {
			m, err := MakeCustomMap[int, int](0, func(key int) SizeType {
				return 3
			}, func(k1, k2 int) bool {
				return k1 == k2
			}, &concurrentMemory)
			test.PanicErr(err)

			test.Assert(m.Length() == 0, m.Length())

			test.PanicErr(m.DirectPut(1, 1))
			test.PanicErr(m.DirectPut(2, 2))
			test.PanicErr(m.DirectPut(3, 3))
			test.PanicErr(m.DirectPut(4, 4))
			test.PanicErr(m.DirectPut(5, 5))

			test.Assert(m.Length() == 5, m.Length())

			logger.Info(m.String())
			m.Delete(i)
			m.Delete(j)
			logger.Info("del", i, j, m.String(), m.String())
			if i == j {
				test.Assert(m.Length() == 4, m.Length())
			} else {
				test.Assert(m.Length() == 3, m.Length())
			}

			for k := 1; k <= 5; k++ {
				val, ok := m.Get2(k)
				if k == i || k == j {
					test.Assert(!ok, k, val)
					test.Assert(val == 0, k, val)
				} else {
					test.Assert(ok, k, val)
					test.Assert(k == val, k, val)
				}
			}

			m.Free()
		}
	}
}

func TestMap_Delete0(t *testing.T) {
	memory := New(256 * 1024 * 1024)
	defer memory.Free()
	concurrentMemory := memory.NewConcurrentMemory()
	defer concurrentMemory.Destroy()

	for i := 1; i <= 5; i++ {
		m, err := MakeMap[int, int](0, &concurrentMemory)
		test.PanicErr(err)

		test.Assert(m.Length() == 0, m.Length())

		test.PanicErr(m.DirectPut(1, 1))
		test.PanicErr(m.DirectPut(2, 2))
		test.PanicErr(m.DirectPut(3, 3))
		test.PanicErr(m.DirectPut(4, 4))
		test.PanicErr(m.DirectPut(5, 5))

		test.Assert(m.Length() == 5, m.Length())

		logger.Info(m.String())
		m.Delete(i)
		logger.Info("del", i, m.String())
		test.Assert(m.Length() == 4, m.Length())

		for k := 1; k <= 5; k++ {
			val, ok := m.Get2(k)
			if k == i {
				test.Assert(!ok, k, val)
				test.Assert(val == 0, k, val)
			} else {
				test.Assert(ok, k, val)
				test.Assert(k == val, k, val)
			}
		}

		m.Free()
	}
}

func TestMap_PutGetDelete(t *testing.T) {
	memory := New(256 * 1024 * 1024)
	defer memory.Free()
	concurrentMemory := memory.NewConcurrentMemory()
	defer concurrentMemory.Destroy()

	m, err := MakeMap[int, int](0, &concurrentMemory)
	test.PanicErr(err)
	defer func() { m.Free() }()

	gm := map[int]int{}

	const sz = 10000
	rand.Seed(1)
	var puts []int
	for i := 0; i < sz; i++ {
		k := rand.Intn(10000)
		test.PanicErr(m.Put(k, 0))
		gm[k] = 0
		puts = append(puts, k)
	}

	test.Assert(m.Length() == len(gm), m.Length(), len(gm))

	// get
	for _, k := range puts {
		{
			v1, ok1 := m.Get2(k)
			v2, ok2 := gm[k]
			test.Assert(v1 == v2, v1, v2)
			test.Assert(ok1 == ok2, v1, v2)
		}
		k = rand.Intn(10000)
		{
			v1, ok1 := m.Get2(k)
			v2, ok2 := gm[k]
			test.Assert(v1 == v2, v1, v2)
			test.Assert(ok1 == ok2, v1, v2)
		}
	}

	// del
	for _, k := range puts {
		{
			m.Delete(k)
			delete(gm, k)
			test.Assert(m.Length() == len(gm), m.Length(), len(gm))
		}
		{
			v1, ok1 := m.Get2(k)
			v2, ok2 := gm[k]
			test.Assert(v1 == v2, v1, v2)
			test.Assert(ok1 == ok2, v1, v2)
		}
		{
			var k2 = rand.Intn(10000)
			m.Delete(k2)
			delete(gm, k2)
			test.Assert(m.Length() == len(gm), m.Length(), len(gm))
		}
	}
}

func TestMap_float64Key(t *testing.T) {
	memory := New(256 * 1024 * 1024)
	defer memory.Free()
	concurrentMemory := memory.NewConcurrentMemory()
	defer concurrentMemory.Destroy()

	m, err := MakeMap[float64, int](0, &concurrentMemory)
	test.PanicErr(err)
	defer func() { m.Free() }()

	test.Assert(m.Length() == 0, m.Length())

	test.PanicErr(m.DirectPut(1, 1))
	test.PanicErr(m.DirectPut(2, 2))
	test.PanicErr(m.DirectPut(3, 3))
	test.PanicErr(m.DirectPut(4, 4))
	test.PanicErr(m.DirectPut(5, 5))

	logger.Info(m.String())
	logger.Info(m.debugString())
}

func TestMap_Iterator(t *testing.T) {
	memory := New(256 * 1024 * 1024)
	defer memory.Free()
	concurrentMemory := memory.NewConcurrentMemory()
	defer concurrentMemory.Destroy()

	m, err := MakeMap[int, int](0, &concurrentMemory)
	test.PanicErr(err)
	defer func() { m.Free() }()

	test.PanicErr(m.Put(1, 1))

	iter := m.Iterator()
	for iter.Next() {
		logger.Info(iter.Key(), iter.Value())
	}
}

func TestMap_Iterator2(t *testing.T) {
	memory := New(256 * 1024 * 1024)
	defer memory.Free()
	concurrentMemory := memory.NewConcurrentMemory()
	defer concurrentMemory.Destroy()

	m, err := MakeCustomMap[int, int](0, func(key int) SizeType {
		return 1
	}, func(i int, i2 int) bool {
		return i == i2
	}, &concurrentMemory)
	test.PanicErr(err)
	defer func() { m.Free() }()

	test.PanicErr(m.Put(1, 1))
	test.PanicErr(m.Put(2, 2))

	iter := m.Iterator()
	for iter.Next() {
		logger.Info(iter.Key(), iter.Value())
	}
	logger.Info(m.debugString())
}

func TestMap_Iterator3(t *testing.T) {
	memory := New(256 * 1024 * 1024)
	defer memory.Free()
	concurrentMemory := memory.NewConcurrentMemory()
	defer concurrentMemory.Destroy()

	m, err := MakeCustomMap[int, int](0, func(key int) SizeType {
		return SizeType(key)
	}, func(i int, i2 int) bool {
		return i == i2
	}, &concurrentMemory)
	test.PanicErr(err)
	defer func() { m.Free() }()

	test.PanicErr(m.Put(7, 7))
	test.PanicErr(m.Put(8, 8))

	iter := m.Iterator()
	for iter.Next() {
		logger.Info(iter.Key(), iter.Value())
	}
	logger.Info(m.debugString())
}

func TestMap_Iterator4(t *testing.T) {
	memory := New(256 * 1024 * 1024)
	defer memory.Free()
	concurrentMemory := memory.NewConcurrentMemory()
	defer concurrentMemory.Destroy()

	const sz = 100_0000
	m := test.PanicErr1(MakeMap[int, int](0, &concurrentMemory))
	gm := map[int]int{}
	for i := 0; i < sz; i++ {
		r := rand.Int()
		test.PanicErr(m.Put(r, 2*r))
		gm[r] = 2 * r

		test.Assert(m.Length() == len(gm), m.Length(), len(gm))
	}
	logger.Info(len(gm), m.Length())
	cnt := 0
	for k, v := range gm {
		v2, ok := m.Get2(k)
		test.Assert(ok)
		test.Assert(v == v2)
		cnt++
	}
	test.Assert(cnt == len(gm))
	iter := m.Iterator()
	for iter.Next() {
		v, ok := gm[iter.Key()]
		test.Assert(ok)
		test.Assert(v == iter.Value())
		cnt--
	}
	test.Assert(cnt == 0)
	m.Free()
}

func TestMap_Move(t *testing.T) {
	memory := New(256 * 1024 * 1024)
	defer memory.Free()
	concurrentMemory := memory.NewConcurrentMemory()
	defer concurrentMemory.Destroy()

	m, err := MakeMap[int, int](0, &concurrentMemory)
	test.PanicErr(err)

	m2 := m.Move()
	m.Free()
	m2.Free()
}

func TestMap_Move2(t *testing.T) {
	memory := New(256 * 1024 * 1024)
	defer memory.Free()
	concurrentMemory := memory.NewConcurrentMemory()
	defer concurrentMemory.Destroy()

	m, err := MakeMap[int, int](0, &concurrentMemory)
	test.PanicErr(err)

	m2 := m.Move()
	m2.Free()
}

func TestMap_Move3(t *testing.T) {
	memory := New(256 * 1024 * 1024)
	concurrentMemory := memory.NewConcurrentMemory()
	m, err := MakeMap[int, int](0, &concurrentMemory)
	test.PanicErr(err)

	_ = m.Move()
	m.Free()

	concurrentMemory.Destroy()
	var recovered any
	func() {
		defer func() {
			recovered = recover()
		}()
		logger.Info("========= NEXT ERROR IS OK =========")
		memory.Free()
	}()

	test.Assert(recovered != nil)
	logger.Info(recovered)
}

type Obj[T any] struct {
	hash func(T) uint64
}

func hasher[T any](v T) uint64 {
	return *((*uint64)(unsafe.Pointer(&v)))
}

func Benchmark_hasher(b *testing.B) {
	o := Obj[int]{hash: hasher[int]}
	s := uint64(0)
	for i := 0; i < b.N; i++ {
		s += o.hash(i)
	}
}

func Benchmark_myhasher(b *testing.B) {
	o := Obj[int]{hash: func(i int) uint64 {
		return uint64(i)
	}}
	s := uint64(0)
	for i := 0; i < b.N; i++ {
		s += o.hash(i)
	}
}
