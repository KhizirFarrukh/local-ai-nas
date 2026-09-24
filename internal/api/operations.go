package api

import (
	"context"

	"github.com/KhizirFarrukh/local-ai-nas/internal/api/gen"
	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/files"
)

// RenameItem renames a file or folder in its folder (S01.3-T07).
func (s *server) RenameItem(ctx context.Context, req gen.RenameItemRequestObject) (gen.RenameItemResponseObject, error) {
	b := req.Body
	if b == nil {
		return nil, apperr.New(apperr.InvalidRequest, "a JSON body is required")
	}
	policy, err := conflictPolicy(b.OnConflict)
	if err != nil {
		return nil, err
	}
	it, err := s.files.Rename(ctx, s.owner, b.Path, b.NewName, files.MoveOptions{OnConflict: policy})
	if err != nil {
		return nil, err
	}
	return gen.RenameItem200JSONResponse(fileItem(it)), nil
}

// MoveItem moves a file or folder to a new path (S01.3-T07).
func (s *server) MoveItem(ctx context.Context, req gen.MoveItemRequestObject) (gen.MoveItemResponseObject, error) {
	b := req.Body
	if b == nil {
		return nil, apperr.New(apperr.InvalidRequest, "a JSON body is required")
	}
	policy, err := conflictPolicy(b.OnConflict)
	if err != nil {
		return nil, err
	}
	it, err := s.files.Move(ctx, s.owner, b.From, b.To, files.MoveOptions{OnConflict: policy})
	if err != nil {
		return nil, err
	}
	return gen.MoveItem200JSONResponse(fileItem(it)), nil
}
