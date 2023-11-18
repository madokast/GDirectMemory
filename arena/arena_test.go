package arena

import (
	"fmt"
	"reflect"
	"testing"
)

func TestAllocate(t *testing.T) {
	var a = New(1024)
	u := a.Allocate(5, 1)
	t.Log(u)
}

func TestAlignTo(t *testing.T) {
	for i := 0; i < 20; i++ {
		p := uintptr(i) + alignPadSize(uintptr(i), 1)
		t.Log("align 1", i, p)
		assertEq(p, uintptr(i))
	}

	for i := 0; i < 20; i++ {
		p := uintptr(i) + alignPadSize(uintptr(i), 2)
		t.Log("align 2", i, p)
		assertEq(p, uintptr(i+i%2))
	}

	for i := 0; i < 20; i++ {
		p := uintptr(i) + alignPadSize(uintptr(i), 4)
		t.Log("align 4", i, p)
		assertTrue(p%4 == 0)
		assertTrue(p >= uintptr(i))
	}

	for i := 0; i < 20; i++ {
		p := uintptr(i) + alignPadSize(uintptr(i), 8)
		t.Log("align 8", i, p)
		assertTrue(p%8 == 0)
		assertTrue(p >= uintptr(i))
	}
}

func assertEq[E any](a, b E, msg ...any) {
	if !reflect.DeepEqual(a, b) {
		panic(fmt.Sprintf("assert fail %v != %v %v", a, b, msg))
	}
}

func assertTrue(b bool, msg ...any) {
	if !b {
		panic(fmt.Sprintf("assert fail %v", msg))
	}
}
