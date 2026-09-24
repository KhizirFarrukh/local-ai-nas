package files

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
)

// CleanTemp removes the server's temporary files and folders
// (storage.TempPrefix) whose last change is older than olderThan from the
// owner's namespace (S01.4-T06). They are left behind when the server
// stops in the middle of an upload or a copy; they are never items, so
// nobody sees them, but they use space. Newer ones may belong to a write
// in progress and are kept. It returns how many it removed.
func (s *Local) CleanTemp(ctx context.Context, owner string, olderThan time.Duration, now time.Time) (int, error) {
	removed := 0
	err := s.withRoot(owner, func(root *os.Root) error {
		return fs.WalkDir(root.FS(), ".", func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				if storage.IsNotFound(err) {
					return nil // removed while walking
				}
				return err
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			if !storage.IsTempName(d.Name()) {
				return nil
			}
			skip := error(nil)
			if d.IsDir() {
				skip = fs.SkipDir // a temporary folder is removed or kept as a whole
			}
			info, err := d.Info()
			if err != nil || now.Sub(info.ModTime()) < olderThan {
				return skip
			}
			if err := root.RemoveAll(filepath.FromSlash(p)); err == nil {
				removed++
			}
			return skip
		})
	})
	return removed, err
}
