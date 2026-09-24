package api

import (
	"context"

	"github.com/KhizirFarrukh/local-ai-nas/internal/api/gen"
	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/files"
)

// CreateFolder creates a folder (S01.3-T04): 201 when it was created, 200
// when on_conflict=overwrite found the folder already there.
func (s *server) CreateFolder(ctx context.Context, req gen.CreateFolderRequestObject) (gen.CreateFolderResponseObject, error) {
	b := req.Body
	if b == nil {
		return nil, apperr.New(apperr.InvalidRequest, "a JSON body is required")
	}
	policy, err := conflictPolicy(b.OnConflict)
	if err != nil {
		return nil, err
	}
	opts := files.FolderOptions{Parents: b.Parents != nil && *b.Parents, OnConflict: policy}
	it, created, err := s.files.CreateFolder(ctx, s.owner, b.Path, opts)
	if err != nil {
		return nil, err
	}
	if created {
		return gen.CreateFolder201JSONResponse(fileItem(it)), nil
	}
	return gen.CreateFolder200JSONResponse(fileItem(it)), nil
}
