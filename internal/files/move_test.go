package files

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
)

func TestRename(t *testing.T) {
	s, l := newService(t, nil, map[string]string{
		"docs/a.txt": "a", "docs/sub/deep.txt": "d", "case.txt": "c",
	})
	tests := []struct {
		path, newName, want string
	}{
		{"/docs/a.txt", "b.txt", "docs/b.txt"},
		{"/docs/sub", "renamed", "docs/renamed"},
		{"/case.txt", "CASE.txt", "CASE.txt"},  // a case-only change, also on Windows
		{"/docs/b.txt", "b.txt", "docs/b.txt"}, // the same name: nothing to do
	}
	for _, tt := range tests {
		it, err := s.Rename(t.Context(), owner, tt.path, tt.newName, MoveOptions{})
		if err != nil || it.RelPath != tt.want {
			t.Errorf("Rename(%s, %s) = %s, %v; want %s", tt.path, tt.newName, it.RelPath, err, tt.want)
		}
	}
	want := map[string]string{"docs/b.txt": "a", "docs/renamed/deep.txt": "d", "CASE.txt": "c"}
	if diff := cmp.Diff(want, onDisk(t, l)); diff != "" {
		t.Errorf("files on disk (-want +got):\n%s", diff)
	}
}

func TestMove(t *testing.T) {
	s, l := newService(t, nil, map[string]string{
		"in/a.txt": "a", "in/tree/x/y.txt": "y", "out/.keep": "", "docs/x": "x", "docs2/.keep": "",
	})
	tests := []struct {
		from, to string
	}{
		{"/in/a.txt", "/out/a.txt"},
		{"/out/a.txt", "/out/renamed.txt"}, // a move can rename
		{"/in/tree", "/out/tree"},          // a folder with its contents
		{"/docs", "/docs2/docs"},           // "docs2" only starts like "docs"
	}
	for _, tt := range tests {
		it, err := s.Move(t.Context(), owner, tt.from, tt.to, MoveOptions{})
		if err != nil || "/"+it.RelPath != tt.to {
			t.Errorf("Move(%s, %s) = %s, %v", tt.from, tt.to, it.RelPath, err)
		}
	}
	want := map[string]string{
		"out/.keep": "", "out/renamed.txt": "a", "out/tree/x/y.txt": "y", "docs2/.keep": "", "docs2/docs/x": "x",
	}
	if diff := cmp.Diff(want, onDisk(t, l)); diff != "" {
		t.Errorf("files on disk (-want +got):\n%s", diff)
	}
}

func TestMoveConflictPolicies(t *testing.T) {
	s, l := newService(t, nil, map[string]string{
		"f1.txt": "1", "f2.txt": "2", "f3.txt": "3", "f4.txt": "4",
		"v1.2/.keep": "", "d2/.keep": "", "target/.keep": "", "t.txt": "old",
	})
	tests := []struct {
		from, to string
		policy   OnConflict
		wantErr  apperr.Kind
		want     string
	}{
		{"/f1.txt", "/t.txt", "", apperr.Conflict, ""},
		{"/f1.txt", "/t.txt", ConflictRename, 0, "t (1).txt"},
		{"/f2.txt", "/t.txt", ConflictRename, 0, "t (2).txt"},
		{"/f3.txt", "/t.txt", ConflictOverwrite, 0, "t.txt"},
		{"/v1.2", "/target", ConflictRename, 0, "target (1)"},
		{"/d2", "/target (1)", ConflictRename, 0, "target (1) (1)"},
		{"/target", "/t.txt", ConflictOverwrite, apperr.Conflict, ""},      // a folder never replaces a file
		{"/f4.txt", "/target", ConflictOverwrite, apperr.Conflict, ""},     // nor is a folder replaced
		{"/target (1)", "/target", ConflictOverwrite, apperr.Conflict, ""}, // nor merged
		{"/target", "/t.txt", ConflictFail, apperr.Conflict, ""},
	}
	for _, tt := range tests {
		it, err := s.Move(t.Context(), owner, tt.from, tt.to, MoveOptions{OnConflict: tt.policy})
		if tt.wantErr != 0 {
			if err == nil || apperr.KindOf(err) != tt.wantErr {
				t.Errorf("%s → %s %q: %v, want %v", tt.from, tt.to, tt.policy, err, tt.wantErr)
			}
			continue
		}
		if err != nil || it.RelPath != tt.want {
			t.Errorf("%s → %s %q: %s, %v; want %s", tt.from, tt.to, tt.policy, it.RelPath, err, tt.want)
		}
	}
	want := map[string]string{
		"t.txt": "3", "t (1).txt": "1", "t (2).txt": "2", "f4.txt": "4",
		"target/.keep": "", "target (1)/.keep": "", "target (1) (1)/.keep": "",
	}
	if diff := cmp.Diff(want, onDisk(t, l)); diff != "" {
		t.Errorf("files on disk (-want +got):\n%s", diff)
	}
}

// TestMoveIntoItself: a folder cannot move into itself or below itself,
// also through a case variant of its path on a case-insensitive disk.
func TestMoveIntoItself(t *testing.T) {
	s, l := newService(t, nil, map[string]string{"docs/sub/x.txt": "x"})
	cases := map[string]apperr.Kind{
		"/docs/inside":     apperr.InvalidRequest,
		"/docs/sub/inside": apperr.InvalidRequest,
	}
	if runtime.GOOS == "windows" {
		cases["/DOCS/sub/inside"] = apperr.InvalidRequest // the same folder on NTFS
	} else {
		cases["/DOCS/sub/inside"] = apperr.NotFound // another (missing) folder on ext4
	}
	for to, want := range cases {
		if _, err := s.Move(t.Context(), owner, "/docs", to, MoveOptions{}); apperr.KindOf(err) != want || err == nil {
			t.Errorf("Move(/docs, %s) = %v, want %v", to, err, want)
		}
	}
	if it, err := s.Move(t.Context(), owner, "/docs", "/docs", MoveOptions{}); err != nil || it.RelPath != "docs" {
		t.Errorf("Move onto itself = %s, %v; want nothing to do", it.RelPath, err)
	}
	if diff := cmp.Diff(map[string]string{"docs/sub/x.txt": "x"}, onDisk(t, l)); diff != "" {
		t.Errorf("files on disk (-want +got):\n%s", diff)
	}
}

func TestMoveErrorsLeaveNothing(t *testing.T) {
	fixture := map[string]string{
		"docs/a.txt": "a", "f.txt": "f", "docs/" + storage.TempPrefix + "1.part": "partial",
	}
	s, l := newService(t, nil, fixture)
	tests := []struct {
		from, to string
		policy   OnConflict
		want     apperr.Kind
	}{
		{"/missing", "/x", "", apperr.NotFound},
		{"/", "/x", "", apperr.InvalidRequest},
		{"/docs/a.txt", "/", "", apperr.InvalidName},
		{"/docs/a.txt", "/nope/a.txt", "", apperr.NotFound},
		{"/docs/a.txt", "/f.txt/a.txt", "", apperr.Conflict},
		{"/docs/a.txt", "/aux", "", apperr.InvalidName},
		{"/docs/a.txt", "/" + storage.TempPrefix + "x", "", apperr.InvalidName},
		{"/docs/" + storage.TempPrefix + "1.part", "/visible.txt", "", apperr.NotFound},
		{"/docs/../../a", "/x", "", apperr.OutsideRoot},
		{"/docs/a.txt", "/../x", "", apperr.OutsideRoot},
		{"/docs/a.txt", "/x", "merge", apperr.InvalidRequest},
	}
	for _, tt := range tests {
		if _, err := s.Move(t.Context(), owner, tt.from, tt.to, MoveOptions{OnConflict: tt.policy}); err == nil || apperr.KindOf(err) != tt.want {
			t.Errorf("Move(%s, %s) = %v, want %v", tt.from, tt.to, err, tt.want)
		}
	}
	for _, name := range []string{"", ".", "..", "a/b", "con.txt", "x.", storage.TempPrefix + "y"} {
		if _, err := s.Rename(t.Context(), owner, "/docs/a.txt", name, MoveOptions{}); apperr.KindOf(err) != apperr.InvalidName {
			t.Errorf("Rename to %q = %v, want invalid_name", name, err)
		}
	}
	if diff := cmp.Diff(fixture, onDisk(t, l)); diff != "" {
		t.Errorf("files on disk changed (-want +got):\n%s", diff)
	}
}

// TestMoveConcurrentToOneName: with fail, exactly one of many concurrent
// moves to one name wins and no file is lost; with rename, all succeed
// under distinct names.
func TestMoveConcurrentToOneName(t *testing.T) {
	const n = 10
	for _, policy := range []OnConflict{ConflictFail, ConflictRename} {
		t.Run(string(policy), func(t *testing.T) {
			files := map[string]string{}
			for i := range n {
				files[fmt.Sprintf("src/f%d.txt", i)] = fmt.Sprintf("file %d", i)
			}
			s, l := newService(t, nil, files)
			errs := make([]error, n)
			var wg sync.WaitGroup
			for i := range n {
				wg.Go(func() {
					_, errs[i] = s.Move(t.Context(), owner, fmt.Sprintf("/src/f%d.txt", i), "/same.txt", MoveOptions{OnConflict: policy})
				})
			}
			wg.Wait()
			ok := 0
			for _, err := range errs {
				switch {
				case err == nil:
					ok++
				case apperr.KindOf(err) != apperr.Conflict:
					t.Errorf("unexpected error %v", err)
				}
			}
			want := map[OnConflict]int{ConflictFail: 1, ConflictRename: n}[policy]
			disk := onDisk(t, l)
			if ok != want || len(disk) != n {
				t.Errorf("%d moves succeeded (want %d); %d files on disk (want %d): %v", ok, want, len(disk), n, disk)
			}
		})
	}
}

func TestMoveWithoutHardLinks(t *testing.T) {
	orig := linkFile
	linkFile = func(*os.Root, string, string) error { return errors.New("operation not supported") }
	t.Cleanup(func() { linkFile = orig })
	s, l := newService(t, nil, map[string]string{"a.txt": "a", "b.txt": "b"})
	if _, err := s.Move(t.Context(), owner, "/a.txt", "/b.txt", MoveOptions{}); apperr.KindOf(err) != apperr.Conflict {
		t.Errorf("fail onto an existing file = %v, want conflict", err)
	}
	if it, err := s.Move(t.Context(), owner, "/a.txt", "/c.txt", MoveOptions{}); err != nil || it.RelPath != "c.txt" {
		t.Errorf("move = %s, %v", it.RelPath, err)
	}
	if diff := cmp.Diff(map[string]string{"b.txt": "b", "c.txt": "a"}, onDisk(t, l)); diff != "" {
		t.Errorf("files on disk (-want +got):\n%s", diff)
	}
}

func TestMoveHooks(t *testing.T) {
	h := &recordingHooks{}
	s, _ := newService(t, h, map[string]string{"d/a.txt": "a"})
	if _, err := s.Rename(t.Context(), owner, "/d/a.txt", "b.txt", MoveOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Move(t.Context(), owner, "/d/b.txt", "/c.txt", MoveOptions{}); err != nil {
		t.Fatal(err)
	}
	want := []Event{
		{Op: OpRename, Owner: owner, Path: "d/a.txt", Target: "d/b.txt"},
		{Op: OpMove, Owner: owner, Path: "d/b.txt", Target: "c.txt"},
	}
	if diff := cmp.Diff(want, h.events); diff != "" {
		t.Errorf("events (-want +got):\n%s", diff)
	}
}

// TestMoveConcurrentFolders: concurrent folder moves to one name with
// rename all succeed under distinct names (on Windows a lost race shows
// up as "access denied" and must still count as a taken name).
func TestMoveConcurrentFolders(t *testing.T) {
	const n = 6
	tree := map[string]string{}
	for i := range n {
		tree[fmt.Sprintf("d%d/f.txt", i)] = fmt.Sprint(i)
	}
	s, l := newService(t, nil, tree)
	errs := make([]error, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			_, errs[i] = s.Move(t.Context(), owner, fmt.Sprintf("/d%d", i), "/dst", MoveOptions{OnConflict: ConflictRename})
		})
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Errorf("move %d: %v", i, err)
		}
	}
	if disk := onDisk(t, l); len(disk) != n {
		t.Errorf("%d files on disk, want %d: %v", len(disk), n, disk)
	}
}
