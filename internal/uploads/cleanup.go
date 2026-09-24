package uploads

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// Cleanup removes abandoned uploads (S01.4-T06): sessions past their
// expiry whose data has not been written to for the expiry time (an
// upload that is still moving is kept), and files in the upload directory
// without a session that are that old (left by a crash). It returns how
// many uploads it removed.
func (s *Server) Cleanup(ctx context.Context) (int, error) {
	now := s.o.Now()
	expired, err := s.index.Expired(ctx, now)
	if err != nil {
		return 0, err
	}
	removed := 0
	for _, sess := range expired {
		data := filepath.Join(s.o.Dir, sess.ID)
		if info, err := os.Stat(data); err == nil && now.Sub(info.ModTime()) < s.o.Expiry {
			continue // still being written
		}
		s.removeFiles(ctx, sess.ID)
		if err := s.index.Remove(ctx, sess.ID); err != nil {
			return removed, err
		}
		removed++
	}

	// Files without a session: a crash between tusd storing and the index
	// (or the other way round) leaves them.
	entries, err := os.ReadDir(s.o.Dir)
	if err != nil {
		return removed, err
	}
	orphans := map[string]bool{}
	for _, e := range entries {
		id := strings.TrimSuffix(strings.TrimSuffix(e.Name(), ".info"), ".lock")
		if orphans[id] {
			continue
		}
		info, err := e.Info()
		if err != nil || now.Sub(info.ModTime()) < s.o.Expiry {
			continue
		}
		if _, err := s.index.Get(ctx, id); !errors.Is(err, ErrNoSession) {
			continue // a session owns it (or the index failed: keep it)
		}
		orphans[id] = true
	}
	for id := range orphans {
		s.removeFiles(ctx, id)
		removed++
	}
	return removed, nil
}

// removeFiles deletes tusd's files of one upload.
func (s *Server) removeFiles(ctx context.Context, id string) {
	base := filepath.Join(s.o.Dir, id)
	for _, p := range []string{base, base + ".info", base + ".lock"} {
		if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
			s.o.Logger.WarnContext(ctx, "removing an abandoned upload failed", "upload", id, "error", err.Error())
		}
	}
}
