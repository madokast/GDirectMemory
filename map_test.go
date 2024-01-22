package direct

import (
	"math/rand"
	"strconv"
	"testing"
	"unsafe"
)

func TestMakeMap(t *testing.T) {
	memory := New(1024 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	m, err := MakeMap[int, string](8, &localMemory)
	PanicErr(err)
	defer func() {
		m.Free(&localMemory)
	}()

	PanicErr(m.DirectPut(0, "a", &localMemory))
	PanicErr(m.DirectPut(1, "a", &localMemory))
	PanicErr(m.DirectPut(2, "a", &localMemory))
	PanicErr(m.DirectPut(3, "a", &localMemory))
	PanicErr(m.DirectPut(4, "a", &localMemory))
	PanicErr(m.DirectPut(5, "a", &localMemory))
	PanicErr(m.DirectPut(8, "a", &localMemory))

	PanicErr(m.Put(3, "q", &localMemory))
	PanicErr(m.Put(5, "b", &localMemory))
	PanicErr(m.Put(4, "c", &localMemory))
	PanicErr(m.Put(8, "d", &localMemory))
	PanicErr(m.Put(12, "k", &localMemory))

	for i, e := range m.header().table.GoSlice() {
		t.Log(i, e)
	}

	t.Log(m.GoMap())
	t.Log(m.String())
}

func TestMapCorrect(t *testing.T) {
	memory := New(256 * 1024 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	const sz = 100_0000
	m := PanicErr1(MakeMap[int, int](0, &localMemory))
	gm := map[int]int{}
	for i := 0; i < sz; i++ {
		r := rand.Int()
		PanicErr(m.Put(r, 2*r, &localMemory))
		gm[r] = 2 * r

		Assert(m.Length() == len(gm), m.Length(), len(gm))
	}
	keys := make([]int, sz)
	for i := 0; i < sz; i++ {
		keys[i] = rand.Int()
	}
	for _, k := range keys {
		v1, ok1 := m.Get2(k)
		v2, ok2 := gm[k]
		Assert(v1 == v2, v1, v2)
		Assert(ok1 == ok2, ok1, ok2)
	}
	m.Free(&localMemory)
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
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	const sz = 100_0000
	rand.Seed(1)
	m := PanicErr1(MakeMap[int, int](sz, &localMemory))
	for i := 0; i < sz; i++ {
		PanicErr(m.Put(rand.Int(), 0, &localMemory))
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
	m.Free(&localMemory)
}

func BenchmarkMyCustomMapGet(b *testing.B) {
	memory := New(128 * 1024 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	const sz = 100_0000
	rand.Seed(1)
	m := PanicErr1(MakeCustomMap[int, int](sz, func(k int) SizeType {
		return SizeType(k)
	}, func(k1 int, k2 int) bool {
		return k1 == k2
	}, &localMemory))
	for i := 0; i < sz; i++ {
		PanicErr(m.Put(rand.Int(), 0, &localMemory))
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
	m.Free(&localMemory)
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
	memory := New(1 * GB)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()
	factory := NewStringFactory()
	defer factory.Destroy(&localMemory)

	const sz = 100_0000
	rand.Seed(1)
	m := PanicErr1(MakeMap[String, int](sz, &localMemory))
	keys := make([]String, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = PanicErr1(factory.CreateFromGoString(strconv.Itoa(rand.Int()), &localMemory))
	}
	b.ResetTimer()
	for _, k := range keys {
		_ = m.Get(k)
	}
	b.StopTimer()
	for _, key := range keys {
		key.Free(&localMemory)
	}
	m.Free(&localMemory)
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
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	b.ResetTimer()
	for ii := 0; ii < b.N; ii++ {
		const sz = 100_0000
		m := PanicErr1(MakeMap[int, int](0, &localMemory))
		for i := 0; i < sz; i++ {
			r := rand.Int()
			PanicErr(m.Put(r, 2*r, &localMemory))
		}
		keys := make([]int, sz)
		for i := 0; i < sz; i++ {
			keys[i] = rand.Int()
		}
		for _, k := range keys {
			_ = m.Get(k)
		}
		m.Free(&localMemory)
	}
}

func Test_simpleHash(t *testing.T) {
	t.Log(Sizeof[simpleHashHelper[byte]]())
	t.Log(Sizeof[simpleHashHelper[int32]]())

	t.Log(simpleHash[byte](1).BitString())
	t.Log(simpleHash[byte](10).BitString())
	t.Log(simpleHash[byte](100).BitString())

	t.Log(simpleHash[int64](1).BitString())
	t.Log(simpleHash[int64](10).BitString())
	t.Log(simpleHash[int64](100).BitString())

	t.Log(simpleHash[int32](1).BitString())
	t.Log(simpleHash[int32](10).BitString())
	t.Log(simpleHash[int32](100).BitString())

	t.Log(simpleHash[float64](1).BitString())
	t.Log(simpleHash[float64](10).BitString())
	t.Log(simpleHash[float64](100).BitString())
}

func BenchmarkHashInt(b *testing.B) {
	var h SizeType
	for i := 0; i < b.N; i++ {
		h += SizeType(i)
	}
}

func BenchmarkSimpleHashInt(b *testing.B) {
	var h SizeType
	for i := 0; i < b.N; i++ {
		h += simpleHash(i)
	}
}

func TestMap_Link(t *testing.T) {
	memory := New(256 * 1024 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	m, err := MakeCustomMap[int, int](0, func(key int) SizeType {
		return 3
	}, func(k1, k2 int) bool {
		return k1 == k2
	}, &localMemory)
	PanicErr(err)
	defer func() { m.Free(&localMemory) }()

	Assert(m.Length() == 0, m.Length())

	PanicErr(m.DirectPut(1, 1, &localMemory))
	PanicErr(m.DirectPut(2, 2, &localMemory))
	PanicErr(m.DirectPut(3, 3, &localMemory))
	PanicErr(m.DirectPut(4, 4, &localMemory))
	PanicErr(m.DirectPut(5, 5, &localMemory))

	t.Log(m.String())
	t.Log(m.debugString())
}

func TestMap_Delete(t *testing.T) {
	memory := New(256 * 1024 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	for i := 1; i <= 5; i++ {
		m, err := MakeCustomMap[int, int](0, func(key int) SizeType {
			return 3
		}, func(k1, k2 int) bool {
			return k1 == k2
		}, &localMemory)
		PanicErr(err)

		Assert(m.Length() == 0, m.Length())

		PanicErr(m.DirectPut(1, 1, &localMemory))
		PanicErr(m.DirectPut(2, 2, &localMemory))
		PanicErr(m.DirectPut(3, 3, &localMemory))
		PanicErr(m.DirectPut(4, 4, &localMemory))
		PanicErr(m.DirectPut(5, 5, &localMemory))

		Assert(m.Length() == 5, m.Length())

		t.Log(m.debugString())
		m.Delete(i)
		t.Log("del", i, m.debugString())
		Assert(m.Length() == 4, m.Length())

		for k := 1; k <= 5; k++ {
			val, ok := m.Get2(k)
			if k == i {
				Assert(!ok, k, val)
				Assert(val == 0, k, val)
			} else {
				Assert(ok, k, val)
				Assert(k == val, k, val)
			}
		}

		m.Free(&localMemory)

	}
}

func TestMap_Delete2(t *testing.T) {
	memory := New(256 * 1024 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	for i := 1; i <= 5; i++ {
		for j := 1; j <= 5; j++ {
			m, err := MakeCustomMap[int, int](0, func(key int) SizeType {
				return 3
			}, func(k1, k2 int) bool {
				return k1 == k2
			}, &localMemory)
			PanicErr(err)

			Assert(m.Length() == 0, m.Length())

			PanicErr(m.DirectPut(1, 1, &localMemory))
			PanicErr(m.DirectPut(2, 2, &localMemory))
			PanicErr(m.DirectPut(3, 3, &localMemory))
			PanicErr(m.DirectPut(4, 4, &localMemory))
			PanicErr(m.DirectPut(5, 5, &localMemory))

			Assert(m.Length() == 5, m.Length())

			t.Log(m.String())
			m.Delete(i)
			m.Delete(j)
			t.Log("del", i, j, m.String(), m.String())
			if i == j {
				Assert(m.Length() == 4, m.Length())
			} else {
				Assert(m.Length() == 3, m.Length())
			}

			for k := 1; k <= 5; k++ {
				val, ok := m.Get2(k)
				if k == i || k == j {
					Assert(!ok, k, val)
					Assert(val == 0, k, val)
				} else {
					Assert(ok, k, val)
					Assert(k == val, k, val)
				}
			}

			m.Free(&localMemory)
		}
	}
}

func TestMap_Delete0(t *testing.T) {
	memory := New(256 * 1024 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	for i := 1; i <= 5; i++ {
		m, err := MakeMap[int, int](0, &localMemory)
		PanicErr(err)

		Assert(m.Length() == 0, m.Length())

		PanicErr(m.DirectPut(1, 1, &localMemory))
		PanicErr(m.DirectPut(2, 2, &localMemory))
		PanicErr(m.DirectPut(3, 3, &localMemory))
		PanicErr(m.DirectPut(4, 4, &localMemory))
		PanicErr(m.DirectPut(5, 5, &localMemory))

		Assert(m.Length() == 5, m.Length())

		t.Log(m.String())
		m.Delete(i)
		t.Log("del", i, m.String())
		Assert(m.Length() == 4, m.Length())

		for k := 1; k <= 5; k++ {
			val, ok := m.Get2(k)
			if k == i {
				Assert(!ok, k, val)
				Assert(val == 0, k, val)
			} else {
				Assert(ok, k, val)
				Assert(k == val, k, val)
			}
		}

		m.Free(&localMemory)
	}
}

func TestMap_PutGetDelete(t *testing.T) {
	memory := New(256 * 1024 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	m, err := MakeMap[int, int](0, &localMemory)
	PanicErr(err)
	defer func() { m.Free(&localMemory) }()

	gm := map[int]int{}

	const sz = 10000
	rand.Seed(1)
	var puts []int
	for i := 0; i < sz; i++ {
		k := rand.Intn(10000)
		PanicErr(m.Put(k, 0, &localMemory))
		gm[k] = 0
		puts = append(puts, k)
	}

	Assert(m.Length() == len(gm), m.Length(), len(gm))

	// get
	for _, k := range puts {
		{
			v1, ok1 := m.Get2(k)
			v2, ok2 := gm[k]
			Assert(v1 == v2, v1, v2)
			Assert(ok1 == ok2, v1, v2)
		}
		k = rand.Intn(10000)
		{
			v1, ok1 := m.Get2(k)
			v2, ok2 := gm[k]
			Assert(v1 == v2, v1, v2)
			Assert(ok1 == ok2, v1, v2)
		}
	}

	// del
	for _, k := range puts {
		{
			m.Delete(k)
			delete(gm, k)
			Assert(m.Length() == len(gm), m.Length(), len(gm))
		}
		{
			v1, ok1 := m.Get2(k)
			v2, ok2 := gm[k]
			Assert(v1 == v2, v1, v2)
			Assert(ok1 == ok2, v1, v2)
		}
		{
			var k2 = rand.Intn(10000)
			m.Delete(k2)
			delete(gm, k2)
			Assert(m.Length() == len(gm), m.Length(), len(gm))
		}
	}
}

func TestMap_float64Key(t *testing.T) {
	memory := New(256 * 1024 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	m, err := MakeMap[float64, int](0, &localMemory)
	PanicErr(err)
	defer func() { m.Free(&localMemory) }()

	Assert(m.Length() == 0, m.Length())

	PanicErr(m.DirectPut(1, 1, &localMemory))
	PanicErr(m.DirectPut(2, 2, &localMemory))
	PanicErr(m.DirectPut(3, 3, &localMemory))
	PanicErr(m.DirectPut(4, 4, &localMemory))
	PanicErr(m.DirectPut(5, 5, &localMemory))

	t.Log(m.String())
	t.Log(m.debugString())
}

func TestMap_Iterator(t *testing.T) {
	memory := New(256 * 1024 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	m, err := MakeMap[int, int](0, &localMemory)
	PanicErr(err)
	defer func() { m.Free(&localMemory) }()

	PanicErr(m.Put(1, 1, &localMemory))

	iter := m.Iterator()
	for iter.Next() {
		t.Log(iter.Key(), iter.Value())
	}
}

func TestMap_Iterator2(t *testing.T) {
	memory := New(256 * 1024 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	m, err := MakeCustomMap[int, int](0, func(key int) SizeType {
		return 1
	}, func(i int, i2 int) bool {
		return i == i2
	}, &localMemory)
	PanicErr(err)
	defer func() { m.Free(&localMemory) }()

	PanicErr(m.Put(1, 1, &localMemory))
	PanicErr(m.Put(2, 2, &localMemory))

	iter := m.Iterator()
	for iter.Next() {
		t.Log(iter.Key(), iter.Value())
	}
	t.Log(m.debugString())
}

func TestMap_Iterator3(t *testing.T) {
	memory := New(256 * 1024 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	m, err := MakeCustomMap[int, int](0, func(key int) SizeType {
		return SizeType(key)
	}, func(i int, i2 int) bool {
		return i == i2
	}, &localMemory)
	PanicErr(err)
	defer func() { m.Free(&localMemory) }()

	PanicErr(m.Put(7, 7, &localMemory))
	PanicErr(m.Put(8, 8, &localMemory))

	iter := m.Iterator()
	for iter.Next() {
		t.Log(iter.Key(), iter.Value())
	}
	t.Log(m.debugString())
}

func TestMap_Iterator4(t *testing.T) {
	memory := New(256 * 1024 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	const sz = 100_0000
	m := PanicErr1(MakeMap[int, int](0, &localMemory))
	gm := map[int]int{}
	for i := 0; i < sz; i++ {
		r := rand.Int()
		PanicErr(m.Put(r, 2*r, &localMemory))
		gm[r] = 2 * r

		Assert(m.Length() == len(gm), m.Length(), len(gm))
	}
	t.Log(len(gm), m.Length())
	cnt := 0
	for k, v := range gm {
		v2, ok := m.Get2(k)
		Assert(ok)
		Assert(v == v2)
		cnt++
	}
	Assert(cnt == len(gm))
	iter := m.Iterator()
	for iter.Next() {
		v, ok := gm[iter.Key()]
		Assert(ok)
		Assert(v == iter.Value())
		cnt--
	}
	Assert(cnt == 0)
	m.Free(&localMemory)
}

func TestMap_Move(t *testing.T) {
	memory := New(256 * 1024 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	m, err := MakeMap[int, int](0, &localMemory)
	PanicErr(err)

	m2 := m.Move()
	m.Free(&localMemory)
	m2.Free(&localMemory)
}

func TestMap_Move2(t *testing.T) {
	memory := New(256 * 1024 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	m, err := MakeMap[int, int](0, &localMemory)
	PanicErr(err)

	m2 := m.Move()
	m2.Free(&localMemory)
}

func TestMap_Move3(t *testing.T) {
	memory := New(256 * 1024 * 1024)
	localMemory := memory.NewLocalMemory()
	m, err := MakeMap[int, int](0, &localMemory)
	PanicErr(err)

	_ = m.Move() // leak
	m.Free(&localMemory)

	localMemory.Destroy()
	Assert(memory.IsMemoryLeak())
	memory.Free()
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

func Test_isMap(t *testing.T) {
	t.Log(isMap[int]())
	t.Log(isMap[Map[int, int]]())
	t.Log(isMap[*Map[int, int]]())

	Assert(!isMap[int]())
	Assert(isMap[Map[int, int]]())
	Assert(!isMap[*Map[int, int]]())
}

func TestMap_MapMap(t *testing.T) {
	memory := New(256 * 1024 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	mm, err := MakeMap[int, Map[int, int]](0, &localMemory)
	PanicErr(err)
	defer func() { mm.Free(&localMemory) }()

	for i := 0; i < 3; i++ {
		m, err := MakeMapFromGoMap(map[int]int{
			i:      i,
			i + 10: i + 10,
		}, &localMemory)
		PanicErr(err)
		defer func() { m.Free(&localMemory) }()

		err = mm.Put(i, m.Move(), &localMemory)
		PanicErr(err)
	}

	t.Log(mm)

	mmIter := mm.Iterator()
	for mmIter.Next() {
		mmIter.Value().Free(&localMemory)
	}
}
