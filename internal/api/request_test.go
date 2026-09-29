package api

import (
	"io"
	"net/http"
	"net/http/httptest"
)

// newTestRequest is httptest.NewRequest with a Host this server answers
// to. httptest's default Host, example.com, is refused by the Host check
// since S03.5-T02 (origin.go, ADR-0042).
func newTestRequest(method, target string, body io.Reader) *http.Request {
	r := httptest.NewRequest(method, target, body)
	r.Host = "127.0.0.1:8080"
	return r
}
