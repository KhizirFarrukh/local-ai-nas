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

// OnConflict says what an operation does when its target already exists
// (docs/api/conventions.md).
type OnConflict string

// The conflict policies.
const (
	ConflictFail      OnConflict = "fail"
	ConflictRename    OnConflict = "rename"
	ConflictOverwrite OnConflict = "overwrite"
)

// maxRenameAttempts bounds the "name (n)" search of ConflictRename.
const maxRenameAttempts = 10000

func (c OnConflict) normalize() (OnConflict, error) {
	switch c {
	case "":
		return ConflictFail, nil
	case ConflictFail, ConflictRename, ConflictOverwrite:
		return c, nil
	}
	return "", apperr.Newf(apperr.InvalidRequest, "on_conflict must be fail, rename, or overwrite, got %q", c)
}

// FolderOptions configure CreateFolder.
type FolderOptions struct {
	// Parents creates missing parent folders too.
	Parents bool
	// OnConflict applies when the folder's name is taken. Overwrite never
	// replaces anything: an existing folder is returned as it is, and an
	// existing file is a conflict.
	OnConflict OnConflict
}

// CreateFolder creates the folder at path (S01.3-T04). It reports whether
// a folder was created; false means that ConflictOverwrite found the
// folder already there.
func (s *Local) CreateFolder(ctx context.Context, owner, apiPath string, o FolderOptions) (Item, bool, error) {
	policy, err := o.OnConflict.normalize()
	if err != nil {
		return Item{}, false, err
	}
	rel, err := s.resolver.Resolve(storage.FilesArea, owner, apiPath)
	if err != nil {
		return Item{}, false, err
	}
	if rel == "." {
		return Item{}, false, apperr.NewRule(apperr.InvalidName, storage.RuleDotName, "the root folder always exists")
	}
	if err := checkNewName(rel); err != nil {
		return Item{}, false, err
	}
	type result struct {
		item    Item
		created bool
	}
	res, err := run(ctx, s.hooks, Event{Op: OpCreateFolder, Owner: owner, Path: rel}, func() (result, error) {
		var r result
		err := s.withRoot(owner, func(root *os.Root) error {
			if err := refuseLinkParents(root, rel, apiPath); err != nil {
				return err
			}
			parent := path.Dir(rel)
			if err := ensureParent(root, parent, o.Parents, apiPath); err != nil {
				return err
			}
			made, created, err := mkdirWithPolicy(root, rel, policy, apiPath)
			if err != nil {
				return err
			}
			info, err := root.Lstat(filepath.FromSlash(made))
			if err != nil {
				return fsError(err, apiPath)
			}
			r = result{NewItem(owner, made, info), created}
			return nil
		})
		return r, err
	})
	return res.item, res.created, err
}

// checkNewName applies the name rules to the last element of rel and the
// path-length rule to all of it. Parents that already exist are not
// checked: they may have been created by other means.
func checkNewName(rel string) error {
	if err := storage.ValidateName(path.Base(rel)); err != nil {
		return err
	}
	if len(rel) > storage.MaxPathBytes {
		return apperr.NewRule(apperr.InvalidName, storage.RulePathTooLong,
			fmt.Sprintf("a path must be at most %d bytes, this one has %d", storage.MaxPathBytes, len(rel)))
	}
	return nil
}

// ensureParent checks that the parent folder exists, or creates the
// missing folders when parents is set (each new name must pass the name
// rules).
func ensureParent(root *os.Root, parent string, parents bool, apiPath string) error {
	if parent == "." {
		return nil
	}
	if !parents {
		info, err := root.Lstat(filepath.FromSlash(parent))
		switch {
		case storage.IsNotFound(err):
			return apperr.Newf(apperr.NotFound, "the parent folder of %s does not exist", apiPath)
		case err != nil:
			return fsError(err, apiPath)
		case !info.IsDir():
			return apperr.Newf(apperr.Conflict, "the parent of %s is not a folder", apiPath)
		}
		return nil
	}
	// First find the missing parents and check their names, so a refused
	// name creates nothing; then create them.
	comps := strings.Split(parent, "/")
	missing := len(comps)
	cur := ""
	for i, name := range comps {
		cur = path.Join(cur, name)
		info, err := root.Lstat(filepath.FromSlash(cur))
		if storage.IsNotFound(err) {
			missing = i
			break
		}
		if err != nil {
			return fsError(err, apiPath)
		}
		if !info.IsDir() {
			return apperr.Newf(apperr.Conflict, "/%s is not a folder", cur)
		}
	}
	for _, name := range comps[missing:] {
		if err := storage.ValidateName(name); err != nil {
			return err
		}
	}
	cur = path.Join(comps[:missing]...)
	for _, name := range comps[missing:] {
		cur = path.Join(cur, name)
		if err := root.Mkdir(filepath.FromSlash(cur), 0o750); err != nil && !errors.Is(err, fs.ErrExist) {
			return fsError(err, apiPath)
		}
	}
	return nil
}

// mkdirWithPolicy creates rel, applying the conflict policy when the name
// is taken. It returns the path it created (or found) and whether it
// created it.
func mkdirWithPolicy(root *os.Root, rel string, policy OnConflict, apiPath string) (string, bool, error) {
	err := root.Mkdir(filepath.FromSlash(rel), 0o750)
	if err == nil {
		return rel, true, nil
	}
	if !errors.Is(err, fs.ErrExist) {
		return "", false, fsError(err, apiPath)
	}
	switch policy {
	case ConflictOverwrite:
		info, err := root.Lstat(filepath.FromSlash(rel))
		if err != nil {
			return "", false, fsError(err, apiPath)
		}
		if !info.IsDir() {
			return "", false, apperr.Newf(apperr.Conflict, "a file already exists at %s; folders never replace files", apiPath)
		}
		return rel, false, nil
	case ConflictRename:
		dir, base := path.Dir(rel), path.Base(rel)
		for n := 1; n <= maxRenameAttempts; n++ {
			candidate := path.Join(dir, fmt.Sprintf("%s (%d)", base, n))
			if err := storage.ValidateName(path.Base(candidate)); err != nil {
				return "", false, err
			}
			err := root.Mkdir(filepath.FromSlash(candidate), 0o750)
			if err == nil {
				return candidate, true, nil
			}
			if !errors.Is(err, fs.ErrExist) {
				return "", false, fsError(err, apiPath)
			}
		}
		return "", false, apperr.Newf(apperr.Conflict, "no free name for %s after %d attempts", apiPath, maxRenameAttempts)
	}
	return "", false, apperr.Newf(apperr.Conflict, "an item already exists at %s", apiPath)
}
