package config

import (
	"math"
	"testing"
	"time"
)

func TestParseByteSize(t *testing.T) {
	tests := []struct {
		in      string
		want    ByteSize
		wantErr bool
	}{
		{"0", 0, false},
		{"1048576", 1 << 20, false},
		{"512B", 512, false},
		{"1KiB", 1 << 10, false},
		{"64MiB", 64 << 20, false},
		{"10 GiB", 10 << 30, false},
		{"2TiB", 2 << 40, false},
		{"1KB", 1000, false},
		{"5MB", 5_000_000, false},
		{"3GB", 3_000_000_000, false},
		{"1TB", 1_000_000_000_000, false},
		{" 7MiB ", 7 << 20, false},
		{"", 0, true},
		{"lots", 0, true},
		{"-1", 0, true},
		{"1.5GiB", 0, true},
		{"10XB", 0, true},
		{"MiB", 0, true},
		{"9223372036854775807", math.MaxInt64, false},
		{"8388608TiB", 0, true}, // 2^23 TiB = 2^63 bytes overflows int64
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := ParseByteSize(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseByteSize(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseByteSize(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestByteSizeString(t *testing.T) {
	tests := []struct {
		in   ByteSize
		want string
	}{
		{0, "0B"},
		{512, "512B"},
		{1 << 10, "1KiB"},
		{1536, "1536B"},
		{64 << 20, "64MiB"},
		{100 << 30, "100GiB"},
		{2 << 40, "2TiB"},
	}
	for _, tt := range tests {
		if got := tt.in.String(); got != tt.want {
			t.Errorf("ByteSize(%d).String() = %q, want %q", int64(tt.in), got, tt.want)
		}
		if tt.in != 0 {
			back, err := ParseByteSize(tt.want)
			if err != nil || back != tt.in {
				t.Errorf("ParseByteSize(%q) = %d, %v; want %d", tt.want, back, err, tt.in)
			}
		}
	}
}

func TestDuration(t *testing.T) {
	var d Duration
	if err := d.UnmarshalText([]byte(" 90s ")); err != nil || d.Duration != 90*time.Second {
		t.Errorf("UnmarshalText(90s) = %s, %v", d, err)
	}
	for _, bad := range []string{"", "30", "soon", "1 hour"} {
		if err := d.UnmarshalText([]byte(bad)); err == nil {
			t.Errorf("UnmarshalText(%q) succeeded, want an error", bad)
		}
	}
	text, err := Duration{2 * time.Minute}.MarshalText()
	if err != nil || string(text) != "2m0s" {
		t.Errorf("MarshalText = %q, %v", text, err)
	}
}
