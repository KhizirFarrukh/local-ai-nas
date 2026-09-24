//go:build unix

package storage

import (
	"fmt"

	"golang.org/x/sys/unix"
)

// DiskFree reports the bytes available to unprivileged users on the file
// system that holds dir (statfs f_bavail × f_bsize).
func DiskFree(dir string) (uint64, error) {
	var st unix.Statfs_t
	if err := unix.Statfs(dir, &st); err != nil {
		return 0, fmt.Errorf("storage: free space of %s: %w", dir, err)
	}
	return uint64(st.Bavail) * uint64(st.Bsize), nil //nolint:gosec // Statfs field types differ between Unix systems; both values are non-negative
}
