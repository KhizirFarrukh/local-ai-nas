package api

import (
	"context"
	"crypto/rand"
	"errors"
	"log/slog"
	"net/http"
	"path"
	"runtime/metrics"
	"strings"
	"sync"
	"time"

	"github.com/KhizirFarrukh/local-ai-nas/internal/api/gen"
	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/files"
)

// Archives (S02.4-T03, FR-006): a POST checks the items and returns a
// ticket, and a GET of the ticket streams the ZIP. The two steps keep the
// request a JSON body (the API's convention) while the browser downloads
// the result like any file, without holding it in memory.

// ArchivesPath is where archive tickets are downloaded.
const ArchivesPath = "/api/v1/files/archives/"

// archiveTTL is how long a ticket can be downloaded.
const archiveTTL = 5 * time.Minute

// maxTickets bounds the tickets kept at once; the oldest goes first.
const maxTickets = 256

// maxArchivePaths is the most paths one archive request names.
const maxArchivePaths = 10000

// archiveBodyBytes is the body limit of the POST: 10,000 paths fit.
const archiveBodyBytes = 8 << 20

type archiveTicket struct {
	plan    files.ArchivePlan
	name    string
	expires time.Time
}

// archiveTickets keeps the tickets in memory; a restart forgets them.
type archiveTickets struct {
	mu      sync.Mutex
	tickets map[string]archiveTicket
	order   []string // oldest first
	now     func() time.Time
}

func newArchiveTickets(now func() time.Time) *archiveTickets {
	if now == nil {
		now = time.Now
	}
	return &archiveTickets{tickets: map[string]archiveTicket{}, now: now}
}

func (a *archiveTickets) add(plan files.ArchivePlan, name string) (string, time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.prune()
	for len(a.order) >= maxTickets {
		delete(a.tickets, a.order[0])
		a.order = a.order[1:]
	}
	id := strings.ToLower(rand.Text())
	expires := a.now().Add(archiveTTL).UTC()
	a.tickets[id] = archiveTicket{plan: plan, name: name, expires: expires}
	a.order = append(a.order, id)
	return id, expires
}

func (a *archiveTickets) get(id string) (archiveTicket, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.prune()
	t, ok := a.tickets[id]
	return t, ok
}

// prune drops expired tickets; the caller holds the lock.
func (a *archiveTickets) prune() {
	now := a.now()
	kept := a.order[:0]
	for _, id := range a.order {
		if now.Before(a.tickets[id].expires) {
			kept = append(kept, id)
		} else {
			delete(a.tickets, id)
		}
	}
	a.order = kept
}

// CreateArchive checks the items and returns a ticket (201).
func (s *server) CreateArchive(ctx context.Context, req gen.CreateArchiveRequestObject) (gen.CreateArchiveResponseObject, error) {
	body := req.Body
	if body == nil || len(body.Paths) == 0 {
		return nil, apperr.New(apperr.InvalidRequest, "an archive needs at least one path")
	}
	if len(body.Paths) > maxArchivePaths {
		return nil, apperr.Newf(apperr.InvalidRequest, "an archive request names at most %d paths, this one has %d", maxArchivePaths, len(body.Paths))
	}
	plan, err := s.files.PlanArchive(ctx, s.owner, body.Paths)
	if err != nil {
		return nil, err
	}
	name := archiveFileName(body.Name, body.Paths, s.archives.now())
	id, expires := s.archives.add(plan, name)
	return gen.CreateArchive201JSONResponse{
		Id:        id,
		Url:       ArchivesPath + id,
		Name:      name,
		ExpiresAt: expires,
		Entries:   len(plan.Entries),
		Size:      plan.Size,
	}, nil
}

// archiveFileName is the download's name: the one asked for, reduced to a
// plain name ending in .zip, or the default.
func archiveFileName(asked *string, paths []string, now time.Time) string {
	if asked != nil {
		name := strings.TrimSpace(path.Base(strings.ReplaceAll(*asked, "\\", "/")))
		name = strings.Map(func(r rune) rune {
			if r < 0x20 || r == 0x7f {
				return -1
			}
			return r
		}, name)
		if name != "" && name != "." && name != "/" {
			if !strings.HasSuffix(strings.ToLower(name), ".zip") {
				name += ".zip"
			}
			return name
		}
	}
	return files.ArchiveName(paths, now)
}

// DownloadArchive streams the archive of a ticket.
func (s *server) DownloadArchive(ctx context.Context, req gen.DownloadArchiveRequestObject) (gen.DownloadArchiveResponseObject, error) {
	t, ok := s.archives.get(req.Id)
	if !ok {
		return nil, apperr.New(apperr.NotFound, "no archive with this ID: tickets last 5 minutes, and a restart forgets them")
	}
	return archiveContent{ctx: ctx, files: s.files, log: s.log, ticket: t}, nil
}

// archiveContent writes the ZIP as the response body.
type archiveContent struct {
	ctx    context.Context
	files  files.Service
	log    *slog.Logger
	ticket archiveTicket
}

func (a archiveContent) VisitDownloadArchiveResponse(w http.ResponseWriter) error {
	h := w.Header()
	h.Set("Content-Type", "application/zip")
	h.Set("Content-Disposition", attachment(a.ticket.name))
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Cache-Control", "no-store")
	debug := a.log.Enabled(a.ctx, slog.LevelDebug)
	var heapStart uint64
	if debug {
		heapStart = heapObjects()
	}
	w.WriteHeader(http.StatusOK)
	if err := a.files.WriteArchive(a.ctx, a.ticket.plan, w); err != nil {
		// The status is sent; only ending the connection tells the client
		// that the archive is incomplete.
		level := slog.LevelError
		if errors.Is(err, context.Canceled) {
			level = slog.LevelInfo // the client went away
		}
		a.log.LogAttrs(a.ctx, level, "archive stopped", slog.String("archive", a.ticket.name), slog.String("error", err.Error()))
		panic(http.ErrAbortHandler)
	}
	if debug {
		// The heap before and after shows that memory stays flat, whatever
		// the archive's size (S02.4-T03).
		a.log.LogAttrs(a.ctx, slog.LevelDebug, "archive sent", slog.String("archive", a.ticket.name),
			slog.Int("entries", len(a.ticket.plan.Entries)), slog.Int64("size", a.ticket.plan.Size),
			slog.Uint64("heap_start_bytes", heapStart), slog.Uint64("heap_end_bytes", heapObjects()))
	}
	return nil
}

// heapObjects is the memory the heap's objects take, in bytes: live ones
// and dead ones not yet freed.
func heapObjects() uint64 {
	s := []metrics.Sample{{Name: "/memory/classes/heap/objects:bytes"}}
	metrics.Read(s)
	return s[0].Value.Uint64()
}
