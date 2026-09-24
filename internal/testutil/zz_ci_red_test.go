package testutil

import "testing"

// TestDeliberateFailure verifies once that CI turns red on a failing test
// (S01.1-T05 acceptance). It is reverted in the next commit.
func TestDeliberateFailure(t *testing.T) {
	t.Fatal("deliberate failure to verify that CI turns red (S01.1-T05)")
}
