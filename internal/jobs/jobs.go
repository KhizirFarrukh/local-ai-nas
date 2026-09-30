// Package jobs is the durable job foundation (S01.4-T08, ADR-0011
// amendment 1, FR-351): long operations and periodic work run as jobs in
// the SQLite database, so they survive a restart. S04.3 grows it
// (priorities, per-type limits, the monitor) instead of replacing it.
//
// A job is claimed by one worker at a time; a job that was running when
// the server stopped is queued again at the next start, so every handler
// must be safe to run again. A failed attempt is retried with backoff up
// to the job's attempts; cancel stops a queued job at once and a running
// one at its next progress report. Progress is written at most once a
// second per job, so a busy job does not flood the disk (NFR-056).
package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/KhizirFarrukh/local-ai-nas/internal/db"
	"github.com/KhizirFarrukh/local-ai-nas/internal/schedule"
)

// State is where a job is in its life.
type State string

// The states.
const (
	Queued    State = "queued"
	Running   State = "running"
	Succeeded State = "succeeded"
	Failed    State = "failed"
	Canceled  State = "canceled"
)

// Finished reports whether the state is final.
func (s State) Finished() bool { return s == Succeeded || s == Failed || s == Canceled }

// Job is one job.
type Job struct {
	ID    string
	Type  string
	Owner string // the namespace it works for; "" for system jobs
	State State
	// Payload is the job's input, and Result its output once it
	// succeeded (JSON).
	Payload, Result json.RawMessage
	// Done and Total are the progress in the job's own units (bytes,
	// items).
	Done, Total     int64
	Attempts        int
	MaxAttempts     int
	CancelRequested bool
	Error           string // the last failure, safe to show
	CreatedAt       time.Time
	UpdatedAt       time.Time
	FinishedAt      time.Time // zero until finished
}

// Handler runs one job and returns its result (marshalled to JSON), or an
// error. It reports progress through progress. It must be safe to run
// again: an interrupted job runs again from the start.
type Handler func(ctx context.Context, j Job, progress func(done, total int64)) (any, error)

// Errors of the queue.
var (
	ErrNotFound = errors.New("jobs: no such job")
	// ErrFinished means the job cannot be canceled any more.
	ErrFinished = errors.New("jobs: the job has finished")
)

// permanent marks an error that is not retried.
type permanent struct{ err error }

func (p permanent) Error() string { return p.err.Error() }
func (p permanent) Unwrap() error { return p.err }

// Permanent wraps err so the job fails at once instead of being retried
// (for errors a retry cannot fix, such as a missing source).
func Permanent(err error) error { return permanent{err} }

// Options configures the queue.
type Options struct {
	// Workers is the number of jobs run at once; 0 means 2, sized for a
	// Raspberry Pi's four cores (NFR-051).
	Workers int
	// Poll is how often idle workers look for due jobs (for example retries
	// whose backoff ended); new jobs wake them at once. 0 means 2 seconds.
	Poll time.Duration
	// Lease is how long a claimed job counts as running without a
	// progress report; 0 means 5 minutes. (With one server process a
	// restart re-queues every running job anyway.)
	Lease time.Duration
	// Keep is how long finished jobs are kept; 0 means 7 days.
	Keep time.Duration
}

type handler struct {
	fn          Handler
	maxAttempts int
}

// Queue is the job queue on the database.
type Queue struct {
	db       *db.DB
	log      *slog.Logger
	opts     Options
	now      func() time.Time
	mu       sync.Mutex
	handlers map[string]handler
	wake     chan struct{}
	wg       sync.WaitGroup
}

var _ schedule.Scheduler = (*Queue)(nil)

// New returns the queue in d.
func New(d *db.DB, log *slog.Logger, o Options) *Queue {
	if o.Workers <= 0 {
		o.Workers = 2
	}
	if o.Poll <= 0 {
		o.Poll = 2 * time.Second
	}
	if o.Lease <= 0 {
		o.Lease = 5 * time.Minute
	}
	if o.Keep <= 0 {
		o.Keep = 7 * 24 * time.Hour
	}
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Queue{db: d, log: log, opts: o, now: time.Now, handlers: map[string]handler{}, wake: make(chan struct{}, 1)}
}

// Handle registers the handler of a job type; maxAttempts (at least 1)
// bounds the retries.
func (q *Queue) Handle(typ string, maxAttempts int, fn Handler) {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	q.handlers[typ] = handler{fn: fn, maxAttempts: maxAttempts}
}

func (q *Queue) handlerFor(typ string) (handler, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	h, ok := q.handlers[typ]
	return h, ok
}

// Enqueue adds a job of a registered type.
func (q *Queue) Enqueue(ctx context.Context, typ, owner string, payload any) (Job, error) {
	h, ok := q.handlerFor(typ)
	if !ok {
		return Job{}, fmt.Errorf("jobs: no handler for %q", typ)
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return Job{}, fmt.Errorf("jobs: payload of %s: %w", typ, err)
	}
	u, err := uuid.NewV7()
	if err != nil {
		return Job{}, err
	}
	now := q.now().UnixMilli()
	if _, err := q.db.Write.ExecContext(ctx,
		`INSERT INTO jobs (id, type, owner, state, payload, max_attempts, run_after, created_at, updated_at)
		 VALUES (?, ?, ?, 'queued', ?, ?, ?, ?, ?)`,
		u.String(), typ, owner, string(body), h.maxAttempts, now, now, now); err != nil {
		return Job{}, fmt.Errorf("jobs: enqueue %s: %w", typ, err)
	}
	q.signal()
	return q.Get(ctx, u.String())
}

// enqueueUnique adds a job of typ unless one is queued or running (for
// periodic work); it reports whether it added one.
func (q *Queue) enqueueUnique(ctx context.Context, typ string) (bool, error) {
	var n int
	if err := q.db.Read.QueryRowContext(ctx,
		`SELECT count(*) FROM jobs WHERE type = ? AND state IN ('queued', 'running')`, typ).Scan(&n); err != nil {
		return false, err
	}
	if n > 0 {
		return false, nil
	}
	_, err := q.Enqueue(ctx, typ, "", struct{}{})
	return err == nil, err
}

func (q *Queue) signal() {
	select {
	case q.wake <- struct{}{}:
	default:
	}
}

const jobColumns = `id, type, owner, state, payload, result, progress_done, progress_total, attempts, max_attempts,
	cancel_requested, error, created_at, updated_at, finished_at`

func scanJob(row interface{ Scan(...any) error }) (Job, error) {
	var (
		j                Job
		payload          string
		result, errText  sql.NullString
		cancel           int
		created, updated int64
		finished         sql.NullInt64
	)
	err := row.Scan(&j.ID, &j.Type, &j.Owner, &j.State, &payload, &result, &j.Done, &j.Total, &j.Attempts, &j.MaxAttempts,
		&cancel, &errText, &created, &updated, &finished)
	if err != nil {
		return Job{}, err
	}
	j.Payload = json.RawMessage(payload)
	if result.Valid {
		j.Result = json.RawMessage(result.String)
	}
	j.CancelRequested = cancel == 1
	j.Error = errText.String
	j.CreatedAt, j.UpdatedAt = time.UnixMilli(created), time.UnixMilli(updated)
	if finished.Valid {
		j.FinishedAt = time.UnixMilli(finished.Int64)
	}
	return j, nil
}

// Get returns a job; ErrNotFound if there is none.
func (q *Queue) Get(ctx context.Context, id string) (Job, error) {
	j, err := scanJob(q.db.Read.QueryRowContext(ctx, `SELECT `+jobColumns+` FROM jobs WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return Job{}, ErrNotFound
	}
	if err != nil {
		return Job{}, fmt.Errorf("jobs: get %s: %w", id, err)
	}
	return j, nil
}

// List returns owner's most recent jobs, newest first, at most limit.
func (q *Queue) List(ctx context.Context, owner string, limit int) ([]Job, error) {
	rows, err := q.db.Read.QueryContext(ctx,
		`SELECT `+jobColumns+` FROM jobs WHERE owner = ? ORDER BY created_at DESC, id DESC LIMIT ?`, owner, limit)
	if err != nil {
		return nil, fmt.Errorf("jobs: list: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []Job
	for rows.Next() {
		j, err := scanJob(rows)
		if err != nil {
			return nil, fmt.Errorf("jobs: list: %w", err)
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

// Cancel asks a job to stop: a queued job is canceled at once, a running
// one when its handler next reports progress or returns.
func (q *Queue) Cancel(ctx context.Context, id string) (Job, error) {
	j, err := q.Get(ctx, id)
	if err != nil {
		return Job{}, err
	}
	if j.State.Finished() {
		return j, ErrFinished
	}
	now := q.now().UnixMilli()
	if _, err := q.db.Write.ExecContext(ctx,
		`UPDATE jobs SET cancel_requested = 1, updated_at = ?,
		   state = CASE WHEN state = 'queued' THEN 'canceled' ELSE state END,
		   finished_at = CASE WHEN state = 'queued' THEN ? ELSE finished_at END
		 WHERE id = ? AND state IN ('queued', 'running')`, now, now, id); err != nil {
		return Job{}, fmt.Errorf("jobs: cancel %s: %w", id, err)
	}
	return q.Get(ctx, id)
}

// Every implements schedule.Scheduler: task becomes the handler of a job
// type called name, and a job of that type is queued at once and then
// every interval unless one is still queued or running. Runs never
// overlap, and a run cut off by a restart runs again.
func (q *Queue) Every(ctx context.Context, name string, interval time.Duration, task func(context.Context)) {
	q.Handle(name, 1, func(ctx context.Context, _ Job, _ func(int64, int64)) (any, error) {
		task(ctx)
		return nil, nil
	})
	q.wg.Go(func() {
		tick := time.NewTicker(interval)
		defer tick.Stop()
		for {
			if _, err := q.enqueueUnique(ctx, name); err != nil && ctx.Err() == nil {
				q.log.WarnContext(ctx, "cannot queue a periodic job", "type", name, "error", err.Error())
			}
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
			}
		}
	})
}

// Start re-queues the jobs a stopped server left running, then starts the
// workers and the pruning of old finished jobs. They stop when ctx ends;
// Wait waits for them.
func (q *Queue) Start(ctx context.Context) error {
	now := q.now().UnixMilli()
	res, err := q.db.Write.ExecContext(ctx,
		`UPDATE jobs SET state = 'queued', lease_until = NULL, run_after = ?, updated_at = ? WHERE state = 'running'`, now, now)
	if err != nil {
		return fmt.Errorf("jobs: recover: %w", err)
	}
	if n, _ := res.RowsAffected(); n > 0 {
		q.log.InfoContext(ctx, "jobs resumed after a restart", "count", n)
	}
	for range q.opts.Workers {
		q.wg.Go(func() { q.work(ctx) })
	}
	q.Every(ctx, "jobs.prune", time.Hour, q.prune)
	return nil
}

// Wait waits until the workers and periodic tasks have stopped; call it
// after the context given to Start ended.
func (q *Queue) Wait() { q.wg.Wait() }

// prune removes finished jobs older than Keep.
func (q *Queue) prune(ctx context.Context) {
	cutoff := q.now().Add(-q.opts.Keep).UnixMilli()
	if _, err := q.db.Write.ExecContext(ctx,
		`DELETE FROM jobs WHERE state IN ('succeeded', 'failed', 'canceled') AND finished_at < ?`, cutoff); err != nil {
		q.log.WarnContext(ctx, "pruning old jobs failed", "error", err.Error())
	}
}

// work runs jobs until ctx ends.
func (q *Queue) work(ctx context.Context) {
	tick := time.NewTicker(q.opts.Poll)
	defer tick.Stop()
	for ctx.Err() == nil {
		j, ok, err := q.claim(ctx)
		if err != nil && ctx.Err() == nil {
			q.log.WarnContext(ctx, "cannot claim a job", "error", err.Error())
		}
		if ok {
			q.run(ctx, j)
			continue
		}
		select {
		case <-ctx.Done():
		case <-q.wake:
		case <-tick.C:
		}
	}
}

// claim takes the oldest due job. It reads first, so an idle queue writes
// nothing.
func (q *Queue) claim(ctx context.Context) (Job, bool, error) {
	now := q.now().UnixMilli()
	var id string
	err := q.db.Read.QueryRowContext(ctx,
		`SELECT id FROM jobs WHERE state = 'queued' AND run_after <= ? ORDER BY created_at, id LIMIT 1`, now).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return Job{}, false, nil
	}
	if err != nil {
		return Job{}, false, err
	}
	res, err := q.db.Write.ExecContext(ctx,
		`UPDATE jobs SET state = 'running', attempts = attempts + 1, lease_until = ?, updated_at = ?
		 WHERE id = ? AND state = 'queued'`, now+q.opts.Lease.Milliseconds(), now, id)
	if err != nil {
		return Job{}, false, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return Job{}, false, nil // another worker took it; look again
	}
	j, err := q.Get(ctx, id)
	return j, err == nil, err
}

// run runs one claimed job and records its outcome.
func (q *Queue) run(ctx context.Context, j Job) {
	h, ok := q.handlerFor(j.Type)
	if !ok {
		q.finish(ctx, j, nil, Permanent(fmt.Errorf("no handler for job type %q", j.Type)))
		return
	}
	jobCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	var (
		mu        sync.Mutex
		lastWrite time.Time
	)
	progress := func(done, total int64) {
		mu.Lock()
		defer mu.Unlock()
		if time.Since(lastWrite) < time.Second {
			return // at most one write a second per job (NFR-056)
		}
		lastWrite = time.Now()
		now := q.now().UnixMilli()
		_, _ = q.db.Write.ExecContext(ctx,
			`UPDATE jobs SET progress_done = ?, progress_total = ?, lease_until = ?, updated_at = ? WHERE id = ?`,
			done, total, now+q.opts.Lease.Milliseconds(), now, j.ID)
		var c int
		if err := q.db.Read.QueryRowContext(ctx, `SELECT cancel_requested FROM jobs WHERE id = ?`, j.ID).Scan(&c); err == nil && c == 1 {
			cancel()
		}
	}
	// A cancel is noticed within a second even while the handler reports
	// no progress (one large file): a read-only check, no writes.
	watchDone := make(chan struct{})
	go func() {
		t := time.NewTicker(time.Second)
		defer t.Stop()
		for {
			select {
			case <-watchDone:
				return
			case <-jobCtx.Done():
				return
			case <-t.C:
				var c int
				if err := q.db.Read.QueryRowContext(jobCtx, `SELECT cancel_requested FROM jobs WHERE id = ?`, j.ID).Scan(&c); err == nil && c == 1 {
					cancel()
					return
				}
			}
		}
	}()
	start := time.Now()
	result, err := h.fn(jobCtx, j, progress)
	close(watchDone)
	q.log.InfoContext(ctx, "job ran", "job", j.ID, "type", j.Type, "attempt", j.Attempts,
		"duration_ms", time.Since(start).Milliseconds(), "ok", err == nil)
	q.finish(ctx, j, result, err)
}

// finish records a job's outcome: done, failed, canceled, or queued again
// after a backoff.
func (q *Queue) finish(ctx context.Context, j Job, result any, runErr error) {
	now := q.now()
	ms := now.UnixMilli()
	var cancelRequested int
	_ = q.db.Read.QueryRowContext(ctx, `SELECT cancel_requested FROM jobs WHERE id = ?`, j.ID).Scan(&cancelRequested)
	var err error
	switch {
	case runErr == nil:
		body, merr := json.Marshal(result)
		if merr != nil || result == nil {
			body = nil
		}
		var res any
		if body != nil {
			res = string(body)
		}
		_, err = q.db.Write.ExecContext(ctx,
			`UPDATE jobs SET state = 'succeeded', result = ?, error = NULL, lease_until = NULL, updated_at = ?, finished_at = ?,
			   progress_done = CASE WHEN progress_total > 0 THEN progress_total ELSE progress_done END
			 WHERE id = ?`, res, ms, ms, j.ID)
	case cancelRequested == 1 || ctx.Err() == nil && errors.Is(runErr, context.Canceled):
		_, err = q.db.Write.ExecContext(ctx,
			`UPDATE jobs SET state = 'canceled', lease_until = NULL, updated_at = ?, finished_at = ? WHERE id = ?`, ms, ms, j.ID)
	case ctx.Err() != nil:
		return // the server is stopping: Start re-queues the job next time
	default:
		var p permanent
		if errors.As(runErr, &p) || j.Attempts >= j.MaxAttempts {
			_, err = q.db.Write.ExecContext(ctx,
				`UPDATE jobs SET state = 'failed', error = ?, lease_until = NULL, updated_at = ?, finished_at = ? WHERE id = ?`,
				runErr.Error(), ms, ms, j.ID)
			break
		}
		backoff := min(time.Duration(1<<min(j.Attempts, 8))*time.Second, 5*time.Minute)
		_, err = q.db.Write.ExecContext(ctx,
			`UPDATE jobs SET state = 'queued', error = ?, lease_until = NULL, run_after = ?, updated_at = ? WHERE id = ?`,
			runErr.Error(), now.Add(backoff).UnixMilli(), ms, j.ID)
	}
	if err != nil {
		q.log.ErrorContext(ctx, "cannot record a job's outcome", "job", j.ID, "error", err.Error())
	}
}
