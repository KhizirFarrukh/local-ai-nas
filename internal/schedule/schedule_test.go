package schedule

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestTickerRunsUntilCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	var runs atomic.Int32
	var s Ticker
	s.Every(ctx, "test", 10*time.Millisecond, func(context.Context) { runs.Add(1) })
	for deadline := time.Now().Add(5 * time.Second); runs.Load() < 3 && time.Now().Before(deadline); {
		time.Sleep(5 * time.Millisecond)
	}
	cancel()
	s.Wait()
	n := runs.Load()
	if n < 3 {
		t.Fatalf("the task ran %d times, want at least 3 (once at once, then on ticks)", n)
	}
	time.Sleep(30 * time.Millisecond)
	if runs.Load() != n {
		t.Error("the task ran after the context ended")
	}
}
