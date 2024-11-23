package pauser

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type Pauser struct {
	cond   *sync.Cond
	paused atomic.Bool
}

func NewPauser() *Pauser {
	var locker sync.Mutex
	return &Pauser{
		cond: sync.NewCond(&locker),
	}
}

func (pauser *Pauser) Pause() {
	if pauser.paused.Load() {
		return
	}
	pauser.cond.L.Lock()
	fmt.Println("Pausing")
	pauser.paused.Store(true)
	pauser.cond.Broadcast()
	pauser.cond.L.Unlock()
}

func (pauser *Pauser) Unpause() {
	if !pauser.paused.Load() {
		return
	}
	pauser.cond.L.Lock()
	fmt.Println("Unpausing")
	pauser.paused.Store(false)
	pauser.cond.Broadcast()
	pauser.cond.L.Unlock()
}

func (pauser *Pauser) Wait() {
	if !pauser.paused.Load() {
		return
	}
	pauser.cond.L.Lock()
	for pauser.paused.Load() {
		fmt.Println("Waiting")
		pauser.cond.Wait()
	}
	pauser.cond.L.Unlock()
}
