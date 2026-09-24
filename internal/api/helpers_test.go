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
	l := storage.NewLayout(testutil.StorageRoot(t), storage.Options{})
	if _, err := l.Init(); err != nil {
		t.Fatal(err)
	}
	if err := testutil.WriteFiles(l.Area(storage.FilesArea, storage.DefaultNamespace), fixtureFiles); err != nil {
		t.Fatal(err)
	}
	return files.NewLocal(storage.NewResolver(l), nil)
}
