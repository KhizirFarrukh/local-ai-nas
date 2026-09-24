package storage

import (
	"fmt"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
)

// FreeFunc reports the bytes available to the server on the file system
// that holds dir. DiskFree is the real one; tests inject others.
type FreeFunc func(dir string) (uint64, error)

// SpaceGuard refuses writes that would leave less free space than the
// configured reserve (S01.2-T04, NFR-026). Writes declare their size up
// front, so the check happens before any byte is stored.
type SpaceGuard struct {
	dir     string
	reserve uint64
	free    FreeFunc
}

// NewSpaceGuard returns a guard for the file system of dir that keeps
// reserve bytes free. A nil free uses DiskFree.
func NewSpaceGuard(dir string, reserve int64, free FreeFunc) *SpaceGuard {
	if free == nil {
		free = DiskFree
	}
	if reserve < 0 {
		reserve = 0
	}
	return &SpaceGuard{dir: dir, reserve: uint64(reserve), free: free}
}

// Free returns the bytes available on the guarded file system.
func (g *SpaceGuard) Free() (uint64, error) {
	return g.free(g.dir)
}

// Reserve returns the bytes that writes must leave free.
func (g *SpaceGuard) Reserve() uint64 { return g.reserve }

// Check returns an insufficient_storage error (HTTP 507) if writing size
// more bytes would leave less than the reserve free.
func (g *SpaceGuard) Check(size int64) error {
	if size < 0 {
		return apperr.Newf(apperr.InvalidRequest, "invalid size %d", size)
	}
	free, err := g.free(g.dir)
	if err != nil {
		return apperr.Wrap(apperr.Internal, "cannot read the free space", err)
	}
	need := uint64(size)
	if free < need || free-need < g.reserve {
		return apperr.Newf(apperr.InsufficientStorage,
			"not enough free space: the write needs %s, %s is free, and %s must stay free",
			formatBytes(need), formatBytes(free), formatBytes(g.reserve))
	}
	return nil
}

// formatBytes renders a byte count for messages, such as "1.5 GiB".
func formatBytes(n uint64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := uint64(unit), 0
	for m := n / unit; m >= unit && exp < 4; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTP"[exp])
}
