package logging

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

// RotatingFile is an io.WriteCloser that appends to a log file and rotates
// it by size: when a write would make the file larger than maxSize, the
// file is renamed to <name>.1, older files move up one number (<name>.1 to
// <name>.2, and so on), and files beyond maxFiles are deleted. A single
// write larger than maxSize still goes into one (new) file. It is safe for
// concurrent use.
type RotatingFile struct {
	mu       sync.Mutex
	path     string
	maxSize  int64
	maxFiles int
	file     *os.File
	size     int64
}

// OpenRotatingFile opens (or creates) the log file at path for appending.
// maxSize is the size in bytes at which it is rotated; maxFiles is the
// number of rotated files kept besides the current one.
func OpenRotatingFile(path string, maxSize int64, maxFiles int) (*RotatingFile, error) {
	if maxSize <= 0 || maxFiles < 1 {
		return nil, fmt.Errorf("logging: invalid rotation settings: max size %d, max files %d", maxSize, maxFiles)
	}
	r := &RotatingFile{path: path, maxSize: maxSize, maxFiles: maxFiles}
	if err := r.open(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *RotatingFile) open() error {
	f, err := os.OpenFile(r.path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o600)
	if err != nil {
		return fmt.Errorf("logging: open log file: %w", err)
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return fmt.Errorf("logging: stat log file: %w", err)
	}
	r.file, r.size = f, info.Size()
	return nil
}

// Write appends p to the current file, rotating first if p would not fit.
func (r *RotatingFile) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.file == nil {
		return 0, fs.ErrClosed
	}
	if r.size > 0 && r.size+int64(len(p)) > r.maxSize {
		if err := r.rotate(); err != nil {
			// Keep logging into the current file; rotation is retried on
			// the next write that does not fit.
			if r.file == nil {
				return 0, err
			}
		}
	}
	n, err := r.file.Write(p)
	r.size += int64(n)
	return n, err
}

// rotate closes the current file, shifts the numbered files, and opens a
// new, empty file. If the shift fails (for example because another program
// holds a file open on Windows), the current file is reopened.
func (r *RotatingFile) rotate() error {
	if err := r.file.Close(); err != nil {
		r.file = nil
		return fmt.Errorf("logging: close log file: %w", err)
	}
	r.file = nil
	shiftErr := r.shift()
	if err := r.open(); err != nil {
		return errors.Join(shiftErr, err)
	}
	return shiftErr
}

func (r *RotatingFile) shift() error {
	numbered := func(i int) string { return r.path + "." + strconv.Itoa(i) }
	if err := os.Remove(numbered(r.maxFiles)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("logging: remove oldest log file: %w", err)
	}
	for i := r.maxFiles - 1; i >= 1; i-- {
		if err := os.Rename(numbered(i), numbered(i+1)); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("logging: rotate log file: %w", err)
		}
	}
	if err := os.Rename(r.path, numbered(1)); err != nil {
		return fmt.Errorf("logging: rotate log file: %w", err)
	}
	return nil
}

// Close closes the current file. Later writes fail with fs.ErrClosed.
func (r *RotatingFile) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.file == nil {
		return nil
	}
	err := r.file.Close()
	r.file = nil
	return err
}

// LogFileName is the name of the current log file in the logs directory.
const LogFileName = "nas.log"

// DefaultPath returns the log file path inside the logs directory.
func DefaultPath(logsDir string) string { return filepath.Join(logsDir, LogFileName) }
