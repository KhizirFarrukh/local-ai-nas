// Package items keeps the stable ID of every item in the items table
// (S01.3-T11, ADR-0040, FR-346). It implements files.Registry for one area.
// IDs are UUIDv7 (RFC 9562): time-ordered, so the table's index stays
// compact. An ID is an identifier, never a permission (I5).
package items

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/google/uuid"

	"github.com/KhizirFarrukh/local-ai-nas/internal/db"
	"github.com/KhizirFarrukh/local-ai-nas/internal/files"
)

// The areas (the items.area column).
const (
	AreaFiles  = "files"
	AreaPhotos = "photos"
)

// Store is the items table for one area.
type Store struct {
	db    *db.DB
	area  string
	now   func() time.Time
	newID func() (string, error)
}

var _ files.Registry = (*Store)(nil)

// New returns the store of area in d.
func New(d *db.DB, area string) *Store {
	return &Store{db: d, area: area, now: time.Now, newID: newV7}
}

func newV7() (string, error) {
	u, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// execer is what ensure needs from a transaction.
type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// ensure inserts the item at rel if no present item has that path, and
// returns the ID of the present item.
func (s *Store) ensure(ctx context.Context, x execer, owner, rel string, kind files.Kind, now int64) (string, error) {
	id, err := s.newID()
	if err != nil {
		return "", err
	}
	if _, err := x.ExecContext(ctx,
		`INSERT INTO items (id, namespace, area, path, kind, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, 'present', ?, ?)
		 ON CONFLICT (namespace, area, path) WHERE status = 'present' DO NOTHING`,
		id, owner, s.area, rel, kindOf(kind), now, now); err != nil {
		return "", err
	}
	var got string
	err = x.QueryRowContext(ctx,
		`SELECT id FROM items WHERE namespace = ? AND area = ? AND path = ? AND status = 'present'`,
		owner, s.area, rel).Scan(&got)
	return got, err
}

// kindOf maps a files kind to the items.kind column; only files and
// folders get IDs (the app never creates links or devices).
func kindOf(k files.Kind) string {
	if k == files.KindDir {
		return "dir"
	}
	return "file"
}

// ensureWithParents ensures every folder above rel, then rel.
func (s *Store) ensureWithParents(ctx context.Context, x execer, owner, rel string, kind files.Kind, now int64) (string, error) {
	var parents []string
	for p := path.Dir(rel); p != "." && p != "/" && p != ""; p = path.Dir(p) {
		parents = append(parents, p)
	}
	for i := len(parents) - 1; i >= 0; i-- {
		if _, err := s.ensure(ctx, x, owner, parents[i], files.KindDir, now); err != nil {
			return "", err
		}
	}
	return s.ensure(ctx, x, owner, rel, kind, now)
}

// Ensure implements files.Registry.
func (s *Store) Ensure(ctx context.Context, owner, rel string, kind files.Kind) (string, error) {
	if rel == "." || rel == "" {
		return "", nil // the namespace root is not an item
	}
	var id string
	err := s.inTx(ctx, func(tx *sql.Tx, now int64) error {
		var err error
		id, err = s.ensureWithParents(ctx, tx, owner, rel, kind, now)
		return err
	})
	if err != nil {
		return "", fmt.Errorf("items: ensure %s/%s: %w", owner, rel, err)
	}
	return id, nil
}

// EnsureAll implements files.Registry.
func (s *Store) EnsureAll(ctx context.Context, owner string, list []files.RegistryEntry) error {
	err := s.inTx(ctx, func(tx *sql.Tx, now int64) error {
		for _, e := range list {
			if e.Rel == "." || e.Rel == "" {
				continue
			}
			if _, err := s.ensureWithParents(ctx, tx, owner, e.Rel, e.Kind, now); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("items: ensure %d items in %s: %w", len(list), owner, err)
	}
	return nil
}

// Moved implements files.Registry.
func (s *Store) Moved(ctx context.Context, owner, from, to string) error {
	if from == to {
		return nil
	}
	err := s.inTx(ctx, func(tx *sql.Tx, now int64) error {
		// What the move replaced is gone.
		if _, err := tx.ExecContext(ctx,
			`UPDATE items SET status = 'retired', updated_at = ?
			 WHERE namespace = ? AND area = ? AND status = 'present' AND `+atOrBelow,
			now, owner, s.area, to, to, to); err != nil {
			return err
		}
		// substr is 1-based and counts characters: the rest of the path
		// after from (empty for the item itself, "/..." below it).
		_, err := tx.ExecContext(ctx,
			`UPDATE items SET path = ? || substr(path, length(?) + 1), updated_at = ?
			 WHERE namespace = ? AND area = ? AND status = 'present' AND `+atOrBelow,
			to, from, now, owner, s.area, from, from, from)
		return err
	})
	if err != nil {
		return fmt.Errorf("items: move %s/%s: %w", owner, from, err)
	}
	return nil
}

// Deleted implements files.Registry. The rows stay, retired, so references
// and the audit trail can still name the item.
func (s *Store) Deleted(ctx context.Context, owner, rel string) error {
	now := s.now().UnixMilli()
	var err error
	if rel == "." {
		_, err = s.db.Write.ExecContext(ctx,
			`UPDATE items SET status = 'retired', updated_at = ? WHERE namespace = ? AND area = ? AND status = 'present'`,
			now, owner, s.area)
	} else {
		_, err = s.db.Write.ExecContext(ctx,
			`UPDATE items SET status = 'retired', updated_at = ?
			 WHERE namespace = ? AND area = ? AND status = 'present' AND `+atOrBelow,
			now, owner, s.area, rel, rel, rel)
	}
	if err != nil {
		return fmt.Errorf("items: delete %s/%s: %w", owner, rel, err)
	}
	return nil
}

// Lookup implements files.Registry.
func (s *Store) Lookup(ctx context.Context, owner, rel string) (string, error) {
	var id string
	err := s.db.Read.QueryRowContext(ctx,
		`SELECT id FROM items WHERE namespace = ? AND area = ? AND path = ? AND status = 'present'`,
		owner, s.area, rel).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("items: look up %s/%s: %w", owner, rel, err)
	}
	return id, nil
}

// inTx runs fn in one write transaction (BEGIN IMMEDIATE, db.Open).
func (s *Store) inTx(ctx context.Context, fn func(tx *sql.Tx, now int64) error) error {
	tx, err := s.db.Write.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }() // no-op after Commit
	if err := fn(tx, s.now().UnixMilli()); err != nil {
		return err
	}
	return tx.Commit()
}

// atOrBelow matches the item at a path and everything inside it; it takes
// the path three times (as in internal/contenthash).
const atOrBelow = `(path = ? OR substr(path, 1, length(?) + 1) = ? || '/')`
