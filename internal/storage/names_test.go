package storage

import (
	"errors"
	"os"
	"strings"
	"testing"

	"golang.org/x/text/unicode/norm"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
)

// ruleOf returns the rule of an invalid_name error, or "" for nil.
func ruleOf(t *testing.T, err error) string {
	t.Helper()
	if err == nil {
		return ""
	}
	var e *apperr.Error
	if !errors.As(err, &e) || e.Kind != apperr.InvalidName {
		t.Fatalf("error %v is not an invalid_name error", err)
	}
	return e.Rule
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name string
		rule string // "" means accepted
	}{
		// Accepted.
		{"report.pdf", ""},
		{"Ω résumé 2026.txt", ""},
		{".hidden", ""},
		{"a..b", ""},
		{" leading space.txt", ""},
		{"CONSOLE.txt", ""}, // only the exact device names are reserved
		{"com10", ""},
		{"lpt", ""},
		{"nul-bytes.txt", ""},
		{strings.Repeat("a", 255), ""},
		{strings.Repeat("é", 127), ""}, // 254 bytes

		// Empty and dot names.
		{"", RuleEmptyName},
		{".", RuleDotName},
		{"..", RuleDotName},

		// Reserved device names, with or without an extension, any case.
		{"CON", RuleReservedName},
		{"con", RuleReservedName},
		{"Prn.txt", RuleReservedName},
		{"aux.tar.gz", RuleReservedName},
		{"NUL", RuleReservedName},
		{"nul .txt", RuleReservedName},
		{"COM1", RuleReservedName},
		{"com9.log", RuleReservedName},
		{"COM0", RuleReservedName},
		{"LPT1", RuleReservedName},
		{"lpt9.doc", RuleReservedName},
		{"COM¹", RuleReservedName},
		{"LPT³.txt", RuleReservedName},
		{"CONIN$", RuleReservedName},
		{"conout$.txt", RuleReservedName},

		// Look-alikes of separators and of dot names; other fullwidth
		// characters are ordinary (common in CJK names).
		{"a\uFF0Fb", RuleLookalikeSep},
		{"a\uFF3Cb", RuleLookalikeSep},
		{"1\u22152", RuleLookalikeSep},
		{"a\u2044b", RuleLookalikeSep},
		{"\uFF0E\uFF0E", RuleDotName},
		{"\uFF0E", RuleDotName},
		{"\u2025", RuleDotName},
		{"\u2026", RuleDotName},
		{"なぜ\uFF1F.txt", ""},
		{"\uFF26\uFF55\uFF4C\uFF4C.txt", ""},
		{"Wait\u2026 more.txt", ""},

		// The prefix of the server's temporary files.
		{".local-ai-nas-tmp-1a2b.part", RuleReservedName},
		{".local-ai-nas-tmp-", RuleReservedName},
		{".local-ai-nas-temp.txt", ""},
		{"x.local-ai-nas-tmp-1", ""},

		// Forbidden characters.
		{"a<b", RuleForbiddenChar},
		{"a>b", RuleForbiddenChar},
		{"a:b", RuleForbiddenChar},
		{"file.txt:stream", RuleForbiddenChar},
		{`a"b`, RuleForbiddenChar},
		{"a/b", RuleForbiddenChar},
		{`a\b`, RuleForbiddenChar},
		{"a|b", RuleForbiddenChar},
		{"a?b", RuleForbiddenChar},
		{"a*b", RuleForbiddenChar},

		// Control characters.
		{"a\x00b", RuleControlChar},
		{"a\tb", RuleControlChar},
		{"line\nbreak", RuleControlChar},
		{"a\x1fb", RuleControlChar},
		{"a\x7fb", RuleControlChar},

		// Trailing dot or space.
		{"report.", RuleTrailingChar},
		{"report ", RuleTrailingChar},
		{"...", RuleTrailingChar},
		{"a. ", RuleTrailingChar},

		// Too long.
		{strings.Repeat("a", 256), RuleNameTooLong},
		{strings.Repeat("é", 128), RuleNameTooLong}, // 256 bytes

		// Not UTF-8.
		{"a\xffb", RuleInvalidUTF8},
	}
	for _, tt := range tests {
		got := ruleOf(t, ValidateName(tt.name))
		if got != tt.rule {
			t.Errorf("ValidateName(%q) rule = %q, want %q", tt.name, got, tt.rule)
		}
	}
}

func TestValidateNewPath(t *testing.T) {
	long := strings.Repeat(strings.Repeat("a", 200)+"/", 21) + "x" // 4222 bytes
	tests := []struct {
		rel  string
		rule string
	}{
		{"docs/2026/report.pdf", ""},
		{"a", ""},
		{".", RuleDotName},
		{"docs/con/x.txt", RuleReservedName},
		{"docs/draft./x", RuleTrailingChar},
		{"docs/a:b", RuleForbiddenChar},
		{long, RulePathTooLong},
	}
	for _, tt := range tests {
		if got := ruleOf(t, ValidateNewPath(tt.rel)); got != tt.rule {
			t.Errorf("ValidateNewPath(%q) rule = %q, want %q", tt.rel, got, tt.rule)
		}
	}
}

// FuzzValidateName checks that every name ValidateName accepts can really
// be created, listed under the same name, and read back on this OS, so the
// rules never let through a name the file system would refuse or change.
func FuzzValidateName(f *testing.F) {
	for _, s := range []string{"report.pdf", "Ω résumé.txt", ".hidden", "aux.txt", "a:b", "x.", "COM¹", "tab\there", " lead"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, name string) {
		if ValidateName(name) != nil {
			return
		}
		dir := t.TempDir()
		root, err := os.OpenRoot(dir)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = root.Close() }()
		if err := root.WriteFile(name, []byte("x"), 0o600); err != nil {
			t.Fatalf("ValidateName accepted %q, but it cannot be created: %v", name, err)
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 1 {
			t.Fatalf("after creating %q the directory holds %d entries", name, len(entries))
		}
		if got := entries[0].Name(); got != name && norm.NFC.String(got) != norm.NFC.String(name) {
			t.Fatalf("created %q but the file system lists %q", name, got)
		}
		if data, err := root.ReadFile(name); err != nil || string(data) != "x" {
			t.Fatalf("created %q but cannot read it back: %v", name, err)
		}
	})
}

// TestNameRuleInProblem checks that a refused name reaches the client with
// its rule.
func TestNameRuleInProblem(t *testing.T) {
	p := apperr.ProblemFor(ValidateName("aux.txt"), "req1")
	if p.Code != "invalid_name" || p.Rule != RuleReservedName || p.Status != 400 {
		t.Errorf("problem = %+v, want invalid_name / reserved_name / 400", p)
	}
}
