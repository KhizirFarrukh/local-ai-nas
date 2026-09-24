package files

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

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

// Local implements Service on the local disk. Every path goes through the
// storage resolver, and every file-system call through the namespace's
// os.Root.
type Local struct {
	resolver *storage.Resolver
	hooks    Hooks
}

var _ Service = (*Local)(nil)

// NewLocal returns the local files service. A nil hooks means NopHooks.
func NewLocal(r *storage.Resolver, hooks Hooks) *Local {
	if hooks == nil {
		hooks = NopHooks{}
	}
	return &Local{resolver: r, hooks: hooks}
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
	rel, err := s.resolver.Resolve(storage.FilesArea, owner, path)
	if err != nil {
		return Item{}, err
	}
	return run(ctx, s.hooks, Event{Op: OpStat, Owner: owner, Path: rel}, func() (Item, error) {
		var it Item
		err := s.withRoot(owner, func(root *os.Root) error {
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
	case storage.IsNotFound(err):
		return apperr.Wrap(apperr.NotFound, "no item at "+path, err)
	case errors.Is(err, fs.ErrExist):
		return apperr.Wrap(apperr.Conflict, "an item already exists at "+path, err)
	default:
		return apperr.Wrap(apperr.Internal, "the storage operation failed", err)
	}
}
