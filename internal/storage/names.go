package storage

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
)

// Limits for names the API creates (S01.6-T02).
const (
	// MaxNameBytes is the longest file or folder name, in UTF-8 bytes; the
	// limit of the common file systems (ext4, NTFS in UTF-16 units, APFS).
	MaxNameBytes = 255
	// MaxPathBytes is the longest path inside a namespace, in UTF-8 bytes.
	MaxPathBytes = 4096
)

// Rules of ValidateName and ValidateNewPath. A refused name is reported
// with code invalid_name and one of these in the problem's "rule" field.
const (
	RuleEmptyName     = "empty_name"
	RuleDotName       = "dot_name"
	RuleReservedName  = "reserved_name"
	RuleForbiddenChar = "forbidden_character"
	RuleControlChar   = "control_character"
	RuleTrailingChar  = "trailing_dot_or_space"
	RuleNameTooLong   = "name_too_long"
	RulePathTooLong   = "path_too_long"
	RuleInvalidUTF8   = "invalid_utf8"
)

// reservedNames are the device names Windows reserves, with or without an
// extension, in any case. They are refused on every OS, so the storage
// stays portable (a disk moved to Windows, or a Windows client over SMB).
var reservedNames = func() map[string]bool {
	m := map[string]bool{"CON": true, "PRN": true, "AUX": true, "NUL": true, "CONIN$": true, "CONOUT$": true}
	for _, d := range []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "¹", "²", "³"} {
		m["COM"+d], m["LPT"+d] = true, true
	}
	return m
}()

// ValidateName checks one file or folder name that the API is about to
// create (upload, new folder, rename, move or copy target). It never
// rewrites a name: an invalid one is refused with the rule it breaks.
// Existing names on disk are not checked, so files put there by other
// means stay readable.
func ValidateName(name string) error {
	switch {
	case name == "":
		return nameErr(RuleEmptyName, "a name must not be empty")
	case name == "." || name == "..":
		return nameErr(RuleDotName, fmt.Sprintf("%q is not a valid name", name))
	case !utf8.ValidString(name):
		return nameErr(RuleInvalidUTF8, "a name must be valid UTF-8")
	case len(name) > MaxNameBytes:
		return nameErr(RuleNameTooLong, fmt.Sprintf("a name must be at most %d bytes, this one has %d", MaxNameBytes, len(name)))
	}
	for _, r := range name {
		switch {
		case r < 0x20 || r == 0x7f:
			return nameErr(RuleControlChar, fmt.Sprintf("a name must not contain control characters (found U+%04X)", r))
		case strings.ContainsRune(`<>:"/\|?*`, r):
			return nameErr(RuleForbiddenChar, fmt.Sprintf(`a name must not contain any of < > : " / \ | ? * (found %q)`, r))
		}
	}
	if strings.HasSuffix(name, ".") || strings.HasSuffix(name, " ") {
		return nameErr(RuleTrailingChar, "a name must not end with a dot or a space")
	}
	if isReservedName(name) {
		return nameErr(RuleReservedName, fmt.Sprintf("%q is a reserved device name on Windows", name))
	}
	return nil
}

// isReservedName reports a Windows device name, with any extension, in
// any case, with spaces before the extension ("nul .txt").
func isReservedName(name string) bool {
	base, _, _ := strings.Cut(name, ".")
	return reservedNames[strings.ToUpper(strings.TrimRight(base, " "))]
}

// ValidateNewPath checks a slash-separated path relative to a namespace
// (as Resolve returns it) that the API is about to create: its length and
// every name in it.
func ValidateNewPath(rel string) error {
	if len(rel) > MaxPathBytes {
		return nameErr(RulePathTooLong, fmt.Sprintf("a path must be at most %d bytes, this one has %d", MaxPathBytes, len(rel)))
	}
	if rel == "." {
		return nameErr(RuleDotName, "the namespace root cannot be created or replaced")
	}
	for _, name := range strings.Split(rel, "/") {
		if err := ValidateName(name); err != nil {
			return err
		}
	}
	return nil
}

func nameErr(rule, detail string) error {
	return apperr.NewRule(apperr.InvalidName, rule, detail)
}
