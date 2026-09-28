package uploads

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	tus "github.com/tus/tusd/v2/pkg/handler"
)

// tusd's file store writes the absolute paths of an upload's data and
// .info files into the .info file and trusts them when the upload
// resumes. After the storage root has moved (S01.2-T07, NFR-036), those
// paths name the old place: a resumed upload would write there, and
// finishing or removing it would read or delete files outside the root.
// So the server derives both paths from its upload directory and the
// upload ID, and repairs the .info files when it starts.

// dataPath is where the file store keeps an upload's bytes.
func (s *Server) dataPath(id string) string { return filepath.Join(s.o.Dir, id) }

// infoPath is where the file store keeps an upload's .info file.
func (s *Server) infoPath(id string) string { return filepath.Join(s.o.Dir, id+".info") }

// repairInfoPaths points every .info file in dir at dir, and returns how
// many it changed. A file that cannot be read or is not an upload of this
// server is left alone (the expiry cleanup removes it); a failed rewrite
// is an error, because tusd would then use the old place.
func repairInfoPaths(ctx context.Context, dir string, log *slog.Logger) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}
	changed := 0
	for _, e := range entries {
		id, ok := strings.CutSuffix(e.Name(), ".info")
		if !ok || !e.Type().IsRegular() || !isUploadID(id) {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path) // #nosec G304 -- a file in the upload directory
		if err != nil {
			log.WarnContext(ctx, "reading an upload's info failed", "upload", id, "error", err.Error())
			continue
		}
		var info tus.FileInfo
		if err := json.Unmarshal(data, &info); err != nil || info.ID != id {
			log.WarnContext(ctx, "an upload's info file is not valid; the cleanup removes it", "upload", id)
			continue
		}
		want := map[string]string{"Path": filepath.Join(dir, id), "InfoPath": path}
		if info.Storage["Path"] == want["Path"] && info.Storage["InfoPath"] == want["InfoPath"] {
			continue
		}
		if info.Storage == nil {
			info.Storage = map[string]string{"Type": "filestore"}
		}
		info.Storage["Path"], info.Storage["InfoPath"] = want["Path"], want["InfoPath"]
		if err := writeInfo(path, info); err != nil {
			return changed, err
		}
		changed++
	}
	return changed, nil
}

// writeInfo replaces an .info file in one step: a temporary file in the
// same directory, flushed, then renamed over it.
func writeInfo(path string, info tus.FileInfo) (err error) {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".info-*")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = os.Remove(tmp.Name())
		}
	}()
	if _, err = tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err = tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
