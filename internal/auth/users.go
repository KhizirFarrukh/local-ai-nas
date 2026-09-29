// Package auth holds the accounts of the NAS: users and their passwords
// (S03.2), sessions and API tokens (S03.3). It stores them in the SQLite
// database (ADR-0007, ADR-0010) and never keeps a secret in plain form:
// passwords as Argon2id hashes, session and token secrets as SHA-256.
package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/db"
	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
)

// Role is what a user may do. S03 has one account, the admin; the user
// role exists so S07 adds users without changing the tables (ADR-0010).
type Role string

// The roles.
const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

// User is an account.
type User struct {
	ID       int64
	Username string // lowercase ASCII (NormalizeUsername)
	// Namespace owns the user's areas, such as u0001 (ADR-0003).
	Namespace         string
	Role              Role
	PasswordHash      string // Argon2id, PHC string format
	PasswordChangedAt time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	// DisabledAt is zero while the account is active.
	DisabledAt time.Time
}

// Disabled reports whether the account is disabled.
func (u User) Disabled() bool { return !u.DisabledAt.IsZero() }

// Errors of the user store.
var (
	// ErrNotFound means no user matches.
	ErrNotFound = errors.New("auth: no such user")
	// ErrUsersExist means first-run setup is over: an account exists.
	ErrUsersExist = errors.New("auth: an account exists already")
)

// maxUsername is the longest username, in bytes (ASCII only).
const maxUsername = 64

// NormalizeUsername checks a username and returns its stored form: 1 to 64
// characters from a-z, 0-9, ".", "_", and "-", starting with a letter or a
// digit. Upper-case letters are accepted and stored in lower case, so
// "Admin" and "admin" are the same account.
func NormalizeUsername(name string) (string, error) {
	n := strings.ToLower(strings.TrimSpace(name))
	switch {
	case n == "":
		return "", apperr.New(apperr.InvalidRequest, "the username is empty")
	case len(n) > maxUsername:
		return "", apperr.Newf(apperr.InvalidRequest, "the username is longer than %d characters", maxUsername)
	case !isAlnum(n[0]):
		return "", apperr.New(apperr.InvalidRequest, "the username must start with a letter or a digit")
	}
	for i := range len(n) {
		if c := n[i]; !isAlnum(c) && c != '.' && c != '_' && c != '-' {
			return "", apperr.New(apperr.InvalidRequest, "the username may contain only the letters a to z, digits, and . _ -")
		}
	}
	return n, nil
}

func isAlnum(c byte) bool { return c >= 'a' && c <= 'z' || c >= '0' && c <= '9' }

// Users is the users table.
type Users struct {
	db  *db.DB
	now func() time.Time
}

// NewUsers returns the user store in d.
func NewUsers(d *db.DB) *Users { return &Users{db: d, now: time.Now} }

// Count returns the number of accounts.
func (s *Users) Count(ctx context.Context) (int, error) {
	var n int
	if err := s.db.Read.QueryRowContext(ctx, `SELECT count(*) FROM users`).Scan(&n); err != nil {
		return 0, fmt.Errorf("auth: count users: %w", err)
	}
	return n, nil
}

// CreateFirstAdmin creates the first account, an admin that owns the S01
// namespace (storage.DefaultNamespace), so no file moves (ADR-0003). The
// check that no account exists and the insert are one write transaction,
// so two setups at once cannot create two admins (threat T-02). It returns
// ErrUsersExist when an account exists. username must be normalized and
// passwordHash an Argon2id PHC string.
func (s *Users) CreateFirstAdmin(ctx context.Context, username, passwordHash string) (User, error) {
	tx, err := s.db.Write.BeginTx(ctx, nil) // BEGIN IMMEDIATE (db.Open): takes the write lock now
	if err != nil {
		return User{}, fmt.Errorf("auth: create the first admin: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // no-op after Commit
	var n int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM users`).Scan(&n); err != nil {
		return User{}, fmt.Errorf("auth: create the first admin: %w", err)
	}
	if n > 0 {
		return User{}, ErrUsersExist
	}
	now := s.now().UnixMilli()
	res, err := tx.ExecContext(ctx,
		`INSERT INTO users (username, namespace, role, password_hash, password_changed_at, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		username, storage.DefaultNamespace, RoleAdmin, passwordHash, now, now, now)
	if err != nil {
		return User{}, fmt.Errorf("auth: create the first admin: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return User{}, fmt.Errorf("auth: create the first admin: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return User{}, fmt.Errorf("auth: create the first admin: %w", err)
	}
	t := time.UnixMilli(now)
	return User{
		ID: id, Username: username, Namespace: storage.DefaultNamespace, Role: RoleAdmin,
		PasswordHash: passwordHash, PasswordChangedAt: t, CreatedAt: t, UpdatedAt: t,
	}, nil
}

const userColumns = `id, username, namespace, role, password_hash, password_changed_at, created_at, updated_at, disabled_at`

// ByUsername returns the account with the given username, which is
// normalized first; ErrNotFound if there is none.
func (s *Users) ByUsername(ctx context.Context, username string) (User, error) {
	n, err := NormalizeUsername(username)
	if err != nil {
		return User{}, ErrNotFound // an invalid name names no account
	}
	return s.one(ctx, `SELECT `+userColumns+` FROM users WHERE username = ?`, n)
}

// ByID returns the account with the given ID; ErrNotFound if there is none.
func (s *Users) ByID(ctx context.Context, id int64) (User, error) {
	return s.one(ctx, `SELECT `+userColumns+` FROM users WHERE id = ?`, id)
}

func (s *Users) one(ctx context.Context, query string, arg any) (User, error) {
	var (
		u                         User
		changed, created, updated int64
		disabled                  sql.NullInt64
	)
	err := s.db.Read.QueryRowContext(ctx, query, arg).Scan(
		&u.ID, &u.Username, &u.Namespace, &u.Role, &u.PasswordHash, &changed, &created, &updated, &disabled)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("auth: look up user: %w", err)
	}
	u.PasswordChangedAt = time.UnixMilli(changed)
	u.CreatedAt = time.UnixMilli(created)
	u.UpdatedAt = time.UnixMilli(updated)
	if disabled.Valid {
		u.DisabledAt = time.UnixMilli(disabled.Int64)
	}
	return u, nil
}

// SetPasswordHash replaces a user's password hash. When changed is true
// the password itself changed (password_changed_at moves); a rehash of the
// same password with new parameters passes false. ErrNotFound if the user
// does not exist.
func (s *Users) SetPasswordHash(ctx context.Context, id int64, passwordHash string, changed bool) error {
	now := s.now().UnixMilli()
	query := `UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?`
	args := []any{passwordHash, now, id}
	if changed {
		query = `UPDATE users SET password_hash = ?, updated_at = ?, password_changed_at = ? WHERE id = ?`
		args = []any{passwordHash, now, now, id}
	}
	res, err := s.db.Write.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("auth: set password: %w", err)
	}
	if n, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("auth: set password: %w", err)
	} else if n == 0 {
		return ErrNotFound
	}
	return nil
}
