package api

import (
	"slices"
	"strings"
	"testing"
)

// TestSpecOperationsAreRouted keeps the spec and the route table in step:
// every operation in api/openapi.yaml has a route with the same method and
// path, and every method-specific route is an operation in the spec. The
// photos and uploads tags are the exceptions: a hand-written catch-all
// serves each (uploads is the tus protocol, served by tusd).
func TestSpecOperationsAreRouted(t *testing.T) {
	doc := loadSpec(t)
	routes := map[string]bool{}
	for _, r := range Routes(Options{}) {
		routes[r.Pattern] = true
	}
	inSpec := map[string]bool{}
	for path, item := range doc.Paths.Map() {
		for method, op := range item.Operations() {
			if slices.Contains(op.Tags, "photos") || slices.Contains(op.Tags, "uploads") {
				continue
			}
			pattern := method + " /api/v1" + path
			inSpec[pattern] = true
			if !routes[pattern] {
				t.Errorf("the spec operation %s (%s) has no route", pattern, op.OperationID)
			}
		}
	}
	if len(inSpec) == 0 {
		t.Fatal("the spec has no operations to route")
	}
	for pattern := range routes {
		if strings.Contains(pattern, " ") && !inSpec[pattern] {
			t.Errorf("the route %q is not an operation in the spec", pattern)
		}
	}
}
