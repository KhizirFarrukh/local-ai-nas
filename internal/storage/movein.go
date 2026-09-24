package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ErrNotSameDevice means that a file could not be renamed into an area
// because it lies on another file system; the caller copies it instead.
var ErrNotSameDevice = errors.New("storage: the file is on another file system")

// MoveInto renames the file src, which lies outside the areas (a finished
// upload in the internal data, S01.4-T03), to rel in the namespace of
// area. rel must be a new name that the caller checked through the
// namespace's os.Root first (no links among its parents): os.Root cannot
// rename across roots, so this is the one move by path. A rename across
// file systems returns ErrNotSameDevice.
func (r *Resolver) MoveInto(area, namespace, rel, src string) error {
	local := filepath.FromSlash(rel)
	if rel == "." || !filepath.IsLocal(local) {
		return fmt.Errorf("storage: %q is not a path inside a namespace", rel)
	}
	dst := filepath.Join(r.layout.Area(area, namespace), local)
	if err := os.Rename(src, dst); err != nil {
		if isCrossDevice(err) {
			return fmt.Errorf("%w: %w", ErrNotSameDevice, err)
		}
		return err
	}
	return nil
}
