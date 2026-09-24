// Package config loads and validates the server configuration: a TOML file,
// LOCALAINAS_* environment variables, and command-line flags, with the
// precedence defaults < file < environment < flags (S01.1-T07).
//
// Every setting has one key, such as "storage.root". The same key names the
// setting in the file ([storage] root = ...), in the environment
// (LOCALAINAS_STORAGE_ROOT), and on the command line (--storage-root).
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
)

// Config is the complete server configuration.
type Config struct {
	Storage Storage
	Server  Server
	Log     Log
	Uploads Uploads
	Copy    Copy
}

// Storage configures the storage root (ADR-0003).
type Storage struct {
	// Root holds files/, photos/, and the internal data .local-ai-nas/.
	Root string
	// FreeSpaceReserve is the free space that writes never use (S01.2-T04).
	FreeSpaceReserve ByteSize
	// DBDir and LogsDir optionally move the database and the logs out of
	// <root>/.local-ai-nas, for example to a faster disk. Empty keeps the
	// default. The storage package checks that they stay out of the areas.
	DBDir   string
	LogsDir string
}

// Server configures the HTTP server.
type Server struct {
	// Bind is the listen address, host:port. S01 allows loopback only,
	// which S01.6-T06 checks.
	Bind              string
	ReadHeaderTimeout Duration
	IdleTimeout       Duration
	ShutdownTimeout   Duration
}

// Log configures logging (S01.1-T08).
type Log struct {
	// Level is debug, info, warn, or error.
	Level string
	// FileMaxSize is the size at which the log file is rotated.
	FileMaxSize ByteSize
	// FileMaxFiles is the number of rotated log files kept.
	FileMaxFiles int
}

// Uploads configures upload limits (S01.3-T05, S01.4).
type Uploads struct {
	// MaxFileSize is the largest file that can be uploaded, by the simple
	// upload and by tus.
	MaxFileSize ByteSize
	// MaxChunkSize is the largest body of one tus upload request.
	MaxChunkSize ByteSize
}

// Copy configures the copy operation (S01.3-T08).
type Copy struct {
	// SyncMaxItems and SyncMaxBytes limit a copy that runs within one
	// request: files and folders, and bytes. A larger copy is refused
	// (too_large_for_sync) until background jobs exist (S04.3).
	SyncMaxItems int
	SyncMaxBytes ByteSize
}

// EnvPrefix starts every environment variable name.
const EnvPrefix = "LOCALAINAS_"

// ConfigEnv names the environment variable with the config file path.
const ConfigEnv = EnvPrefix + "CONFIG"

// Default returns the configuration used when nothing overrides it. The
// storage root has no default and must be configured.
func Default() Config {
	return Config{
		Storage: Storage{FreeSpaceReserve: 1 << 30},
		Server: Server{
			Bind:              "127.0.0.1:8080",
			ReadHeaderTimeout: Duration{10 * time.Second},
			IdleTimeout:       Duration{2 * time.Minute},
			ShutdownTimeout:   Duration{30 * time.Second},
		},
		Log: Log{Level: "info", FileMaxSize: 10 << 20, FileMaxFiles: 5},
		Uploads: Uploads{
			MaxFileSize:  100 << 30,
			MaxChunkSize: 64 << 20,
		},
		Copy: Copy{SyncMaxItems: 1000, SyncMaxBytes: 1 << 30},
	}
}

// kind is the type of a setting's value.
type kind int

const (
	kindString   kind = iota // a TOML string
	kindSize                 // a TOML integer (bytes) or a string with a unit
	kindDuration             // a TOML string such as "30s"
	kindInt                  // a TOML integer
)

// setting describes one configuration key.
type setting struct {
	key  string
	kind kind
	// set parses a value in text form into c.
	set func(c *Config, v string) error
}

func stringSetting(key string, field func(*Config) *string) setting {
	return setting{key, kindString, func(c *Config, v string) error {
		*field(c) = v
		return nil
	}}
}

func sizeSetting(key string, field func(*Config) *ByteSize) setting {
	return setting{key, kindSize, func(c *Config, v string) error {
		return field(c).UnmarshalText([]byte(v))
	}}
}

func durationSetting(key string, field func(*Config) *Duration) setting {
	return setting{key, kindDuration, func(c *Config, v string) error {
		return field(c).UnmarshalText([]byte(v))
	}}
}

func intSetting(key string, field func(*Config) *int) setting {
	return setting{key, kindInt, func(c *Config, v string) error {
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return fmt.Errorf("invalid integer %q", v)
		}
		*field(c) = n
		return nil
	}}
}

// settings lists every key. A new setting needs a field in Config, a row
// here, a validation rule if it has constraints, and a line in the example
// config file (S01.1-T06).
var settings = []setting{
	stringSetting("storage.root", func(c *Config) *string { return &c.Storage.Root }),
	sizeSetting("storage.free_space_reserve", func(c *Config) *ByteSize { return &c.Storage.FreeSpaceReserve }),
	stringSetting("storage.db_dir", func(c *Config) *string { return &c.Storage.DBDir }),
	stringSetting("storage.logs_dir", func(c *Config) *string { return &c.Storage.LogsDir }),
	stringSetting("server.bind", func(c *Config) *string { return &c.Server.Bind }),
	durationSetting("server.read_header_timeout", func(c *Config) *Duration { return &c.Server.ReadHeaderTimeout }),
	durationSetting("server.idle_timeout", func(c *Config) *Duration { return &c.Server.IdleTimeout }),
	durationSetting("server.shutdown_timeout", func(c *Config) *Duration { return &c.Server.ShutdownTimeout }),
	stringSetting("log.level", func(c *Config) *string { return &c.Log.Level }),
	sizeSetting("log.file_max_size", func(c *Config) *ByteSize { return &c.Log.FileMaxSize }),
	intSetting("log.file_max_files", func(c *Config) *int { return &c.Log.FileMaxFiles }),
	sizeSetting("uploads.max_file_size", func(c *Config) *ByteSize { return &c.Uploads.MaxFileSize }),
	sizeSetting("uploads.max_chunk_size", func(c *Config) *ByteSize { return &c.Uploads.MaxChunkSize }),
	intSetting("copy.sync_max_items", func(c *Config) *int { return &c.Copy.SyncMaxItems }),
	sizeSetting("copy.sync_max_bytes", func(c *Config) *ByteSize { return &c.Copy.SyncMaxBytes }),
}

// EnvName returns the environment variable for a key:
// "storage.root" -> "LOCALAINAS_STORAGE_ROOT".
func EnvName(key string) string {
	return EnvPrefix + strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
}

// FlagName returns the command-line flag for a key:
// "storage.root" -> "storage-root".
func FlagName(key string) string {
	return strings.NewReplacer(".", "-", "_", "-").Replace(key)
}

// Sources are the inputs of Load besides the defaults.
type Sources struct {
	// ConfigFile is the --config flag value, or "" when it was not given.
	ConfigFile string
	// LookupEnv reads an environment variable (os.LookupEnv in production).
	// A nil LookupEnv reads nothing.
	LookupEnv func(string) (string, bool)
	// Flags holds the settings given on the command line, by key.
	Flags map[string]string
	// DefaultConfigFile is used when neither --config nor LOCALAINAS_CONFIG
	// is set. A missing default file is not an error.
	DefaultConfigFile string
}

// DefaultConfigFile returns the OS default location of the config file:
// %ProgramData%\local-ai-nas\config.toml on Windows and
// /etc/local-ai-nas/config.toml elsewhere. The file lives outside the
// storage root, because it is what tells the server where the root is.
func DefaultConfigFile() string {
	if runtime.GOOS == "windows" {
		base := os.Getenv("ProgramData")
		if base == "" {
			base = `C:\ProgramData`
		}
		return filepath.Join(base, "local-ai-nas", "config.toml")
	}
	return "/etc/local-ai-nas/config.toml"
}

// Loaded is a validated configuration together with where it came from.
type Loaded struct {
	Config
	// File is the config file that was read, or "" if none was.
	File string
	// Origins maps each key that was not left at its default to its
	// source: "file", "env LOCALAINAS_…", or "flag --…".
	Origins map[string]string
}

// Load builds the configuration from the defaults, the config file, the
// environment, and the flags, in that order, and validates it. The error
// lists every problem, each naming its key and source.
func Load(src Sources) (*Loaded, error) {
	lookup := src.LookupEnv
	if lookup == nil {
		lookup = func(string) (string, bool) { return "", false }
	}
	l := &Loaded{Config: Default(), Origins: map[string]string{}}

	path, explicit := src.ConfigFile, true
	if path == "" {
		path, _ = lookup(ConfigEnv)
	}
	if path == "" {
		path, explicit = src.DefaultConfigFile, false
	}
	if path != "" {
		if err := l.readFile(path, explicit); err != nil {
			return nil, err
		}
	}

	var errs []error
	for _, s := range settings {
		if v, ok := lookup(EnvName(s.key)); ok {
			errs = append(errs, l.apply(s, v, "env "+EnvName(s.key)))
		}
	}
	for _, s := range settings {
		if v, ok := src.Flags[s.key]; ok {
			errs = append(errs, l.apply(s, v, "flag --"+FlagName(s.key)))
		}
	}
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	if err := l.validate(); err != nil {
		return nil, err
	}
	return l, nil
}

func (l *Loaded) apply(s setting, v, origin string) error {
	if err := s.set(&l.Config, v); err != nil {
		return &Error{Key: s.key, Origin: origin, Err: err}
	}
	l.Origins[s.key] = origin
	return nil
}

// readFile applies the settings in the TOML file at path. The file is
// decoded into a generic map and every value goes through the same parsers
// as environment and flag values, so each error names its key.
func (l *Loaded) readFile(path string, explicit bool) error {
	data, err := os.ReadFile(path) // #nosec G304 -- the operator chooses the config file.
	if errors.Is(err, fs.ErrNotExist) && !explicit {
		return nil
	}
	if err != nil {
		return fmt.Errorf("config: read %s: %w", path, err)
	}
	var doc map[string]any
	if err := toml.Unmarshal(data, &doc); err != nil {
		var de *toml.DecodeError
		if errors.As(err, &de) {
			row, col := de.Position()
			return fmt.Errorf("config: %s:%d:%d: %s", path, row, col, de.Error())
		}
		return fmt.Errorf("config: %s: %w", path, err)
	}
	l.File = path
	origin := "file " + path

	byKey := make(map[string]setting, len(settings))
	for _, s := range settings {
		byKey[s.key] = s
	}
	var errs []error
	for _, section := range sortedKeys(doc) {
		table, ok := doc[section].(map[string]any)
		if !ok {
			errs = append(errs, fmt.Errorf("config: %s: unknown setting %q (settings live in sections such as [storage])", path, section))
			continue
		}
		for _, name := range sortedKeys(table) {
			key := section + "." + name
			s, ok := byKey[key]
			if !ok {
				errs = append(errs, fmt.Errorf("config: %s: unknown setting %q", path, key))
				continue
			}
			v, err := s.fromTOML(table[name])
			if err != nil {
				errs = append(errs, &Error{Key: key, Origin: origin, Err: err})
				continue
			}
			if err := l.apply(s, v, origin); err != nil {
				errs = append(errs, err)
				continue
			}
			l.Origins[key] = "file"
		}
	}
	return errors.Join(errs...)
}

// fromTOML checks that a value from the file has a TOML type that fits the
// setting and returns it in the text form that set parses.
func (s setting) fromTOML(v any) (string, error) {
	switch x := v.(type) {
	case string:
		if s.kind == kindInt {
			return "", fmt.Errorf("must be an integer, got the string %q", x)
		}
		return x, nil
	case int64:
		switch s.kind {
		case kindInt, kindSize:
			return strconv.FormatInt(x, 10), nil
		case kindDuration:
			return "", fmt.Errorf("must be a string such as \"30s\", got the integer %d", x)
		}
		return "", fmt.Errorf("must be a string, got the integer %d", x)
	}
	return "", fmt.Errorf("has the wrong type (%s)", tomlType(v))
}

func tomlType(v any) string {
	switch v.(type) {
	case bool:
		return "boolean"
	case float64:
		return "float"
	case []any:
		return "array"
	case map[string]any:
		return "table"
	}
	return "date or time"
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// Error is a problem with one setting.
type Error struct {
	Key    string // the setting, such as "storage.root"
	Origin string // where the value came from, such as "env LOCALAINAS_STORAGE_ROOT"
	Err    error
}

func (e *Error) Error() string {
	if e.Origin == "" {
		return fmt.Sprintf("config: %s: %v", e.Key, e.Err)
	}
	return fmt.Sprintf("config: %s (from %s): %v", e.Key, e.Origin, e.Err)
}

func (e *Error) Unwrap() error { return e.Err }

// validate checks every rule, normalizes values, and returns all problems.
func (l *Loaded) validate() error {
	c := &l.Config
	var errs []error
	fail := func(key, format string, args ...any) {
		errs = append(errs, &Error{Key: key, Origin: l.Origins[key], Err: fmt.Errorf(format, args...)})
	}

	switch {
	case c.Storage.Root == "":
		fail("storage.root", "required: set it in the config file, with $%s, or with --%s",
			EnvName("storage.root"), FlagName("storage.root"))
	case !filepath.IsAbs(c.Storage.Root):
		fail("storage.root", "must be an absolute path, got %q", c.Storage.Root)
	default:
		c.Storage.Root = filepath.Clean(c.Storage.Root)
	}
	for _, d := range []struct {
		key string
		dir *string
	}{
		{"storage.db_dir", &c.Storage.DBDir},
		{"storage.logs_dir", &c.Storage.LogsDir},
	} {
		switch {
		case *d.dir == "":
		case !filepath.IsAbs(*d.dir):
			fail(d.key, "must be an absolute path or empty, got %q", *d.dir)
		default:
			*d.dir = filepath.Clean(*d.dir)
		}
	}

	if host, port, err := net.SplitHostPort(c.Server.Bind); err != nil || host == "" {
		fail("server.bind", "must be host:port, such as 127.0.0.1:8080, got %q", c.Server.Bind)
	} else if n, err := strconv.Atoi(port); err != nil || n < 1 || n > 65535 {
		fail("server.bind", "port must be 1 to 65535, got %q", port)
	}
	for _, d := range []struct {
		key string
		val Duration
	}{
		{"server.read_header_timeout", c.Server.ReadHeaderTimeout},
		{"server.idle_timeout", c.Server.IdleTimeout},
		{"server.shutdown_timeout", c.Server.ShutdownTimeout},
	} {
		if d.val.Duration <= 0 {
			fail(d.key, "must be positive, got %s", d.val)
		}
	}

	switch strings.ToLower(c.Log.Level) {
	case "debug", "info", "warn", "error":
		c.Log.Level = strings.ToLower(c.Log.Level)
	default:
		fail("log.level", "must be debug, info, warn, or error, got %q", c.Log.Level)
	}
	if c.Log.FileMaxSize < 1<<20 {
		fail("log.file_max_size", "must be at least 1MiB, got %s", c.Log.FileMaxSize)
	}
	if c.Log.FileMaxFiles < 1 {
		fail("log.file_max_files", "must be at least 1, got %d", c.Log.FileMaxFiles)
	}

	if c.Uploads.MaxFileSize <= 0 {
		fail("uploads.max_file_size", "must be positive")
	}
	switch {
	case c.Uploads.MaxChunkSize <= 0:
		fail("uploads.max_chunk_size", "must be positive")
	case c.Uploads.MaxFileSize > 0 && c.Uploads.MaxChunkSize > c.Uploads.MaxFileSize:
		fail("uploads.max_chunk_size", "must not exceed uploads.max_file_size (%s), got %s",
			c.Uploads.MaxFileSize, c.Uploads.MaxChunkSize)
	}

	if c.Copy.SyncMaxItems < 1 {
		fail("copy.sync_max_items", "must be at least 1, got %d", c.Copy.SyncMaxItems)
	}
	if c.Copy.SyncMaxBytes <= 0 {
		fail("copy.sync_max_bytes", "must be positive")
	}

	return errors.Join(errs...)
}
