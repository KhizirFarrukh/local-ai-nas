// Package schedule runs periodic background tasks (S01.4-T06). S01 has a
// simple ticker; the job system of stage S04.3 replaces it behind the same
// interface.
package schedule

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Scheduler runs tasks in the background.
type Scheduler interface {
	// Every runs task once at once and then every interval until ctx
	// ends. A task that is still running when the next run is due delays
	// it; runs never overlap.
	Every(ctx context.Context, name string, interval time.Duration, task func(context.Context))
}

// Ticker is the S01 Scheduler: one goroutine and time.Ticker per task.
type Ticker struct {
	Log *slog.Logger
	wg  sync.WaitGroup
}

var _ Scheduler = (*Ticker)(nil)

// Every implements Scheduler.
func (t *Ticker) Every(ctx context.Context, name string, interval time.Duration, task func(context.Context)) {
	t.wg.Go(func() {
		tick := time.NewTicker(interval)
		defer tick.Stop()
		for {
			start := time.Now()
			task(ctx)
			if t.Log != nil {
				t.Log.DebugContext(ctx, "background task ran", "task", name, "duration_ms", time.Since(start).Milliseconds())
			}
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
			}
		}
	})
}

// Wait waits until every task has stopped; call it after ctx ended.
func (t *Ticker) Wait() { t.wg.Wait() }
