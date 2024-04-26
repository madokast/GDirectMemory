package direct

import (
	"github.com/madokast/direct/utils"
	"math/rand"
	"strconv"
	"testing"
)

func TestStringFactory(t *testing.T) {
	Global.Init(1 * MB)
	defer Global.Free()

	factory := NewStringFactory()
	defer factory.Destroy()

	s1, err := factory.CreateFromGoString("hello")
	utils.PanicErr(err)
	defer s1.Free()

	t.Log(s1.Length())         // 5
	t.Log(s1)                  // hello
	t.Log(s1.AsGoString())     // hello
	t.Log(s1.CopyToGoString()) // hello

	utils.Assert(s1.CopyToGoString() == "hello")
}

func TestString_Hashcode(t *testing.T) {
	Global.Init(1 * MB)
	defer Global.Free()

	factory := NewStringFactory()
	defer factory.Destroy()

	s1, err := factory.CreateFromGoString("hello")
	utils.PanicErr(err)
	defer s1.Free()

	utils.Assert(s1.CopyToGoString() == "hello")

	t.Log(s1.Hashcode())
}

func Test_IsString(t *testing.T) {
	utils.Assert(isString[String]())
	utils.Assert(!isString[string]())
}

func Benchmark_IsString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		utils.Assert(isString[String]())
	}
}

func TestStringFactory_CreateFromGoString(t *testing.T) {
	Global.Init(1 * MB)
	defer Global.Free()

	factory := NewStringFactory()
	defer factory.Destroy()

	m, err := MakeMap[String, SizeType](0)
	utils.PanicErr(err)
	defer m.Free()

	for _, s := range []string{"hello", ",", " ", "world", "!"} {
		ss, err := factory.CreateFromGoString(s)
		utils.PanicErr(err)

		err = m.Put(ss, ss.Length())
		utils.PanicErr(err)

		t.Logf("%v", m)
	}

	iter := m.Iterator()
	for iter.Next() {
		iter.KeyRef().Free()
	}
}

func TestStringFactory_CreateFromGoString1000(t *testing.T) {
	Global.Init(128 * MB)
	defer Global.Free()

	factory := NewStringFactory()
	defer factory.Destroy()

	m, err := MakeMap[String, SizeType](0)
	utils.PanicErr(err)
	defer m.Free()

	for i := 0; i < 1000; i++ {
		s := strconv.Itoa(rand.Int())
		ss, err := factory.CreateFromGoString(s)
		utils.PanicErr(err)
		err = m.Put(ss, ss.Length())
		utils.PanicErr(err)
	}

	iter := m.Iterator()
	for iter.Next() {
		iter.KeyRef().Free()
	}
}

func TestStringFactory_CreateFromGoString_bigString(t *testing.T) {
	Global.Init(16 * MB)
	defer Global.Free()

	factory := NewStringFactory()
	defer factory.Destroy()

	m, err := MakeMap[String, SizeType](0)
	utils.PanicErr(err)
	defer m.Free()

	for _, s := range []string{"hello", string(make([]byte, 1024)), " ", string(make([]byte, 4096)), "!"} {
		ss, err := factory.CreateFromGoString(s)
		utils.PanicErr(err)

		err = m.Put(ss, ss.Length())
		utils.PanicErr(err)

		t.Logf("%v", m)
	}

	iter := m.Iterator()
	for iter.Next() {
		iter.KeyRef().Free()
	}
}
