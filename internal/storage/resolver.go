package storage

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
)

// Resolver maps API paths to paths inside a namespace directory and opens
// namespace directories as os.Root (ADR-0003, S01.2-T03). It is the only
// way from a request to the disk.
//
// Two layers keep every operation inside its namespace: Resolve refuses any
// path that could leave it, and the os.Root from OpenRoot refuses to leave
// the directory even for a path Resolve did not catch. S01.6-T01 extends
// the path rules with a full attack corpus, and S01.6-T02 adds the file
// name rules.
type Resolver struct {
	layout Layout
}

// NewResolver returns the resolver for a layout.
func NewResolver(l Layout) *Resolver {
	return &Resolver{layout: l}
}

// namespacePattern matches namespace IDs such as u0001 (ADR-0003).
var namespacePattern = regexp.MustCompile(`^u[0-9]{4}$`)

// Resolve maps userPath, a slash-separated API path that starts with "/"
// (the namespace root), to the clean path relative to the namespace
// directory of the area: "/docs/a.txt" gives "docs/a.txt", and "/" gives
// ".". The result is slash-separated.
func (r *Resolver) Resolve(area, namespace, userPath string) (string, error) {
	if err := checkNamespace(area, namespace); err != nil {
		return "", err
	}
	switch {
	case !strings.HasPrefix(userPath, "/"):
		return "", apperr.Newf(apperr.InvalidRequest, "a path must start with /, got %q", userPath)
	case strings.ContainsRune(userPath, 0):
		return "", apperr.New(apperr.InvalidName, "a path must not contain a NUL character")
	case strings.Contains(userPath, `\`):
		return "", apperr.New(apperr.InvalidName, `a path must use / as the separator, not \`)
	}
	for _, seg := range strings.Split(userPath[1:], "/") {
		if seg == ".." {
			return "", apperr.New(apperr.OutsideRoot, "a path must not contain \"..\"")
		}
	}
	rel := strings.TrimPrefix(path.Clean(userPath), "/")
	if rel == "" {
		return ".", nil
	}
	// filepath.IsLocal also refuses names that are not plain local names
	// on this OS, such as NUL or COM1 on Windows.
	if !filepath.IsLocal(filepath.FromSlash(rel)) {
		return "", apperr.Newf(apperr.InvalidName, "%q is not a valid path on this server", userPath)
	}
	return rel, nil
}

// OpenRoot opens the namespace directory of an area as an os.Root. Every
// file operation for the namespace must go through it.
func (r *Resolver) OpenRoot(area, namespace string) (*os.Root, error) {
	if err := checkNamespace(area, namespace); err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(r.layout.Area(area, namespace))
	if err != nil {
		return nil, apperr.Wrap(apperr.Internal, "the storage area is not available", err)
	}
	return root, nil
}

// checkNamespace accepts the files area and a well-formed namespace. The
// photos area exists on disk but has no API before S04 (S01.2-T06).
func checkNamespace(area, namespace string) error {
	switch area {
	case FilesArea:
	case PhotosArea:
		return apperr.New(apperr.NotAvailable, "the photos area is not available yet")
	default:
		return apperr.Newf(apperr.InvalidRequest, "unknown storage area %q", area)
	}
	if !namespacePattern.MatchString(namespace) {
		return apperr.Newf(apperr.InvalidRequest, "invalid namespace %q", namespace)
	}
	return nil
}
