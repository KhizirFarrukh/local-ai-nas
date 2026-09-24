//go:build unix

package storage

import (
	"fmt"
	"os"
	"syscall"
)

// FileID identifies the file behind an open handle on its file system:
// device and inode. A file replaced by another one (for example through
// an atomic rename) gets a different ID, even with the same size and time.
func FileID(f *os.File) (string, error) {
	info, err := f.Stat()
	if err != nil {
		return "", fmt.Errorf("storage: file ID: %w", err)
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return "", fmt.Errorf("storage: file ID: no stat data for %s", f.Name())
	}
	return fmt.Sprintf("%x:%x", st.Dev, st.Ino), nil
}
