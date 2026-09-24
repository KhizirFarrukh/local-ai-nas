//go:build unix

package storage

import (
	"fmt"

	"golang.org/x/sys/unix"
)

// DeviceID identifies the file system that holds path (stat st_dev). Two
// paths with the same ID are on the same file system, so a rename between
// them is atomic.
func DeviceID(path string) (uint64, error) {
	var st unix.Stat_t
	if err := unix.Stat(path, &st); err != nil {
		return 0, fmt.Errorf("storage: device of %s: %w", path, err)
	}
	return uint64(st.Dev), nil //nolint:gosec,unconvert // st_dev's type differs between Unix systems
}
