package api

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/KhizirFarrukh/local-ai-nas/internal/api/gen"
	"github.com/KhizirFarrukh/local-ai-nas/internal/files"
	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

// clock is a settable time for the ticket tests.
type clock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *clock) add(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

func createArchive(t *testing.T, c client, body string) (int, gen.ArchiveTicket, []byte) {
	t.Helper()
	status, _, b := c.do(http.MethodPost, "/api/v1/files/archives", "application/json", []byte(body), nil)
	var ticket gen.ArchiveTicket
	if status == http.StatusCreated {
		if err := json.Unmarshal(b, &ticket); err != nil {
			t.Fatal(err)
		}
	}
	return status, ticket, b
}

// TestArchivesOverHTTP covers the two steps of an archive download
// (S02.4-T03, FR-006) through a real server: the ticket, then the ZIP.
func TestArchivesOverHTTP(t *testing.T) {
	svc, _ := testFilesDir(t)
	clk := &clock{now: time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)}
	c := client{t, testutil.NewServer(t, New(Options{Files: svc, Now: clk.Now})).URL}

	status, ticket, body := createArchive(t, c, `{"paths":["/docs","/readme.md"]}`)
	if status != http.StatusCreated {
		t.Fatalf("create: %d %s", status, body)
	}
	if ticket.Name != "download-2026-09-28.zip" || ticket.Entries != 4 || ticket.Size != int64(len("hellohi# readme")) {
		t.Errorf("ticket %+v", ticket)
	}
	if !ticket.ExpiresAt.Equal(clk.Now().Add(archiveTTL)) || ticket.Url != ArchivesPath+ticket.Id {
		t.Errorf("ticket expiry %v, url %q", ticket.ExpiresAt, ticket.Url)
	}

	status, h, zipped := c.do(http.MethodGet, ticket.Url, "", nil, nil)
	if status != http.StatusOK || h.Get("Content-Type") != "application/zip" || h.Get("Cache-Control") != "no-store" {
		t.Fatalf("download: %d %v", status, h)
	}
	if cd := h.Get("Content-Disposition"); !strings.Contains(cd, `filename="download-2026-09-28.zip"`) {
		t.Errorf("Content-Disposition %q", cd)
	}
	r, err := zip.NewReader(bytes.NewReader(zipped), int64(len(zipped)))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, f := range r.File {
		names = append(names, f.Name)
	}
	sort.Strings(names)
	if got := strings.Join(names, " "); got != "docs/ docs/a.txt docs/b.txt readme.md" {
		t.Errorf("entries %q", got)
	}

	// A ticket can be downloaded again (browsers retry) until it expires.
	if status, _, _ := c.do(http.MethodGet, ticket.Url, "", nil, nil); status != http.StatusOK {
		t.Errorf("second download: %d", status)
	}
	clk.add(archiveTTL)
	if status, _, b := c.do(http.MethodGet, ticket.Url, "", nil, nil); status != http.StatusNotFound || !strings.Contains(string(b), "tickets last 5 minutes") {
		t.Errorf("expired ticket: %d %s", status, b)
	}
}

func TestArchiveNames(t *testing.T) {
	svc, _ := testFilesDir(t)
	c := client{t, testutil.NewServer(t, New(Options{Files: svc})).URL}
	for _, tc := range []struct{ body, want string }{
		{`{"paths":["/docs"]}`, "docs.zip"},
		{`{"paths":["/docs"],"name":"photos"}`, "photos.zip"},
		{`{"paths":["/docs"],"name":"Trip.ZIP"}`, "Trip.ZIP"},
		{`{"paths":["/docs"],"name":"..\\..\\evil\u0007.zip"}`, "evil.zip"},
		{`{"paths":["/docs"],"name":"  "}`, "docs.zip"},
	} {
		status, ticket, body := createArchive(t, c, tc.body)
		if status != http.StatusCreated || ticket.Name != tc.want {
			t.Errorf("%s: %d %q, want %q (%s)", tc.body, status, ticket.Name, tc.want, body)
		}
	}
}

func TestArchiveRefusals(t *testing.T) {
	svc, _ := testFilesDir(t)
	c := client{t, testutil.NewServer(t, New(Options{Files: svc})).URL}
	many := make([]string, maxArchivePaths+1)
	for i := range many {
		many[i] = fmt.Sprintf("/f%d", i)
	}
	tooMany, _ := json.Marshal(map[string]any{"paths": many})
	for _, tc := range []struct {
		body   string
		status int
		code   string
	}{
		{`{"paths":[]}`, 400, "invalid_request"},
		{`{"paths":["/missing"]}`, 404, "not_found"},
		{`{"paths":["/../../photos"]}`, 400, "outside_root"},
		{string(tooMany), 400, "invalid_request"},
	} {
		status, _, body := createArchive(t, c, tc.body)
		if status != tc.status || !strings.Contains(string(body), `"code":"`+tc.code+`"`) {
			t.Errorf("%.60s: %d %s, want %d %s", tc.body, status, body, tc.status, tc.code)
		}
	}
	if status, _, _ := c.do(http.MethodGet, ArchivesPath+"nosuchticket", "", nil, nil); status != http.StatusNotFound {
		t.Errorf("unknown ticket: %d", status)
	}
}

func TestArchiveTicketsAreBounded(t *testing.T) {
	a := newArchiveTickets(nil)
	first, _ := a.add(files.ArchivePlan{}, "a.zip")
	for i := 0; i < maxTickets; i++ {
		a.add(files.ArchivePlan{}, "b.zip")
	}
	if _, ok := a.get(first); ok {
		t.Error("the oldest ticket was kept past the bound")
	}
	if len(a.tickets) != maxTickets {
		t.Errorf("%d tickets kept, want %d", len(a.tickets), maxTickets)
	}
}

// syncBuffer is a bytes.Buffer safe for writes from other goroutines.
type syncBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

// brokenArchive plans archives as usual and fails while writing one, as a
// disk read error would.
type brokenArchive struct{ files.Service }

func (brokenArchive) WriteArchive(_ context.Context, _ files.ArchivePlan, w io.Writer) error {
	_, _ = w.Write([]byte{'P', 'K', 3, 4})
	return errors.New("read error on the disk") // after the ZIP's first bytes
}

// TestArchiveWriteFailureEndsTheConnection: once the ZIP has started, a
// failure can only end the connection, so the client sees an incomplete
// download instead of a complete-looking archive; the log says why. A good
// archive is recorded at debug level with the heap before and after.
func TestArchiveWriteFailureEndsTheConnection(t *testing.T) {
	svc, _ := testFilesDir(t)
	var logs syncBuffer // written by the server's goroutines
	log := slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))

	good := client{t, testutil.NewServer(t, New(Options{Files: svc, Logger: log})).URL}
	_, ticket, _ := createArchive(t, good, `{"paths":["/docs"]}`)
	if status, _, _ := good.do(http.MethodGet, ticket.Url, "", nil, nil); status != http.StatusOK {
		t.Fatal(status)
	}
	if !strings.Contains(logs.String(), "archive sent") || !strings.Contains(logs.String(), "heap_end_bytes") {
		t.Errorf("no debug record of the archive: %s", logs.String())
	}

	bad := client{t, testutil.NewServer(t, New(Options{Files: brokenArchive{svc}, Logger: log})).URL}
	_, ticket, _ = createArchive(t, bad, `{"paths":["/docs"]}`)
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, bad.url+ticket.Url, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err == nil {
		_, err = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
	}
	if err == nil {
		t.Error("the client got a complete answer for a failed archive")
	}
	if !strings.Contains(logs.String(), "archive stopped") || !strings.Contains(logs.String(), "read error on the disk") {
		t.Errorf("the failure is not logged: %s", logs.String())
	}
}
