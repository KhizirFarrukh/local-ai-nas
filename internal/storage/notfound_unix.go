//go:build unix

package storage

import (
	"errors"
	"io/fs"

	"golang.org/x/sys/unix"
)

// IsNotFound reports whether err means that a path does not exist,
// including a path that goes through a file ("a.txt/b"), which Unix
// reports as ENOTDIR rather than ENOENT, and a name too long to exist
// (ENAMETOOLONG; found by the S01.7-T02 attack suite).
func IsNotFound(err error) bool {
	return errors.Is(err, fs.ErrNotExist) || errors.Is(err, unix.ENOTDIR) || errors.Is(err, unix.ENAMETOOLONG)
}
