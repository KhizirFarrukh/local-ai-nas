package auth

import "testing"

// BenchmarkPasswordHash measures one Argon2id hash with the ADR-0010
// parameters (S03.2-T02). A login costs one of these; the result on each
// machine is recorded in the session log, and the Raspberry Pi profile
// checks the login time (stage document 4.9).
func BenchmarkPasswordHash(b *testing.B) {
	h, err := NewHasher(DefaultParams, 1, 0)
	if err != nil {
		b.Fatal(err)
	}
	ctx := b.Context()
	for b.Loop() {
		if _, err := h.Hash(ctx, "correct horse battery staple"); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPasswordVerify measures a login's password check.
func BenchmarkPasswordVerify(b *testing.B) {
	h, err := NewHasher(DefaultParams, 1, 0)
	if err != nil {
		b.Fatal(err)
	}
	ctx := b.Context()
	stored, err := h.Hash(ctx, "correct horse battery staple")
	if err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		if ok, _, err := h.Verify(ctx, "correct horse battery staple", stored); err != nil || !ok {
			b.Fatal(ok, err)
		}
	}
}
