package db

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

func openTest(t *testing.T) *DB {
	t.Helper()
	path := filepath.Join(testutil.StorageRoot(t), ".local-ai-nas", "db", FileName)
	d, err := Open(t.Context(), path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() {
		if err := d.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	})
	return d
}

func pragma(t *testing.T, db *sql.DB, name string) string {
	t.Helper()
	var v string
	if err := db.QueryRowContext(t.Context(), "PRAGMA "+name).Scan(&v); err != nil {
		t.Fatalf("PRAGMA %s: %v", name, err)
	}
	return v
}

func TestOpenPragmas(t *testing.T) {
	d := openTest(t)
	if _, err := os.Stat(d.Path()); err != nil {
		t.Fatalf("database file: %v", err)
	}
	want := map[string]string{"journal_mode": "wal", "foreign_keys": "1", "busy_timeout": "5000", "synchronous": "1"}
	for name, pool := range map[string]*sql.DB{"writer": d.Write, "reader": d.Read} {
		for p, w := range want {
			if got := pragma(t, pool, p); !strings.EqualFold(got, w) {
				t.Errorf("%s: PRAGMA %s = %q, want %q", name, p, got, w)
			}
		}
	}
	if got := pragma(t, d.Read, "query_only"); got != "1" {
		t.Errorf("reader: query_only = %q, want 1", got)
	}
	if got := pragma(t, d.Write, "query_only"); got != "0" {
		t.Errorf("writer: query_only = %q, want 0", got)
	}
}

func TestReadDuringWriteTransaction(t *testing.T) {
	d := openTest(t)
	ctx := t.Context()
	if _, err := d.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	insert := "INSERT INTO settings (key, value, updated_at) VALUES (?, ?, ?)"
	if _, err := d.Write.ExecContext(ctx, insert, "a", "1", time.Now().UnixMilli()); err != nil {
		t.Fatal(err)
	}

	tx, err := d.Write.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(ctx, insert, "b", "2", time.Now().UnixMilli()); err != nil {
		t.Fatal(err)
	}

	// The write transaction is open. A reader must not block or fail, and
	// it sees the last committed state.
	readCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	var n int
	if err := d.Read.QueryRowContext(readCtx, "SELECT count(*) FROM settings").Scan(&n); err != nil {
		t.Fatalf("read during the write transaction: %v", err)
	}
	if n != 1 {
		t.Errorf("reader saw %d rows during the transaction, want 1 (the committed state)", n)
	}

	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := d.Read.QueryRowContext(ctx, "SELECT count(*) FROM settings").Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Errorf("reader saw %d rows after the commit, want 2", n)
	}
}

func TestReaderCannotWrite(t *testing.T) {
	d := openTest(t)
	if _, err := d.Migrate(t.Context()); err != nil {
		t.Fatal(err)
	}
	_, err := d.Read.ExecContext(t.Context(), "INSERT INTO settings (key, value, updated_at) VALUES ('x', 'y', 0)")
	if err == nil {
		t.Error("a write through the read pool succeeded")
	}
}

func TestMigrateIdempotent(t *testing.T) {
	d := openTest(t)
	ctx := t.Context()

	first, err := d.Migrate(ctx)
	if err != nil {
		t.Fatalf("first Migrate: %v", err)
	}
	if len(first) != 1 || first[0].Source.Version != 1 {
		t.Fatalf("first Migrate applied %d migrations, want version 1 only", len(first))
	}
	second, err := d.Migrate(ctx)
	if err != nil {
		t.Fatalf("second Migrate: %v", err)
	}
	if len(second) != 0 {
		t.Errorf("second Migrate applied %d migrations, want 0", len(second))
	}

	st, err := d.MigrationStatus(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(st) != 1 || st[0].State != "applied" {
		t.Errorf("status = %d migrations, first state %q; want 1 applied", len(st), st[0].State)
	}

	// The schema exists and its constraints hold.
	for _, table := range []string{"settings", "uploads"} {
		var name string
		if err := d.Read.QueryRowContext(ctx, "SELECT name FROM sqlite_schema WHERE type = 'table' AND name = ?", table).Scan(&name); err != nil {
			t.Errorf("table %s: %v", table, err)
		}
	}
	_, err = d.Write.ExecContext(ctx, `INSERT INTO uploads (id, namespace, target_path, on_conflict, declared_size, created_at, expires_at)
		VALUES ('u1', 'u0001', 'a.bin', 'sometimes', 1, 0, 0)`)
	if err == nil {
		t.Error("an invalid on_conflict value was accepted")
	}
	_, err = d.Write.ExecContext(ctx, `INSERT INTO settings (key, value, updated_at) VALUES ('k', 'v', 'not a number')`)
	if err == nil {
		t.Error("STRICT table accepted text in an INTEGER column")
	}
}

func TestMigrationStatusBeforeMigrate(t *testing.T) {
	d := openTest(t)
	st, err := d.MigrationStatus(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(st) != 1 || st[0].State != "pending" {
		t.Errorf("status before Migrate: %d migrations, want 1 pending", len(st))
	}
}

func TestReopenKeepsData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db", FileName)
	ctx := t.Context()
	d, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Write.ExecContext(ctx, "INSERT INTO settings (key, value, updated_at) VALUES ('k', 'v', 0)"); err != nil {
		t.Fatal(err)
	}
	if err := d.Close(); err != nil {
		t.Fatal(err)
	}

	d, err = Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = d.Close() }()
	if res, err := d.Migrate(ctx); err != nil || len(res) != 0 {
		t.Errorf("Migrate after reopen = %d, %v; want nothing to do", len(res), err)
	}
	var v string
	if err := d.Read.QueryRowContext(ctx, "SELECT value FROM settings WHERE key = 'k'").Scan(&v); err != nil || v != "v" {
		t.Errorf("value after reopen = %q, %v", v, err)
	}
}

func TestOpenErrors(t *testing.T) {
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(t.Context(), filepath.Join(file, "db", FileName)); err == nil {
		t.Error("Open under a regular file succeeded")
	}
	notDB := filepath.Join(t.TempDir(), "garbage.db")
	if err := os.WriteFile(notDB, []byte(strings.Repeat("this is not an SQLite database ", 200)), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(t.Context(), notDB); err == nil {
		t.Error("Open of a non-database file succeeded")
	}
}
