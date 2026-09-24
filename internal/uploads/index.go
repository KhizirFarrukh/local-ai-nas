package uploads

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/KhizirFarrukh/local-ai-nas/internal/db"
)

// Session is one upload that was created and has not finished or expired
// yet: a row of the uploads table (S01.1-T10 migration).
type Session struct {
	ID           string
	Namespace    string // the owner, such as u0001
	TargetPath   string // the API path of the file to create, such as /docs/a.bin
	OnConflict   string // fail, rename, or overwrite
	DeclaredSize int64
	SHA256       string // the expected checksum in lowercase hex, or ""
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

// ErrNoSession means that the index has no session with the ID.
var ErrNoSession = errors.New("uploads: no such upload session")

// Index is the uploads table.
type Index struct{ db *db.DB }

// NewIndex returns the index in d.
func NewIndex(d *db.DB) Index { return Index{db: d} }

// Add records a new session.
func (x Index) Add(ctx context.Context, s Session) error {
	var sum any
	if s.SHA256 != "" {
		sum = s.SHA256
	}
	_, err := x.db.Write.ExecContext(ctx,
		`INSERT INTO uploads (id, namespace, target_path, on_conflict, declared_size, sha256, created_at, expires_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		s.ID, s.Namespace, s.TargetPath, s.OnConflict, s.DeclaredSize, sum, s.CreatedAt.UnixMilli(), s.ExpiresAt.UnixMilli())
	if err != nil {
		return fmt.Errorf("uploads: record session %s: %w", s.ID, err)
	}
	return nil
}

// Get returns the session with the ID, or ErrNoSession.
func (x Index) Get(ctx context.Context, id string) (Session, error) {
	var s Session
	var sum sql.NullString
	var created, expires int64
	err := x.db.Read.QueryRowContext(ctx,
		`SELECT id, namespace, target_path, on_conflict, declared_size, sha256, created_at, expires_at
		 FROM uploads WHERE id = ?`, id).
		Scan(&s.ID, &s.Namespace, &s.TargetPath, &s.OnConflict, &s.DeclaredSize, &sum, &created, &expires)
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, ErrNoSession
	}
	if err != nil {
		return Session{}, fmt.Errorf("uploads: read session %s: %w", id, err)
	}
	s.SHA256 = sum.String
	s.CreatedAt, s.ExpiresAt = time.UnixMilli(created).UTC(), time.UnixMilli(expires).UTC()
	return s, nil
}

// Remove deletes the session with the ID; a missing session is not an
// error.
func (x Index) Remove(ctx context.Context, id string) error {
	if _, err := x.db.Write.ExecContext(ctx, `DELETE FROM uploads WHERE id = ?`, id); err != nil {
		return fmt.Errorf("uploads: remove session %s: %w", id, err)
	}
	return nil
}
