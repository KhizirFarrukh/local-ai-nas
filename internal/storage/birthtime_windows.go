//go:build windows

package storage

import (
	"io/fs"
	"os"
	"syscall"
	"time"
)

// birthNeedsDir says BirthTime asks the file system through the folder;
// Windows reports the creation time with every Lstat already.
const birthNeedsDir = false

// BirthTime returns when the item was created on its file system, for the
// date an item was added (FR-215), from info (Lstat of the item). ok is
// false when info has no creation time; callers use the modification time
// then. Note: NTFS "tunneling" can give a file that replaces another one
// within 15 seconds the old file's creation time.
func BirthTime(_ *os.File, _ string, info fs.FileInfo) (time.Time, bool) {
	d, ok := info.Sys().(*syscall.Win32FileAttributeData)
	if !ok || d.CreationTime == (syscall.Filetime{}) {
		return time.Time{}, false
	}
	return time.Unix(0, d.CreationTime.Nanoseconds()), true
}
