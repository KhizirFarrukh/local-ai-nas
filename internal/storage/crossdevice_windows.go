//go:build windows

package storage

import (
	"errors"

	"golang.org/x/sys/windows"
)

// isCrossDevice reports a rename that failed because source and target
// are on different volumes.
func isCrossDevice(err error) bool { return errors.Is(err, windows.ERROR_NOT_SAME_DEVICE) }

// crossDeviceErrno is the error a cross-device rename returns (tests).
var crossDeviceErrno error = windows.ERROR_NOT_SAME_DEVICE
