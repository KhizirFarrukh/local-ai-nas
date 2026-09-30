package api

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/KhizirFarrukh/local-ai-nas/internal/api/gen"
	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/files"
	"github.com/KhizirFarrukh/local-ai-nas/internal/jobs"
)

// JobQueue is what the API needs of the background job queue
// (S01.4-T08, internal/jobs).
type JobQueue interface {
	Enqueue(ctx context.Context, typ, owner string, payload any) (jobs.Job, error)
	Get(ctx context.Context, id string) (jobs.Job, error)
	List(ctx context.Context, owner string, limit int) ([]jobs.Job, error)
	Cancel(ctx context.Context, id string) (jobs.Job, error)
}

// CopyJobType is the job that runs a copy over the synchronous limits.
const CopyJobType = "files.copy"

// CopyJobPayload is the input of a CopyJobType job; the job works for the
// owner recorded with it.
type CopyJobPayload struct {
	From       string           `json:"from"`
	To         string           `json:"to"`
	OnConflict files.OnConflict `json:"on_conflict"`
}

// CopyJobResult is the result of a CopyJobType job.
type CopyJobResult struct {
	Path    string `json:"path"`
	Created bool   `json:"created"`
}

// jobListLimit is how many jobs GET /jobs returns.
const jobListLimit = 50

var errNoJobs = apperr.New(apperr.NotAvailable, "background jobs are not set up on this server")

// ListJobs lists the caller's recent jobs.
func (s *server) ListJobs(ctx context.Context, _ gen.ListJobsRequestObject) (gen.ListJobsResponseObject, error) {
	if s.jobs == nil {
		return nil, errNoJobs
	}
	list, err := s.jobs.List(ctx, s.owner, jobListLimit)
	if err != nil {
		return nil, err
	}
	out := gen.JobList{Items: make([]gen.Job, 0, len(list))}
	for _, j := range list {
		out.Items = append(out.Items, jobOut(j))
	}
	return gen.ListJobs200JSONResponse(out), nil
}

// GetJob returns one of the caller's jobs.
func (s *server) GetJob(ctx context.Context, req gen.GetJobRequestObject) (gen.GetJobResponseObject, error) {
	j, err := s.ownJob(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return gen.GetJob200JSONResponse(jobOut(j)), nil
}

// CancelJob asks one of the caller's jobs to stop.
func (s *server) CancelJob(ctx context.Context, req gen.CancelJobRequestObject) (gen.CancelJobResponseObject, error) {
	if _, err := s.ownJob(ctx, req.Id); err != nil {
		return nil, err
	}
	j, err := s.jobs.Cancel(ctx, req.Id)
	switch {
	case errors.Is(err, jobs.ErrFinished):
		return nil, apperr.Newf(apperr.Conflict, "the job has %s and cannot be canceled", j.State)
	case err != nil:
		return nil, err
	}
	return gen.CancelJob200JSONResponse(jobOut(j)), nil
}

// ownJob returns a job of the caller. Another owner's job is reported as
// not found, so a job ID never reveals someone else's work.
func (s *server) ownJob(ctx context.Context, id string) (jobs.Job, error) {
	if s.jobs == nil {
		return jobs.Job{}, errNoJobs
	}
	j, err := s.jobs.Get(ctx, id)
	if errors.Is(err, jobs.ErrNotFound) || err == nil && j.Owner != s.owner {
		return jobs.Job{}, apperr.New(apperr.NotFound, "no job with this id")
	}
	return j, err
}

// copyAsJob queues a copy that is over the synchronous limits.
func (s *server) copyAsJob(ctx context.Context, from, to string, policy files.OnConflict) (gen.CopyItemResponseObject, error) {
	j, err := s.jobs.Enqueue(ctx, CopyJobType, s.owner, CopyJobPayload{From: from, To: to, OnConflict: policy})
	if err != nil {
		return nil, err
	}
	return gen.CopyItem202JSONResponse(jobOut(j)), nil
}

// jobOut converts a job for the API.
func jobOut(j jobs.Job) gen.Job {
	out := gen.Job{
		Id:              j.ID,
		Type:            j.Type,
		State:           gen.JobState(j.State),
		Attempts:        j.Attempts,
		MaxAttempts:     j.MaxAttempts,
		CancelRequested: j.CancelRequested,
		CreatedAt:       j.CreatedAt.UTC(),
		UpdatedAt:       j.UpdatedAt.UTC(),
	}
	out.Progress.Done, out.Progress.Total = j.Done, j.Total
	if j.Error != "" {
		msg := j.Error
		out.Error = &msg
	}
	if !j.FinishedAt.IsZero() {
		t := j.FinishedAt.UTC()
		out.FinishedAt = &t
	}
	if len(j.Result) > 0 {
		var r map[string]any
		if json.Unmarshal(j.Result, &r) == nil {
			out.Result = &r
		}
	}
	return out
}
