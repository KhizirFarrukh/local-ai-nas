package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/KhizirFarrukh/local-ai-nas/internal/api"
	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/files"
	"github.com/KhizirFarrukh/local-ai-nas/internal/jobs"
)

// copyJobAttempts bounds the retries of a background copy: a failure the
// files service reports (not found, conflict, no space) is final at once;
// an unexpected one is tried again.
const copyJobAttempts = 3

// copyJob runs a copy over the synchronous limits (S01.4-T08, FR-351).
// The copy is built under a temporary name and appears only when complete,
// so running it again after a restart starts over cleanly; the leftovers
// of a cut-off attempt are removed by the temporary-file cleanup.
func copyJob(fsvc files.Service, log *slog.Logger) jobs.Handler {
	return func(ctx context.Context, j jobs.Job, progress func(done, total int64)) (any, error) {
		var p api.CopyJobPayload
		if err := json.Unmarshal(j.Payload, &p); err != nil {
			return nil, jobs.Permanent(err)
		}
		it, created, err := fsvc.Copy(ctx, j.Owner, p.From, p.To, files.CopyOptions{
			OnConflict: p.OnConflict, NoLimits: true, Progress: progress,
		})
		if err != nil {
			return nil, jobError(ctx, log, j, err)
		}
		return api.CopyJobResult{Path: "/" + it.RelPath, Created: created}, nil
	}
}

// jobError turns a files error into a job failure: the detail of an
// expected error is shown to the user and not retried; an unexpected one
// is logged, retried, and shown only as a generic message.
func jobError(ctx context.Context, log *slog.Logger, j jobs.Job, err error) error {
	var ae *apperr.Error
	if errors.As(err, &ae) && ae.Kind != apperr.Internal {
		return jobs.Permanent(errors.New(ae.Detail))
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	log.ErrorContext(ctx, "a background job failed", "job", j.ID, "type", j.Type, "error", err.Error())
	return errors.New("the job failed unexpectedly; see the server log")
}
