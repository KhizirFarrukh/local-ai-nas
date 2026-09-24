package files

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
)

// Download implements Service. Only regular files are opened; a link is
// never followed. The size, time, ETag, and media type come from the open
// file, so they describe exactly the bytes that are read.
func (s *Local) Download(ctx context.Context, owner, apiPath string) (Item, io.ReadSeekCloser, error) {
	rel, err := s.resolveVisible(owner, apiPath)
	if err != nil {
		return Item{}, nil, err
	}
	type result struct {
		item Item
		f    *os.File
	}
	res, err := run(ctx, s.hooks, Event{Op: OpDownload, Owner: owner, Path: rel}, func() (result, error) {
		var r result
		err := s.withRoot(owner, func(root *os.Root) error {
			if err := refuseLinkParents(root, rel, apiPath); err != nil {
				return err
			}
			var err error
			r.f, r.item, err = openFile(root, owner, rel, apiPath)
			return err
		})
		return r, err
	})
	if err != nil {
		return Item{}, nil, err
	}
	// The file stays open after its os.Root is closed.
	return res.item, res.f, nil
}

// openFile opens the regular file rel. It checks the item without
// following links first, and then that the opened file is that same item,
// so a link or another file swapped in between is refused.
func openFile(root *os.Root, owner, rel, apiPath string) (*os.File, Item, error) {
	name := filepath.FromSlash(rel)
	info, err := root.Lstat(name)
	if err != nil {
		return nil, Item{}, fsError(err, apiPath)
	}
	if it := NewItem(owner, rel, info); it.Kind != KindFile {
		return nil, Item{}, apperr.Newf(apperr.InvalidRequest, "%s is not a file (it is a %s)", apiPath, it.Kind)
	}
	f, err := root.Open(name)
	if err != nil {
		return nil, Item{}, fsError(err, apiPath)
	}
	opened, err := f.Stat()
	switch {
	case err != nil:
		err = fsError(err, apiPath)
	case !os.SameFile(info, opened):
		err = apperr.Newf(apperr.Conflict, "%s changed while it was opened; try again", apiPath)
	}
	var it Item
	if err == nil {
		it, err = fileDetails(f, NewItem(owner, rel, opened), apiPath)
	}
	if err != nil {
		_ = f.Close() // read-only
		return nil, Item{}, err
	}
	return f, it, nil
}
