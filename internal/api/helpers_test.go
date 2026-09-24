package api

import (
	"testing"

	"github.com/KhizirFarrukh/local-ai-nas/internal/files"
	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
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
func testFilesDir(t testing.TB) (files.Service, string) {
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
