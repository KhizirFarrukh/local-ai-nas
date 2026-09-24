package files

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
)

// DeleteOptions configure Delete.
type DeleteOptions struct {
	// Recursive deletes a folder with everything in it. Without it only
	// an empty folder is deleted.
	Recursive bool
}

// Delete deletes the file, link, or folder at path permanently
// (S01.3-T09; a trash comes in S08). A link is deleted itself, never what
// it points to. A recursive delete is not atomic: if it fails part-way,
// the rest stays, and a retry finishes it.
func (s *Local) Delete(ctx context.Context, owner, apiPath string, o DeleteOptions) error {
	rel, err := s.resolveVisible(owner, apiPath)
	if err != nil {
		return err
	}
	if rel == "." {
		return apperr.New(apperr.InvalidRequest, "the root folder cannot be deleted")
	}
	_, err = run(ctx, s.hooks, Event{Op: OpDelete, Owner: owner, Path: rel}, func() (struct{}, error) {
		return struct{}{}, s.withRoot(owner, func(root *os.Root) error {
			name := filepath.FromSlash(rel)
			info, err := root.Lstat(name)
			if err != nil {
				return fsError(err, apiPath)
			}
			if info.IsDir() && o.Recursive {
				if err := root.RemoveAll(name); err != nil {
					return fsError(err, apiPath)
				}
				return nil
			}
			err = root.Remove(name)
			if err != nil && info.IsDir() && errors.Is(err, fs.ErrExist) {
				return notEmpty(root, rel, apiPath)
			}
			if err != nil {
				return fsError(err, apiPath)
			}
			return nil
		})
	})
	return err
}

// notEmpty explains why the folder rel was not deleted. A folder that
// holds only the server's temporary files looks empty to clients: an
// upload or copy into it has not finished (or was cut off by a crash).
func notEmpty(root *os.Root, rel, apiPath string) error {
	names, err := readNames(root, rel)
	if err == nil && len(names) == 0 {
		return apperr.Newf(apperr.Conflict, "%s has an unfinished upload or copy; try again later, or delete it with recursive", apiPath)
	}
	return apperr.Newf(apperr.Conflict, "%s is not empty; set recursive to delete it with everything in it", apiPath)
}
