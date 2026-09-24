package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/KhizirFarrukh/local-ai-nas/internal/api/gen"
	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/files"
)

// getItems calls GET /api/v1/files/items with the query and decodes a
// successful response.
func getItems(t *testing.T, h http.Handler, query url.Values) (*httptest.ResponseRecorder, gen.ItemsResponse) {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/files/items?"+query.Encode(), nil))
	var resp gen.ItemsResponse
	if rec.Code == http.StatusOK {
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("response is not ItemsResponse: %v\n%s", err, rec.Body)
		}
	}
	return rec, resp
}

func names(items *[]gen.FileItem) []string {
	if items == nil {
		return nil
	}
	var out []string
	for _, it := range *items {
		out = append(out, it.Name)
	}
	return out
}

func TestGetItemsFolder(t *testing.T) {
	doc := loadSpec(t)
	h := New(Options{Files: testFiles(t)})
	rec, resp := getItems(t, h, url.Values{"path": {"/"}})
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	var body any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if err := doc.Components.Schemas["ItemsResponse"].Value.VisitJSON(body); err != nil {
		t.Errorf("the response does not match ItemsResponse: %v\n%s", err, rec.Body)
	}
	if resp.Item.Path != "/" || resp.Item.Kind != "dir" || resp.NextCursor != nil {
		t.Errorf("item = %+v, next = %v", resp.Item, resp.NextCursor)
	}
	if diff := cmp.Diff([]string{"docs", "empty", "readme.md"}, names(resp.Items)); diff != "" {
		t.Errorf("items (-want +got):\n%s", diff)
	}
	for _, it := range *resp.Items {
		if it.Path != "/"+it.Name || it.ModTime.Location().String() != "UTC" {
			t.Errorf("item %+v: want path /%s and a UTC time", it, it.Name)
		}
	}
}

func TestGetItemsFile(t *testing.T) {
	h := New(Options{Files: testFiles(t)})
	rec, resp := getItems(t, h, url.Values{"path": {"/docs/a.txt"}, "limit": {"1"}})
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	// The media type of .txt comes from the OS's table (it varies), so only
	// its presence is checked; the ETag format is fixed.
	want := gen.FileItem{Path: "/docs/a.txt", Name: "a.txt", Kind: "file", Size: 5, ModTime: resp.Item.ModTime, Mime: resp.Item.Mime, Etag: resp.Item.Etag}
	if diff := cmp.Diff(want, resp.Item); diff != "" || resp.Items != nil || resp.NextCursor != nil {
		t.Errorf("file details (-want +got):\n%s items=%v next=%v", diff, resp.Items, resp.NextCursor)
	}
	if resp.Item.Mime == nil || *resp.Item.Mime == "" {
		t.Error("file details without a media type")
	}
	if resp.Item.Etag == nil || !regexp.MustCompile(`^"[0-9a-f]{24}"$`).MatchString(*resp.Item.Etag) {
		t.Errorf("ETag = %v, want a quoted 24-digit hex tag", resp.Item.Etag)
	}
}

func TestGetItemsListingHasNoETags(t *testing.T) {
	h := New(Options{Files: testFiles(t)})
	_, resp := getItems(t, h, url.Values{"path": {"/docs"}})
	for _, it := range *resp.Items {
		if it.Etag != nil {
			t.Errorf("listing item %s has an ETag", it.Name)
		}
	}
}

func TestGetItemsPaging(t *testing.T) {
	h := New(Options{Files: testFiles(t)})
	query := url.Values{"path": {"/"}, "limit": {"2"}, "sort": {"name"}, "order": {"desc"}}
	var got []string
	for page := 0; page < 10; page++ {
		rec, resp := getItems(t, h, query)
		if rec.Code != http.StatusOK {
			t.Fatalf("status %d: %s", rec.Code, rec.Body)
		}
		got = append(got, names(resp.Items)...)
		if resp.NextCursor == nil {
			break
		}
		query.Set("cursor", *resp.NextCursor)
	}
	if diff := cmp.Diff([]string{"readme.md", "empty", "docs"}, got); diff != "" {
		t.Errorf("pages (-want +got):\n%s", diff)
	}
}

// recordingFiles counts calls; used to prove that invalid input never
// reaches the service.
type recordingFiles struct {
	calls int
}

func (r *recordingFiles) Stat(context.Context, string, string) (files.Item, error) {
	r.calls++
	return files.Item{}, apperr.New(apperr.NotFound, "fake")
}

func (r *recordingFiles) List(context.Context, string, string, files.ListOptions) (files.ListPage, error) {
	r.calls++
	return files.ListPage{}, apperr.New(apperr.NotFound, "fake")
}

func (r *recordingFiles) CreateFolder(context.Context, string, string, files.FolderOptions) (files.Item, bool, error) {
	r.calls++
	return files.Item{}, false, apperr.New(apperr.NotFound, "fake")
}

func (r *recordingFiles) Upload(context.Context, string, string, io.Reader, int64, files.UploadOptions) (files.Item, bool, error) {
	r.calls++
	return files.Item{}, false, apperr.New(apperr.NotFound, "fake")
}

func (r *recordingFiles) Download(context.Context, string, string) (files.Item, io.ReadSeekCloser, error) {
	r.calls++
	return files.Item{}, nil, apperr.New(apperr.NotFound, "fake")
}

func (r *recordingFiles) Rename(context.Context, string, string, string, files.MoveOptions) (files.Item, error) {
	r.calls++
	return files.Item{}, apperr.New(apperr.NotFound, "fake")
}

func (r *recordingFiles) Move(context.Context, string, string, string, files.MoveOptions) (files.Item, error) {
	r.calls++
	return files.Item{}, apperr.New(apperr.NotFound, "fake")
}

func TestGetItemsInvalidInputNeverReachesService(t *testing.T) {
	svc := &recordingFiles{}
	h := New(Options{Files: svc})
	for _, q := range []string{
		"",                       // no path
		"path=/&limit=0",         // out of range
		"path=/&limit=1001",      //
		"path=/&limit=abc",       // not a number
		"path=/&limit=1&limit=2", // repeated
		"path=/&sort=color",      // not in the enumeration
		"path=/&order=up",        //
		"path=/&sort=NAME",       // enumerations are case-sensitive
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/files/items?"+q, nil))
		var p apperr.Problem
		if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil || rec.Code != http.StatusBadRequest || p.Code != "invalid_request" {
			t.Errorf("?%s → %d %s, want 400 invalid_request", q, rec.Code, rec.Body)
		}
	}
	if svc.calls != 0 {
		t.Errorf("the service was called %d times with invalid input", svc.calls)
	}
}

func TestGetItemsWithoutFilesService(t *testing.T) {
	rec, _ := getItems(t, New(Options{}), url.Values{"path": {"/"}})
	if rec.Code != http.StatusNotImplemented {
		t.Errorf("no files service: status %d, want 501 not_available", rec.Code)
	}
}
