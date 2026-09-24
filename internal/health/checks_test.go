package health

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KhizirFarrukh/local-ai-nas/internal/db"
	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

// setupLayout returns an initialized layout in a new storage root.
func setupLayout(t *testing.T) storage.Layout {
	t.Helper()
	l := storage.NewLayout(testutil.StorageRoot(t), storage.Options{})
	if _, err := l.Init(); err != nil {
		t.Fatal(err)
	}
	return l
}

// runOne runs a single check and returns its result.
func runOne(t *testing.T, c Check) Result {
	t.Helper()
	rep := Run(t.Context(), "dev", []Check{c})
	if len(rep.Checks) != 1 || rep.Checks[0].Name != c.Name {
		t.Fatalf("report = %+v", rep)
	}
	return rep.Checks[0]
}

func TestConfigCheck(t *testing.T) {
	if r := runOne(t, Config("/etc/local-ai-nas/config.toml")); r.Status != "ok" || !strings.Contains(r.Detail, "config.toml") {
		t.Errorf("config with a file = %+v", r)
	}
	if r := runOne(t, Config("")); r.Status != "ok" || !strings.Contains(r.Detail, "no config file") {
		t.Errorf("config without a file = %+v", r)
	}
}

func TestStorageWritableCheck(t *testing.T) {
	l := setupLayout(t)
	if r := runOne(t, StorageWritable(l)); r.Status != "ok" {
		t.Errorf("healthy layout = %+v", r)
	}
	missing := storage.NewLayout(filepath.Join(testutil.StorageRoot(t), "never-created"), storage.Options{})
	if r := runOne(t, StorageWritable(missing)); r.Status != "fail" || r.Name != "storage_writable" {
		t.Errorf("missing layout = %+v, want storage_writable to fail", r)
	}
}

func TestSameFilesystemCheck(t *testing.T) {
	l := setupLayout(t)

	// The real device IDs: a fresh temp root is on one file system.
	if r := runOne(t, SameFilesystem(l, nil)); r.Status != "ok" || r.Detail != "upload_finalize_mode=rename" {
		t.Errorf("real devices = %+v", r)
	}

	// Simulated split: tmp/uploads on one device, the files area on another.
	split := func(path string) (uint64, error) {
		if strings.Contains(path, filepath.Join("tmp", "uploads")) {
			return 1, nil
		}
		return 2, nil
	}
	r := runOne(t, SameFilesystem(l, split))
	if r.Status != "warn" || !strings.Contains(r.Detail, "upload_finalize_mode=copy") {
		t.Errorf("split devices = %+v, want a warning with upload_finalize_mode=copy", r)
	}

	failing := func(string) (uint64, error) { return 0, errors.New("stat failed") }
	if r := runOne(t, SameFilesystem(l, failing)); r.Status != "fail" || r.Error != "stat failed" {
		t.Errorf("device error = %+v", r)
	}
}

func TestFreeSpaceCheck(t *testing.T) {
	const gib = 1 << 30
	fixed := func(n uint64) storage.FreeFunc { return func(string) (uint64, error) { return n, nil } }

	if r := runOne(t, FreeSpace(storage.NewSpaceGuard("x", gib, fixed(5*gib)))); r.Status != "ok" || !strings.Contains(r.Detail, "reserve") {
		t.Errorf("enough space = %+v", r)
	}
	if r := runOne(t, FreeSpace(storage.NewSpaceGuard("x", gib, fixed(gib-1)))); r.Status != "fail" || !strings.Contains(r.Error, "below the reserve") {
		t.Errorf("low space = %+v, want free_space to fail", r)
	}
	broken := storage.NewSpaceGuard("x", gib, func(string) (uint64, error) { return 0, errors.New("statfs failed") })
	if r := runOne(t, FreeSpace(broken)); r.Status != "fail" {
		t.Errorf("provider error = %+v", r)
	}
}

func TestDatabaseCheck(t *testing.T) {
	d, err := db.Open(t.Context(), filepath.Join(t.TempDir(), "db", db.FileName))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = d.Close() }()

	if r := runOne(t, Database(d)); r.Status != "fail" || !strings.Contains(r.Error, "not applied") {
		t.Errorf("before migrating = %+v, want database to fail with pending migrations", r)
	}
	if _, err := d.Migrate(t.Context()); err != nil {
		t.Fatal(err)
	}
	if r := runOne(t, Database(d)); r.Status != "ok" || !strings.Contains(r.Detail, "migrations applied") {
		t.Errorf("after migrating = %+v", r)
	}
	_ = d.Close()
	if r := runOne(t, Database(d)); r.Status != "fail" {
		t.Errorf("closed database = %+v, want a failure", r)
	}
}
