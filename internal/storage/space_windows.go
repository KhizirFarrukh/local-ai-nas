//go:build windows

package storage

import (
	"fmt"

	"golang.org/x/sys/windows"
)

// DiskFree reports the bytes available to the current user on the volume
// that holds dir (GetDiskFreeSpaceEx; quotas are taken into account).
func DiskFree(dir string) (uint64, error) {
	p, err := windows.UTF16PtrFromString(dir)
	if err != nil {
		return 0, fmt.Errorf("storage: free space of %s: %w", dir, err)
	}
	var avail, total, totalFree uint64
	if err := windows.GetDiskFreeSpaceEx(p, &avail, &total, &totalFree); err != nil {
		return 0, fmt.Errorf("storage: free space of %s: %w", dir, err)
	}
	return avail, nil
}
