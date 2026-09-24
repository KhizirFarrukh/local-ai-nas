package files

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
)

// sniffBytes is how much of a file is read to guess its type, the amount
// http.DetectContentType looks at.
const sniffBytes = 512

// mimeByExtension returns the media type for a name's extension, or "".
// The table is Go's built-in one plus what the OS adds (/etc/mime.types on
// Linux, the registry on Windows).
func mimeByExtension(name string) string {
	ext := filepath.Ext(name)
	if ext == "" {
		return ""
	}
	return mime.TypeByExtension(strings.ToLower(ext))
}

// withDetails adds the media type and the ETag to a file item, reading the
// start of the file and its identity through root. Folders and links are
// returned unchanged (S01.3-T03).
func withDetails(root *os.Root, it Item, apiPath string) (Item, error) {
	if it.Kind != KindFile {
		return it, nil
	}
	f, err := root.Open(filepath.FromSlash(it.RelPath))
	if err != nil {
		return it, fsError(err, apiPath)
	}
	defer func() { _ = f.Close() }() // read-only
	id, err := storage.FileID(f)
	if err != nil {
		return it, fsError(err, apiPath)
	}
	it.ETag = etag(it.Size, it.ModTime, id)
	it.MIME = mimeByExtension(it.Name)
	if it.MIME == "" {
		head := make([]byte, sniffBytes)
		n, err := io.ReadFull(f, head)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return it, fsError(err, apiPath)
		}
		it.MIME = http.DetectContentType(head[:n])
	}
	return it, nil
}

// etag is a strong entity tag for a file version: a hash of its size,
// modification time, and file ID. Writing the file changes the time;
// replacing it (an atomic rename) changes the ID.
func etag(size int64, mod time.Time, id string) string {
	sum := sha256.Sum256(fmt.Appendf(nil, "%d|%d|%s", size, mod.UnixNano(), id))
	return `"` + hex.EncodeToString(sum[:12]) + `"`
}
