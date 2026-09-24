package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KhizirFarrukh/local-ai-nas/internal/api/gen"
	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/health"
	"github.com/KhizirFarrukh/local-ai-nas/internal/logging"
)

func checkOf(name, detail string, err error) health.Check {
	return health.Check{Name: name, Run: func(context.Context) (string, error) { return detail, err }}
}

func TestGetHealthStatuses(t *testing.T) {
	tests := []struct {
		name       string
		checks     []health.Check
		wantCode   int
		wantStatus string
	}{
		{"healthy", []health.Check{checkOf("config", "valid", nil)}, 200, "ok"},
		{"warning", []health.Check{checkOf("same_filesystem", "", health.Warnf("copy"))}, 200, "warn"},
		{"failure", []health.Check{checkOf("config", "valid", nil), checkOf("database", "", errors.New("locked"))}, 503, "fail"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			New(Options{Version: "v", Checks: tt.checks}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/system/health", nil))
			var rep gen.HealthReport
			if err := json.Unmarshal(rec.Body.Bytes(), &rep); err != nil {
				t.Fatal(err)
			}
			if rec.Code != tt.wantCode || string(rep.Status) != tt.wantStatus || len(rep.Checks) != len(tt.checks) {
				t.Errorf("got %d %s with %d checks, want %d %s", rec.Code, rep.Status, len(rep.Checks), tt.wantCode, tt.wantStatus)
			}
			if rec.Header().Get("Cache-Control") != "no-store" {
				t.Errorf("Cache-Control = %q, want no-store", rec.Header().Get("Cache-Control"))
			}
		})
	}
}

// problemOf runs h and decodes the problem it writes.
func problemOf(t *testing.T, h http.HandlerFunc) (int, apperr.Problem) {
	t.Helper()
	rec := httptest.NewRecorder()
	logging.RequestID(h).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/x", nil))
	var p apperr.Problem
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("body %q is not a problem: %v", rec.Body, err)
	}
	return rec.Code, p
}

// TestErrorHooks drives the error hooks of the generated handlers: a
// parameter binding error and a body decoding error become 400
// invalid_request problems, and an operation's error keeps its kind.
func TestErrorHooks(t *testing.T) {
	log := slog.New(slog.DiscardHandler)

	bindErr := &gen.InvalidParamFormatError{ParamName: "limit", Err: errors.New("not a number")}
	code, p := problemOf(t, func(w http.ResponseWriter, r *http.Request) { bindingError(log)(w, r, bindErr) })
	if code != 400 || p.Code != "invalid_request" || !strings.Contains(p.Detail, "limit") {
		t.Errorf("binding error → %d %+v", code, p)
	}

	code, p = problemOf(t, func(w http.ResponseWriter, r *http.Request) { bodyError(log)(w, r, errors.New("unexpected EOF")) })
	if code != 400 || p.Code != "invalid_request" || !strings.Contains(p.Detail, "body") {
		t.Errorf("body error → %d %+v", code, p)
	}

	strict := gen.NewStrictHandlerWithOptions(&failingServer{err: apperr.New(apperr.Conflict, "exists")}, nil,
		gen.StrictHTTPServerOptions{ResponseErrorHandlerFunc: operationError(log)})
	code, p = problemOf(t, strict.GetHealth)
	if code != 409 || p.Code != "conflict" {
		t.Errorf("operation error → %d %+v", code, p)
	}

	strict = gen.NewStrictHandlerWithOptions(&failingServer{err: errors.New("disk on fire")}, nil,
		gen.StrictHTTPServerOptions{ResponseErrorHandlerFunc: operationError(log)})
	code, p = problemOf(t, strict.GetHealth)
	if code != 500 || p.Code != "internal" || strings.Contains(p.Detail, "fire") {
		t.Errorf("plain operation error → %d %+v, want a generic 500", code, p)
	}
}

// failingServer returns a fixed error from every operation.
type failingServer struct{ err error }

func (f *failingServer) GetHealth(context.Context, gen.GetHealthRequestObject) (gen.GetHealthResponseObject, error) {
	return nil, f.err
}

func (f *failingServer) GetItems(context.Context, gen.GetItemsRequestObject) (gen.GetItemsResponseObject, error) {
	return nil, f.err
}

func (f *failingServer) CreateFolder(context.Context, gen.CreateFolderRequestObject) (gen.CreateFolderResponseObject, error) {
	return nil, f.err
}

func (f *failingServer) UploadFile(context.Context, gen.UploadFileRequestObject) (gen.UploadFileResponseObject, error) {
	return nil, f.err
}

func (f *failingServer) DownloadFile(context.Context, gen.DownloadFileRequestObject) (gen.DownloadFileResponseObject, error) {
	return nil, f.err
}

func (f *failingServer) RenameItem(context.Context, gen.RenameItemRequestObject) (gen.RenameItemResponseObject, error) {
	return nil, f.err
}

func (f *failingServer) MoveItem(context.Context, gen.MoveItemRequestObject) (gen.MoveItemResponseObject, error) {
	return nil, f.err
}
