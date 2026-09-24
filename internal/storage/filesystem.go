package storage

// DeviceFunc identifies the file system that holds a path. DeviceID is the
// real one; tests inject others to simulate a split layout.
type DeviceFunc func(path string) (uint64, error)

// How finished uploads move from tmp/uploads into the files area (plan
// 8.18, S01.4).
const (
	// FinalizeRename is an atomic rename: both are on one file system.
	FinalizeRename = "rename"
	// FinalizeCopy copies, syncs, and then renames within the target
	// directory: slower, used when the two are on different file systems.
	FinalizeCopy = "copy"
)

// FinalizeMode reports how finished uploads reach the files area:
// FinalizeRename when tmp/uploads and the default namespace of the files
// area are on the same file system, otherwise FinalizeCopy. A nil dev uses
// DeviceID.
func (l Layout) FinalizeMode(dev DeviceFunc) (string, error) {
	if dev == nil {
		dev = DeviceID
	}
	tmp, err := dev(l.TmpUploads)
	if err != nil {
		return "", err
	}
	files, err := dev(l.Area(FilesArea, DefaultNamespace))
	if err != nil {
		return "", err
	}
	if tmp == files {
		return FinalizeRename, nil
	}
	return FinalizeCopy, nil
}

// Writable checks that every layout directory still exists, is a real
// directory, and accepts new files, without leaving a trace (see Init).
func (l Layout) Writable() error {
	for _, dir := range l.dirs() {
		if err := ensureExistingDir(dir); err != nil {
			return err
		}
		if err := checkWritable(dir); err != nil {
			return err
		}
	}
	return nil
}
