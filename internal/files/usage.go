package files

import (
	"context"
	"os"
	"path"
	"path/filepath"
)

// Usage is how much a folder holds, at any depth (FR-214).
type Usage struct {
	Item    Item  // the folder (or file) asked about
	Size    int64 // the total size of the files, in bytes
	Files   int   // how many files
	Folders int   // how many folders below it
}

// Usage adds up the files in the folder at path, at any depth. Links are
// neither followed nor counted, and neither are other special files. For
// a file, it is the file's own size. The walk stops when ctx ends, such as
// when the client goes away (S02.3-T05).
func (s *Local) Usage(ctx context.Context, owner, path string) (Usage, error) {
	rel, err := s.resolveVisible(owner, path)
	if err != nil {
		return Usage{}, err
	}
	return run(ctx, s.hooks, Event{Op: OpUsage, Owner: owner, Path: rel}, func() (Usage, error) {
		var u Usage
		err := s.withRoot(owner, func(root *os.Root) error {
			if err := refuseLinkParents(root, rel, path); err != nil {
				return err
			}
			info, err := root.Lstat(filepath.FromSlash(rel))
			if err != nil {
				return fsError(err, path)
			}
			u.Item = NewItem(owner, rel, info)
			stampAdded(root, &u.Item, info)
			switch u.Item.Kind {
			case KindFile:
				u.Size, u.Files = u.Item.Size, 1
			case KindDir:
				return addUp(ctx, root, rel, &u)
			}
			return nil
		})
		return u, err
	})
}

// addUp walks the folder rel, one folder at a time, into u.
func addUp(ctx context.Context, root *os.Root, rel string, u *Usage) error {
	pending := []string{rel}
	for len(pending) > 0 {
		if err := ctx.Err(); err != nil {
			return err // the client went away
		}
		dir := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		infos, err := readInfos(root, dir)
		if err != nil {
			return err
		}
		for _, info := range infos {
			switch m := info.Mode(); {
			case m.IsRegular():
				u.Size += info.Size()
				u.Files++
			case m.IsDir():
				u.Folders++
				pending = append(pending, path.Join(dir, info.Name()))
			}
		}
	}
	return nil
}
