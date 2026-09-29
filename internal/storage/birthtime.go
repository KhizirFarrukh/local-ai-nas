package storage

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"time"
)

// BirthTimeAt is BirthTime for the item at rel, a slash-separated path in
// root ("." for the root itself), with info from Lstat of it.
func BirthTimeAt(root *os.Root, rel string, info fs.FileInfo) (time.Time, bool) {
	if !birthNeedsDir {
		return BirthTime(nil, "", info)
	}
	dir, name := path.Split(rel)
	if dir == "" {
		dir = "."
	}
	d, err := root.Open(filepath.FromSlash(dir))
	if err != nil {
		return time.Time{}, false
	}
	defer func() { _ = d.Close() }() // read-only
	return BirthTime(d, name, info)
}
