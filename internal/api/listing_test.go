package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/KhizirFarrukh/local-ai-nas/internal/api/gen"
	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

// TestListingTotalAndOffset covers the S02.3-T02 additions over the real
// server: every page says the folder's total, a page can start at an offset
// (the GUI's virtual list loads pages where the user scrolls), and an offset
// past the end gives an empty page, not an error.
func TestListingTotalAndOffset(t *testing.T) {
	svc, dir := testFilesDir(t)
	names := map[string]string{}
	for i := 0; i < 25; i++ {
		names[fmt.Sprintf("big/f%02d.txt", i)] = strings.Repeat("x", i)
	}
	if err := testutil.WriteFiles(dir, names); err != nil {
		t.Fatal(err)
	}
	c := client{t, testutil.NewServer(t, New(Options{Files: svc})).URL}
	page := func(query string) (int, gen.ItemsResponse, []byte) {
		t.Helper()
		status, _, body := c.do(http.MethodGet, "/api/v1/files/items?path=/big&"+query, "", nil, nil)
		var r gen.ItemsResponse
		if status == http.StatusOK {
			if err := json.Unmarshal(body, &r); err != nil {
				t.Fatal(err)
			}
		}
		return status, r, body
	}
	firstNames := func(r gen.ItemsResponse) string {
		var out []string
		if r.Items != nil {
			for _, it := range *r.Items {
				out = append(out, it.Name)
			}
		}
		return strings.Join(out, " ")
	}

	_, r, _ := page("limit=10")
	if r.Total == nil || *r.Total != 25 || firstNames(r) != "f00.txt f01.txt f02.txt f03.txt f04.txt f05.txt f06.txt f07.txt f08.txt f09.txt" {
		t.Errorf("first page: total %v, items %q", r.Total, firstNames(r))
	}
	_, r, _ = page("limit=3&offset=20")
	if *r.Total != 25 || firstNames(r) != "f20.txt f21.txt f22.txt" || r.NextCursor == nil {
		t.Errorf("offset 20: total %v, items %q, next %v", *r.Total, firstNames(r), r.NextCursor)
	}
	// The cursor from an offset page continues after it.
	_, r2, _ := page("limit=10&cursor=" + q(*r.NextCursor))
	if firstNames(r2) != "f23.txt f24.txt" || r2.NextCursor != nil {
		t.Errorf("after the offset page: %q, next %v", firstNames(r2), r2.NextCursor)
	}
	// Sorted by size, descending: offset follows the sort.
	_, r, _ = page("limit=2&offset=1&sort=size&order=desc")
	if firstNames(r) != "f23.txt f22.txt" {
		t.Errorf("offset in size order: %q", firstNames(r))
	}
	// Past the end: an empty page with the total.
	_, r, _ = page("offset=500")
	if *r.Total != 25 || firstNames(r) != "" || r.NextCursor != nil {
		t.Errorf("past the end: total %v, items %q", *r.Total, firstNames(r))
	}
	// Refusals.
	for _, bad := range []string{"offset=-1", "offset=2&cursor=abc"} {
		if status, _, body := page(bad); status != http.StatusBadRequest || !strings.Contains(string(body), "invalid_request") {
			t.Errorf("%s: %d %s", bad, status, body)
		}
	}
}
