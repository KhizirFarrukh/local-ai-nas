package storage

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

// TestInitRejectsOverlaps covers every way internal data could overlap an
// area (invariant I2), one case per test. Validation runs before anything
// is created, so a rejected layout leaves the root untouched.
func TestInitRejectsOverlaps(t *testing.T) {
	tests := []struct {
		name string
		opts func(root string) Options
		want string
	}{
		{
			name: "database inside the files area",
			opts: func(root string) Options { return Options{DBDir: filepath.Join(root, "files", "u0001", "db")} },
			want: "the database directory (storage.db_dir) (%ROOT%/files/u0001/db) must not be inside the files area",
		},
		{
			name: "logs inside the photos area",
			opts: func(root string) Options { return Options{LogsDir: filepath.Join(root, "photos", "logs")} },
			want: "the logs directory (storage.logs_dir) (%ROOT%/photos/logs) must not be inside the photos area",
		},
		{
			name: "database is the files area",
			opts: func(root string) Options { return Options{DBDir: filepath.Join(root, "files")} },
			want: "the database directory (storage.db_dir) and the files area are the same directory",
		},
		{
			name: "logs are the photos area",
			opts: func(root string) Options { return Options{LogsDir: filepath.Join(root, "photos")} },
			want: "the logs directory (storage.logs_dir) and the photos area are the same directory",
		},
		{
			name: "an area inside the logs directory",
			opts: func(root string) Options { return Options{LogsDir: root} },
			want: "the logs directory (storage.logs_dir) (%ROOT%) must not contain the files area",
		},
		{
			name: "an area inside the database directory",
			opts: func(root string) Options { return Options{DBDir: filepath.Dir(root)} },
			want: "must not contain the files area",
		},
		{
			name: "database is the uploads directory",
			opts: func(root string) Options {
				return Options{DBDir: filepath.Join(root, ".local-ai-nas", "tmp", "uploads")}
			},
			want: "the database directory (storage.db_dir) and the uploads directory are the same directory",
		},
		{
			name: "logs contain the uploads directory",
			opts: func(root string) Options { return Options{LogsDir: filepath.Join(root, ".local-ai-nas", "tmp")} },
			want: "must not contain the uploads directory",
		},
		{
			name: "database elsewhere in the storage root",
			opts: func(root string) Options { return Options{DBDir: filepath.Join(root, "db")} },
			want: "is inside the storage root but not in .local-ai-nas",
		},
		{
			name: "relative database directory",
			opts: func(string) Options { return Options{DBDir: filepath.Join("relative", "db")} },
			want: "must be an absolute path",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := filepath.Join(testutil.StorageRoot(t), "nas")
			_, err := NewLayout(root, tt.opts(root)).Init()
			want := strings.ReplaceAll(filepath.FromSlash(tt.want), "%ROOT%", root)
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Errorf("Init error = %v\nwant it to contain: %s", err, want)
			}
			if _, statErr := os.Stat(root); !os.IsNotExist(statErr) {
				t.Errorf("a rejected layout created %s", root)
			}
		})
	}
}

func TestInitRelocatesDatabaseAndLogs(t *testing.T) {
	root := testutil.StorageRoot(t)
	elsewhere := testutil.StorageRoot(t)
	opts := Options{
		DBDir:   filepath.Join(elsewhere, "fast-disk", "db"),         // outside the root, parents missing
		LogsDir: filepath.Join(root, ".local-ai-nas", "custom-logs"), // inside the internal directory
	}
	l := NewLayout(root, opts)
	if l.DB != opts.DBDir || l.Logs != opts.LogsDir {
		t.Fatalf("layout DB=%q Logs=%q, want the configured paths", l.DB, l.Logs)
	}
	unknown, err := l.Init()
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if len(unknown) != 0 {
		t.Errorf("unknown = %v", unknown)
	}
	for _, dir := range []string{opts.DBDir, opts.LogsDir, l.TmpUploads} {
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			t.Errorf("%s was not created: %v", dir, err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, ".local-ai-nas", "db")); !os.IsNotExist(err) {
		t.Error("the default database directory was created although the database was moved")
	}
}

func TestWithin(t *testing.T) {
	sep := string(filepath.Separator)
	base := filepath.Join(testutil.StorageRoot(t), "nas")
	tests := []struct {
		dir, path string
		want      bool
	}{
		{base, base, true},
		{base, base + sep + "files", true},
		{base, base + "-other", false}, // a sibling with the same prefix
		{base + sep + "files", base, false},
		{filepath.VolumeName(base) + sep, base, true}, // the filesystem root contains everything
	}
	if runtime.GOOS == "windows" {
		tests = append(tests, struct {
			dir, path string
			want      bool
		}{strings.ToUpper(base), base + sep + "files", true}) // case-insensitive
	}
	for _, tt := range tests {
		if got := within(tt.dir, tt.path); got != tt.want {
			t.Errorf("within(%q, %q) = %v, want %v", tt.dir, tt.path, got, tt.want)
		}
	}
}
