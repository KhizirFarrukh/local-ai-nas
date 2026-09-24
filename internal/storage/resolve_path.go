package storage

import (
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/text/unicode/norm"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
)

// cleanUserPath turns an API path into a clean, NFC-normalized path
// relative to the namespace, or refuses it (S01.6-T01). The rules, in
// order:
//
//  1. It starts with exactly one "/": "//server/share" (UNC-like) is
//     refused.
//  2. No NUL byte, no backslash (not a separator here, but one on
//     Windows), and no percent-encoded separator or NUL (%2F, %5C, %00),
//     which would mean the client encoded the path twice.
//  3. No segment is "..", a run of three or more dots, or dots mixed with
//     spaces (".. ", ". ."): Windows may treat those as "..".
//  4. No segment is a drive letter ("C:").
//  5. The result is NFC-normalized, so the same visible name is always
//     the same file (macOS clients often send decomposed names).
//  6. filepath.IsLocal must accept it on this OS (it also refuses names
//     such as NUL or COM1 on Windows).
//
// The os.Root from OpenRoot is the second layer: it refuses to leave the
// namespace even if these rules ever miss something.
func cleanUserPath(userPath string) (string, error) {
	switch {
	case !strings.HasPrefix(userPath, "/"):
		return "", apperr.Newf(apperr.InvalidRequest, "a path must start with /, got %q", userPath)
	case strings.HasPrefix(userPath, "//"):
		return "", apperr.New(apperr.InvalidName, "a path must not start with // (network paths are not allowed)")
	case strings.ContainsRune(userPath, 0):
		return "", apperr.New(apperr.InvalidName, "a path must not contain a NUL character")
	case strings.Contains(userPath, `\`):
		return "", apperr.New(apperr.InvalidName, `a path must use / as the separator, not \`)
	}
	lower := strings.ToLower(userPath)
	for _, enc := range []string{"%2f", "%5c", "%00"} {
		if strings.Contains(lower, enc) {
			return "", apperr.Newf(apperr.InvalidName, "a path must not contain the encoded character %s; send the path decoded once", strings.ToUpper(enc))
		}
	}
	for _, seg := range strings.Split(userPath[1:], "/") {
		switch {
		case seg == "..":
			return "", apperr.New(apperr.OutsideRoot, `a path must not contain ".."`)
		case isDotsAndSpaces(seg):
			return "", apperr.Newf(apperr.OutsideRoot, "a path must not contain the segment %q", seg)
		case isDriveLetter(seg):
			return "", apperr.Newf(apperr.InvalidName, "a path must not contain a drive letter (%q)", seg)
		}
	}
	rel := strings.TrimPrefix(path.Clean(norm.NFC.String(userPath)), "/")
	if rel == "" {
		return ".", nil
	}
	if !filepath.IsLocal(filepath.FromSlash(rel)) {
		return "", apperr.Newf(apperr.InvalidName, "%q is not a valid path on this server", userPath)
	}
	return rel, nil
}

// isDotsAndSpaces reports segments made only of dots and spaces that are
// not a plain "." or "": "...", ".. ", ". .", " .". Windows removes
// trailing dots and spaces from names, so such segments are ambiguous.
func isDotsAndSpaces(seg string) bool {
	if seg == "" || seg == "." {
		return false
	}
	return strings.Trim(seg, ". ") == "" && strings.Contains(seg, ".")
}

// isDriveLetter reports segments such as "C:" or "c:".
func isDriveLetter(seg string) bool {
	return len(seg) == 2 && seg[1] == ':' &&
		(seg[0] >= 'a' && seg[0] <= 'z' || seg[0] >= 'A' && seg[0] <= 'Z')
}
