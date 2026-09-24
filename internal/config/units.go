package config

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// ByteSize is a size in bytes. In the config file it is either a TOML
// integer (bytes) or a string with a unit, such as "512MiB" or "10GB".
// Binary units (KiB, MiB, GiB, TiB) are powers of 1024; decimal units
// (KB, MB, GB, TB) are powers of 1000.
type ByteSize int64

var byteUnits = []struct {
	suffix string
	factor int64
}{
	// Longer suffixes first, so "MiB" is not read as "B".
	{"KiB", 1 << 10}, {"MiB", 1 << 20}, {"GiB", 1 << 30}, {"TiB", 1 << 40},
	{"KB", 1e3}, {"MB", 1e6}, {"GB", 1e9}, {"TB", 1e12},
	{"B", 1},
}

// ParseByteSize parses a size such as "1048576", "512MiB", or "10 GB".
func ParseByteSize(s string) (ByteSize, error) {
	t := strings.TrimSpace(s)
	factor := int64(1)
	for _, u := range byteUnits {
		if strings.HasSuffix(t, u.suffix) {
			factor = u.factor
			t = strings.TrimSpace(strings.TrimSuffix(t, u.suffix))
			break
		}
	}
	n, err := strconv.ParseInt(t, 10, 64)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("invalid size %q: use a whole number of bytes or a unit such as 512MiB or 10GB", s)
	}
	if n > math.MaxInt64/factor {
		return 0, fmt.Errorf("invalid size %q: too large", s)
	}
	return ByteSize(n * factor), nil
}

// UnmarshalText implements encoding.TextUnmarshaler for TOML strings.
func (b *ByteSize) UnmarshalText(text []byte) error {
	v, err := ParseByteSize(string(text))
	if err != nil {
		return err
	}
	*b = v
	return nil
}

// String formats the size with the largest binary unit that divides it.
func (b ByteSize) String() string {
	for _, u := range []struct {
		suffix string
		factor int64
	}{{"TiB", 1 << 40}, {"GiB", 1 << 30}, {"MiB", 1 << 20}, {"KiB", 1 << 10}} {
		if b != 0 && int64(b)%u.factor == 0 {
			return strconv.FormatInt(int64(b)/u.factor, 10) + u.suffix
		}
	}
	return strconv.FormatInt(int64(b), 10) + "B"
}

// Duration is a time span written as a string such as "30s" or "2m", in
// the time.ParseDuration format. It is a struct so that a bare TOML
// integer is rejected instead of being read as nanoseconds.
type Duration struct {
	time.Duration
}

// UnmarshalText implements encoding.TextUnmarshaler for TOML strings.
func (d *Duration) UnmarshalText(text []byte) error {
	v, err := time.ParseDuration(strings.TrimSpace(string(text)))
	if err != nil {
		return fmt.Errorf("invalid duration %q: use a value such as 30s, 2m, or 1h", text)
	}
	d.Duration = v
	return nil
}

// MarshalText implements encoding.TextMarshaler.
func (d Duration) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}
