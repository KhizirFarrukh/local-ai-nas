package files

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
)

// UploadOptions configure Upload.
type UploadOptions struct {
	// OnConflict applies when the name is taken. Overwrite replaces a file
	// atomically and never replaces a folder.
	OnConflict OnConflict
}

// copyBufferSize is the buffer of a streamed write: large enough for fast
// sequential writes, small enough for many uploads at once.
const copyBufferSize = 1 << 20

// Upload writes the file at path (S01.3-T05). The bytes go to a temporary
// file in the target folder, which is synced and only then given its name
// in one step, so the name never shows a partial file (NFR-006). It
// reports whether it created a new file; false means that
// ConflictOverwrite replaced one.
//
// Everything that can be checked before the body is read is checked
// first: the path, the name rules, the free space for size bytes, the
// parent folder, and the conflict policy against the current target.
func (s *Local) Upload(ctx context.Context, owner, apiPath string, body io.Reader, size int64, o UploadOptions) (Item, bool, error) {
	policy, err := o.OnConflict.normalize()
	if err != nil {
		return Item{}, false, err
	}
	if size < 0 {
		return Item{}, false, apperr.Newf(apperr.InvalidRequest, "the upload size must be known, got %d", size)
	}
	rel, err := s.resolver.Resolve(storage.FilesArea, owner, apiPath)
	if err != nil {
		return Item{}, false, err
	}
	if rel == "." {
		return Item{}, false, apperr.NewRule(apperr.InvalidName, storage.RuleDotName, "the root folder cannot be replaced by a file")
	}
	if err := checkNewName(rel); err != nil {
		return Item{}, false, err
	}
	if s.space != nil {
		if err := s.space.Check(size); err != nil {
			return Item{}, false, err
		}
	}
	type result struct {
		item    Item
		created bool
	}
	res, err := run(ctx, s.hooks, Event{Op: OpUpload, Owner: owner, Path: rel}, func() (result, error) {
		var r result
		err := s.withRoot(owner, func(root *os.Root) error {
			dir := path.Dir(rel)
			if err := ensureParent(root, dir, false, apiPath); err != nil {
				return err
			}
			if err := checkTarget(root, rel, policy, apiPath); err != nil {
				return err
			}
			tmp, err := writeTemp(ctx, root, dir, body, size)
			if err != nil {
				return err
			}
			final, created, err := commitFile(root, tmp, rel, policy, apiPath)
			if err != nil {
				_ = root.Remove(filepath.FromSlash(tmp)) // hidden either way; best effort
				return err
			}
			syncFolder(root, dir)
			info, err := root.Lstat(filepath.FromSlash(final))
			if err != nil {
				return fsError(err, apiPath)
			}
			it, err := withDetails(root, NewItem(owner, final, info), "/"+final)
			r = result{it, created}
			return err
		})
		return r, err
	})
	return res.item, res.created, err
}

// checkTarget refuses, before the body is read, what the commit would
// refuse anyway: with ConflictFail an existing item, and with
// ConflictOverwrite anything that is not a regular file. The commit
// checks again, atomically.
func checkTarget(root *os.Root, rel string, policy OnConflict, apiPath string) error {
	if policy == ConflictRename {
		return nil
	}
	info, err := root.Lstat(filepath.FromSlash(rel))
	switch {
	case storage.IsNotFound(err):
		return nil
	case err != nil:
		return fsError(err, apiPath)
	case policy == ConflictFail:
		return apperr.Newf(apperr.Conflict, "an item already exists at %s", apiPath)
	case !info.Mode().IsRegular():
		return apperr.Newf(apperr.Conflict, "%s is not a file; only files can be overwritten", apiPath)
	}
	return nil
}

// writeTemp copies exactly size bytes from body into a new, synced
// temporary file in dir and returns its path. On any error the temporary
// file is removed.
func writeTemp(ctx context.Context, root *os.Root, dir string, body io.Reader, size int64) (tmp string, err error) {
	tmp = path.Join(dir, storage.TempPrefix+rand.Text()+".part")
	name := filepath.FromSlash(tmp)
	f, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		return "", apperr.Wrap(apperr.Internal, "cannot create the temporary file", err)
	}
	defer func() {
		if err != nil {
			_ = f.Close() // may already be closed
			_ = root.Remove(name)
		}
	}()
	if err := copyExactly(ctx, onlyWriter{f}, body, size); err != nil {
		return "", err
	}
	if err := f.Sync(); err != nil {
		return "", apperr.Wrap(apperr.Internal, "syncing the upload failed", err)
	}
	if err := f.Close(); err != nil {
		return "", apperr.Wrap(apperr.Internal, "closing the upload failed", err)
	}
	return tmp, nil
}

// onlyWriter hides the ReadFrom method of *os.File, so the copy uses its
// own fixed buffer and counts every byte.
type onlyWriter struct{ w io.Writer }

func (o onlyWriter) Write(p []byte) (int, error) { return o.w.Write(p) }

// copyExactly copies src to dst and requires exactly size bytes: a body
// that ends early (a dropped connection) or runs longer than declared is
// refused.
func copyExactly(ctx context.Context, dst io.Writer, src io.Reader, size int64) error {
	buf := make([]byte, min(size+1, copyBufferSize)) // at least 1 byte, to see a longer body
	var n int64
	for {
		if err := ctx.Err(); err != nil {
			return apperr.Wrap(apperr.InvalidRequest, "the upload was cancelled", err)
		}
		m, rerr := src.Read(buf)
		if m > 0 {
			if n+int64(m) > size {
				return apperr.Newf(apperr.InvalidRequest, "the body is longer than the declared %d bytes", size)
			}
			if _, err := dst.Write(buf[:m]); err != nil {
				return apperr.Wrap(apperr.Internal, "writing the upload failed", err)
			}
			n += int64(m)
		}
		if errors.Is(rerr, io.EOF) {
			break
		}
		if rerr != nil {
			return apperr.Wrap(apperr.InvalidRequest, fmt.Sprintf("reading the body failed after %d of %d bytes", n, size), rerr)
		}
	}
	if n < size {
		return apperr.Newf(apperr.InvalidRequest, "the body ended after %d of the declared %d bytes", n, size)
	}
	return nil
}

// commitFile gives the synced temporary file tmp its name under the
// conflict policy. It returns the name used and whether the file is new.
func commitFile(root *os.Root, tmp, rel string, policy OnConflict, apiPath string) (string, bool, error) {
	switch policy {
	case ConflictOverwrite:
		info, err := root.Lstat(filepath.FromSlash(rel))
		exists := err == nil
		switch {
		case err != nil && !storage.IsNotFound(err):
			return "", false, fsError(err, apiPath)
		case exists && !info.Mode().IsRegular():
			return "", false, apperr.Newf(apperr.Conflict, "%s is not a file; only files can be overwritten", apiPath)
		}
		// Rename replaces the old file in one step: a reader sees the old
		// content or the new, never a mix.
		if err := root.Rename(filepath.FromSlash(tmp), filepath.FromSlash(rel)); err != nil {
			return "", false, fsError(err, apiPath)
		}
		return rel, !exists, nil
	case ConflictRename:
		dir, base := path.Dir(rel), path.Base(rel)
		for n := 0; n <= maxRenameAttempts; n++ {
			candidate := rel
			if n > 0 {
				candidate = path.Join(dir, numberedName(base, n))
				if err := storage.ValidateName(path.Base(candidate)); err != nil {
					return "", false, err
				}
			}
			err := placeNew(root, tmp, candidate)
			if err == nil {
				return candidate, true, nil
			}
			if !errors.Is(err, fs.ErrExist) {
				return "", false, fsError(err, apiPath)
			}
		}
		return "", false, apperr.Newf(apperr.Conflict, "no free name for %s after %d attempts", apiPath, maxRenameAttempts)
	}
	if err := placeNew(root, tmp, rel); err != nil {
		return "", false, fsError(err, apiPath)
	}
	return rel, true, nil
}

// numberedName puts " (n)" before the extension: "a.txt" → "a (1).txt".
// A name that is all extension keeps it whole: ".env" → ".env (1)".
func numberedName(name string, n int) string {
	ext := path.Ext(name)
	stem := strings.TrimSuffix(name, ext)
	if stem == "" {
		stem, ext = name, ""
	}
	return fmt.Sprintf("%s (%d)%s", stem, n, ext)
}

// linkFile is os.Root.Link. Tests replace it to simulate a file system
// without hard links.
var linkFile = (*os.Root).Link

// placeNew moves the regular file from to name only if nothing has that
// name, and returns an fs.ErrExist error otherwise. A hard link fails
// atomically when the name exists, so concurrent writers cannot replace
// each other; the old name is then removed (if that fails, the new link
// is removed too, so the file is never left under both names). File
// systems without hard links (FAT, exFAT) fall back to a check and a
// rename, which a concurrent write to the same name can race.
func placeNew(root *os.Root, from, name string) error {
	src, dst := filepath.FromSlash(from), filepath.FromSlash(name)
	err := linkFile(root, src, dst)
	if err == nil {
		if err := root.Remove(src); err != nil {
			_ = root.Remove(dst) // undo; the file is still at from
			return err
		}
		return nil
	}
	if errors.Is(err, fs.ErrExist) {
		return err
	}
	switch _, serr := root.Lstat(dst); {
	case serr == nil:
		return &fs.PathError{Op: "rename", Path: name, Err: fs.ErrExist}
	case !storage.IsNotFound(serr):
		return serr
	}
	return root.Rename(src, dst)
}

// syncFolder makes a new name in dir durable, best effort: the file itself
// is already synced, and Windows cannot sync a folder.
func syncFolder(root *os.Root, dir string) {
	if runtime.GOOS == "windows" {
		return
	}
	d, err := root.Open(filepath.FromSlash(dir))
	if err != nil {
		return
	}
	_ = d.Sync()
	_ = d.Close()
}
