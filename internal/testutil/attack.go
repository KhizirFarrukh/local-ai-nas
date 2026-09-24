package testutil

// AttackPaths holds API paths that must be refused on every OS (the
// S01.6-T01 attack corpus). The resolver tests, FuzzResolve's seeds, and
// the API attack suite (S01.7-T02) use it.
var AttackPaths = []string{
	// Not an API path at all.
	"", "docs", "docs/../x", `\..\x`,
	// Parent segments.
	"/..", "/../", "/../x", "/../../etc/passwd", "/../../../../../../windows/win.ini",
	"/a/../../x", "/a/b/../../../x", "/a/..", "/./../x", "/a/./../../x", "/a/b/c/../../../../x",
	// Other namespaces, the photos area, internal data.
	"/files/../photos/u0001/a.jpg", "/../u0002/secret.txt", "/../../photos/u0001",
	"/../../.local-ai-nas/db/nas.db",
	// Dot runs and dots with spaces (Windows may read them as "..").
	"/...", "/..../x", "/....", "/.../.../x", "/.. ", "/. .", "/ ..", "/a/.. /b", "/a/... /b", "/a/ ../b",
	// Backslashes.
	`/..\x`, `/a\..\..\x`, `/a\b`, `\\server\share\x`,
	// UNC-like and device paths.
	"//server/share/x", "///x", "//./C:/x", "//?/C:/x",
	// Drive letters.
	"/C:", "/C:/Windows/System32", "/c:/x", "/a/D:/x", "/Z:",
	// Percent-encoded separators and NUL (the path was encoded twice).
	"/..%2Fx", "/..%2fx", "/%2F..%2Fx", "/a%5C..%5Cx", "/a%5c..%5cx", "/a%00b", "/x%00.txt",
	"/%2e%2e%2f", "/a/%2F/b", "/C:%5Cx",
	// NUL bytes.
	"/a\x00b", "/\x00", "/..\x00/x", "/docs/\x00../x",
	// Unicode look-alikes of dot segments (fullwidth full stop, one and
	// two dot leaders, ellipsis), which folding tools read as dots.
	"/\uFF0E\uFF0E/x", "/a/\uFF0E\uFF0E/\uFF0E\uFF0E/x", "/\u2025/x", "/\u2024\u2024/x", "/\uFF0E/\uFF0E\uFF0E", "/\u2026/x",
}

// WindowsAttackPaths holds paths that only Windows refuses at the resolver
// (S01.6-T02 refuses them on every OS with its own error codes, including
// the forms with an extension such as "aux.txt", which Go's IsLocal accepts
// since current Windows versions allow them).
var WindowsAttackPaths = []string{"/NUL", "/con", "/COM1", "/a/LPT1", "/ /x", "/docs./x", "/docs /x", "/a.", "/a "}
