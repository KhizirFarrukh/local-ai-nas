package api

import (
	"context"
	"io"
	"strings"

	"github.com/KhizirFarrukh/local-ai-nas/internal/api/gen"
	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/files"
)

// DeleteItem deletes a file or folder (S01.3-T09).
func (s *server) DeleteItem(ctx context.Context, req gen.DeleteItemRequestObject) (gen.DeleteItemResponseObject, error) {
	opts := files.DeleteOptions{Recursive: req.Params.Recursive != nil && *req.Params.Recursive}
	if err := s.files.Delete(ctx, s.owner, req.Params.Path, opts); err != nil {
		return nil, err
	}
	return gen.DeleteItem204Response{}, nil
}

// GetItems lists a folder, or returns the details of a file (S01.3-T02).
// Every parameter is checked before the service is called, so invalid
// input never reaches it.
func (s *server) GetItems(ctx context.Context, req gen.GetItemsRequestObject) (gen.GetItemsResponseObject, error) {
	opts, err := listOptions(req.Params)
	if err != nil {
		return nil, err
	}
	path := string(req.Params.Path)
	it, err := s.files.Stat(ctx, s.owner, path)
	if err != nil {
		return nil, err
	}
	resp := gen.ItemsResponse{Item: fileItem(it)}
	if it.Kind == files.KindDir {
		page, err := s.files.List(ctx, s.owner, path, opts)
		if err != nil {
			return nil, err
		}
		items := make([]gen.FileItem, 0, len(page.Items))
		for _, x := range page.Items {
			items = append(items, fileItem(x))
		}
		resp.Items = &items
		if page.NextCursor != "" {
			resp.NextCursor = &page.NextCursor
		}
		resp.Total = &page.Total
		if page.Position >= 0 {
			resp.Position = &page.Position
		}
	}
	return gen.GetItems200JSONResponse(resp), nil
}

// GetUsage adds up the size of a folder (S02.3-T05, FR-214).
func (s *server) GetUsage(ctx context.Context, req gen.GetUsageRequestObject) (gen.GetUsageResponseObject, error) {
	u, err := s.files.Usage(ctx, s.owner, string(req.Params.Path))
	if err != nil {
		return nil, err
	}
	return gen.GetUsage200JSONResponse{Path: fileItem(u.Item).Path, Size: u.Size, Files: u.Files, Folders: u.Folders}, nil
}

// listOptions checks the paging parameters. The generated binding checks
// their types; this checks their values.
func listOptions(p gen.GetItemsParams) (files.ListOptions, error) {
	var o files.ListOptions
	if p.Offset != nil {
		if *p.Offset < 0 {
			return o, apperr.Newf(apperr.InvalidRequest, "offset must be 0 or more, got %d", *p.Offset)
		}
		if p.Cursor != nil {
			return o, apperr.New(apperr.InvalidRequest, "offset and cursor cannot be used together")
		}
		o.Offset = *p.Offset
	}
	if p.Limit != nil {
		if *p.Limit < 1 || *p.Limit > files.MaxLimit {
			return o, apperr.Newf(apperr.InvalidRequest, "limit must be between 1 and %d, got %d", files.MaxLimit, *p.Limit)
		}
		o.Limit = *p.Limit
	}
	if p.Sort != nil {
		if !p.Sort.Valid() {
			return o, apperr.Newf(apperr.InvalidRequest, "sort must be name, size, mod_time, added_time, or kind, got %q", *p.Sort)
		}
		o.Sort = files.SortKey(*p.Sort)
	}
	if p.Order != nil {
		if !p.Order.Valid() {
			return o, apperr.Newf(apperr.InvalidRequest, "order must be asc or desc, got %q", *p.Order)
		}
		o.Order = files.Order(*p.Order)
	}
	if p.Cursor != nil {
		o.Cursor = *p.Cursor
	}
	if p.Locate != nil {
		if strings.Contains(*p.Locate, "/") {
			return o, apperr.Newf(apperr.InvalidRequest, "locate is the name of an item in the folder, not a path: %q", *p.Locate)
		}
		o.Locate = *p.Locate
	}
	return o, nil
}

// fileItem converts an item to its API form: the path starts with "/" and
// times are UTC.
func fileItem(it files.Item) gen.FileItem {
	p := "/" + it.RelPath
	if it.RelPath == "." {
		p = "/"
	}
	fi := gen.FileItem{
		Path:      p,
		Name:      it.Name,
		Kind:      gen.ItemKind(it.Kind),
		Size:      it.Size,
		ModTime:   it.ModTime.UTC(),
		AddedTime: it.AddedTime.UTC(),
	}
	if it.MIME != "" {
		fi.Mime = &it.MIME
	}
	if it.ETag != "" {
		fi.Etag = &it.ETag
	}
	if it.ContentHash != "" {
		fi.ContentHash = &it.ContentHash
	}
	return fi
}

// noFiles stands in when no files service is configured: every call
// answers not_available instead of crashing.
type noFiles struct{}

var errNoFiles = apperr.New(apperr.NotAvailable, "the files area is not configured on this server")

func (noFiles) Stat(context.Context, string, string) (files.Item, error) {
	return files.Item{}, errNoFiles
}

func (noFiles) List(context.Context, string, string, files.ListOptions) (files.ListPage, error) {
	return files.ListPage{}, errNoFiles
}

func (noFiles) CreateFolder(context.Context, string, string, files.FolderOptions) (files.Item, bool, error) {
	return files.Item{}, false, errNoFiles
}

func (noFiles) Upload(context.Context, string, string, io.Reader, int64, files.UploadOptions) (files.Item, bool, error) {
	return files.Item{}, false, errNoFiles
}

func (noFiles) Download(context.Context, string, string) (files.Item, io.ReadSeekCloser, error) {
	return files.Item{}, nil, errNoFiles
}

func (noFiles) Rename(context.Context, string, string, string, files.MoveOptions) (files.Item, error) {
	return files.Item{}, errNoFiles
}

func (noFiles) Move(context.Context, string, string, string, files.MoveOptions) (files.Item, error) {
	return files.Item{}, errNoFiles
}

func (noFiles) Copy(context.Context, string, string, string, files.CopyOptions) (files.Item, bool, error) {
	return files.Item{}, false, errNoFiles
}

func (noFiles) Delete(context.Context, string, string, files.DeleteOptions) error {
	return errNoFiles
}

func (noFiles) PlanArchive(context.Context, string, []string) (files.ArchivePlan, error) {
	return files.ArchivePlan{}, errNoFiles
}

func (noFiles) WriteArchive(context.Context, files.ArchivePlan, io.Writer) error {
	return errNoFiles
}

func (noFiles) Usage(context.Context, string, string) (files.Usage, error) {
	return files.Usage{}, errNoFiles
}
