package files

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
)

// Service is the only way the API reaches the files area (S01.3). Paths
// are API paths ("/docs/a.txt", with "/" as the namespace root); every
// method resolves them in the owner's namespace, so no call can reach
// another namespace or the photos area. The methods grow with the S01.3
// tasks, each together with its endpoint.
type Service interface {
	// Stat returns the item at path.
	Stat(ctx context.Context, owner, path string) (Item, error)
	// List returns one page of the folder at path (S01.3-T02).
	List(ctx context.Context, owner, path string, opts ListOptions) (ListPage, error)
	// CreateFolder creates the folder at path and reports whether it
	// created it (S01.3-T04).
	CreateFolder(ctx context.Context, owner, path string, opts FolderOptions) (Item, bool, error)
	// Upload writes exactly size bytes from body to the file at path, all
	// or nothing, and reports whether it created a new file (false: it
	// replaced one) (S01.3-T05).
	Upload(ctx context.Context, owner, path string, body io.Reader, size int64, opts UploadOptions) (Item, bool, error)
	// Download opens the file at path for reading; the caller closes it.
	// The item describes exactly the bytes that are read (S01.3-T06).
	Download(ctx context.Context, owner, path string) (Item, io.ReadSeekCloser, error)
	// Rename gives the item at path a new name in its folder, and Move
	// moves it to another path; both return the item at its new path
	// (S01.3-T07).
	Rename(ctx context.Context, owner, path, newName string, opts MoveOptions) (Item, error)
	Move(ctx context.Context, owner, from, to string, opts MoveOptions) (Item, error)
	// Copy copies the file or folder at from to the path to and reports
	// whether it created a new item (false: it replaced a file)
	// (S01.3-T08).
	Copy(ctx context.Context, owner, from, to string, opts CopyOptions) (Item, bool, error)
	// Delete deletes the item at path permanently (S01.3-T09).
	Delete(ctx context.Context, owner, path string, opts DeleteOptions) error
}

// Op names a file operation, for hooks and logs.
type Op string

// The operations.
const (
	OpStat         Op = "stat"
	OpList         Op = "list"
	OpCreateFolder Op = "create_folder"
	OpUpload       Op = "upload"
	OpDownload     Op = "download"
	OpRename       Op = "rename"
	OpMove         Op = "move"
	OpCopy         Op = "copy"
	OpDelete       Op = "delete"
)

// Event describes one operation for the hooks.
type Event struct {
	Op    Op
	Owner string // the owner namespace, such as u0001
	// Path is the resolved path in the namespace ("." is the root).
	Path string
	// Target is the second resolved path of a rename, move, or copy.
	Target string
}

// Hooks are called around every operation. They are the attachment points
// for later stages: policy checks (S03) and quotas (S10) in Before, and
// indexing (S06), trash (S08), and auditing in After.
type Hooks interface {
	// Before runs before the operation. An error cancels the operation
	// and is returned to the caller; After is then not called.
	Before(ctx context.Context, e Event) error
	// After runs once the operation has run, with its error (nil on
	// success).
	After(ctx context.Context, e Event, err error)
}

// NopHooks does nothing; S01 has no subscribers.
type NopHooks struct{}

// Before implements Hooks.
func (NopHooks) Before(context.Context, Event) error { return nil }

// After implements Hooks.
func (NopHooks) After(context.Context, Event, error) {}

// Options configures the local files service.
type Options struct {
	// Hooks are called around every operation; nil means NopHooks.
	Hooks Hooks
	// Space refuses writes that would use the free-space reserve. Nil
	// means no check (tests).
	Space *storage.SpaceGuard
	// CopyLimits bound a copy within one request; zero means no limit.
	CopyLimits CopyLimits
}

// Local implements Service on the local disk. Every path goes through the
// storage resolver, and every file-system call through the namespace's
// os.Root.
type Local struct {
	resolver   *storage.Resolver
	hooks      Hooks
	space      *storage.SpaceGuard
	copyLimits CopyLimits
}

var _ Service = (*Local)(nil)

// NewLocal returns the local files service.
func NewLocal(r *storage.Resolver, o Options) *Local {
	if o.Hooks == nil {
		o.Hooks = NopHooks{}
	}
	return &Local{resolver: r, hooks: o.Hooks, space: o.Space, copyLimits: o.CopyLimits}
}

// run calls the hooks around op.
func run[T any](ctx context.Context, h Hooks, e Event, op func() (T, error)) (T, error) {
	if err := h.Before(ctx, e); err != nil {
		var zero T
		return zero, err
	}
	res, err := op()
	h.After(ctx, e, err)
	return res, err
}

// resolveVisible resolves path for an operation on an existing item. The
// server's temporary files (storage.TempPrefix) are not items: a path
// through one is not found, so a partial upload can never be read.
func (s *Local) resolveVisible(owner, path string) (string, error) {
	rel, err := s.resolver.Resolve(storage.FilesArea, owner, path)
	if err != nil {
		return "", err
	}
	for _, name := range strings.Split(rel, "/") {
		if storage.IsTempName(name) {
			return "", apperr.New(apperr.NotFound, "no item at "+path)
		}
	}
	return rel, nil
}

// withRoot opens the owner's namespace root for the duration of fn.
func (s *Local) withRoot(owner string, fn func(root *os.Root) error) (err error) {
	root, err := s.resolver.OpenRoot(storage.FilesArea, owner)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := root.Close(); cerr != nil && err == nil {
			err = apperr.Wrap(apperr.Internal, "closing the storage area failed", cerr)
		}
	}()
	return fn(root)
}

// Stat implements Service.
func (s *Local) Stat(ctx context.Context, owner, path string) (Item, error) {
	rel, err := s.resolveVisible(owner, path)
	if err != nil {
		return Item{}, err
	}
	return run(ctx, s.hooks, Event{Op: OpStat, Owner: owner, Path: rel}, func() (Item, error) {
		var it Item
		err := s.withRoot(owner, func(root *os.Root) error {
			if err := refuseLinkParents(root, rel, path); err != nil {
				return err
			}
			info, err := root.Lstat(filepath.FromSlash(rel))
			if err != nil {
				return fsError(err, path)
			}
			it, err = withDetails(root, NewItem(owner, rel, info), path)
			return err
		})
		return it, err
	})
}

// fsError maps a file-system error to an apperr error for the API path
// the client used.
func fsError(err error, path string) error {
	switch {
	case storage.IsEscape(err):
		return apperr.Wrap(apperr.OutsideRoot, path+" leads outside your storage area (through a symbolic link)", err)
	case storage.IsNotFound(err):
		return apperr.Wrap(apperr.NotFound, "no item at "+path, err)
	case errors.Is(err, fs.ErrExist):
		return apperr.Wrap(apperr.Conflict, "an item already exists at "+path, err)
	default:
		return apperr.Wrap(apperr.Internal, "the storage operation failed", err)
	}
}
