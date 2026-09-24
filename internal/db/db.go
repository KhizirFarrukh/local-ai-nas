// Package db opens the SQLite database in WAL mode and applies the embedded
// goose migrations (S01.1-T10, ADR-0007). The driver is modernc.org/sqlite,
// which is pure Go, so builds need no C compiler (CGO_ENABLED=0).
package db

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite" // registers the "sqlite" database/sql driver
)

// FileName is the database file name inside the db directory.
const FileName = "nas.db"

// readers is the size of the read connection pool.
const readers = 4

//go:embed migrations/*.sql
var migrationFiles embed.FS

// pragmas are set on every connection. WAL lets readers work while one
// writer writes; busy_timeout makes a connection wait for a lock instead of
// failing at once; NORMAL synchronous is safe with WAL (a power cut can
// lose the last transactions, never corrupt the file).
var pragmas = []string{
	"journal_mode(WAL)",
	"foreign_keys(1)",
	"busy_timeout(5000)",
	"synchronous(NORMAL)",
}

// DB is the application database: one writer connection, because SQLite
// allows one writer at a time, and a pool of read-only connections.
type DB struct {
	// Write is for every statement that changes data. It has one connection
	// and starts transactions with BEGIN IMMEDIATE, so a transaction that
	// will write takes the write lock at once instead of failing later.
	Write *sql.DB
	// Read is for queries. Its connections are query-only.
	Read *sql.DB

	path string
}

// Open opens the database at path, creating the file and its directory if
// needed, and checks that WAL mode is active. It does not migrate; call
// Migrate for that.
func Open(ctx context.Context, path string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return nil, fmt.Errorf("db: create directory: %w", err)
	}
	w, err := sql.Open("sqlite", dsn(path, "_txlock=immediate"))
	if err != nil {
		return nil, fmt.Errorf("db: open writer: %w", err)
	}
	w.SetMaxOpenConns(1)
	d := &DB{Write: w, path: path}

	var mode string
	if err := w.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&mode); err != nil {
		_ = w.Close()
		return nil, fmt.Errorf("db: open %s: %w", path, err)
	}
	if !strings.EqualFold(mode, "wal") {
		_ = w.Close()
		return nil, fmt.Errorf("db: %s: journal mode is %q, want wal", path, mode)
	}

	r, err := sql.Open("sqlite", dsn(path, "_pragma=query_only(1)"))
	if err != nil {
		_ = w.Close()
		return nil, fmt.Errorf("db: open readers: %w", err)
	}
	r.SetMaxOpenConns(readers)
	d.Read = r
	if err := r.PingContext(ctx); err != nil {
		_ = d.Close()
		return nil, fmt.Errorf("db: open readers: %w", err)
	}
	return d, nil
}

// dsn builds the modernc.org/sqlite data source name: the file path, the
// pragmas for every connection, and extra parameters.
func dsn(path string, extra ...string) string {
	q := make([]string, 0, len(pragmas)+len(extra))
	for _, p := range pragmas {
		q = append(q, "_pragma="+url.QueryEscape(p))
	}
	q = append(q, extra...)
	return path + "?" + strings.Join(q, "&")
}

// Path returns the database file path.
func (d *DB) Path() string { return d.path }

// Close closes both pools.
func (d *DB) Close() error {
	var errs []error
	if d.Read != nil {
		errs = append(errs, d.Read.Close())
	}
	if d.Write != nil {
		errs = append(errs, d.Write.Close())
	}
	return errors.Join(errs...)
}

func (d *DB) provider() (*goose.Provider, error) {
	sub, err := fs.Sub(migrationFiles, "migrations")
	if err != nil {
		return nil, err
	}
	return goose.NewProvider(goose.DialectSQLite3, d.Write, sub)
}

// Migrate applies every pending migration and returns the ones it applied.
// Running it again when nothing is pending does nothing.
func (d *DB) Migrate(ctx context.Context) ([]*goose.MigrationResult, error) {
	p, err := d.provider()
	if err != nil {
		return nil, fmt.Errorf("db: migrations: %w", err)
	}
	res, err := p.Up(ctx)
	if err != nil {
		return res, fmt.Errorf("db: migrate: %w", err)
	}
	return res, nil
}

// MigrationStatus reports every known migration and whether it is applied.
func (d *DB) MigrationStatus(ctx context.Context) ([]*goose.MigrationStatus, error) {
	p, err := d.provider()
	if err != nil {
		return nil, fmt.Errorf("db: migrations: %w", err)
	}
	st, err := p.Status(ctx)
	if err != nil {
		return nil, fmt.Errorf("db: migration status: %w", err)
	}
	return st, nil
}
