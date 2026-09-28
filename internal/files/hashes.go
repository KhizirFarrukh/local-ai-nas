package files

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"path"
)

// HashPrefix starts every content hash: the algorithm, then the digest in
// lowercase hex (ADR-0021). A later algorithm gets its own prefix, so old
// and new hashes never mix.
const HashPrefix = "sha256:"

// Hashes keeps the content hash of files (FR-211, S01.3-T10) with the ETag
// of the version it was computed from. A hash counts only while the file
// still has that ETag (size, modification time, and file ID), so a file
// changed outside the app is never reported with a stale hash. Rename and
// move keep a file's ETag, so its hash moves with it.
//
// Every method is best effort for the files service: an operation that
// succeeded is not undone because its hash could not be kept, and a
// missing hash is made again by the backfill (S11.1).
type Hashes interface {
	// Record stores hash for the file at rel in owner's namespace, whose
	// version is etag, replacing what rel had.
	Record(ctx context.Context, owner, rel, etag, hash string) error
	// Lookup returns the hash of the file at rel if it was recorded for
	// the version etag, and "" otherwise.
	Lookup(ctx context.Context, owner, rel, etag string) (string, error)
	// Moved follows a rename or move of the file or folder from to to;
	// whatever to had before is forgotten.
	Moved(ctx context.Context, owner, from, to string) error
	// Deleted forgets the file or folder at rel and everything in it.
	Deleted(ctx context.Context, owner, rel string) error
}

// NopHashes keeps nothing; the service then reports no hashes.
type NopHashes struct{}

// Record implements Hashes.
func (NopHashes) Record(context.Context, string, string, string, string) error { return nil }

// Lookup implements Hashes.
func (NopHashes) Lookup(context.Context, string, string, string) (string, error) { return "", nil }

// Moved implements Hashes.
func (NopHashes) Moved(context.Context, string, string, string) error { return nil }

// Deleted implements Hashes.
func (NopHashes) Deleted(context.Context, string, string) error { return nil }

// newHash returns the hash the service computes while bytes stream.
func newHash() hash.Hash { return sha256.New() }

// hashString formats a finished hash with its prefix.
func hashString(h hash.Hash) string { return HashPrefix + hex.EncodeToString(h.Sum(nil)) }

// hashedFile is a file whose content hash was computed while it was
// written: its path (relative to a copy's root while the copy is built)
// and its version.
type hashedFile struct {
	rel, etag, hash string
}

// recordAll stores the hashes of files written below base; failures are
// left to the backfill (see Hashes).
func (s *Local) recordAll(ctx context.Context, owner, base string, list []hashedFile) {
	for _, f := range list {
		if f.etag == "" {
			continue // the copy's version could not be read
		}
		_ = s.hashes.Record(ctx, owner, path.Join(base, f.rel), f.etag, f.hash) // best effort
	}
}
