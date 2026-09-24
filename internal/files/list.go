package files

import (
	"cmp"
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
)

// SortKey orders a folder listing (docs/api/conventions.md).
type SortKey string

// The sort keys.
const (
	SortName    SortKey = "name"
	SortSize    SortKey = "size"
	SortModTime SortKey = "mod_time"
	SortKind    SortKey = "kind"
)

// Order is the direction of a listing.
type Order string

// The orders.
const (
	Asc  Order = "asc"
	Desc Order = "desc"
)

// Page sizes of a listing.
const (
	DefaultLimit = 100
	MaxLimit     = 1000
)

// ListOptions select one page of a folder listing. Zero values mean the
// defaults: the first page, DefaultLimit items, by name, ascending.
type ListOptions struct {
	Cursor string
	Limit  int
	Sort   SortKey
	Order  Order
}

// ListPage is one page of a folder listing.
type ListPage struct {
	Folder Item   // the listed folder
	Items  []Item // this page, in order
	// NextCursor continues the listing; "" when this is the last page.
	NextCursor string
}

// normalize fills in the defaults and checks every option.
func (o ListOptions) normalize() (ListOptions, error) {
	switch {
	case o.Limit == 0:
		o.Limit = DefaultLimit
	case o.Limit < 1 || o.Limit > MaxLimit:
		return o, apperr.Newf(apperr.InvalidRequest, "limit must be between 1 and %d, got %d", MaxLimit, o.Limit)
	}
	switch o.Sort {
	case "":
		o.Sort = SortName
	case SortName, SortSize, SortModTime, SortKind:
	default:
		return o, apperr.Newf(apperr.InvalidRequest, "sort must be name, size, mod_time, or kind, got %q", o.Sort)
	}
	switch o.Order {
	case "":
		o.Order = Asc
	case Asc, Desc:
	default:
		return o, apperr.Newf(apperr.InvalidRequest, "order must be asc or desc, got %q", o.Order)
	}
	return o, nil
}

// List returns one page of the folder at path.
func (s *Local) List(ctx context.Context, owner, path string, opts ListOptions) (ListPage, error) {
	opts, err := opts.normalize()
	if err != nil {
		return ListPage{}, err
	}
	var after *cursor
	if opts.Cursor != "" {
		if after, err = decodeCursor(opts.Cursor, opts); err != nil {
			return ListPage{}, err
		}
	}
	rel, err := s.resolveVisible(owner, path)
	if err != nil {
		return ListPage{}, err
	}
	return run(ctx, s.hooks, Event{Op: OpList, Owner: owner, Path: rel}, func() (ListPage, error) {
		var page ListPage
		err := s.withRoot(owner, func(root *os.Root) error {
			items, folder, err := readFolder(root, owner, rel, path)
			if err != nil {
				return err
			}
			page.Folder = folder
			less := compareItems(opts.Sort, opts.Order)
			slices.SortFunc(items, less)
			start := 0
			if after != nil {
				pos := after.item()
				start, _ = slices.BinarySearchFunc(items, pos, less)
				for start < len(items) && less(items[start], pos) <= 0 {
					start++
				}
			}
			end := min(start+opts.Limit, len(items))
			page.Items = items[start:end]
			if end < len(items) {
				page.NextCursor = encodeCursor(items[end-1], opts)
			}
			return nil
		})
		return page, err
	})
}

// readFolder returns the entries of the folder rel (as items) and the
// folder itself.
func readFolder(root *os.Root, owner, rel, apiPath string) ([]Item, Item, error) {
	name := filepath.FromSlash(rel)
	info, err := root.Lstat(name)
	if err != nil {
		return nil, Item{}, fsError(err, apiPath)
	}
	folder := NewItem(owner, rel, info)
	if folder.Kind != KindDir {
		return nil, Item{}, apperr.Newf(apperr.InvalidRequest, "%s is not a folder", apiPath)
	}
	f, err := root.Open(name)
	if err != nil {
		return nil, Item{}, fsError(err, apiPath)
	}
	defer func() { _ = f.Close() }() // read-only
	entries, err := f.ReadDir(-1)
	if err != nil {
		return nil, Item{}, fsError(err, apiPath)
	}
	items := make([]Item, 0, len(entries))
	for _, e := range entries {
		if storage.IsTempName(e.Name()) {
			continue // a write in progress, or left by a crash: never an item
		}
		info, err := e.Info() // Lstat: links are listed, never followed
		if storage.IsNotFound(err) {
			continue // removed since ReadDir
		}
		if err != nil {
			return nil, Item{}, fsError(err, apiPath)
		}
		it := NewItem(owner, path.Join(rel, e.Name()), info)
		if it.Kind == KindFile {
			it.MIME = mimeByExtension(it.Name) // no sniffing in listings
		}
		items = append(items, it)
	}
	return items, folder, nil
}

// kindRank orders kinds: folders first, then files, links, and others.
var kindRank = map[Kind]int{KindDir: 0, KindFile: 1, KindSymlink: 2, KindOther: 3}

// compareNames orders names case-insensitively, then exactly, so the order
// is total.
func compareNames(a, b string) int {
	if c := strings.Compare(strings.ToLower(a), strings.ToLower(b)); c != 0 {
		return c
	}
	return strings.Compare(a, b)
}

// compareItems returns the total order of a listing: the sort key, then
// the name; Desc reverses the whole order.
func compareItems(key SortKey, order Order) func(a, b Item) int {
	return func(a, b Item) int {
		var c int
		switch key {
		case SortSize:
			c = cmp.Compare(a.Size, b.Size)
		case SortModTime:
			c = a.ModTime.Compare(b.ModTime)
		case SortKind:
			c = cmp.Compare(kindRank[a.Kind], kindRank[b.Kind])
		}
		if c == 0 {
			c = compareNames(a.Name, b.Name)
		}
		if order == Desc {
			c = -c
		}
		return c
	}
}

// cursor is the position after the last item of a page. It records the
// sort, so it cannot be used with another one.
type cursor struct {
	Sort    SortKey `json:"s"`
	Order   Order   `json:"o"`
	Name    string  `json:"n"`
	Size    int64   `json:"z,omitempty"`
	ModTime int64   `json:"t,omitempty"` // Unix nanoseconds
	Kind    Kind    `json:"k,omitempty"`
}

func (c *cursor) item() Item {
	return Item{Name: c.Name, Size: c.Size, ModTime: time.Unix(0, c.ModTime), Kind: c.Kind}
}

func encodeCursor(last Item, opts ListOptions) string {
	c := cursor{Sort: opts.Sort, Order: opts.Order, Name: last.Name, Size: last.Size, ModTime: last.ModTime.UnixNano(), Kind: last.Kind}
	data, _ := json.Marshal(c) // strings and numbers always marshal
	return base64.RawURLEncoding.EncodeToString(data)
}

func decodeCursor(s string, opts ListOptions) (*cursor, error) {
	data, err := base64.RawURLEncoding.DecodeString(s)
	var c cursor
	if err == nil {
		err = json.Unmarshal(data, &c)
	}
	if err != nil || c.Name == "" {
		return nil, apperr.New(apperr.InvalidRequest, "the cursor is not valid; start again without a cursor")
	}
	if c.Sort != opts.Sort || c.Order != opts.Order {
		return nil, apperr.Newf(apperr.InvalidRequest, "the cursor belongs to a listing sorted by %s %s; use the same sort and order, or start again", c.Sort, c.Order)
	}
	return &c, nil
}
