package files

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
)

// listAll pages through a folder and returns the names in order and the
// number of pages.
func listAll(t *testing.T, s *Local, path string, opts ListOptions) ([]string, int) {
	t.Helper()
	var names []string
	pages := 0
	for {
		page, err := s.List(t.Context(), owner, path, opts)
		if err != nil {
			t.Fatalf("List(%q, %+v): %v", path, opts, err)
		}
		pages++
		for _, it := range page.Items {
			names = append(names, it.Name)
		}
		if page.NextCursor == "" {
			return names, pages
		}
		opts.Cursor = page.NextCursor
		if pages > 100000 {
			t.Fatal("the listing does not end")
		}
	}
}

// TestListLargeFolderPaginates is the S01.3-T02 acceptance test: a
// 10,000-entry folder pages without duplicates or gaps in every sort order,
// and the pages concatenate to the fully sorted listing.
func TestListLargeFolderPaginates(t *testing.T) {
	if testing.Short() {
		t.Skip("creates 10,000 files")
	}
	s, l := newService(t, nil, nil)
	dir := filepath.Join(l.Area(storage.FilesArea, owner), "big")
	if err := os.Mkdir(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	const n = 10000
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range n {
		// Repeating sizes and times make ties, so the tie-break matters;
		// names in mixed case check the case-insensitive name order.
		name := fmt.Sprintf("File-%05d.txt", (i*7919)%n)
		if i%3 == 0 {
			name = strings.ToLower(name)
		}
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(strings.Repeat("x", i%97)), 0o600); err != nil {
			t.Fatal(err)
		}
		mt := base.Add(time.Duration(i%50) * time.Minute)
		if err := os.Chtimes(p, mt, mt); err != nil {
			t.Fatal(err)
		}
	}
	for i := range 20 { // some folders, for the kind order
		if err := os.Mkdir(filepath.Join(dir, fmt.Sprintf("folder-%02d", i)), 0o750); err != nil {
			t.Fatal(err)
		}
	}

	// The reference: every entry with its file information, read from disk
	// directly (not through the service).
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	ref := map[string]Item{}
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			t.Fatal(err)
		}
		ref[e.Name()] = NewItem(owner, "big/"+e.Name(), info)
	}

	for _, key := range []SortKey{SortName, SortSize, SortModTime, SortKind} {
		for _, order := range []Order{Asc, Desc} {
			t.Run(string(key)+"_"+string(order), func(t *testing.T) {
				got, pages := listAll(t, s, "/big", ListOptions{Limit: 333, Sort: key, Order: order})
				if len(got) != len(ref) {
					t.Fatalf("got %d names, want %d", len(got), len(ref))
				}
				seen := map[string]bool{}
				for _, name := range got {
					if seen[name] {
						t.Fatalf("duplicate %q", name)
					}
					if _, ok := ref[name]; !ok {
						t.Fatalf("unknown name %q", name)
					}
					seen[name] = true
				}
				if want := (len(ref) + 332) / 333; pages != want {
					t.Errorf("pages = %d, want %d", pages, want)
				}
				less := compareItems(key, order)
				for i := 1; i < len(got); i++ {
					if less(ref[got[i-1]], ref[got[i]]) >= 0 {
						t.Fatalf("out of order at %d: %q before %q", i, got[i-1], got[i])
					}
				}
			})
		}
	}
}

func TestListOrders(t *testing.T) {
	// Case-insensitive name order puts "banana" between "Apple" and
	// "Cherry"; byte order would put it last. (No two names differ only in
	// case: Windows cannot store both.)
	s, l := newService(t, nil, map[string]string{
		"banana.txt": "22", "Apple.txt": "1", "Cherry.txt": "333", "date.txt": "4444",
	})
	if err := os.Mkdir(filepath.Join(l.Area(storage.FilesArea, owner), "zeta"), 0o750); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		opts ListOptions
		want []string
	}{
		{ListOptions{}, []string{"Apple.txt", "banana.txt", "Cherry.txt", "date.txt", "zeta"}},
		{ListOptions{Order: Desc}, []string{"zeta", "date.txt", "Cherry.txt", "banana.txt", "Apple.txt"}},
		{ListOptions{Sort: SortSize}, []string{"zeta", "Apple.txt", "banana.txt", "Cherry.txt", "date.txt"}},
		{ListOptions{Sort: SortKind}, []string{"zeta", "Apple.txt", "banana.txt", "Cherry.txt", "date.txt"}},
		{ListOptions{Sort: SortKind, Order: Desc}, []string{"date.txt", "Cherry.txt", "banana.txt", "Apple.txt", "zeta"}},
	}
	for _, tt := range tests {
		got, _ := listAll(t, s, "/", tt.opts)
		if diff := cmp.Diff(tt.want, got); diff != "" {
			t.Errorf("List(%+v) (-want +got):\n%s", tt.opts, diff)
		}
	}
	if compareNames("A", "a") >= 0 || compareNames("a", "B") >= 0 {
		t.Error("compareNames: want exact order within a case-insensitive tie, and case-insensitive order otherwise")
	}
}

func TestListPageAndCursor(t *testing.T) {
	s, l := newService(t, nil, map[string]string{"a": "", "b": "", "c": "", "d": "", "e": ""})
	page, err := s.List(t.Context(), owner, "/", ListOptions{Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 2 || page.NextCursor == "" || page.Folder.RelPath != "." || page.Folder.Kind != KindDir {
		t.Fatalf("first page = %+v", page)
	}
	// The folder changes between pages: an item already shown is removed
	// and a new one appears after the cursor. The cursor is a position, so
	// the next page continues correctly.
	ns := l.Area(storage.FilesArea, owner)
	if err := os.Remove(filepath.Join(ns, "a")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ns, "bb"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	next, err := s.List(t.Context(), owner, "/", ListOptions{Limit: 2, Cursor: page.NextCursor})
	if err != nil {
		t.Fatal(err)
	}
	if got := []string{next.Items[0].Name, next.Items[1].Name}; !slices.Equal(got, []string{"bb", "c"}) {
		t.Errorf("second page = %v, want [bb c]", got)
	}
	// The last page has no cursor.
	last, err := s.List(t.Context(), owner, "/", ListOptions{Limit: 2, Cursor: next.NextCursor})
	if err != nil || len(last.Items) != 2 || last.NextCursor != "" {
		t.Errorf("last page = %+v, %v", last, err)
	}
	// An empty folder.
	s2, _ := newService(t, nil, nil)
	empty, err := s2.List(t.Context(), owner, "/", ListOptions{})
	if err != nil || len(empty.Items) != 0 || empty.NextCursor != "" {
		t.Errorf("empty folder = %+v, %v", empty, err)
	}
}

func TestListErrors(t *testing.T) {
	s, _ := newService(t, nil, map[string]string{"a.txt": "x", "b": "", "c": ""})
	firstPage, err := s.List(t.Context(), owner, "/", ListOptions{Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		path string
		opts ListOptions
		want apperr.Kind
	}{
		{"limit too small", "/", ListOptions{Limit: -1}, apperr.InvalidRequest},
		{"limit too large", "/", ListOptions{Limit: MaxLimit + 1}, apperr.InvalidRequest},
		{"unknown sort", "/", ListOptions{Sort: "color"}, apperr.InvalidRequest},
		{"unknown order", "/", ListOptions{Order: "up"}, apperr.InvalidRequest},
		{"garbage cursor", "/", ListOptions{Cursor: "%%%"}, apperr.InvalidRequest},
		{"cursor of another sort", "/", ListOptions{Limit: 1, Sort: SortSize, Cursor: firstPage.NextCursor}, apperr.InvalidRequest},
		{"a file", "/a.txt", ListOptions{}, apperr.InvalidRequest},
		{"missing", "/missing", ListOptions{}, apperr.NotFound},
		{"outside", "/../u0002", ListOptions{}, apperr.OutsideRoot},
	}
	for _, tt := range tests {
		_, err := s.List(t.Context(), owner, tt.path, tt.opts)
		if k := apperr.KindOf(err); err == nil || k != tt.want {
			t.Errorf("%s: List = %v (kind %s), want %s", tt.name, err, k, tt.want)
		}
	}
}
