package files

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"strings"
	"sync"
	"testing"
	"testing/iotest"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

// upload calls Upload with a body of the declared size.
func upload(s *Local, apiPath, content string, policy OnConflict) (Item, bool, error) {
	return s.Upload(context.Background(), owner, apiPath, strings.NewReader(content), int64(len(content)), UploadOptions{OnConflict: policy})
}

// onDisk returns every regular file of the namespace, read directly from
// the disk, hidden temporary files included.
func onDisk(t *testing.T, l storage.Layout) map[string]string {
	t.Helper()
	files, err := testutil.ReadFiles(l.Area(storage.FilesArea, owner))
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func TestUploadIsByteIdentical(t *testing.T) {
	s, l := newService(t, nil, map[string]string{"up/.keep": ""})
	rng := rand.New(rand.NewPCG(1, 2))
	for _, size := range []int{0, 1, copyBufferSize - 1, copyBufferSize, 3*copyBufferSize + 17} {
		data := make([]byte, size)
		for i := range data {
			data[i] = byte(rng.Uint32())
		}
		name := fmt.Sprintf("/up/f%d.bin", size)
		it, created, err := s.Upload(t.Context(), owner, name, bytes.NewReader(data), int64(size), UploadOptions{})
		if err != nil {
			t.Fatalf("%d bytes: %v", size, err)
		}
		if !created || "/"+it.RelPath != name || it.Kind != KindFile || it.Size != int64(size) || it.ETag == "" {
			t.Errorf("%d bytes: item %+v, created %v", size, it, created)
		}
		got := onDisk(t, l)[strings.TrimPrefix(name, "/")]
		if sha256.Sum256([]byte(got)) != sha256.Sum256(data) {
			t.Errorf("%d bytes: the stored file differs from the upload", size)
		}
	}
	if n := len(onDisk(t, l)); n != 6 {
		t.Errorf("%d files on disk, want 6 (no temporary files left)", n)
	}
}

func TestUploadConflictPolicies(t *testing.T) {
	s, l := newService(t, nil, map[string]string{
		"docs/a.txt": "old", "docs/.env": "e", "docs/x.tar.gz": "z", "docs/sub/.keep": "",
	})
	tests := []struct {
		path        string
		policy      OnConflict
		wantErr     apperr.Kind
		wantPath    string
		wantCreated bool
	}{
		{"/docs/a.txt", "", apperr.Conflict, "", false},
		{"/docs/a.txt", ConflictFail, apperr.Conflict, "", false},
		{"/docs/sub", ConflictFail, apperr.Conflict, "", false},
		{"/docs/a.txt", ConflictRename, 0, "/docs/a (1).txt", true},
		{"/docs/a.txt", ConflictRename, 0, "/docs/a (2).txt", true},
		{"/docs/.env", ConflictRename, 0, "/docs/.env (1)", true},
		{"/docs/x.tar.gz", ConflictRename, 0, "/docs/x.tar (1).gz", true},
		{"/docs/sub", ConflictRename, 0, "/docs/sub (1)", true},
		{"/docs/new.txt", ConflictRename, 0, "/docs/new.txt", true},
		{"/docs/a.txt", ConflictOverwrite, 0, "/docs/a.txt", false},
		{"/docs/fresh.txt", ConflictOverwrite, 0, "/docs/fresh.txt", true},
		{"/docs/sub", ConflictOverwrite, apperr.Conflict, "", false},
		{"/docs/a.txt", "merge", apperr.InvalidRequest, "", false},
	}
	for _, tt := range tests {
		it, created, err := upload(s, tt.path, "new", tt.policy)
		if tt.wantErr != 0 {
			if apperr.KindOf(err) != tt.wantErr || err == nil {
				t.Errorf("%s %q: error %v, want %v", tt.path, tt.policy, err, tt.wantErr)
			}
			continue
		}
		if err != nil || "/"+it.RelPath != tt.wantPath || created != tt.wantCreated {
			t.Errorf("%s %q: %s created=%v err=%v, want %s created=%v", tt.path, tt.policy, it.RelPath, created, err, tt.wantPath, tt.wantCreated)
		}
	}
	want := map[string]string{
		"docs/a.txt": "new", "docs/a (1).txt": "new", "docs/a (2).txt": "new",
		"docs/.env": "e", "docs/.env (1)": "new", "docs/x.tar.gz": "z", "docs/x.tar (1).gz": "new",
		"docs/sub/.keep": "", "docs/sub (1)": "new", "docs/new.txt": "new", "docs/fresh.txt": "new",
	}
	if diff := cmp.Diff(want, onDisk(t, l)); diff != "" {
		t.Errorf("files on disk (-want +got):\n%s", diff)
	}
}

// countingReader counts the bytes read from r.
type countingReader struct {
	r io.Reader
	n int
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += n
	return n, err
}

// TestUploadFailuresLeaveNothing: every refused or failed upload leaves the
// disk as it was, without a temporary file, and a request refused on its
// path, size, or target is refused before its body is read.
func TestUploadFailuresLeaveNothing(t *testing.T) {
	fixture := map[string]string{"docs/a.txt": "old", "f.txt": "file"}
	lowSpace := func(string) (uint64, error) { return 1 << 20, nil }
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	tests := []struct {
		name       string
		path       string
		body       io.Reader
		size       int64
		ctx        context.Context
		space      storage.FreeFunc
		want       apperr.Kind
		readsFirst bool // the body is read before the failure
	}{
		{"shorter body", "/n.txt", strings.NewReader("abc"), 10, nil, nil, apperr.InvalidRequest, true},
		{"longer body", "/n.txt", strings.NewReader("abcdef"), 3, nil, nil, apperr.InvalidRequest, true},
		{"read error", "/n.txt", io.MultiReader(strings.NewReader("abc"), iotest.ErrReader(errors.New("connection reset"))), 10, nil, nil, apperr.InvalidRequest, true},
		{"canceled", "/n.txt", strings.NewReader("abc"), 3, canceled, nil, apperr.InvalidRequest, false},
		{"negative size", "/n.txt", strings.NewReader(""), -1, nil, nil, apperr.InvalidRequest, false},
		{"no free space", "/n.txt", strings.NewReader("abc"), 3, nil, lowSpace, apperr.InsufficientStorage, false},
		{"exists", "/docs/a.txt", strings.NewReader("abc"), 3, nil, nil, apperr.Conflict, false},
		{"missing parent", "/nope/n.txt", strings.NewReader("abc"), 3, nil, nil, apperr.NotFound, false},
		{"file as parent", "/f.txt/n.txt", strings.NewReader("abc"), 3, nil, nil, apperr.Conflict, false},
		{"root", "/", strings.NewReader("abc"), 3, nil, nil, apperr.InvalidName, false},
		{"reserved name", "/docs/con.txt", strings.NewReader("abc"), 3, nil, nil, apperr.InvalidName, false},
		{"temp prefix", "/" + storage.TempPrefix + "x", strings.NewReader("abc"), 3, nil, nil, apperr.InvalidName, false},
		{"outside", "/docs/../../x", strings.NewReader("abc"), 3, nil, nil, apperr.OutsideRoot, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := storage.NewLayout(testutil.StorageRoot(t), storage.Options{})
			if _, err := l.Init(); err != nil {
				t.Fatal(err)
			}
			if err := testutil.WriteFiles(l.Area(storage.FilesArea, owner), fixture); err != nil {
				t.Fatal(err)
			}
			o := Options{}
			if tt.space != nil {
				o.Space = storage.NewSpaceGuard(l.Root, 1<<30, tt.space)
			}
			s := NewLocal(storage.NewResolver(l), o)
			ctx := tt.ctx
			if ctx == nil {
				ctx = t.Context()
			}
			body := &countingReader{r: tt.body}
			_, _, err := s.Upload(ctx, owner, tt.path, body, tt.size, UploadOptions{})
			if err == nil || apperr.KindOf(err) != tt.want {
				t.Errorf("error %v, want %v", err, tt.want)
			}
			if !tt.readsFirst && body.n != 0 {
				t.Errorf("%d bytes of the body were read before the refusal", body.n)
			}
			if diff := cmp.Diff(fixture, onDisk(t, l)); diff != "" {
				t.Errorf("files on disk changed (-want +got):\n%s", diff)
			}
		})
	}
}

// TestUploadInvisibleWhileWriting: during an upload neither the target
// nor the temporary file is an item; the file appears complete at once.
func TestUploadInvisibleWhileWriting(t *testing.T) {
	s, l := newService(t, nil, map[string]string{"up/other.txt": "o"})
	pr, pw := io.Pipe()
	done := make(chan error, 1)
	go func() {
		_, _, err := s.Upload(context.Background(), owner, "/up/big.bin", pr, 8, UploadOptions{})
		done <- err
	}()
	if _, err := pw.Write([]byte("half")); err != nil {
		t.Fatal(err)
	}
	// The first half is on disk, under a temporary name only.
	var temp string
	for deadline := time.Now().Add(5 * time.Second); temp == "" && time.Now().Before(deadline); {
		for name := range onDisk(t, l) {
			if strings.HasPrefix(name, "up/"+storage.TempPrefix) {
				temp = name
			}
		}
	}
	if temp == "" {
		t.Fatal("no temporary file appeared")
	}
	page, err := s.List(t.Context(), owner, "/up", ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if got := itemNames(page.Items); !cmp.Equal(got, []string{"other.txt"}) {
		t.Errorf("listing during the upload = %v, want only other.txt", got)
	}
	for _, p := range []string{"/up/big.bin", "/" + temp} {
		if _, err := s.Stat(t.Context(), owner, p); apperr.KindOf(err) != apperr.NotFound {
			t.Errorf("Stat(%s) during the upload = %v, want not_found", p, err)
		}
	}
	if _, err := s.List(t.Context(), owner, "/"+temp, ListOptions{}); apperr.KindOf(err) != apperr.NotFound {
		t.Errorf("List of the temporary file = %v, want not_found", err)
	}

	if _, err := pw.Write([]byte("done")); err != nil {
		t.Fatal(err)
	}
	if err := pw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"up/other.txt": "o", "up/big.bin": "halfdone"}
	if diff := cmp.Diff(want, onDisk(t, l)); diff != "" {
		t.Errorf("files after the upload (-want +got):\n%s", diff)
	}
}

// TestUploadRenameRace: concurrent uploads to one name with
// ConflictRename all succeed, each under its own name with its own
// content.
func TestUploadRenameRace(t *testing.T) {
	s, l := newService(t, nil, map[string]string{"up/.keep": ""})
	const n = 10
	paths := make([]string, n)
	errs := make([]error, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			var it Item
			it, _, errs[i] = upload(s, "/up/same.txt", fmt.Sprintf("writer %d", i), ConflictRename)
			paths[i] = it.RelPath
		})
	}
	wg.Wait()
	disk := onDisk(t, l)
	seen := map[string]bool{}
	for i := range n {
		if errs[i] != nil {
			t.Fatalf("writer %d: %v", i, errs[i])
		}
		if seen[paths[i]] {
			t.Errorf("two writers got %s", paths[i])
		}
		seen[paths[i]] = true
		if got := disk[paths[i]]; got != fmt.Sprintf("writer %d", i) {
			t.Errorf("%s holds %q, want writer %d's content", paths[i], got, i)
		}
	}
	if len(disk) != n+1 {
		t.Errorf("%d files on disk, want %d", len(disk), n+1)
	}
}

// TestUploadWithoutHardLinks simulates a file system without hard links
// (FAT, exFAT): the commit falls back to a check and a rename.
func TestUploadWithoutHardLinks(t *testing.T) {
	orig := linkFile
	linkFile = func(*os.Root, string, string) error { return errors.New("operation not supported") }
	t.Cleanup(func() { linkFile = orig })

	s, l := newService(t, nil, map[string]string{"docs/a.txt": "old"})
	if it, _, err := upload(s, "/docs/b.txt", "b", ConflictFail); err != nil || it.RelPath != "docs/b.txt" {
		t.Errorf("new name: %s, %v", it.RelPath, err)
	}
	if it, _, err := upload(s, "/docs/a.txt", "a2", ConflictRename); err != nil || it.RelPath != "docs/a (1).txt" {
		t.Errorf("rename: %s, %v", it.RelPath, err)
	}
	// A name taken between the early check and the commit is still refused.
	if _, err := commitFileAfterRace(t, s, l); apperr.KindOf(err) != apperr.Conflict {
		t.Errorf("a name taken during the upload = %v, want conflict", err)
	}
	want := map[string]string{"docs/a.txt": "old", "docs/b.txt": "b", "docs/a (1).txt": "a2", "docs/late.txt": "first"}
	if diff := cmp.Diff(want, onDisk(t, l)); diff != "" {
		t.Errorf("files on disk (-want +got):\n%s", diff)
	}
}

// commitFileAfterRace uploads docs/late.txt while another writer creates
// the same name after the early check: the second writer's file appears
// once the body has been read.
func commitFileAfterRace(t *testing.T, s *Local, l storage.Layout) (Item, error) {
	t.Helper()
	racer := readerFunc(func(p []byte) (int, error) {
		if err := testutil.WriteFiles(l.Area(storage.FilesArea, owner), map[string]string{"docs/late.txt": "first"}); err != nil {
			t.Error(err)
		}
		return copy(p, "second"), io.EOF
	})
	it, _, err := s.Upload(t.Context(), owner, "/docs/late.txt", racer, 6, UploadOptions{})
	return it, err
}

// itemNames returns the names of items.
func itemNames(items []Item) []string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = it.Name
	}
	return out
}

type readerFunc func([]byte) (int, error)

func (f readerFunc) Read(p []byte) (int, error) { return f(p) }

func TestUploadRaceWithHardLinks(t *testing.T) {
	s, l := newService(t, nil, map[string]string{"docs/.keep": ""})
	if _, err := commitFileAfterRace(t, s, l); apperr.KindOf(err) != apperr.Conflict {
		t.Errorf("a name taken during the upload = %v, want conflict", err)
	}
	if got := onDisk(t, l)["docs/late.txt"]; got != "first" {
		t.Errorf("docs/late.txt = %q, want the first writer's content", got)
	}
	if n := len(onDisk(t, l)); n != 2 {
		t.Errorf("%d files on disk, want 2 (no temporary file)", n)
	}
}

func TestUploadHook(t *testing.T) {
	h := &recordingHooks{}
	s, l := newService(t, h, map[string]string{"docs/.keep": ""})
	if _, _, err := upload(s, "/docs/a.txt", "a", ""); err != nil {
		t.Fatal(err)
	}
	if len(h.events) != 1 || h.events[0].Op != OpUpload || h.events[0].Path != "docs/a.txt" {
		t.Errorf("events = %+v", h.events)
	}
	h.vetoErr = apperr.New(apperr.Conflict, "vetoed")
	if _, _, err := upload(s, "/docs/b.txt", "b", ""); apperr.KindOf(err) != apperr.Conflict {
		t.Errorf("vetoed upload = %v", err)
	}
	if _, ok := onDisk(t, l)["docs/b.txt"]; ok {
		t.Error("a vetoed upload was written")
	}
}

func TestNumberedName(t *testing.T) {
	for name, want := range map[string]string{
		"a.txt": "a (3).txt", ".env": ".env (3)", "x.tar.gz": "x.tar (3).gz", "noext": "noext (3)", "a.": "a (3).",
	} {
		if got := numberedName(name, 3); got != want {
			t.Errorf("numberedName(%q) = %q, want %q", name, got, want)
		}
	}
}
