// Package contenthash keeps the content hashes of files in the database
// (S01.3-T10, FR-211, ADR-0021). It implements files.Hashes on the
// content_hashes table.
package contenthash

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/KhizirFarrukh/local-ai-nas/internal/db"
	"github.com/KhizirFarrukh/local-ai-nas/internal/files"
)

// Store is the content_hashes table.
type Store struct {
	db  *db.DB
	now func() time.Time
}

var _ files.Hashes = (*Store)(nil)

// New returns the store in d.
func New(d *db.DB) *Store { return &Store{db: d, now: time.Now} }

// Record implements files.Hashes.
func (s *Store) Record(ctx context.Context, owner, rel, etag, hash string) error {
	_, err := s.db.Write.ExecContext(ctx,
		`INSERT INTO content_hashes (namespace, path, etag, hash, hashed_at) VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT (namespace, path) DO UPDATE SET etag = excluded.etag, hash = excluded.hash, hashed_at = excluded.hashed_at`,
		owner, rel, etag, hash, s.now().UnixMilli())
	if err != nil {
		return fmt.Errorf("contenthash: record %s/%s: %w", owner, rel, err)
	}
	return nil
}

// Lookup implements files.Hashes.
func (s *Store) Lookup(ctx context.Context, owner, rel, etag string) (string, error) {
	var hash string
	err := s.db.Read.QueryRowContext(ctx,
		`SELECT hash FROM content_hashes WHERE namespace = ? AND path = ? AND etag = ?`,
		owner, rel, etag).Scan(&hash)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("contenthash: look up %s/%s: %w", owner, rel, err)
	}
	return hash, nil
}

// Moved implements files.Hashes: the rows at from and below it move to
// to, after the rows at to and below it are dropped (what a move with
// overwrite replaced). One transaction, so a reader sees either state.
func (s *Store) Moved(ctx context.Context, owner, from, to string) error {
	if from == to {
		return nil
	}
	tx, err := s.db.Write.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("contenthash: move %s/%s: %w", owner, from, err)
	}
	defer func() { _ = tx.Rollback() }() // no-op after Commit
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM content_hashes WHERE namespace = ? AND `+atOrBelow, owner, to, to, to); err != nil {
		return fmt.Errorf("contenthash: move %s/%s: %w", owner, from, err)
	}
	// substr is 1-based and counts characters: the rest of the path after
	// "from" (empty for the item itself, "/..." for what is below it).
	if _, err := tx.ExecContext(ctx,
		`UPDATE content_hashes SET path = ? || substr(path, length(?) + 1)
		 WHERE namespace = ? AND `+atOrBelow, to, from, owner, from, from, from); err != nil {
		return fmt.Errorf("contenthash: move %s/%s: %w", owner, from, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("contenthash: move %s/%s: %w", owner, from, err)
	}
	return nil
}

// Deleted implements files.Hashes.
func (s *Store) Deleted(ctx context.Context, owner, rel string) error {
	var err error
	if rel == "." {
		_, err = s.db.Write.ExecContext(ctx, `DELETE FROM content_hashes WHERE namespace = ?`, owner)
	} else {
		_, err = s.db.Write.ExecContext(ctx,
			`DELETE FROM content_hashes WHERE namespace = ? AND `+atOrBelow, owner, rel, rel, rel)
	}
	if err != nil {
		return fmt.Errorf("contenthash: delete %s/%s: %w", owner, rel, err)
	}
	return nil
}

// atOrBelow matches the item at a path and everything inside it; it takes
// the path three times. It compares exactly (case-sensitive, no wildcards)
// in characters, as SQLite's substr and length count them.
const atOrBelow = `(path = ? OR substr(path, 1, length(?) + 1) = ? || '/')`
