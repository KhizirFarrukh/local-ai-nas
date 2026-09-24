package config

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

// absRoot is an absolute path on the current OS.
func absRoot(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "nas")
}

func env(vars map[string]string) func(string) (string, bool) {
	return func(k string) (string, bool) {
		v, ok := vars[k]
		return v, ok
	}
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

// tomlPath quotes a path for a TOML basic string (Windows backslashes).
func tomlPath(p string) string {
	return `"` + strings.ReplaceAll(p, `\`, `\\`) + `"`
}

func TestLoadDefaults(t *testing.T) {
	root := absRoot(t)
	l, err := Load(Sources{Flags: map[string]string{"storage.root": root}})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := Default()
	want.Storage.Root = root
	if diff := cmp.Diff(want, l.Config); diff != "" {
		t.Errorf("config mismatch (-want +got):\n%s", diff)
	}
	if l.File != "" {
		t.Errorf("File = %q, want none", l.File)
	}
}

func TestLoadPrecedence(t *testing.T) {
	root := absRoot(t)
	file := writeConfig(t, "[storage]\nroot = "+tomlPath(root)+"\n[log]\nlevel = \"debug\"\nfile_max_files = 3\n[server]\nidle_timeout = \"5m\"\n")

	tests := []struct {
		name        string
		env         map[string]string
		flags       map[string]string
		wantLevel   string
		wantFiles   int
		wantIdle    time.Duration
		wantOrigins map[string]string
	}{
		{
			name:      "file over defaults",
			wantLevel: "debug", wantFiles: 3, wantIdle: 5 * time.Minute,
			wantOrigins: map[string]string{"storage.root": "file", "log.level": "file", "log.file_max_files": "file", "server.idle_timeout": "file"},
		},
		{
			name:      "env over file",
			env:       map[string]string{"LOCALAINAS_LOG_LEVEL": "warn", "LOCALAINAS_SERVER_IDLE_TIMEOUT": "1m"},
			wantLevel: "warn", wantFiles: 3, wantIdle: time.Minute,
			wantOrigins: map[string]string{"storage.root": "file", "log.level": "env LOCALAINAS_LOG_LEVEL", "log.file_max_files": "file", "server.idle_timeout": "env LOCALAINAS_SERVER_IDLE_TIMEOUT"},
		},
		{
			name:      "flag over env and file",
			env:       map[string]string{"LOCALAINAS_LOG_LEVEL": "warn"},
			flags:     map[string]string{"log.level": "error", "log.file_max_files": "9"},
			wantLevel: "error", wantFiles: 9, wantIdle: 5 * time.Minute,
			wantOrigins: map[string]string{"storage.root": "file", "log.level": "flag --log-level", "log.file_max_files": "flag --log-file-max-files", "server.idle_timeout": "file"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, err := Load(Sources{ConfigFile: file, LookupEnv: env(tt.env), Flags: tt.flags})
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if l.Log.Level != tt.wantLevel || l.Log.FileMaxFiles != tt.wantFiles || l.Server.IdleTimeout.Duration != tt.wantIdle {
				t.Errorf("got level=%q files=%d idle=%s, want %q %d %s",
					l.Log.Level, l.Log.FileMaxFiles, l.Server.IdleTimeout, tt.wantLevel, tt.wantFiles, tt.wantIdle)
			}
			if diff := cmp.Diff(tt.wantOrigins, l.Origins); diff != "" {
				t.Errorf("origins mismatch (-want +got):\n%s", diff)
			}
			if l.File != file {
				t.Errorf("File = %q, want %q", l.File, file)
			}
		})
	}
}

func TestLoadConfigFileLocation(t *testing.T) {
	root := absRoot(t)
	good := writeConfig(t, "[storage]\nroot = "+tomlPath(root)+"\n")
	other := writeConfig(t, "[log]\nlevel = \"nonsense\"\n")
	missing := filepath.Join(t.TempDir(), "missing.toml")

	tests := []struct {
		name     string
		src      Sources
		wantFile string
		wantErr  string
	}{
		{"flag wins over env and default", Sources{ConfigFile: good, LookupEnv: env(map[string]string{ConfigEnv: other}), DefaultConfigFile: other}, good, ""},
		{"env wins over default", Sources{LookupEnv: env(map[string]string{ConfigEnv: good}), DefaultConfigFile: other}, good, ""},
		{"default is used last", Sources{DefaultConfigFile: good}, good, ""},
		{"missing default is fine", Sources{DefaultConfigFile: missing, Flags: map[string]string{"storage.root": root}}, "", ""},
		{"missing file from flag fails", Sources{ConfigFile: missing}, "", "missing.toml"},
		{"missing file from env fails", Sources{LookupEnv: env(map[string]string{ConfigEnv: missing})}, "", "missing.toml"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, err := Load(tt.src)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Load error = %v, want one containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if l.File != tt.wantFile {
				t.Errorf("File = %q, want %q", l.File, tt.wantFile)
			}
		})
	}
}

func TestLoadFileErrors(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{"unknown key", "[storage]\nroots = \"/x\"\n", []string{"unknown setting", "roots"}},
		{"unknown section", "[stroage]\nroot = \"/x\"\n", []string{"unknown setting", "stroage"}},
		{"wrong type", "[log]\nfile_max_files = \"many\"\n", []string{"log.file_max_files"}},
		{"integer duration rejected", "[server]\nidle_timeout = 30\n", []string{"server.idle_timeout"}},
		{"bad size", "[uploads]\nmax_file_size = \"lots\"\n", []string{"uploads.max_file_size", "invalid size"}},
		{"bad duration", "[server]\nidle_timeout = \"soon\"\n", []string{"server.idle_timeout", "invalid duration"}},
		{"not TOML", "storage = [\n", []string{"config", "config.toml:1:"}},
		{"boolean for a string", "[storage]\nroot = true\n", []string{"storage.root", "wrong type (boolean)"}},
		{"integer for a string", "[server]\nbind = 8080\n", []string{"server.bind", "must be a string, got the integer 8080"}},
		{"string for an integer", "[log]\nfile_max_files = \"5\"\n", []string{"log.file_max_files", "must be an integer"}},
		{"string for a boolean", "[server]\nallow_container_bind = \"yes\"\n", []string{"server.allow_container_bind", "true or false"}},
		{"integer for a boolean", "[server]\nallow_container_bind = 1\n", []string{"server.allow_container_bind", "true or false"}},
		{"float", "[uploads]\nmax_file_size = 1.5\n", []string{"uploads.max_file_size", "wrong type (float)"}},
		{"array", "[log]\nlevel = [\"info\"]\n", []string{"log.level", "wrong type (array)"}},
		{"table", "[log.level]\nx = 1\n", []string{"log.level", "wrong type (table)"}},
		{"date", "[log]\nlevel = 2026-09-24\n", []string{"log.level", "wrong type (date or time)"}},
		{"value outside a section", "storage = \"/x\"\n", []string{"unknown setting \"storage\""}},
		{"several problems at once", "[log]\nlevel = 1\nfile_max_files = \"x\"\n", []string{"log.level", "log.file_max_files"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(Sources{ConfigFile: writeConfig(t, tt.content)})
			if err == nil {
				t.Fatal("Load succeeded, want an error")
			}
			for _, w := range tt.want {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("error %q does not contain %q", err, w)
				}
			}
		})
	}
}

func TestLoadValidation(t *testing.T) {
	root := absRoot(t)
	tests := []struct {
		name    string
		flags   map[string]string
		env     map[string]string
		wantKey string
		wantMsg string
	}{
		{"root missing", map[string]string{}, nil, "storage.root", "required"},
		{"root relative", map[string]string{"storage.root": "data/nas"}, nil, "storage.root", "absolute"},
		{"db dir relative", map[string]string{"storage.db_dir": "db"}, nil, "storage.db_dir", "absolute path or empty"},
		{"logs dir relative", map[string]string{"storage.logs_dir": "logs"}, nil, "storage.logs_dir", "absolute path or empty"},
		{"bind without port", map[string]string{"server.bind": "127.0.0.1"}, nil, "server.bind", "host:port"},
		{"bind without host", map[string]string{"server.bind": ":8080"}, nil, "server.bind", "host:port"},
		{"bind port zero", map[string]string{"server.bind": "127.0.0.1:0"}, nil, "server.bind", "1 to 65535"},
		{"bind port too big", map[string]string{"server.bind": "127.0.0.1:70000"}, nil, "server.bind", "1 to 65535"},
		{"bind port not a number", map[string]string{"server.bind": "127.0.0.1:http"}, nil, "server.bind", "1 to 65535"},
		{"read header timeout zero", map[string]string{"server.read_header_timeout": "0s"}, nil, "server.read_header_timeout", "positive"},
		{"idle timeout negative", map[string]string{"server.idle_timeout": "-1s"}, nil, "server.idle_timeout", "positive"},
		{"shutdown timeout zero", map[string]string{"server.shutdown_timeout": "0s"}, nil, "server.shutdown_timeout", "positive"},
		{"log level unknown", map[string]string{"log.level": "verbose"}, nil, "log.level", "debug, info, warn, or error"},
		{"log file too small", map[string]string{"log.file_max_size": "1KiB"}, nil, "log.file_max_size", "at least 1MiB"},
		{"log files zero", map[string]string{"log.file_max_files": "0"}, nil, "log.file_max_files", "at least 1"},
		{"max file size zero", map[string]string{"uploads.max_file_size": "0"}, nil, "uploads.max_file_size", "positive"},
		{"max chunk size zero", map[string]string{"uploads.max_chunk_size": "0"}, nil, "uploads.max_chunk_size", "positive"},
		{"chunk larger than file", map[string]string{"uploads.max_file_size": "1MiB", "uploads.max_chunk_size": "2MiB"}, nil, "uploads.max_chunk_size", "must not exceed"},
		{"upload expiry too short", map[string]string{"uploads.expiry": "30s"}, nil, "uploads.expiry", "at least 1m"},
		{"copy items zero", map[string]string{"copy.sync_max_items": "0"}, nil, "copy.sync_max_items", "at least 1"},
		{"copy bytes zero", map[string]string{"copy.sync_max_bytes": "0"}, nil, "copy.sync_max_bytes", "positive"},
		{"bad env value names the variable", nil, map[string]string{"LOCALAINAS_LOG_FILE_MAX_FILES": "x"}, "log.file_max_files", "env LOCALAINAS_LOG_FILE_MAX_FILES"},
		{"bad flag value names the flag", map[string]string{"storage.free_space_reserve": "-5"}, nil, "storage.free_space_reserve", "flag --storage-free-space-reserve"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := map[string]string{"storage.root": root}
			if tt.name == "root missing" {
				delete(flags, "storage.root")
			}
			for k, v := range tt.flags {
				flags[k] = v
			}
			_, err := Load(Sources{Flags: flags, LookupEnv: env(tt.env)})
			if err == nil {
				t.Fatal("Load succeeded, want an error")
			}
			var ce *Error
			if !errors.As(err, &ce) || ce.Key != tt.wantKey {
				t.Errorf("error %q: want a config.Error for key %q", err, tt.wantKey)
			}
			if !strings.Contains(err.Error(), tt.wantKey) || !strings.Contains(err.Error(), tt.wantMsg) {
				t.Errorf("error %q, want it to name %q and contain %q", err, tt.wantKey, tt.wantMsg)
			}
		})
	}
}

func TestErrorMessage(t *testing.T) {
	root := absRoot(t)
	_, err := Load(Sources{Flags: map[string]string{"storage.root": root, "log.level": "loud"}})
	want := `config: log.level (from flag --log-level): must be debug, info, warn, or error, got "loud"`
	if err == nil || err.Error() != want {
		t.Errorf("error = %v\nwant    %s", err, want)
	}

	_, err = Load(Sources{})
	want = "config: storage.root: required: set it in the config file, with $LOCALAINAS_STORAGE_ROOT, or with --storage-root"
	if err == nil || err.Error() != want {
		t.Errorf("error = %v\nwant    %s", err, want)
	}

	_, err = Load(Sources{Flags: map[string]string{"storage.root": root, "log.file_max_files": "many"}})
	var ce *Error
	if !errors.As(err, &ce) || errors.Unwrap(ce) == nil {
		t.Errorf("error %v: want a config.Error that wraps the parse error", err)
	}
}

func TestLoadReportsAllProblems(t *testing.T) {
	_, err := Load(Sources{Flags: map[string]string{"storage.root": "rel", "log.level": "loud"}})
	if err == nil {
		t.Fatal("Load succeeded, want an error")
	}
	for _, key := range []string{"storage.root", "log.level"} {
		if !strings.Contains(err.Error(), key) {
			t.Errorf("error %q does not name %q", err, key)
		}
	}
}

func TestLoadNormalizes(t *testing.T) {
	root := absRoot(t)
	l, err := Load(Sources{Flags: map[string]string{"storage.root": root + string(filepath.Separator), "log.level": "WARN"}})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if l.Storage.Root != root {
		t.Errorf("Storage.Root = %q, want the cleaned %q", l.Storage.Root, root)
	}
	if l.Log.Level != "warn" {
		t.Errorf("Log.Level = %q, want %q", l.Log.Level, "warn")
	}

	db := filepath.Join(root, ".local-ai-nas", "db")
	l, err = Load(Sources{Flags: map[string]string{"storage.root": root, "storage.db_dir": db + string(filepath.Separator)}})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if l.Storage.DBDir != db || l.Storage.LogsDir != "" {
		t.Errorf("DBDir = %q, LogsDir = %q; want the cleaned %q and empty", l.Storage.DBDir, l.Storage.LogsDir, db)
	}
}

func TestNames(t *testing.T) {
	if got := EnvName("storage.free_space_reserve"); got != "LOCALAINAS_STORAGE_FREE_SPACE_RESERVE" {
		t.Errorf("EnvName = %q", got)
	}
	if got := FlagName("storage.free_space_reserve"); got != "storage-free-space-reserve" {
		t.Errorf("FlagName = %q", got)
	}
	seen := map[string]bool{}
	for _, s := range settings {
		if seen[s.key] {
			t.Errorf("duplicate setting %q", s.key)
		}
		seen[s.key] = true
	}
}

func TestFlags(t *testing.T) {
	root := absRoot(t)
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	RegisterFlags(fs)
	file := writeConfig(t, "[log]\nlevel = \"debug\"\n")
	if err := fs.Parse([]string{"--config", file, "--storage-root", root, "--log-file-max-files", "7"}); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	src := SourcesFromFlags(fs)
	src.LookupEnv = env(nil) // keep the test independent of the real environment
	want := map[string]string{"storage.root": root, "log.file_max_files": "7"}
	if diff := cmp.Diff(want, src.Flags); diff != "" {
		t.Errorf("flags mismatch (-want +got):\n%s", diff)
	}
	l, err := Load(src)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if l.File != file || l.Log.Level != "debug" || l.Log.FileMaxFiles != 7 || l.Storage.Root != root {
		t.Errorf("got file=%q level=%q files=%d root=%q", l.File, l.Log.Level, l.Log.FileMaxFiles, l.Storage.Root)
	}
}

func TestDefaultConfigFile(t *testing.T) {
	p := DefaultConfigFile()
	if !filepath.IsAbs(p) || filepath.Base(p) != "config.toml" {
		t.Errorf("DefaultConfigFile() = %q, want an absolute path to config.toml", p)
	}
}
