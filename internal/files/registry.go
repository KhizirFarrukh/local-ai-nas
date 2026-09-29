package files

import (
	"context"
	"path"
)

// Registry keeps the stable ID of every file and folder (S01.3-T11,
// ADR-0040, FR-346). An item gets an ID when the app creates it, or when
// the app first sees it; rename and move keep the ID; a copy gets new IDs;
// writing new content to an existing path keeps that item's ID (a new
// version of the same file); a delete retires the ID. Paths are relative
// to the namespace, as in Hashes.
//
// Like Hashes, every method is best effort for the files service: an
// operation that succeeded is not undone because its ID could not be
// recorded, and an item without an ID gets one when it is next seen
// (S01.3-T12) or by the reconciliation scan (S05.7). The operation journal
// (S01.4-T09, ADR-0041) makes the two consistent after a crash.
type Registry interface {
	// Ensure gives the item at rel, and every folder above it, an ID if it
	// has none, and returns the item's ID. It is idempotent.
	Ensure(ctx context.Context, owner, rel string, kind Kind) (string, error)
	// EnsureAll does Ensure for many items in one transaction (a copied
	// tree).
	EnsureAll(ctx context.Context, owner string, items []RegistryEntry) error
	// Moved follows a rename or move of the item at from, and everything
	// in it, to to: the IDs stay. Items that were at to are retired (a
	// move that replaced them).
	Moved(ctx context.Context, owner, from, to string) error
	// Deleted retires the item at rel and everything in it.
	Deleted(ctx context.Context, owner, rel string) error
	// Lookup returns the ID of the present item at rel, or "".
	Lookup(ctx context.Context, owner, rel string) (string, error)
}

// RegistryEntry is one item for EnsureAll.
type RegistryEntry struct {
	Rel  string
	Kind Kind
}

// NopRegistry keeps nothing.
type NopRegistry struct{}

// Ensure implements Registry.
func (NopRegistry) Ensure(context.Context, string, string, Kind) (string, error) { return "", nil }

// EnsureAll implements Registry.
func (NopRegistry) EnsureAll(context.Context, string, []RegistryEntry) error { return nil }

// Moved implements Registry.
func (NopRegistry) Moved(context.Context, string, string, string) error { return nil }

// Deleted implements Registry.
func (NopRegistry) Deleted(context.Context, string, string) error { return nil }

// Lookup implements Registry.
func (NopRegistry) Lookup(context.Context, string, string) (string, error) { return "", nil }

// registerCopy records the items a copy created below final: the entries
// of a copied folder (relative to the copy), or the one copied file.
func (s *Local) registerCopy(ctx context.Context, owner, final string, entries []copyEntry, isDir bool) {
	if !isDir {
		_, _ = s.registry.Ensure(ctx, owner, final, KindFile) // best effort (Registry)
		return
	}
	list := make([]RegistryEntry, 0, len(entries))
	for _, e := range entries {
		k := KindFile
		if e.info.IsDir() {
			k = KindDir
		}
		list = append(list, RegistryEntry{Rel: path.Join(final, e.rel), Kind: k})
	}
	_ = s.registry.EnsureAll(ctx, owner, list) // best effort (Registry)
}
