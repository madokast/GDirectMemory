package spin

import (
	"github.com/madokast/direct/utils"
	"math"
	"sync/atomic"
	"time"
)

const (
	unlock = 0
	lock   = 1
)

// Mutex will not yield the processor instead of sync.Mutex.
// Used when DisablePreempt
type Mutex struct {
	state uintptr
}

func (mu *Mutex) Lock() {
	for !atomic.CompareAndSwapUintptr(&mu.state, unlock, lock) {
		keepRunning()
	}
}

func (mu *Mutex) Unlock() {
	if utils.Asserted {
		if !atomic.CompareAndSwapUintptr(&mu.state, lock, unlock) {
			panic("unlock an unlock locker")
		}
		return
	}
	atomic.StoreUintptr(&mu.state, unlock)
}

func keepRunning() {
	var s float64
	for i := 0; i < int(time.Now().UnixMilli())%1000; i++ {
		a := math.Sin(float64(i))
		if math.IsNaN(a) {
			continue
		}
		s += a
	}
	if math.IsNaN(s) {
		panic(s)
	}
}
