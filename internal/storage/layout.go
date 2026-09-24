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
	"runtime"
	"slices"
	"strings"
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

// Options relocates internal data (ADR-0003, S01.2-T02). An empty field
// keeps the default location under <root>/.local-ai-nas. Uploads in progress
// cannot be relocated: they must be renamed into the files area at the end,
// which needs the same filesystem.
type Options struct {
	DBDir   string // storage.db_dir
	LogsDir string // storage.logs_dir
}

// NewLayout returns the layout for the storage root, with the database and
// logs moved where opts says.
func NewLayout(root string, opts Options) Layout {
	internal := filepath.Join(root, InternalDir)
	l := Layout{
		Root:       root,
		Internal:   internal,
		TmpUploads: filepath.Join(internal, "tmp", "uploads"),
		DB:         filepath.Join(internal, "db"),
		Logs:       filepath.Join(internal, "logs"),
	}
	if opts.DBDir != "" {
		l.DB = filepath.Clean(opts.DBDir)
	}
	if opts.LogsDir != "" {
		l.Logs = filepath.Clean(opts.LogsDir)
	}
	return l
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
	if err := l.validate(); err != nil {
		return nil, err
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

// namedDir is a layout directory with the name used in error messages.
type namedDir struct {
	name, path string
}

// validate checks the layout before anything is created: absolute paths,
// and internal data that never overlaps the areas (invariant I2).
func (l Layout) validate() error {
	if !filepath.IsAbs(l.Root) {
		return fmt.Errorf("storage: the storage root must be an absolute path, got %q", l.Root)
	}
	areas := []namedDir{
		{"the files area", filepath.Join(l.Root, FilesArea)},
		{"the photos area", filepath.Join(l.Root, PhotosArea)},
	}
	uploads := namedDir{"the uploads directory", l.TmpUploads}
	db := namedDir{"the database directory (storage.db_dir)", l.DB}
	logs := namedDir{"the logs directory (storage.logs_dir)", l.Logs}
	internal := []namedDir{{"the internal data directory", l.Internal}, uploads, db, logs}

	for _, d := range []namedDir{db, logs} {
		if !filepath.IsAbs(d.path) {
			return fmt.Errorf("storage: %s must be an absolute path, got %q", d.name, d.path)
		}
	}
	for _, in := range internal {
		for _, a := range areas {
			if err := noOverlap(in, a); err != nil {
				return err
			}
		}
	}
	for _, d := range []namedDir{db, logs} {
		if err := noOverlap(d, uploads); err != nil {
			return err
		}
		// Inside the storage root, internal data belongs in .local-ai-nas, so
		// the root holds nothing but the two areas and the internal directory.
		if within(l.Root, d.path) && !within(l.Internal, d.path) {
			return fmt.Errorf("storage: %s (%s) is inside the storage root but not in %s; use a folder under %s or outside the root",
				d.name, d.path, InternalDir, l.Internal)
		}
	}
	return nil
}

// noOverlap reports an error if a and b are the same directory or one
// contains the other.
func noOverlap(a, b namedDir) error {
	switch {
	case samePath(a.path, b.path):
		return fmt.Errorf("storage: %s and %s are the same directory (%s)", a.name, b.name, a.path)
	case within(b.path, a.path):
		return fmt.Errorf("storage: %s (%s) must not be inside %s (%s)", a.name, a.path, b.name, b.path)
	case within(a.path, b.path):
		return fmt.Errorf("storage: %s (%s) must not contain %s (%s)", a.name, a.path, b.name, b.path)
	}
	return nil
}

// within reports whether path is dir or lies inside it. Both must be clean
// absolute paths.
func within(dir, path string) bool {
	if samePath(dir, path) {
		return true
	}
	prefix := dir
	if !strings.HasSuffix(prefix, string(filepath.Separator)) {
		prefix += string(filepath.Separator)
	}
	return len(path) > len(prefix) && samePath(path[:len(prefix)], prefix)
}

// samePath compares paths the way the operating system's default file
// system does: case-insensitively on Windows and macOS.
func samePath(a, b string) bool {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		return strings.EqualFold(a, b)
	}
	return a == b
}

// ensureDir creates dir (and missing parents, for relocated directories)
// if it is missing and checks that it is a real directory.
func ensureDir(dir string) error {
	if _, err := os.Lstat(dir); errors.Is(err, fs.ErrNotExist) {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("storage: cannot create %s: %w", dir, err)
		}
		return nil
	}
	return ensureExistingDir(dir)
}

// ensureExistingDir checks that dir exists and is a real directory (not a
// file or a symbolic link). It never creates anything.
func ensureExistingDir(dir string) error {
	info, err := os.Lstat(dir)
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
