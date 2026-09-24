package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TestExampleConfig keeps deploy/config.example.toml in step with the code:
// it must parse, set every setting, and show the default values.
func TestExampleConfig(t *testing.T) {
	path := filepath.Join("..", "..", "deploy", "config.example.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "[storage]\n") || !strings.Contains(string(data), "\nroot = ") {
		t.Error("the example does not set storage.root")
	}

	// The example's root is a Linux path; on Windows it is not absolute, so
	// the test overrides it with a flag. Every other value comes from the file.
	root := absRoot(t)
	l, err := Load(Sources{ConfigFile: path, Flags: map[string]string{"storage.root": root}})
	if err != nil {
		t.Fatalf("the example does not load: %v", err)
	}
	for _, s := range settings {
		if s.key == "storage.root" {
			continue
		}
		if l.Origins[s.key] != "file" {
			t.Errorf("the example does not set %s", s.key)
		}
	}
	want := Default()
	want.Storage.Root = root
	if diff := cmp.Diff(want, l.Config); diff != "" {
		t.Errorf("the example's values differ from the defaults (-default +example):\n%s", diff)
	}
}
