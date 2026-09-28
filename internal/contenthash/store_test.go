package contenthash

import (
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/KhizirFarrukh/local-ai-nas/internal/db"
)

func newStore(t *testing.T) *Store {
	t.Helper()
	d, err := db.Open(t.Context(), filepath.Join(t.TempDir(), db.FileName))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = d.Close() })
	if _, err := d.Migrate(t.Context()); err != nil {
		t.Fatal(err)
	}
	return New(d)
}

// h returns a valid hash made of one repeated hex digit.
func h(c string) string { return "sha256:" + strings.Repeat(c, 64) }

// rows lists the paths of owner's rows, sorted.
func rows(t *testing.T, s *Store, owner string) []string {
	t.Helper()
	r, err := s.db.Read.QueryContext(t.Context(), `SELECT path, hash FROM content_hashes WHERE namespace = ?`, owner)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = r.Close() }()
	var out []string
	for r.Next() {
		var p, hash string
		if err := r.Scan(&p, &hash); err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(hash, "sha256:") {
			t.Errorf("%s: hash %q", p, hash)
		}
		out = append(out, p)
	}
	if err := r.Err(); err != nil {
		t.Fatal(err)
	}
	sort.Strings(out)
	return out
}

func TestRecordAndLookup(t *testing.T) {
	s := newStore(t)
	ctx := t.Context()
	if err := s.Record(ctx, "u0001", "a.txt", `"v1"`, h("1")); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		owner, rel, etag, want string
	}{
		{"u0001", "a.txt", `"v1"`, h("1")},
		{"u0001", "a.txt", `"v2"`, ""}, // another version: stale
		{"u0002", "a.txt", `"v1"`, ""}, // another owner
		{"u0001", "A.txt", `"v1"`, ""}, // paths are exact
	} {
		got, err := s.Lookup(ctx, c.owner, c.rel, c.etag)
		if err != nil || got != c.want {
			t.Errorf("Lookup(%s, %s, %s) = %q, %v; want %q", c.owner, c.rel, c.etag, got, err, c.want)
		}
	}
	// A new version replaces the row.
	if err := s.Record(ctx, "u0001", "a.txt", `"v2"`, h("2")); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.Lookup(ctx, "u0001", "a.txt", `"v2"`); got != h("2") {
		t.Errorf("after a new version: %q", got)
	}
	if got, _ := s.Lookup(ctx, "u0001", "a.txt", `"v1"`); got != "" {
		t.Errorf("the old version still answers: %q", got)
	}
	if err := s.Record(ctx, "u0001", "b.txt", `"v"`, "sha256:short"); err == nil {
		t.Error("a malformed hash was stored")
	}
}

func TestMovedAndDeleted(t *testing.T) {
	seed := []string{
		"a", "a/x.txt", "a/sub/y.txt", "ab/z.txt", "a_b/w.txt", "A/q.txt",
		"Fotos/été.jpg", "b.txt", "dest/old.txt", "dest",
	}
	tests := []struct {
		name string
		run  func(s *Store) error
		want []string
	}{
		{
			name: "file renamed",
			run:  func(s *Store) error { return s.Moved(t.Context(), "u0001", "b.txt", "c.txt") },
			want: []string{"A/q.txt", "Fotos/été.jpg", "a", "a/sub/y.txt", "a/x.txt", "a_b/w.txt", "ab/z.txt", "c.txt", "dest", "dest/old.txt"},
		},
		{
			// "ab", "a_b", and "A" are other folders: neither a prefix match
			// nor a case-insensitive one may move them.
			name: "folder moved with everything below",
			run:  func(s *Store) error { return s.Moved(t.Context(), "u0001", "a", "moved/a2") },
			want: []string{"A/q.txt", "Fotos/été.jpg", "a_b/w.txt", "ab/z.txt", "b.txt", "dest", "dest/old.txt", "moved/a2", "moved/a2/sub/y.txt", "moved/a2/x.txt"},
		},
		{
			name: "move replacing a target drops what it had",
			run:  func(s *Store) error { return s.Moved(t.Context(), "u0001", "b.txt", "dest") },
			want: []string{"A/q.txt", "Fotos/été.jpg", "a", "a/sub/y.txt", "a/x.txt", "a_b/w.txt", "ab/z.txt", "dest"},
		},
		{
			name: "non-ASCII folder renamed",
			run:  func(s *Store) error { return s.Moved(t.Context(), "u0001", "Fotos", "Fotos été") },
			want: []string{"A/q.txt", "Fotos été/été.jpg", "a", "a/sub/y.txt", "a/x.txt", "a_b/w.txt", "ab/z.txt", "b.txt", "dest", "dest/old.txt"},
		},
		{
			name: "folder deleted with everything below",
			run:  func(s *Store) error { return s.Deleted(t.Context(), "u0001", "a") },
			want: []string{"A/q.txt", "Fotos/été.jpg", "a_b/w.txt", "ab/z.txt", "b.txt", "dest", "dest/old.txt"},
		},
		{
			name: "file deleted",
			run:  func(s *Store) error { return s.Deleted(t.Context(), "u0001", "Fotos/été.jpg") },
			want: []string{"A/q.txt", "a", "a/sub/y.txt", "a/x.txt", "a_b/w.txt", "ab/z.txt", "b.txt", "dest", "dest/old.txt"},
		},
		{
			name: "root deleted: the namespace is empty",
			run:  func(s *Store) error { return s.Deleted(t.Context(), "u0001", ".") },
			want: nil,
		},
		{
			name: "moving onto itself changes nothing",
			run:  func(s *Store) error { return s.Moved(t.Context(), "u0001", "a", "a") },
			want: []string{"A/q.txt", "Fotos/été.jpg", "a", "a/sub/y.txt", "a/x.txt", "a_b/w.txt", "ab/z.txt", "b.txt", "dest", "dest/old.txt"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newStore(t)
			for _, p := range seed {
				for _, owner := range []string{"u0001", "u0002"} {
					if err := s.Record(t.Context(), owner, p, `"e"`, h("0")); err != nil {
						t.Fatal(err)
					}
				}
			}
			other := rows(t, s, "u0002")
			if err := tt.run(s); err != nil {
				t.Fatal(err)
			}
			got := rows(t, s, "u0001")
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("rows (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(other, rows(t, s, "u0002")); diff != "" {
				t.Errorf("another namespace changed (-before +after):\n%s", diff)
			}
		})
	}
}
