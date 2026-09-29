//go:build linux

package storage

import (
	"io/fs"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

// birthNeedsDir says BirthTime asks the file system through the folder.
const birthNeedsDir = true

// BirthTime returns when the item name in the open folder dir was created
// on its file system ("." for dir itself), for the date an item was added
// (FR-215). ok is false where the file system does not record it (statx
// without STATX_BTIME); callers use the modification time then. Links
// are not followed.
func BirthTime(dir *os.File, name string, _ fs.FileInfo) (time.Time, bool) {
	var st unix.Statx_t
	err := unix.Statx(int(dir.Fd()), name, unix.AT_SYMLINK_NOFOLLOW|unix.AT_STATX_DONT_SYNC, unix.STATX_BTIME, &st)
	if err != nil || st.Mask&unix.STATX_BTIME == 0 {
		return time.Time{}, false
	}
	return time.Unix(st.Btime.Sec, int64(st.Btime.Nsec)), true
}
