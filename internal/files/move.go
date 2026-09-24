package files

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
)

// MoveOptions configure Rename and Move.
type MoveOptions struct {
	// OnConflict applies when the target name is taken. Overwrite replaces
	// a file with a file; folders are never replaced or merged, and a
	// folder never replaces a file.
	OnConflict OnConflict
}

// Rename gives the item at path the name newName in the same folder
// (S01.3-T07).
func (s *Local) Rename(ctx context.Context, owner, apiPath, newName string, o MoveOptions) (Item, error) {
	if err := storage.ValidateName(newName); err != nil {
		return Item{}, err
	}
	rel, err := s.resolveVisible(owner, apiPath)
	if err != nil {
		return Item{}, err
	}
	// The target goes through the resolver like any path, so it is
	// normalized the same way.
	return s.move(ctx, OpRename, owner, apiPath, "/"+path.Join(path.Dir(rel), newName), o)
}

// Move moves the item at from to the path to, in the same namespace
// (S01.3-T07). The parent of to must exist.
func (s *Local) Move(ctx context.Context, owner, from, to string, o MoveOptions) (Item, error) {
	return s.move(ctx, OpMove, owner, from, to, o)
}

func (s *Local) move(ctx context.Context, op Op, owner, fromAPI, toAPI string, o MoveOptions) (Item, error) {
	policy, err := o.OnConflict.normalize()
	if err != nil {
		return Item{}, err
	}
	from, err := s.resolveVisible(owner, fromAPI)
	if err != nil {
		return Item{}, err
	}
	if from == "." {
		return Item{}, apperr.New(apperr.InvalidRequest, "the root folder cannot be moved or renamed")
	}
	to, err := s.resolver.Resolve(storage.FilesArea, owner, toAPI)
	if err != nil {
		return Item{}, err
	}
	if to == "." {
		return Item{}, apperr.NewRule(apperr.InvalidName, storage.RuleDotName, "the root folder cannot be replaced")
	}
	if err := checkNewName(to); err != nil {
		return Item{}, err
	}
	return run(ctx, s.hooks, Event{Op: op, Owner: owner, Path: from, Target: to}, func() (Item, error) {
		var it Item
		err := s.withRoot(owner, func(root *os.Root) error {
			src, err := root.Lstat(filepath.FromSlash(from))
			if err != nil {
				return fsError(err, fromAPI)
			}
			if err := ensureParent(root, path.Dir(to), false, toAPI); err != nil {
				return err
			}
			if src.IsDir() {
				if err := refuseIntoItself(root, src, path.Dir(to), fromAPI); err != nil {
					return err
				}
			}
			final, err := moveWithPolicy(root, from, to, src, policy, toAPI)
			if err != nil {
				return err
			}
			info, err := root.Lstat(filepath.FromSlash(final))
			if err != nil {
				return fsError(err, toAPI)
			}
			it, err = withDetails(root, NewItem(owner, final, info), "/"+final)
			return err
		})
		return it, err
	})
}

// refuseIntoItself refuses to move the folder src into dir when dir is
// src or lies below it. It compares each ancestor of dir with src as a
// file-system object, so a case variant of the path on a case-insensitive
// disk ("/Docs" for "/docs") cannot slip past.
func refuseIntoItself(root *os.Root, src fs.FileInfo, dir, fromAPI string) error {
	for {
		info, err := root.Lstat(filepath.FromSlash(dir))
		if err != nil {
			return fsError(err, "/"+dir)
		}
		if os.SameFile(info, src) {
			return apperr.Newf(apperr.InvalidRequest, "the folder %s cannot be moved into itself", fromAPI)
		}
		if dir == "." {
			return nil
		}
		dir = path.Dir(dir)
	}
}

// moveWithPolicy moves from to to under the conflict policy and returns
// the path used.
func moveWithPolicy(root *os.Root, from, to string, src fs.FileInfo, policy OnConflict, toAPI string) (string, error) {
	dst, err := root.Lstat(filepath.FromSlash(to))
	switch {
	case err == nil && os.SameFile(src, dst) && strings.EqualFold(from, to):
		// The item itself: the same path is nothing to do, and a
		// case-only change on a case-insensitive disk is a plain rename.
		if from == to {
			return from, nil
		}
		if err := root.Rename(filepath.FromSlash(from), filepath.FromSlash(to)); err != nil {
			return "", fsError(err, toAPI)
		}
		return to, nil
	case err != nil && !storage.IsNotFound(err):
		return "", fsError(err, toAPI)
	}
	exists := err == nil

	switch policy {
	case ConflictOverwrite:
		if exists {
			switch {
			case src.IsDir() && dst.IsDir():
				return "", apperr.Newf(apperr.Conflict, "a folder already exists at %s; folders are never replaced or merged", toAPI)
			case !src.Mode().IsRegular() || !dst.Mode().IsRegular():
				return "", apperr.Newf(apperr.Conflict, "an item already exists at %s; only a file can replace a file", toAPI)
			}
		}
		// Rename replaces the old file in one step.
		if err := root.Rename(filepath.FromSlash(from), filepath.FromSlash(to)); err != nil {
			return "", fsError(err, toAPI)
		}
		return to, nil
	case ConflictRename:
		dir, base := path.Dir(to), path.Base(to)
		for n := 0; n <= maxRenameAttempts; n++ {
			candidate := to
			if n > 0 {
				candidate = path.Join(dir, numbered(base, n, src.IsDir()))
				if err := storage.ValidateName(path.Base(candidate)); err != nil {
					return "", err
				}
			}
			err := moveNew(root, from, candidate, src)
			if err == nil {
				return candidate, nil
			}
			if !errors.Is(err, fs.ErrExist) {
				return "", fsError(err, toAPI)
			}
		}
		return "", apperr.Newf(apperr.Conflict, "no free name for %s after %d attempts", toAPI, maxRenameAttempts)
	}
	if err := moveNew(root, from, to, src); err != nil {
		return "", fsError(err, toAPI)
	}
	return to, nil
}

// numbered is the n-th free-name candidate: a file keeps its extension
// ("a (1).txt"); a folder name is taken whole ("v1.2 (1)"), as for new
// folders.
func numbered(base string, n int, dir bool) string {
	if dir {
		return fmt.Sprintf("%s (%d)", base, n)
	}
	return numberedName(base, n)
}

// moveNew moves from to name only if nothing has that name (fs.ErrExist
// otherwise). A regular file moves with placeNew, which is atomic on file
// systems with hard links. A folder or link is checked and then renamed:
// os.Root has no rename that refuses an existing target, so a concurrent
// write of the same name can race it.
func moveNew(root *os.Root, from, name string, src fs.FileInfo) error {
	if src.Mode().IsRegular() {
		return placeNew(root, from, name)
	}
	switch _, err := root.Lstat(filepath.FromSlash(name)); {
	case err == nil:
		return &fs.PathError{Op: "rename", Path: name, Err: fs.ErrExist}
	case !storage.IsNotFound(err):
		return err
	}
	return root.Rename(filepath.FromSlash(from), filepath.FromSlash(name))
}
