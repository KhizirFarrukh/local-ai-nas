// Package files is the files area: the item model now, and the
// FilesService with its operations from S01.3. All file I/O in this package
// goes through the storage resolver and os.Root (an architecture test
// checks this).
package files

import (
	"io/fs"
	"path"
	"time"
)

// Kind is the type of an item.
type Kind string

// The item kinds. Symbolic links are listed but never followed (S01.6-T03).
const (
	KindFile    Kind = "file"
	KindDir     Kind = "dir"
	KindSymlink Kind = "symlink"
	KindOther   Kind = "other" // devices, sockets, pipes: listed, never opened
)

// Item is a file, folder, or link in a namespace of the files area. Every
// item carries its owner, the namespace it lives in (ADR-0003), so policy
// checks (S03) and sharing (S07) can rely on it.
type Item struct {
	OwnerID string    // the owner namespace, such as u0001
	Area    string    // the storage area: "files"
	RelPath string    // slash-separated path in the namespace; "." is the root
	Name    string    // the last path element; "" for the root
	Kind    Kind      // file, dir, symlink, or other
	Size    int64     // bytes, for files
	ModTime time.Time // last modification
	MIME    string    // media type, from S01.3-T03
	ETag    string    // content version, from S01.3-T03
}

// NewItem builds an Item from what Lstat returned for relPath in a
// namespace of the files area.
func NewItem(namespace, relPath string, info fs.FileInfo) Item {
	it := Item{
		OwnerID: namespace,
		Area:    "files",
		RelPath: relPath,
		ModTime: info.ModTime(),
	}
	if relPath != "." {
		it.Name = path.Base(relPath)
	}
	switch m := info.Mode(); {
	case m.IsRegular():
		it.Kind, it.Size = KindFile, info.Size()
	case m.IsDir():
		it.Kind = KindDir
	case m&fs.ModeSymlink != 0:
		it.Kind = KindSymlink
	default:
		it.Kind = KindOther
	}
	return it
}
