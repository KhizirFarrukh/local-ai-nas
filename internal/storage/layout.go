// Package storage owns the storage root: its directory layout (ADR-0003),
// and later the namespace resolver, the free-space guard, and the path
// locks (S01.2, S01.6).
package storage

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
)

// Names in the storage root (ADR-0003). The two areas never overlap (I1),
// and internal data lives outside both (I2).
const (
	FilesArea   = "files"
	PhotosArea  = "photos"
	InternalDir = ".local-ai-nas"
	// DefaultNamespace is the single owner's namespace in S01. S03.2 binds
	// the first admin to it; S07 adds u0002 and so on.
	DefaultNamespace = "u0001"
)

// Layout is the set of directories under a storage root.
type Layout struct {
	// Root is the absolute path of the storage root.
	Root string
	// Internal is the internal data directory, <root>/.local-ai-nas.
	Internal string
	// TmpUploads holds uploads in progress. It is always on the root's
	// filesystem, so a finished upload can be renamed into place.
	TmpUploads string
	// DB holds the SQLite database.
	DB string
	// Logs holds the log files.
	Logs string
}

// NewLayout returns the default layout for the storage root.
func NewLayout(root string) Layout {
	internal := filepath.Join(root, InternalDir)
	return Layout{
		Root:       root,
		Internal:   internal,
		TmpUploads: filepath.Join(internal, "tmp", "uploads"),
		DB:         filepath.Join(internal, "db"),
		Logs:       filepath.Join(internal, "logs"),
	}
}

// Area returns the directory of a namespace in an area, such as
// <root>/files/u0001.
func (l Layout) Area(area, namespace string) string {
	return filepath.Join(l.Root, area, namespace)
}

// dirs lists every directory the layout needs, parents before children.
func (l Layout) dirs() []string {
	return []string{
		filepath.Join(l.Root, FilesArea),
		l.Area(FilesArea, DefaultNamespace),
		filepath.Join(l.Root, PhotosArea),
		l.Area(PhotosArea, DefaultNamespace),
		l.Internal,
		filepath.Dir(l.TmpUploads),
		l.TmpUploads,
		l.DB,
		l.Logs,
	}
}

// Init creates the missing layout directories and checks that each one is
// a real directory (not a file or a symbolic link) that the server can
// write to. It returns the names of root entries it does not know; they are
// reported, never touched. Running Init again on a complete layout changes
// nothing.
func (l Layout) Init() (unknown []string, err error) {
	if !filepath.IsAbs(l.Root) {
		return nil, fmt.Errorf("storage: the storage root must be an absolute path, got %q", l.Root)
	}
	if err := os.MkdirAll(l.Root, 0o750); err != nil {
		return nil, fmt.Errorf("storage: cannot create the storage root: %w", err)
	}
	for _, dir := range l.dirs() {
		if err := ensureDir(dir); err != nil {
			return nil, err
		}
	}
	for _, dir := range l.dirs() {
		if err := checkWritable(dir); err != nil {
			return nil, err
		}
	}

	entries, err := os.ReadDir(l.Root)
	if err != nil {
		return nil, fmt.Errorf("storage: cannot list the storage root: %w", err)
	}
	for _, e := range entries {
		switch e.Name() {
		case FilesArea, PhotosArea, InternalDir:
		default:
			unknown = append(unknown, e.Name())
		}
	}
	slices.Sort(unknown)
	return unknown, nil
}

// ensureDir creates dir if it is missing and checks that it is a real
// directory.
func ensureDir(dir string) error {
	info, err := os.Lstat(dir)
	if errors.Is(err, fs.ErrNotExist) {
		if err := os.Mkdir(dir, 0o750); err != nil {
			return fmt.Errorf("storage: cannot create %s: %w", dir, err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("storage: cannot check %s: %w", dir, err)
	}
	switch {
	case info.Mode()&fs.ModeSymlink != 0:
		return fmt.Errorf("storage: %s is a symbolic link; the storage layout needs real directories", dir)
	case !info.IsDir():
		return fmt.Errorf("storage: %s exists but is not a directory", dir)
	}
	return nil
}

// checkWritable creates and removes a probe file in dir. The directory's
// modification time is put back afterwards, so the check leaves no trace.
func checkWritable(dir string) error {
	before, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("storage: cannot check %s: %w", dir, err)
	}
	f, err := os.CreateTemp(dir, ".write-check-*")
	if err != nil {
		return fmt.Errorf("storage: %s is not writable: %w", dir, err)
	}
	name := f.Name()
	errs := []error{f.Close(), os.Remove(name)}
	if err := errors.Join(errs...); err != nil {
		return fmt.Errorf("storage: write check in %s: %w", dir, err)
	}
	mtime := before.ModTime()
	if err := os.Chtimes(dir, mtime, mtime); err != nil {
		return fmt.Errorf("storage: restore the modification time of %s: %w", dir, err)
	}
	return nil
}
