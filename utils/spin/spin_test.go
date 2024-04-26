package spin

import (
	"sync"
	"testing"
)

func BenchmarkKeepRun(b *testing.B) {
	for i := 0; i < b.N; i++ {
		keepRunning()
	}
}

func TestMutex_Lock(t *testing.T) {
	var mu Mutex
	var s int
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for k := 0; k < 1000; k++ {
				mu.Lock()
				s += 1
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	t.Log(s)
	if s != 1000*1000 {
		t.Fail()
	}
}

var mu Mutex

func TestMutex_Lock2(t *testing.T) {
	var s int
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for k := 0; k < 1000; k++ {
				mu.Lock()
				s += 1
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	t.Log(s)
	if s != 1000*1000 {
		t.Fail()
	}
}
