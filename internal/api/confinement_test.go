package api

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/KhizirFarrukh/local-ai-nas/internal/files"
	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

// secret is the content of every file the caller must not reach.
const secret = "SECRET outside the namespace"

// snapshotOutside returns every file and folder under root except the
// caller's namespace, with the content of files.
func snapshotOutside(t *testing.T, root, namespace string) map[string]string {
	t.Helper()
	snap := map[string]string{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == namespace {
			return filepath.SkipDir
		}
		rel, _ := filepath.Rel(root, p)
		switch {
		case d.IsDir():
			snap[filepath.ToSlash(rel)+"/"] = ""
		case d.Type().IsRegular():
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			snap[filepath.ToSlash(rel)] = string(b)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return snap
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// TestNoOperationLeavesTheNamespace is the S01.3 acceptance test (plan
// S01.3, criterion 4): every endpoint, given paths that try to leave the
// caller's namespace, answers 4xx, returns nothing from outside, and
// changes nothing outside.
func TestNoOperationLeavesTheNamespace(t *testing.T) {
	root := testutil.StorageRoot(t)
	l := storage.NewLayout(root, storage.Options{})
	if _, err := l.Init(); err != nil {
		t.Fatal(err)
	}
	ns := l.Area(storage.FilesArea, storage.DefaultNamespace)
	if err := testutil.WriteFiles(ns, fixtureFiles); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{
		filepath.Join(l.Area(storage.PhotosArea, storage.DefaultNamespace), "secret.jpg"),
		filepath.Join(root, "files", "u0002", "other.txt"),
		filepath.Join(root, ".local-ai-nas", "db", "secret.db"),
		filepath.Join(root, "root-secret.txt"),
	} {
		if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(secret), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	outside := filepath.Join(t.TempDir(), "outside.txt")
	if err := os.WriteFile(outside, []byte(secret), 0o600); err != nil {
		t.Fatal(err)
	}
	photos := l.Area(storage.PhotosArea, storage.DefaultNamespace)
	links := true
	for name, target := range map[string]string{"escape-abs": photos, "escape-rel": filepath.FromSlash("../../photos/u0001"), "escape-file": outside} {
		links = links && testutil.TrySymlink(t, target, filepath.Join(ns, name))
	}
	before := snapshotOutside(t, root, ns)

	escapes := []string{
		"/..", "/../u0002/other.txt", "/../../photos/u0001/secret.jpg", "/docs/../../../root-secret.txt",
		"/../../.local-ai-nas/db/secret.db", "/..%2Fu0002%2Fother.txt", `/..\u0002\other.txt`, `\..\x`,
		"//server/share/x", "/C:/Windows/win.ini", "/c:", "/" + filepath.ToSlash(outside),
		"/. ./x", "/.../x", "/docs/. /x", "/../u0002", "/docs/%00", "/docs/a.txt\x00.jpg",
	}
	// Paths through a link to the outside (os.Root refuses them).
	var linkItems []string
	if links {
		escapes = append(escapes, "/escape-abs/secret.jpg", "/escape-rel/secret.jpg", "/escape-file/x",
			"/escape-abs/new", "/escape-rel/sub/new")
		linkItems = []string{"/escape-abs", "/escape-rel", "/escape-file"}
	}
	svc := files.NewLocal(storage.NewResolver(l), files.Options{})
	c := client{t, testutil.NewServer(t, New(Options{Files: svc})).URL}
	type request struct {
		method, path, body string
		anyStatus          bool // a request on a link item itself: any answer but 5xx
	}
	// every aims each endpoint at p; created or moved items go to target.
	every := func(p, target string, anyStatus bool) []request {
		return []request{
			{"GET", "/api/v1/files/items?path=" + q(p), "", anyStatus},
			{"GET", "/api/v1/files/content?path=" + q(p), "", anyStatus},
			{"PUT", "/api/v1/files/content?path=" + q(p) + "&on_conflict=overwrite", "EVIL", anyStatus},
			{"POST", "/api/v1/files/folders", `{"path":` + jsonString(p) + `,"parents":true}`, anyStatus},
			{"POST", "/api/v1/files/operations/copy", `{"from":` + jsonString(p) + `,"to":` + jsonString(target) + `}`, anyStatus},
			{"POST", "/api/v1/files/operations/copy", `{"from":"/readme.md","to":` + jsonString(p) + `,"on_conflict":"overwrite"}`, anyStatus},
			{"POST", "/api/v1/files/operations/move", `{"from":"/docs/b.txt","to":` + jsonString(p) + `,"on_conflict":"overwrite"}`, anyStatus},
			{"POST", "/api/v1/files/operations/rename", `{"path":` + jsonString(p) + `,"new_name":` + jsonString(target[1:]) + `}`, anyStatus},
			{"POST", "/api/v1/files/operations/move", `{"from":` + jsonString(p) + `,"to":` + jsonString(target) + `}`, anyStatus},
			{"DELETE", "/api/v1/files/items?path=" + q(p) + "&recursive=true", "", anyStatus},
		}
	}
	var requests []request
	for _, e := range escapes {
		requests = append(requests, every(e, "/stolen", false)...)
	}
	for _, name := range []string{"..", "../x", `..\x`, "a/../../x", "/x"} {
		requests = append(requests, request{"POST", "/api/v1/files/operations/rename", `{"path":"/readme.md","new_name":` + jsonString(name) + `}`, false})
	}
	// Link items are items of the namespace: they may be listed, renamed,
	// or deleted, but no operation may reach what they point to.
	for i, p := range linkItems {
		requests = append(requests, every(p, fmt.Sprintf("/link-op-%d", i), true)...)
	}
	for _, r := range requests {
		contentType := "application/json"
		if r.method == "PUT" {
			contentType = "application/octet-stream"
		}
		var body []byte
		if r.body != "" {
			body = []byte(r.body)
		}
		st, _, b := c.do(r.method, r.path, contentType, body, nil)
		if st >= 500 || !r.anyStatus && st < 400 {
			t.Errorf("%s %s %s → %d %s, want 4xx", r.method, r.path, r.body, st, b)
		}
		if strings.Contains(string(b), secret) {
			t.Errorf("%s %s %s returned content from outside the namespace", r.method, r.path, r.body)
		}
	}
	if diff := cmp.Diff(before, snapshotOutside(t, root, ns)); diff != "" {
		t.Errorf("something outside the namespace changed (-before +after):\n%s", diff)
	}
	if b, err := os.ReadFile(outside); err != nil || string(b) != secret {
		t.Errorf("the file outside the storage root changed: %q, %v", b, err)
	}
	if _, err := os.Lstat(filepath.Join(ns, "stolen")); !os.IsNotExist(err) {
		t.Errorf("something was stolen into the namespace: %v", err)
	}
}
