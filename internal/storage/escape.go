package storage

import "strings"

// escapeMessage is the text of the error os.Root returns for a path that
// would leave the root, for example through a symbolic link that points
// outside. The os package does not export that error; TestIsEscape
// notices if its text changes.
const escapeMessage = "path escapes from parent"

// IsEscape reports whether err is os.Root refusing a path that leaves the
// root. The refusal is correct; this lets callers report it as the
// client's error (outside_root) instead of an internal one.
func IsEscape(err error) bool {
	return err != nil && strings.Contains(err.Error(), escapeMessage)
}
