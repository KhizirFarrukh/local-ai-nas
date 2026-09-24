//go:build windows

package storage

import (
	"errors"
	"io/fs"

	"golang.org/x/sys/windows"
)

// IsNotFound reports whether err means that a path does not exist,
// including a path that goes through a file ("a.txt\b"), which Windows may
// report as ERROR_DIRECTORY or ERROR_INVALID_NAME, and a name too long to
// exist (ERROR_FILENAME_EXCED_RANGE).
func IsNotFound(err error) bool {
	return errors.Is(err, fs.ErrNotExist) ||
		errors.Is(err, windows.ERROR_DIRECTORY) ||
		errors.Is(err, windows.ERROR_INVALID_NAME) ||
		errors.Is(err, windows.ERROR_FILENAME_EXCED_RANGE)
}
