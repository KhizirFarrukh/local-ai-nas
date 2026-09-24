package storage

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestLocksExclude: holders of one key never overlap, holders of other
// keys do not wait for each other, and keys are dropped when released.
func TestLocksExclude(t *testing.T) {
	l := NewLocks()
	var inside, maxInside atomic.Int32
	var wg sync.WaitGroup
	for i := range 50 {
		key := "u0001:docs"
		if i%2 == 0 {
			key = "U0001:DOCS" // the same key without case
		}
		wg.Go(func() {
			unlock := l.Lock(key)
			n := inside.Add(1)
			for {
				m := maxInside.Load()
				if n <= m || maxInside.CompareAndSwap(m, n) {
					break
				}
			}
			time.Sleep(time.Millisecond)
			inside.Add(-1)
			unlock()
		})
	}
	wg.Wait()
	if maxInside.Load() != 1 {
		t.Errorf("%d holders of one key at once, want 1", maxInside.Load())
	}
	if l.Len() != 0 {
		t.Errorf("%d keys left after all unlocks, want 0", l.Len())
	}

	// Another key is not blocked by a held one.
	unlockA := l.Lock("a")
	done := make(chan struct{})
	go func() {
		l.Lock("b")()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("key b waited for key a")
	}
	if l.Len() != 1 {
		t.Errorf("%d keys held, want 1", l.Len())
	}
	unlockA()
}
