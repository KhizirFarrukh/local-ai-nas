package api

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/KhizirFarrukh/local-ai-nas/internal/db"
	"github.com/KhizirFarrukh/local-ai-nas/internal/files"
	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
	"github.com/KhizirFarrukh/local-ai-nas/internal/uploads"
)

// fixtureFiles is the content of the files area in API tests.
var fixtureFiles = map[string]string{
	"docs/a.txt":  "hello",
	"docs/b.txt":  "hi",
	"readme.md":   "# readme",
	"empty/.keep": "",
}

// testFiles returns a real files service on a fresh storage root holding
// fixtureFiles in the default namespace.
func testFiles(t testing.TB) files.Service {
	t.Helper()
	svc, _ := testFilesDir(t)
	return svc
}

// testFilesDir is testFiles that also returns the namespace's directory,
// so a test can check the disk.
func testFilesDir(t testing.TB) (*files.Local, string) {
	t.Helper()
	l := storage.NewLayout(testutil.StorageRoot(t), storage.Options{})
	if _, err := l.Init(); err != nil {
		t.Fatal(err)
	}
	dir := l.Area(storage.FilesArea, storage.DefaultNamespace)
	if err := testutil.WriteFiles(dir, fixtureFiles); err != nil {
		t.Fatal(err)
	}
	return files.NewLocal(storage.NewResolver(l), files.Options{}), dir
}

// testUploads returns a real tus server on a fresh upload directory and
// database, mounted at UploadsPath, finishing uploads into svc.
func testUploads(t testing.TB, svc uploads.Target) http.Handler {
	t.Helper()
	d, err := db.Open(context.Background(), filepath.Join(t.TempDir(), db.FileName))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = d.Close() })
	if _, err := d.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}
	s, err := uploads.New(uploads.Options{Dir: t.TempDir(), DB: d, Files: svc, Namespace: storage.DefaultNamespace, BasePath: UploadsPath})
	if err != nil {
		t.Fatal(err)
	}
	return s
}
