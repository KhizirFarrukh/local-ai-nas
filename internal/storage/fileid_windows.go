//go:build windows

package storage

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

// FileID identifies the file behind an open handle on its volume: the
// volume serial number and the file index. A file replaced by another one
// (for example through an atomic rename) gets a different ID, even with
// the same size and time.
func FileID(f *os.File) (string, error) {
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(windows.Handle(f.Fd()), &info); err != nil {
		return "", fmt.Errorf("storage: file ID of %s: %w", f.Name(), err)
	}
	return fmt.Sprintf("%x:%x:%x", info.VolumeSerialNumber, info.FileIndexHigh, info.FileIndexLow), nil
}
