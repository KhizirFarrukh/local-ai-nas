package files

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

const owner = storage.DefaultNamespace

// newService returns a service on a fresh storage root with files placed
// in the default namespace, and the layout.
func newService(t testing.TB, hooks Hooks, files map[string]string) (*Local, storage.Layout) {
	t.Helper()
	l := storage.NewLayout(testutil.StorageRoot(t), storage.Options{})
	if _, err := l.Init(); err != nil {
		t.Fatal(err)
	}
	if err := testutil.WriteFiles(l.Area(storage.FilesArea, owner), files); err != nil {
		t.Fatal(err)
	}
	return NewLocal(storage.NewResolver(l), Options{Hooks: hooks}), l
}

func TestStat(t *testing.T) {
	s, _ := newService(t, nil, map[string]string{"docs/report.pdf": "12345"})
	tests := []struct {
		path     string
		wantRel  string
		wantKind Kind
		wantSize int64
	}{
		{"/", ".", KindDir, 0},
		{"/docs", "docs", KindDir, 0},
		{"/docs/report.pdf", "docs/report.pdf", KindFile, 5},
	}
	for _, tt := range tests {
		it, err := s.Stat(t.Context(), owner, tt.path)
		if err != nil {
			t.Fatalf("Stat(%q): %v", tt.path, err)
		}
		if it.RelPath != tt.wantRel || it.Kind != tt.wantKind || it.Size != tt.wantSize || it.OwnerID != owner {
			t.Errorf("Stat(%q) = %+v", tt.path, it)
		}
	}
}

func TestStatErrors(t *testing.T) {
	s, l := newService(t, nil, map[string]string{"a.txt": "x"})
	// Another namespace and the photos area exist, with files in them.
	other := filepath.Join(l.Root, "files", "u0002")
	if err := os.MkdirAll(other, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := testutil.WriteFiles(other, map[string]string{"secret.txt": "other user"}); err != nil {
		t.Fatal(err)
	}
	if err := testutil.WriteFiles(l.Area(storage.PhotosArea, owner), map[string]string{"p.jpg": "photo"}); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		path string
		want apperr.Kind
	}{
		{"/missing.txt", apperr.NotFound},
		{"/missing/deeper.txt", apperr.NotFound},
		{"/a.txt/inside", apperr.NotFound}, // through a file (ENOTDIR on Unix)
		{"/../u0002/secret.txt", apperr.OutsideRoot},
		{"/../../photos/u0001/p.jpg", apperr.OutsideRoot},
		{"relative.txt", apperr.InvalidRequest},
		{`/a\b`, apperr.InvalidName},
	}
	for _, tt := range tests {
		_, err := s.Stat(t.Context(), owner, tt.path)
		if k := apperr.KindOf(err); err == nil || k != tt.want {
			t.Errorf("Stat(%q) = %v (kind %s), want %s", tt.path, err, k, tt.want)
		}
	}
	if _, err := s.Stat(t.Context(), "u0003", "/"); apperr.KindOf(err) != apperr.Internal {
		t.Errorf("Stat in a namespace that does not exist = %v, want internal", err)
	}
}

// recordingHooks records the hook calls; before returns vetoErr.
type recordingHooks struct {
	calls     []string
	events    []Event
	vetoErr   error
	afterErrs []error
}

func (h *recordingHooks) Before(_ context.Context, e Event) error {
	h.calls = append(h.calls, "before")
	h.events = append(h.events, e)
	return h.vetoErr
}

func (h *recordingHooks) After(_ context.Context, e Event, err error) {
	h.calls = append(h.calls, "after")
	h.afterErrs = append(h.afterErrs, err)
}

func TestHooksOrderAndEvents(t *testing.T) {
	h := &recordingHooks{}
	s, _ := newService(t, h, map[string]string{"docs/a.txt": "x"})

	if _, err := s.Stat(t.Context(), owner, "/docs/a.txt"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Stat(t.Context(), owner, "/missing"); err == nil {
		t.Fatal("missing item found")
	}
	if diff := cmp.Diff([]string{"before", "after", "before", "after"}, h.calls); diff != "" {
		t.Errorf("hook calls (-want +got):\n%s", diff)
	}
	want := Event{Op: OpStat, Owner: owner, Path: "docs/a.txt"}
	if diff := cmp.Diff(want, h.events[0]); diff != "" {
		t.Errorf("event (-want +got):\n%s", diff)
	}
	if h.afterErrs[0] != nil || apperr.KindOf(h.afterErrs[1]) != apperr.NotFound {
		t.Errorf("After errors = %v, want nil then not_found", h.afterErrs)
	}
}

func TestBeforeHookVetoes(t *testing.T) {
	veto := apperr.New(apperr.InvalidRequest, "not today")
	h := &recordingHooks{vetoErr: veto}
	s, _ := newService(t, h, map[string]string{"a.txt": "x"})
	_, err := s.Stat(t.Context(), owner, "/a.txt")
	if !errors.Is(err, veto) {
		t.Errorf("Stat = %v, want the veto", err)
	}
	if diff := cmp.Diff([]string{"before"}, h.calls); diff != "" {
		t.Errorf("hook calls (-want +got):\n%s", diff)
	}
}

// TestHooksNeverSeeInvalidPaths: paths are resolved before the hooks run,
// so a hook only ever sees a clean path inside the namespace.
func TestHooksNeverSeeInvalidPaths(t *testing.T) {
	h := &recordingHooks{}
	s, _ := newService(t, h, nil)
	for _, p := range []string{"/../x", `/a\b`, "nope"} {
		_, _ = s.Stat(t.Context(), owner, p)
	}
	if len(h.calls) != 0 {
		t.Errorf("hooks ran for invalid paths: %v %v", h.calls, h.events)
	}
}

func TestFSError(t *testing.T) {
	tests := []struct {
		err  error
		want apperr.Kind
	}{
		{os.ErrNotExist, apperr.NotFound},
		{os.ErrExist, apperr.Conflict},
		{os.ErrPermission, apperr.Internal},
		{errors.New("disk on fire"), apperr.Internal},
	}
	for _, tt := range tests {
		err := fsError(tt.err, "/docs/a.txt")
		if k := apperr.KindOf(err); k != tt.want {
			t.Errorf("fsError(%v) kind = %s, want %s", tt.err, k, tt.want)
		}
		if !errors.Is(err, tt.err) {
			t.Errorf("fsError(%v) lost the cause", tt.err)
		}
	}
	if p := apperr.ProblemFor(fsError(os.ErrPermission, "/x"), "id"); p.Detail == "" || p.Code != "internal" {
		t.Errorf("internal storage errors must give a generic problem, got %+v", p)
	}
}

func TestNopHooks(t *testing.T) {
	var h NopHooks
	if err := h.Before(t.Context(), Event{}); err != nil {
		t.Error(err)
	}
	h.After(t.Context(), Event{}, nil)
}
