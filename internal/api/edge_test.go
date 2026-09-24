package api

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/text/unicode/norm"

	"github.com/KhizirFarrukh/local-ai-nas/internal/api/gen"
	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

// listAll pages through a folder with the API and returns every name.
func listAll(t *testing.T, c client, path string, limit int) []string {
	t.Helper()
	var names []string
	cursor := ""
	for {
		q := "/api/v1/files/items?path=" + url.QueryEscape(path) + "&limit=" + strconv.Itoa(limit)
		if cursor != "" {
			q += "&cursor=" + url.QueryEscape(cursor)
		}
		st, _, b := c.do("GET", q, "", nil, nil)
		var r gen.ItemsResponse
		if st != http.StatusOK || json.Unmarshal(b, &r) != nil {
			t.Fatalf("list %s: %d %s", path, st, b)
		}
		if r.Items != nil {
			for _, it := range *r.Items {
				names = append(names, it.Name)
			}
		}
		if r.NextCursor == nil {
			return names
		}
		cursor = *r.NextCursor
	}
}

// TestEdgeUnicode: names in decomposed form, emoji, right-to-left scripts,
// and combining marks work end to end, and one visible name is one file.
func TestEdgeUnicode(t *testing.T) {
	c, area := attackServer(t)
	names := []string{
		"café résumé.txt",  // NFD, as macOS sends it
		"🎉 party 👨‍👩‍👧‍👦.txt", // emoji, with a ZWJ sequence
		"מסמך חשוב.pdf",       // Hebrew (right to left)
		"مرحبا بالعالم.txt",   // Arabic (right to left)
		"Ångström ø ß ﬁ.txt",  // Latin with a ligature
		"ź̧̈alg̶o.txt",       // stacked combining marks
		"日本語のファイル.txt",
	}
	for _, name := range names {
		nfc := norm.NFC.String(name)
		st, _, b := c.do("PUT", "/api/v1/files/content?path="+url.QueryEscape("/"+name), "", []byte(name), nil)
		var it gen.FileItem
		if st != http.StatusCreated || json.Unmarshal(b, &it) != nil || it.Name != nfc {
			t.Errorf("upload %q: %d %s, want the NFC name", name, st, b)
			continue
		}
		// Stored under the NFC name, and reachable by both forms.
		if _, err := os.Stat(filepath.Join(area, nfc)); err != nil {
			t.Errorf("%q is not stored under its NFC name: %v", name, err)
		}
		for _, form := range []string{name, nfc, norm.NFD.String(name)} {
			st, hdr, body := c.do("GET", "/api/v1/files/content?path="+url.QueryEscape("/"+form), "", nil, nil)
			_, params, err := mime.ParseMediaType(hdr.Get("Content-Disposition"))
			if st != http.StatusOK || string(body) != name || err != nil || params["filename"] != nfc {
				t.Errorf("download %q by %q: %d, Content-Disposition %q", name, form, st, hdr.Get("Content-Disposition"))
			}
		}
		// A second upload in the other form is the same name: a conflict.
		if st, _, _ := c.do("PUT", "/api/v1/files/content?path="+url.QueryEscape("/"+norm.NFD.String(name)), "", []byte("x"), nil); st != http.StatusConflict {
			t.Errorf("the NFD form of %q is a new file (%d), want 409", name, st)
		}
		js, _ := json.Marshal("/" + nfc)
		copyTo, _ := json.Marshal("/copy of " + name)
		if st, _, b := c.do("POST", "/api/v1/files/operations/copy", "application/json", []byte(`{"from":`+string(js)+`,"to":`+string(copyTo)+`}`), nil); st != http.StatusCreated {
			t.Errorf("copy %q: %d %s", name, st, b)
		}
		newName, _ := json.Marshal("renamed " + name)
		if st, _, b := c.do("POST", "/api/v1/files/operations/rename", "application/json", []byte(`{"path":`+string(js)+`,"new_name":`+string(newName)+`}`), nil); st != http.StatusOK {
			t.Errorf("rename %q: %d %s", name, st, b)
		}
		if st, _, b := c.do("DELETE", "/api/v1/files/items?path="+url.QueryEscape("/renamed "+name), "", nil, nil); st != http.StatusNoContent {
			t.Errorf("delete %q: %d %s", name, st, b)
		}
	}
	listed := strings.Join(listAll(t, c, "/", 100), "|")
	for _, name := range names {
		if !strings.Contains(listed, "copy of "+norm.NFC.String(name)) {
			t.Errorf("the listing lacks the copy of %q", name)
		}
	}
}

// TestEdgeEmpty: empty files and folders behave like any other.
func TestEdgeEmpty(t *testing.T) {
	c, _ := attackServer(t)
	if st, _, b := c.do("PUT", "/api/v1/files/content?path=/empty.txt", "", nil, nil); st != http.StatusCreated {
		t.Fatalf("empty upload: %d %s", st, b)
	}
	st, hdr, b := c.do("POST", UploadsPath, "", nil, map[string]string{"Tus-Resumable": "1.0.0", "Upload-Length": "0", "Upload-Metadata": "target_path " + b64("/empty-tus.txt")})
	if st != http.StatusCreated || hdr.Get("Item-Path") != "/empty-tus.txt" {
		t.Errorf("empty tus upload: %d %q %s", st, hdr.Get("Item-Path"), b)
	}
	for _, p := range []string{"/empty.txt", "/empty-tus.txt"} {
		st, hdr, b := c.do("GET", "/api/v1/files/content?path="+p, "", nil, nil)
		if st != http.StatusOK || len(b) != 0 || hdr.Get("Content-Length") != "0" {
			t.Errorf("download %s: %d, %d bytes, Content-Length %q", p, st, len(b), hdr.Get("Content-Length"))
		}
		// A range over an empty file is ignored: 200, and no Content-Range
		// (Go's own answer to a suffix range here is a 206 with the
		// invalid "bytes 0--1/0").
		for _, r := range []string{"bytes=0-", "bytes=0-0", "bytes=-1", "bytes=-0"} {
			st, hdr, b := c.do("GET", "/api/v1/files/content?path="+p, "", nil, map[string]string{"Range": r})
			if st != http.StatusOK || hdr.Get("Content-Range") != "" || len(b) != 0 {
				t.Errorf("range %s of an empty file: %d, Content-Range %q, %d bytes", r, st, hdr.Get("Content-Range"), len(b))
			}
		}
	}
	if st, _, b := c.do("POST", "/api/v1/files/operations/copy", "application/json", []byte(`{"from":"/empty.txt","to":"/empty-copy.txt"}`), nil); st != http.StatusCreated {
		t.Errorf("copy of an empty file: %d %s", st, b)
	}
	if st, _, b := c.do("POST", "/api/v1/files/folders", "application/json", []byte(`{"path":"/hollow"}`), nil); st != http.StatusCreated {
		t.Fatalf("empty folder: %d %s", st, b)
	}
	if names := listAll(t, c, "/hollow", 10); len(names) != 0 {
		t.Errorf("the empty folder lists %v", names)
	}
	if st, _, b := c.do("DELETE", "/api/v1/files/items?path=/hollow", "", nil, nil); st != http.StatusNoContent {
		t.Errorf("delete an empty folder: %d %s", st, b)
	}
}

// TestEdgeSparseFile: a multi-GB sparse file (Linux only: on Windows,
// truncate does not make a sparse file, so this would write gigabytes)
// has the right details, and its end can be read without reading the
// rest.
func TestEdgeSparseFile(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("sparse files: Linux only (truncate on Windows allocates the whole file)")
	}
	c, area := attackServer(t)
	const size = 5 << 30
	f, err := os.Create(filepath.Join(area, "sparse.bin"))
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(size); err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteAt([]byte("END"), size-3); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	st, _, b := c.do("GET", "/api/v1/files/items?path=/sparse.bin", "", nil, nil)
	var r gen.ItemsResponse
	if st != http.StatusOK || json.Unmarshal(b, &r) != nil || r.Item.Size != size {
		t.Errorf("details: %d %s", st, b)
	}
	st, hdr, b := c.do("GET", "/api/v1/files/content?path=/sparse.bin", "", nil, map[string]string{"Range": "bytes=-3"})
	if st != http.StatusPartialContent || string(b) != "END" || hdr.Get("Content-Range") != fmt.Sprintf("bytes %d-%d/%d", size-3, size-1, size) {
		t.Errorf("the last 3 bytes: %d %q %s", st, b, hdr.Get("Content-Range"))
	}
	st, hdr, _ = c.do("HEAD", "/api/v1/files/content?path=/sparse.bin", "", nil, nil)
	if st != http.StatusOK || hdr.Get("Content-Length") != strconv.Itoa(size) {
		t.Errorf("HEAD: %d, Content-Length %q", st, hdr.Get("Content-Length"))
	}
	if st, _, b := c.do("DELETE", "/api/v1/files/items?path=/sparse.bin", "", nil, nil); st != http.StatusNoContent {
		t.Errorf("delete: %d %s", st, b)
	}
}

// TestEdgeDeepNesting: a tree 100 levels deep can be made, read, moved,
// copied, and deleted.
func TestEdgeDeepNesting(t *testing.T) {
	c, _ := attackServer(t)
	deep := strings.Repeat("/level", 100)
	js, _ := json.Marshal("/deep" + deep)
	if st, _, b := c.do("POST", "/api/v1/files/folders", "application/json", []byte(`{"path":`+string(js)+`,"parents":true}`), nil); st != http.StatusCreated {
		t.Fatalf("make 100 levels: %d %s", st, b)
	}
	bottom := "/deep" + deep + "/bottom.txt"
	if st, _, b := c.do("PUT", "/api/v1/files/content?path="+url.QueryEscape(bottom), "", []byte("deep"), nil); st != http.StatusCreated {
		t.Fatalf("upload at the bottom: %d %s", st, b)
	}
	if names := listAll(t, c, "/deep"+deep, 10); len(names) != 1 || names[0] != "bottom.txt" {
		t.Errorf("the bottom lists %v", names)
	}
	if st, _, b := c.do("POST", "/api/v1/files/operations/move", "application/json", []byte(`{"from":"/deep","to":"/moved"}`), nil); st != http.StatusOK {
		t.Fatalf("move the tree: %d %s", st, b)
	}
	if st, _, b := c.do("POST", "/api/v1/files/operations/copy", "application/json", []byte(`{"from":"/moved","to":"/copied"}`), nil); st != http.StatusCreated {
		t.Fatalf("copy the tree: %d %s", st, b)
	}
	if st, _, b := c.do("GET", "/api/v1/files/content?path="+url.QueryEscape("/copied"+deep+"/bottom.txt"), "", nil, nil); st != http.StatusOK || string(b) != "deep" {
		t.Errorf("the copied bottom file: %d %q", st, b)
	}
	for _, p := range []string{"/moved", "/copied"} {
		if st, _, b := c.do("DELETE", "/api/v1/files/items?path="+p+"&recursive=true", "", nil, nil); st != http.StatusNoContent {
			t.Errorf("delete %s: %d %s", p, st, b)
		}
	}
}

// TestEdgeLargeFolder: a folder of 10,000 files pages through the API
// without duplicates or gaps and is deleted in one request.
func TestEdgeLargeFolder(t *testing.T) {
	if testing.Short() {
		t.Skip("10,000 files: not in -short")
	}
	c, area := attackServer(t)
	const n = 10000
	files := make(map[string]string, n)
	for i := range n {
		files[fmt.Sprintf("big/f%05d.txt", i)] = ""
	}
	if err := testutil.WriteFiles(area, files); err != nil {
		t.Fatal(err)
	}
	names := listAll(t, c, "/big", 1000)
	seen := make(map[string]bool, n)
	for i, name := range names {
		if seen[name] {
			t.Fatalf("%s listed twice", name)
		}
		seen[name] = true
		if want := fmt.Sprintf("f%05d.txt", i); name != want {
			t.Fatalf("item %d is %s, want %s", i, name, want)
		}
	}
	if len(names) != n {
		t.Errorf("%d items listed, want %d", len(names), n)
	}
	if st, _, b := c.do("DELETE", "/api/v1/files/items?path=/big&recursive=true", "", nil, nil); st != http.StatusNoContent {
		t.Errorf("delete the folder: %d %s", st, b)
	}
}
