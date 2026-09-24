// Package apispec embeds the REST API contract, openapi.yaml, so the
// server can serve it with the offline API documentation (S01.5-T06).
package apispec

import _ "embed" // for go:embed

// OpenAPI is the content of openapi.yaml.
//
//go:embed openapi.yaml
var OpenAPI []byte
