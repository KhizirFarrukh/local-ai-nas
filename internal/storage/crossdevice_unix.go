//go:build unix

package storage

import (
	"errors"

	"golang.org/x/sys/unix"
)

// isCrossDevice reports a rename that failed because source and target
// are on different file systems.
func isCrossDevice(err error) bool { return errors.Is(err, unix.EXDEV) }

// crossDeviceErrno is the error a cross-device rename returns (tests).
var crossDeviceErrno error = unix.EXDEV
