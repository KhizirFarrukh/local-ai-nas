package db

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

// BenchmarkCommit measures one small committed write, as the NAS makes
// for a session, a token, an audit event, or an item (S01.1-T12): with the
// FULL synchronous mode of Open, and with NORMAL for comparison. The
// results on each machine are recorded in the session log; the Raspberry
// Pi profile measures it again in S03.10-T03.
func BenchmarkCommit(b *testing.B) {
	ctx := context.Background()
	b.Run("full", func(b *testing.B) {
		d, err := Open(ctx, filepath.Join(b.TempDir(), FileName))
		if err != nil {
			b.Fatal(err)
		}
		defer func() { _ = d.Close() }()
		benchInserts(ctx, b, d.Write)
	})
	b.Run("normal", func(b *testing.B) {
		w, err := sql.Open("sqlite", dsn(filepath.Join(b.TempDir(), FileName), "_pragma=synchronous(NORMAL)"))
		if err != nil {
			b.Fatal(err)
		}
		w.SetMaxOpenConns(1)
		defer func() { _ = w.Close() }()
		benchInserts(ctx, b, w)
	})
}

func benchInserts(ctx context.Context, b *testing.B, w *sql.DB) {
	if _, err := w.ExecContext(ctx, `CREATE TABLE t (id INTEGER PRIMARY KEY, v TEXT) STRICT`); err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		if _, err := w.ExecContext(ctx, `INSERT INTO t (v) VALUES ('a small row, like a session or an audit event')`); err != nil {
			b.Fatal(err)
		}
	}
}
