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
	"time"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
)

// CopyLimits bound a copy that runs within one request (S01.3-T08). A
// larger copy is refused with too_large_for_sync until background jobs
// exist (S04.3). Zero means no limit.
type CopyLimits struct {
	MaxItems int   // files and folders, the copied item included
	MaxBytes int64 // the sum of the file sizes
}

// CopyOptions configure Copy.
type CopyOptions struct {
	// OnConflict applies to the target name. Overwrite replaces a file
	// with a file; folders are never replaced or merged.
	OnConflict OnConflict
}

// copyEntry is one item of the source tree, relative to the copied item
// ("" is the item itself). Parents come before their children.
type copyEntry struct {
	rel  string
	info fs.FileInfo // from Lstat
}

// Copy copies the file or folder at from to the path to (S01.3-T08) and
// reports whether it created a new item (false: it replaced a file). The
// whole source is scanned before anything is written: the limits, the
// name rules for every new name, and the free space. The copy stays
// hidden until it is complete: a file is written to a temporary file, a
// folder is built under a temporary name, and either is then given its
// name in one step.
func (s *Local) Copy(ctx context.Context, owner, fromAPI, toAPI string, o CopyOptions) (Item, bool, error) {
	policy, err := o.OnConflict.normalize()
	if err != nil {
		return Item{}, false, err
	}
	from, err := s.resolveVisible(owner, fromAPI)
	if err != nil {
		return Item{}, false, err
	}
	to, err := s.resolver.Resolve(storage.FilesArea, owner, toAPI)
	if err != nil {
		return Item{}, false, err
	}
	if to == "." {
		return Item{}, false, apperr.NewRule(apperr.InvalidName, storage.RuleDotName, "the root folder cannot be replaced")
	}
	if err := checkNewName(to); err != nil {
		return Item{}, false, err
	}
	type result struct {
		item    Item
		created bool
	}
	res, err := run(ctx, s.hooks, Event{Op: OpCopy, Owner: owner, Path: from, Target: to}, func() (result, error) {
		var r result
		err := s.withRoot(owner, func(root *os.Root) error {
			if err := refuseLinkParents(root, from, fromAPI); err != nil {
				return err
			}
			if err := refuseLinkParents(root, to, toAPI); err != nil {
				return err
			}
			src, err := root.Lstat(filepath.FromSlash(from))
			if err != nil {
				return fsError(err, fromAPI)
			}
			if err := ensureParent(root, path.Dir(to), false, toAPI); err != nil {
				return err
			}
			if src.IsDir() {
				if err := refuseIntoItself(root, src, path.Dir(to), fromAPI, "copied"); err != nil {
					return err
				}
			}
			entries, size, err := scanCopy(root, from, to, src, s.copyLimits)
			if err != nil {
				return err
			}
			if s.space != nil {
				if err := s.space.Check(size); err != nil {
					return err
				}
			}
			lock := func() func() { return s.lockFolder(owner, path.Dir(to)) }
			var final string
			if src.IsDir() {
				final, err = copyFolder(ctx, root, from, to, entries, policy, toAPI, lock)
				r.created = true
			} else {
				final, r.created, err = copyOneFile(ctx, root, from, to, src, policy, toAPI, lock)
			}
			if err != nil {
				return err
			}
			info, err := root.Lstat(filepath.FromSlash(final))
			if err != nil {
				return fsError(err, toAPI)
			}
			r.item, err = withDetails(root, NewItem(owner, final, info), "/"+final)
			return err
		})
		return r, err
	})
	return res.item, res.created, err
}

// scanCopy lists the source tree without following links and checks it
// against the limits and the name rules. It returns the entries and the
// total size of the files.
func scanCopy(root *os.Root, from, to string, src fs.FileInfo, limits CopyLimits) ([]copyEntry, int64, error) {
	var entries []copyEntry
	var total int64
	var walk func(rel string, info fs.FileInfo) error
	walk = func(rel string, info fs.FileInfo) error {
		srcPath := path.Join(from, rel)
		switch {
		case info.Mode().IsRegular():
			total += info.Size()
		case !info.IsDir():
			return apperr.Newf(apperr.InvalidRequest, "/%s is a %s; only files and folders can be copied (links are never followed)",
				srcPath, NewItem("", srcPath, info).Kind)
		}
		entries = append(entries, copyEntry{rel, info})
		switch {
		case limits.MaxItems > 0 && len(entries) > limits.MaxItems:
			return apperr.Newf(apperr.TooLargeForSync, "the copy has more than %d files and folders, the most one request copies", limits.MaxItems)
		case limits.MaxBytes > 0 && total > limits.MaxBytes:
			return apperr.Newf(apperr.TooLargeForSync, "the copy has more than %d bytes, the most one request copies", limits.MaxBytes)
		}
		if rel != "" {
			if err := checkCopiedName(path.Join(to, rel), srcPath); err != nil {
				return err
			}
		}
		if !info.IsDir() {
			return nil
		}
		children, err := readInfos(root, srcPath)
		if err != nil {
			return err
		}
		for _, info := range children {
			if err := walk(path.Join(rel, info.Name()), info); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk("", src); err != nil {
		return nil, 0, err
	}
	return entries, total, nil
}

// readNames returns the names in the folder dir, in order, without the
// server's temporary files.
func readNames(root *os.Root, dir string) ([]string, error) {
	f, err := root.Open(filepath.FromSlash(dir))
	if err != nil {
		return nil, fsError(err, "/"+dir)
	}
	defer func() { _ = f.Close() }() // read-only
	entries, err := f.ReadDir(-1)
	if err != nil {
		return nil, fsError(err, "/"+dir)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !storage.IsTempName(e.Name()) {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// readInfos returns the items in the folder dir as the folder read reports
// them, in name order, without the server's temporary files; links are
// described, never followed. One folder read costs far less than an Lstat
// per item, which on Windows took about 1 ms each (bug S02-B09).
func readInfos(root *os.Root, dir string) ([]fs.FileInfo, error) {
	f, err := root.Open(filepath.FromSlash(dir))
	if err != nil {
		return nil, fsError(err, "/"+dir)
	}
	defer func() { _ = f.Close() }() // read-only
	entries, err := f.ReadDir(-1)
	if err != nil {
		return nil, fsError(err, "/"+dir)
	}
	infos := make([]fs.FileInfo, 0, len(entries))
	for _, e := range entries {
		if storage.IsTempName(e.Name()) {
			continue
		}
		info, err := e.Info() // Lstat: links are never followed
		if storage.IsNotFound(err) {
			continue // removed since the folder was read
		}
		if err != nil {
			return nil, fsError(err, "/"+path.Join(dir, e.Name()))
		}
		infos = append(infos, info)
	}
	return infos, nil
}

// checkCopiedName applies the name rules to a name the copy creates below
// the target, naming the source item in the error.
func checkCopiedName(dst, srcPath string) error {
	err := checkNewName(dst)
	var e *apperr.Error
	if errors.As(err, &e) {
		return apperr.NewRule(e.Kind, e.Rule, fmt.Sprintf("/%s cannot be copied: %s", srcPath, e.Detail))
	}
	return err
}

// copyOneFile copies the regular file from to a temporary file next to
// to, and commits it like an upload, holding lock during the commit.
func copyOneFile(ctx context.Context, root *os.Root, from, to string, src fs.FileInfo, policy OnConflict, toAPI string, lock func() func()) (string, bool, error) {
	if err := checkTarget(root, to, policy, toAPI); err != nil {
		return "", false, err
	}
	tmp := path.Join(path.Dir(to), storage.TempPrefix+rand.Text()+".part")
	if err := copyFileTo(ctx, root, from, tmp, src, make([]byte, copyBufferSize)); err != nil {
		_ = root.Remove(filepath.FromSlash(tmp))
		return "", false, err
	}
	unlock := lock()
	final, created, err := commitFile(root, tmp, to, policy, toAPI)
	unlock()
	if err != nil {
		_ = root.Remove(filepath.FromSlash(tmp))
		return "", false, err
	}
	syncFolder(root, path.Dir(to))
	return final, created, nil
}

// copyFolder builds the copy of the tree under a temporary name next to
// to, then gives it its name, holding lock during the commit. On any error
// the temporary tree is removed.
func copyFolder(ctx context.Context, root *os.Root, from, to string, entries []copyEntry, policy OnConflict, toAPI string, lock func() func()) (string, error) {
	switch _, err := root.Lstat(filepath.FromSlash(to)); {
	case err == nil && policy != ConflictRename:
		return "", apperr.Newf(apperr.Conflict, "an item already exists at %s; folders are never replaced or merged", toAPI)
	case err != nil && !storage.IsNotFound(err):
		return "", fsError(err, toAPI)
	}
	tmp := path.Join(path.Dir(to), storage.TempPrefix+rand.Text()+".part")
	if err := buildTree(ctx, root, from, tmp, entries); err != nil {
		_ = root.RemoveAll(filepath.FromSlash(tmp))
		return "", err
	}
	unlock := lock()
	final, err := commitFolder(root, tmp, to, entries[0].info, policy, toAPI)
	unlock()
	if err != nil {
		_ = root.RemoveAll(filepath.FromSlash(tmp))
		return "", err
	}
	syncFolder(root, path.Dir(to))
	return final, nil
}

// buildTree creates the entries below dst: folders, then each file
// synced, then the folders' modification times, deepest first, because
// adding entries changes them.
func buildTree(ctx context.Context, root *os.Root, from, dst string, entries []copyEntry) error {
	buf := make([]byte, copyBufferSize)
	for _, e := range entries {
		target := path.Join(dst, e.rel)
		if !e.info.IsDir() {
			if err := copyFileTo(ctx, root, path.Join(from, e.rel), target, e.info, buf); err != nil {
				return err
			}
			continue
		}
		if err := root.Mkdir(filepath.FromSlash(target), 0o750); err != nil {
			return fsError(err, "/"+target)
		}
	}
	for i := len(entries) - 1; i >= 0; i-- {
		if e := entries[i]; e.info.IsDir() {
			if err := root.Chtimes(filepath.FromSlash(path.Join(dst, e.rel)), time.Time{}, e.info.ModTime()); err != nil {
				return fsError(err, "/"+path.Join(dst, e.rel))
			}
		}
	}
	return nil
}

// commitFolder gives the complete temporary tree tmp its name. Folders
// have no atomic rename that refuses an existing target, so this is a
// check and a rename (see moveNew).
func commitFolder(root *os.Root, tmp, to string, info fs.FileInfo, policy OnConflict, toAPI string) (string, error) {
	if policy == ConflictRename {
		dir, base := path.Dir(to), path.Base(to)
		for n := 0; n <= maxRenameAttempts; n++ {
			candidate := to
			if n > 0 {
				candidate = path.Join(dir, numbered(base, n, true))
				if err := storage.ValidateName(path.Base(candidate)); err != nil {
					return "", err
				}
			}
			err := moveNew(root, tmp, candidate, info)
			if err == nil {
				return candidate, nil
			}
			if !errors.Is(err, fs.ErrExist) {
				return "", fsError(err, toAPI)
			}
		}
		return "", apperr.Newf(apperr.Conflict, "no free name for %s after %d attempts", toAPI, maxRenameAttempts)
	}
	if err := moveNew(root, tmp, to, info); err != nil {
		if errors.Is(err, fs.ErrExist) {
			return "", apperr.Newf(apperr.Conflict, "an item already exists at %s; folders are never replaced or merged", toAPI)
		}
		return "", fsError(err, toAPI)
	}
	return to, nil
}

// copyFileTo copies the regular file from to the new file dst, synced,
// with the source's modification time. The opened source must be the file
// that was scanned (info), so a link swapped in is never followed.
func copyFileTo(ctx context.Context, root *os.Root, from, dst string, info fs.FileInfo, buf []byte) (err error) {
	in, err := root.Open(filepath.FromSlash(from))
	if err != nil {
		return fsError(err, "/"+from)
	}
	defer func() { _ = in.Close() }() // read-only
	if opened, err := in.Stat(); err != nil || !os.SameFile(info, opened) {
		return apperr.Newf(apperr.Conflict, "/%s changed during the copy; try again", from)
	}
	out, err := root.OpenFile(filepath.FromSlash(dst), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		return fsError(err, "/"+dst)
	}
	defer func() {
		if cerr := out.Close(); cerr != nil && err == nil {
			err = apperr.Wrap(apperr.Internal, "closing the copy failed", cerr)
		}
	}()
	if err := copyStream(ctx, out, in, buf); err != nil {
		return err
	}
	if err := out.Sync(); err != nil {
		return apperr.Wrap(apperr.Internal, "syncing the copy failed", err)
	}
	if err := root.Chtimes(filepath.FromSlash(dst), time.Time{}, info.ModTime()); err != nil {
		return fsError(err, "/"+dst)
	}
	return nil
}

// copyStream copies src to dst through buf, stopping when ctx is done.
func copyStream(ctx context.Context, dst io.Writer, src io.Reader, buf []byte) error {
	for {
		if err := ctx.Err(); err != nil {
			return apperr.Wrap(apperr.InvalidRequest, "the copy was cancelled", err)
		}
		n, rerr := src.Read(buf)
		if n > 0 {
			if _, err := dst.Write(buf[:n]); err != nil {
				return apperr.Wrap(apperr.Internal, "writing the copy failed", err)
			}
		}
		if errors.Is(rerr, io.EOF) {
			return nil
		}
		if rerr != nil {
			return apperr.Wrap(apperr.Internal, "reading the source failed", rerr)
		}
	}
}
