package files

import (
	"context"
	"fmt"
	"os"
	"sort"
	"sync"
	"testing"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
)

// testCtx is the context of the table-driven operations below.
var testCtx = context.Background()

// A conflict case's target state.
const (
	targetFree   = "free"
	targetFile   = "file"
	targetFolder = "folder"
)

// outcome is what a conflict case must give: an error kind, or the path
// used and (for operations that report it) whether the item is new.
type outcome struct {
	err     apperr.Kind
	path    string
	created bool
}

var conflict = outcome{err: apperr.Conflict}

// Expected outcomes by policy and target state, for items that are files
// (upload, copy, move, rename of a file), for folders moved or copied, and
// for new folders. A file keeps its extension when numbered ("t (1).txt");
// a folder name is numbered whole ("t.txt (1)").
var (
	fileOutcomes = map[OnConflict]map[string]outcome{
		ConflictFail:      {targetFree: {path: "t.txt", created: true}, targetFile: conflict, targetFolder: conflict},
		ConflictRename:    {targetFree: {path: "t.txt", created: true}, targetFile: {path: "t (1).txt", created: true}, targetFolder: {path: "t (1).txt", created: true}},
		ConflictOverwrite: {targetFree: {path: "t.txt", created: true}, targetFile: {path: "t.txt", created: false}, targetFolder: conflict},
	}
	folderOutcomes = map[OnConflict]map[string]outcome{
		ConflictFail:      {targetFree: {path: "t.txt", created: true}, targetFile: conflict, targetFolder: conflict},
		ConflictRename:    {targetFree: {path: "t.txt", created: true}, targetFile: {path: "t.txt (1)", created: true}, targetFolder: {path: "t.txt (1)", created: true}},
		ConflictOverwrite: {targetFree: {path: "t.txt", created: true}, targetFile: conflict, targetFolder: conflict},
	}
	newFolderOutcomes = map[OnConflict]map[string]outcome{
		ConflictFail:      folderOutcomes[ConflictFail],
		ConflictRename:    folderOutcomes[ConflictRename],
		ConflictOverwrite: {targetFree: {path: "t.txt", created: true}, targetFile: conflict, targetFolder: {path: "t.txt", created: false}},
	}
)

// conflictOp runs one operation onto the target /t.txt. created is nil
// for operations that do not report it (move, rename).
type conflictOp struct {
	name     string
	outcomes map[OnConflict]map[string]outcome
	run      func(s *Local, p OnConflict) (Item, *bool, error)
	isFolder bool // the result is a folder copied or moved from srcdir
}

func created(it Item, c bool, err error) (Item, *bool, error) { return it, &c, err }

var conflictOps = []conflictOp{
	{"upload", fileOutcomes, func(s *Local, p OnConflict) (Item, *bool, error) {
		return created(upload(s, "/t.txt", "new", p))
	}, false},
	{"create folder", newFolderOutcomes, func(s *Local, p OnConflict) (Item, *bool, error) {
		return created(s.CreateFolder(testCtx, owner, "/t.txt", FolderOptions{OnConflict: p}))
	}, false},
	{"copy file", fileOutcomes, func(s *Local, p OnConflict) (Item, *bool, error) {
		return created(s.Copy(testCtx, owner, "/src.txt", "/t.txt", CopyOptions{OnConflict: p}))
	}, false},
	{"copy folder", folderOutcomes, func(s *Local, p OnConflict) (Item, *bool, error) {
		return created(s.Copy(testCtx, owner, "/srcdir", "/t.txt", CopyOptions{OnConflict: p}))
	}, true},
	{"tus finalize", fileOutcomes, func(s *Local, p OnConflict) (Item, *bool, error) {
		// A finished upload outside the area (S01.4-T03).
		f, err := os.CreateTemp("", "finished-upload-*")
		if err != nil {
			return Item{}, nil, err
		}
		defer func() { _ = os.Remove(f.Name()) }() // gone after a commit
		if _, err := f.WriteString("new"); err != nil {
			return Item{}, nil, err
		}
		if err := f.Close(); err != nil {
			return Item{}, nil, err
		}
		return created(s.CommitUpload(testCtx, owner, "/t.txt", f.Name(), UploadOptions{OnConflict: p}))
	}, false},
	{"move file", fileOutcomes, func(s *Local, p OnConflict) (Item, *bool, error) {
		it, err := s.Move(testCtx, owner, "/in/src.txt", "/t.txt", MoveOptions{OnConflict: p})
		return it, nil, err
	}, false},
	{"move folder", folderOutcomes, func(s *Local, p OnConflict) (Item, *bool, error) {
		it, err := s.Move(testCtx, owner, "/in/srcdir", "/t.txt", MoveOptions{OnConflict: p})
		return it, nil, err
	}, true},
	{"rename file", fileOutcomes, func(s *Local, p OnConflict) (Item, *bool, error) {
		it, err := s.Rename(testCtx, owner, "/src.txt", "t.txt", MoveOptions{OnConflict: p})
		return it, nil, err
	}, false},
	{"rename folder", folderOutcomes, func(s *Local, p OnConflict) (Item, *bool, error) {
		it, err := s.Rename(testCtx, owner, "/srcdir", "t.txt", MoveOptions{OnConflict: p})
		return it, nil, err
	}, true},
}

// TestConflictMatrix is the S01.6-T04 acceptance test: every conflict
// policy for every operation that creates an item, against a free name,
// an existing file, and an existing folder. The tus finalize rows came
// with S01.4-T03.
func TestConflictMatrix(t *testing.T) {
	for _, op := range conflictOps {
		for _, policy := range []OnConflict{ConflictFail, ConflictRename, ConflictOverwrite} {
			for _, state := range []string{targetFree, targetFile, targetFolder} {
				t.Run(fmt.Sprintf("%s/%s/%s", op.name, policy, state), func(t *testing.T) {
					tree := map[string]string{
						"src.txt": "new", "srcdir/f.txt": "new", "in/src.txt": "new", "in/srcdir/f.txt": "new",
					}
					switch state {
					case targetFile:
						tree["t.txt"] = "old"
					case targetFolder:
						tree["t.txt/keep"] = "old"
					}
					s, l := newService(t, nil, tree)
					want := op.outcomes[policy][state]
					it, gotCreated, err := op.run(s, policy)
					if want.err != 0 {
						if err == nil || apperr.KindOf(err) != want.err {
							t.Fatalf("error %v, want %v", err, want.err)
						}
						return
					}
					if err != nil || it.RelPath != want.path {
						t.Fatalf("result %q, %v; want %q", it.RelPath, err, want.path)
					}
					if gotCreated != nil && *gotCreated != want.created {
						t.Errorf("created = %v, want %v", *gotCreated, want.created)
					}
					disk := onDisk(t, l)
					switch {
					case op.isFolder:
						if disk[want.path+"/f.txt"] != "new" {
							t.Errorf("the folder at %s is incomplete: %v", want.path, disk)
						}
					case op.name != "create folder":
						if disk[want.path] != "new" {
							t.Errorf("%s holds %q, want the new content", want.path, disk[want.path])
						}
					}
					// The existing item survives unless it was overwritten.
					if state == targetFile && want.path != "t.txt" && disk["t.txt"] != "old" {
						t.Errorf("the existing file changed: %q", disk["t.txt"])
					}
					if state == targetFolder && disk["t.txt/keep"] != "old" {
						t.Errorf("the existing folder changed: %v", disk)
					}
				})
			}
		}
	}
}

// TestConcurrentRenameCollisions: for every operation, concurrent calls
// to one name with ConflictRename all succeed under distinct names.
func TestConcurrentRenameCollisions(t *testing.T) {
	const n = 8
	type opN func(s *Local, i int) (Item, error)
	ops := map[string]opN{
		"upload": func(s *Local, i int) (Item, error) {
			it, _, err := upload(s, "/t.txt", fmt.Sprint(i), ConflictRename)
			return it, err
		},
		"create folder": func(s *Local, i int) (Item, error) {
			it, _, err := s.CreateFolder(testCtx, owner, "/t.txt", FolderOptions{OnConflict: ConflictRename})
			return it, err
		},
		"copy file": func(s *Local, i int) (Item, error) {
			it, _, err := s.Copy(testCtx, owner, fmt.Sprintf("/f%d.txt", i), "/t.txt", CopyOptions{OnConflict: ConflictRename})
			return it, err
		},
		"copy folder": func(s *Local, i int) (Item, error) {
			it, _, err := s.Copy(testCtx, owner, fmt.Sprintf("/d%d", i), "/t.txt", CopyOptions{OnConflict: ConflictRename})
			return it, err
		},
		"move file": func(s *Local, i int) (Item, error) {
			return s.Move(testCtx, owner, fmt.Sprintf("/f%d.txt", i), "/t.txt", MoveOptions{OnConflict: ConflictRename})
		},
		"move folder": func(s *Local, i int) (Item, error) {
			return s.Move(testCtx, owner, fmt.Sprintf("/d%d", i), "/t.txt", MoveOptions{OnConflict: ConflictRename})
		},
		"rename file": func(s *Local, i int) (Item, error) {
			return s.Rename(testCtx, owner, fmt.Sprintf("/f%d.txt", i), "t.txt", MoveOptions{OnConflict: ConflictRename})
		},
		"rename folder": func(s *Local, i int) (Item, error) {
			return s.Rename(testCtx, owner, fmt.Sprintf("/d%d", i), "t.txt", MoveOptions{OnConflict: ConflictRename})
		},
	}
	for name, op := range ops {
		t.Run(name, func(t *testing.T) {
			tree := map[string]string{}
			for i := range n {
				tree[fmt.Sprintf("f%d.txt", i)] = fmt.Sprint(i)
				tree[fmt.Sprintf("d%d/x", i)] = fmt.Sprint(i)
			}
			s, _ := newService(t, nil, tree)
			paths := make([]string, n)
			var wg sync.WaitGroup
			for i := range n {
				wg.Go(func() {
					it, err := op(s, i)
					if err != nil {
						t.Errorf("call %d: %v", i, err)
					}
					paths[i] = it.RelPath
				})
			}
			wg.Wait()
			sort.Strings(paths)
			for i := 1; i < n; i++ {
				if paths[i] == paths[i-1] {
					t.Errorf("two calls got %s: %v", paths[i], paths)
				}
			}
		})
	}
}
