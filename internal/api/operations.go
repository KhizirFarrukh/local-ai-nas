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

// CopyItem copies a file or folder (S01.3-T08): 201 for a new item, 200
// when on_conflict=overwrite replaced a file.
func (s *server) CopyItem(ctx context.Context, req gen.CopyItemRequestObject) (gen.CopyItemResponseObject, error) {
	b := req.Body
	if b == nil {
		return nil, apperr.New(apperr.InvalidRequest, "a JSON body is required")
	}
	policy, err := conflictPolicy(b.OnConflict)
	if err != nil {
		return nil, err
	}
	it, created, err := s.files.Copy(ctx, s.owner, b.From, b.To, files.CopyOptions{OnConflict: policy})
	if err != nil {
		return nil, err
	}
	if created {
		return gen.CopyItem201JSONResponse(fileItem(it)), nil
	}
	return gen.CopyItem200JSONResponse(fileItem(it)), nil
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
