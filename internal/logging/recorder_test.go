package logging

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// readFromWriter is a ResponseWriter that has ReadFrom, like the one of a
// real connection.
type readFromWriter struct {
	*httptest.ResponseRecorder
	used bool
}

func (w *readFromWriter) ReadFrom(r io.Reader) (int64, error) {
	w.used = true
	return io.Copy(w.ResponseRecorder, r)
}

// TestRecorderReadFrom: a copy through the access log reaches the
// underlying writer's ReadFrom (sendfile for downloads) and is counted.
func TestRecorderReadFrom(t *testing.T) {
	for _, withReadFrom := range []bool{true, false} {
		base := httptest.NewRecorder()
		var w http.ResponseWriter = base
		rf := &readFromWriter{ResponseRecorder: base}
		if withReadFrom {
			w = rf
		}
		r := &recorder{ResponseWriter: w}
		n, err := r.ReadFrom(io.LimitReader(strings.NewReader("hello"), 5)) // as ServeContent does
		if err != nil || n != 5 || r.bytes != 5 || r.status != http.StatusOK || base.Body.String() != "hello" {
			t.Errorf("ReadFrom (underlying ReadFrom %v): %d %v, counted %d, status %d, body %q", withReadFrom, n, err, r.bytes, r.status, base.Body)
		}
		if rf.used != withReadFrom {
			t.Errorf("the underlying ReadFrom was used: %v, want %v", rf.used, withReadFrom)
		}
	}
}
