package arena

import (
	"math/rand"
	"testing"
)

func BenchmarkGoAllocate(t *testing.B) {
	var s = make([][]byte, t.N)
	for i := 0; i < t.N; i++ {
		s[i] = make([]byte, rand.Int31n(100))
	}
}

func BenchmarkArenaAllocate(t *testing.B) {
	var a = New(64 * 1024 * 1024)
	var s = make([]uintptr, t.N)
	for i := 0; i < t.N; i++ {
		s[i] = a.Allocate(uint(rand.Int31n(100)), 1)
	}
}
