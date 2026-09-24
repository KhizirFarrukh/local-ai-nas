//go:build windows

package storage

import (
	"fmt"

	"golang.org/x/sys/windows"
)

// DeviceID identifies the volume that holds path (its volume serial
// number). Two paths with the same ID are on the same volume, so a rename
// between them is atomic.
func DeviceID(path string) (uint64, error) {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, fmt.Errorf("storage: device of %s: %w", path, err)
	}
	// FILE_FLAG_BACKUP_SEMANTICS is needed to open a directory.
	h, err := windows.CreateFile(p, 0,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return 0, fmt.Errorf("storage: device of %s: %w", path, err)
	}
	defer func() { _ = windows.CloseHandle(h) }()
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(h, &info); err != nil {
		return 0, fmt.Errorf("storage: device of %s: %w", path, err)
	}
	return uint64(info.VolumeSerialNumber), nil
}
