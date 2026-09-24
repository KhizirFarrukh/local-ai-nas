package uploads

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// clock is a settable time source.
type clock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *clock) set(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = t
}

// TestCleanup is the S01.4-T06 acceptance test: expired, idle uploads are
// removed with their session; active ones stay, and so does an expired
// upload that is still being written. Orphaned files go too.
func TestCleanup(t *testing.T) {
	start := time.Now()
	c := &clock{now: start}
	e := newTestEnv(t, nil, c.Now)
	create := func(target string) string {
		t.Helper()
		loc := createUpload(t, e.url, []byte("0123456789"), MetaTargetPath, target)
		return loc[strings.LastIndex(loc, "/")+1:]
	}
	idle := create("/idle.bin")     // created first, then left alone
	moving := create("/moving.bin") // expired, but its data was written just now
	c.set(start.Add(20 * time.Hour))
	fresh := create("/fresh.bin") // not expired yet

	// Files without a session: old ones go, new ones stay.
	oldOrphan := filepath.Join(e.dir, "orphan-old")
	newOrphan := filepath.Join(e.dir, "orphan-new")
	for _, p := range []string{oldOrphan, oldOrphan + ".info", newOrphan} {
		if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	// A day and a bit later. The data files' times are the real "now"; age
	// all of them but the moving one.
	later := start.Add(25 * time.Hour)
	c.set(later)
	past := later.Add(-25 * time.Hour)
	for _, p := range []string{idle, idle + ".info", "orphan-old", "orphan-old.info"} {
		if err := os.Chtimes(filepath.Join(e.dir, p), time.Time{}, past); err != nil {
			t.Fatal(err)
		}
	}
	for _, p := range []string{filepath.Join(e.dir, moving), newOrphan} {
		if err := os.Chtimes(p, time.Time{}, later.Add(-time.Minute)); err != nil {
			t.Fatal(err)
		}
	}

	n, err := e.server.Cleanup(t.Context())
	if err != nil || n != 2 {
		t.Errorf("Cleanup removed %d (%v), want 2 (the idle upload and the old orphan)", n, err)
	}
	for id, want := range map[string]bool{idle: false, moving: true, fresh: true} {
		_, err := e.server.index.Get(t.Context(), id)
		if got := !errors.Is(err, ErrNoSession); got != want {
			t.Errorf("session %s kept: %v, want %v", id, got, want)
		}
		_, err = os.Stat(filepath.Join(e.dir, id))
		if got := err == nil; got != want {
			t.Errorf("data of %s kept: %v, want %v", id, got, want)
		}
	}
	for p, want := range map[string]bool{oldOrphan: false, oldOrphan + ".info": false, newOrphan: true} {
		if _, err := os.Stat(p); (err == nil) != want {
			t.Errorf("%s kept: %v, want %v", filepath.Base(p), err == nil, want)
		}
	}

	// The removed upload is gone for its client too.
	resp, err := tusRequest(t, t.Context(), http.MethodHead, e.url+basePath+idle, nil, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("HEAD of the removed upload: %d, want 404", resp.StatusCode)
	}
}
