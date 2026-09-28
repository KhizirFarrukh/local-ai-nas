//go:build !linux && !windows

package storage

import (
	"io/fs"
	"os"
	"time"
)

// birthNeedsDir says BirthTime asks the file system through the folder.
const birthNeedsDir = false

// BirthTime reports no creation time on this platform, so callers use the
// modification time (FR-215). Linux and Windows are the supported
// platforms (Q1).
func BirthTime(*os.File, string, fs.FileInfo) (time.Time, bool) {
	return time.Time{}, false
}
