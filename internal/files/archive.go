package files

import (
	"archive/zip"
	"context"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
)

// MaxArchiveItems is the most files and folders one archive holds (S02.4-T03),
// so that an archive's plan stays small in memory. It matches the library
// size the project plans for (100,000 files, Q18).
const MaxArchiveItems = 100_000

// ArchiveEntry is one file or folder of an archive.
type ArchiveEntry struct {
	// Name is the path inside the archive, "/" between names.
	Name string
	// Rel is the path in the owner's namespace.
	Rel     string
	Dir     bool
	Size    int64
	ModTime time.Time
}

// ArchivePlan lists what an archive holds, checked before any byte of it is
// sent (S02.4-T03, FR-006).
type ArchivePlan struct {
	Owner   string
	Entries []ArchiveEntry
	// Size is the total size of the files, in bytes.
	Size int64
}

// PlanArchive checks the items at paths and lists them, and everything in
// the folders among them, without following links. Each item is at the top
// of the archive under its own name; a name that repeats gets a number, as
// "name (1).ext". A link or special file anywhere is refused, as for a copy.
func (s *Local) PlanArchive(ctx context.Context, owner string, paths []string) (ArchivePlan, error) {
	if len(paths) == 0 {
		return ArchivePlan{}, apperr.New(apperr.InvalidRequest, "an archive needs at least one path")
	}
	rels := make([]string, 0, len(paths))
	for _, p := range paths {
		rel, err := s.resolveVisible(owner, p)
		if err != nil {
			return ArchivePlan{}, err
		}
		rels = append(rels, rel)
	}
	return run(ctx, s.hooks, Event{Op: OpArchive, Owner: owner, Path: rels[0]}, func() (ArchivePlan, error) {
		plan := ArchivePlan{Owner: owner}
		err := s.withRoot(owner, func(root *os.Root) error {
			used := map[string]bool{}
			for i, rel := range rels {
				if err := refuseLinkParents(root, rel, paths[i]); err != nil {
					return err
				}
				info, err := root.Lstat(filepath.FromSlash(rel))
				if err != nil {
					return fsError(err, paths[i])
				}
				top := topName(rel, used)
				if err := planTree(root, rel, top, info, &plan); err != nil {
					return err
				}
			}
			return nil
		})
		return plan, err
	})
}

// topName is the name an item gets at the top of the archive: its own
// name, "files" for the root, and a numbered name when it is taken.
func topName(rel string, used map[string]bool) string {
	name := path.Base(rel)
	if rel == "." {
		name = "files"
	}
	candidate := name
	for n := 1; used[strings.ToLower(candidate)]; n++ {
		candidate = numberedName(name, n)
	}
	used[strings.ToLower(candidate)] = true
	return candidate
}

// planTree adds the item rel, named name in the archive, and everything in
// it when it is a folder.
func planTree(root *os.Root, rel, name string, info fs.FileInfo, plan *ArchivePlan) error {
	switch {
	case info.Mode().IsRegular():
		plan.Size += info.Size()
		plan.Entries = append(plan.Entries, ArchiveEntry{Name: name, Rel: rel, Size: info.Size(), ModTime: info.ModTime()})
	case info.IsDir():
		plan.Entries = append(plan.Entries, ArchiveEntry{Name: name, Rel: rel, Dir: true, ModTime: info.ModTime()})
	default:
		return apperr.Newf(apperr.InvalidRequest, "/%s is a %s; only files and folders can be archived (links are never followed)",
			rel, NewItem("", rel, info).Kind)
	}
	if len(plan.Entries) > MaxArchiveItems {
		return apperr.Newf(apperr.TooLargeForSync, "the archive has more than %d files and folders, the most one archive holds", MaxArchiveItems)
	}
	if !info.IsDir() {
		return nil
	}
	names, err := readNames(root, rel)
	if err != nil {
		return err
	}
	for _, child := range names {
		childRel := path.Join(rel, child)
		childInfo, err := root.Lstat(filepath.FromSlash(childRel))
		if storage.IsNotFound(err) {
			continue // removed since the listing
		}
		if err != nil {
			return fsError(err, "/"+childRel)
		}
		if err := planTree(root, childRel, name+"/"+child, childInfo, plan); err != nil {
			return err
		}
	}
	return nil
}

// WriteArchive writes the planned items to w as a ZIP archive: entries are
// stored, not compressed (fast, and most large files are compressed
// already), with ZIP64 where sizes need it, UTF-8 names, and the items'
// modification times. Each file is opened as a download opens it, so a
// link swapped in is never read; a file removed since the plan is left
// out. Memory stays bounded: one copy buffer, whatever the sizes.
func (s *Local) WriteArchive(ctx context.Context, plan ArchivePlan, w io.Writer) error {
	return s.withRoot(plan.Owner, func(root *os.Root) error {
		zw := zip.NewWriter(w)
		buf := make([]byte, copyBufferSize)
		for _, e := range plan.Entries {
			if err := ctx.Err(); err != nil {
				return err
			}
			if e.Dir {
				if _, err := zw.CreateHeader(&zip.FileHeader{Name: e.Name + "/", Method: zip.Store, Modified: e.ModTime}); err != nil {
					return err
				}
				continue
			}
			f, _, err := openFile(root, plan.Owner, e.Rel, "/"+e.Rel)
			if apperr.KindOf(err) == apperr.NotFound {
				continue // removed since the plan
			}
			if err != nil {
				return err
			}
			entry, err := zw.CreateHeader(&zip.FileHeader{Name: e.Name, Method: zip.Store, Modified: e.ModTime})
			if err == nil {
				err = copyStream(ctx, entry, f, buf)
			}
			_ = f.Close() // read-only
			if err != nil {
				return err
			}
		}
		return zw.Close()
	})
}

// ArchiveName is a file name for an archive of the items at paths: the
// name of a single item, or "download" with the date for several.
func ArchiveName(paths []string, now time.Time) string {
	if len(paths) == 1 {
		if base := path.Base(paths[0]); base != "/" && base != "." && base != "" {
			return base + ".zip"
		}
	}
	return "download-" + now.Format("2006-01-02") + ".zip"
}
