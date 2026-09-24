package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/KhizirFarrukh/local-ai-nas/internal/api/gen"
	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/files"
	"github.com/KhizirFarrukh/local-ai-nas/internal/health"
)

//go:generate go tool oapi-codegen -config oapi-codegen.yaml ../../api/openapi.yaml

// server implements the operations of the generated strict server
// interface (spec-first: a missing operation does not compile).
type server struct {
	version string
	checks  []health.Check
	files   files.Service
	// owner is the namespace requests act in. S01 has one owner; S03 takes
	// it from the session.
	owner string
	// log is for responses that write problems themselves (downloads).
	log *slog.Logger
}

var _ gen.StrictServerInterface = (*server)(nil)

// GetHealth runs the health checks: 200 when healthy (possibly with
// warnings), 503 when a check failed.
func (s *server) GetHealth(ctx context.Context, _ gen.GetHealthRequestObject) (gen.GetHealthResponseObject, error) {
	rep := health.Run(ctx, s.version, s.checks)
	body := gen.HealthReport{
		Status:  gen.HealthStatus(rep.Status),
		Version: rep.Version,
		Checks:  make([]gen.HealthCheck, 0, len(rep.Checks)),
	}
	for _, c := range rep.Checks {
		gc := gen.HealthCheck{Name: c.Name, Status: gen.HealthStatus(c.Status)}
		if c.Detail != "" {
			gc.Detail = &c.Detail
		}
		if c.Error != "" {
			gc.Error = &c.Error
		}
		body.Checks = append(body.Checks, gc)
	}
	if rep.Status == health.StatusFail {
		return gen.GetHealth503JSONResponse(body), nil
	}
	return gen.GetHealth200JSONResponse(body), nil
}

// generated wraps the server in the generated net/http handlers, with
// hooks that turn every failure into a problem response.
func generated(s *server, log *slog.Logger) *gen.ServerInterfaceWrapper {
	strict := gen.NewStrictHandlerWithOptions(s, nil, gen.StrictHTTPServerOptions{
		RequestErrorHandlerFunc:  bodyError(log),
		ResponseErrorHandlerFunc: operationError(log),
	})
	return &gen.ServerInterfaceWrapper{Handler: strict, ErrorHandlerFunc: bindingError(log)}
}

// bindingError handles a query or path parameter that the generated code
// cannot bind (a wrong type or format): 400 invalid_request.
func bindingError(log *slog.Logger) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		apperr.Write(w, r, log, apperr.Wrap(apperr.InvalidRequest, err.Error(), err))
	}
}

// bodyError handles a request body that cannot be decoded: 400
// invalid_request.
func bodyError(log *slog.Logger) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		apperr.Write(w, r, log, apperr.Wrap(apperr.InvalidRequest, "the request body is not valid: "+err.Error(), err))
	}
}

// operationError handles an error returned by an operation: its apperr
// kind decides the response (500 internal for anything else).
func operationError(log *slog.Logger) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		apperr.Write(w, r, log, err)
	}
}
