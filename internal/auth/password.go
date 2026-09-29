package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/argon2"
	"golang.org/x/text/unicode/norm"
)

// Params are the Argon2id parameters (ADR-0010).
type Params struct {
	Time    uint32 // iterations (t)
	Memory  uint32 // KiB (m)
	Threads uint8  // parallelism (p)
	SaltLen int    // bytes
	KeyLen  uint32 // bytes of the tag
}

// DefaultParams are RFC 9106's second recommended option, the choice of
// ADR-0010: t=3, m=64 MiB, p=4, a 16-byte salt, and a 32-byte tag. S03.2
// benchmarks them; they may be lowered for a Raspberry Pi, never below the
// OWASP minimum (m=19 MiB, t=2, p=1), with the reason in the ADR.
var DefaultParams = Params{Time: 3, Memory: 64 * 1024, Threads: 4, SaltLen: 16, KeyLen: 32}

// Limits on the parameters read from a stored hash, so a damaged or
// crafted row cannot make one check use unbounded memory or time.
const (
	maxMemoryKiB = 1 << 20 // 1 GiB
	maxTime      = 20
	maxSaltLen   = 64
	maxKeyLen    = 64
)

// DefaultConcurrency is how many hashes run at once. Each needs Memory KiB
// (64 MiB by default), so two bound the memory of logins to 128 MiB, which
// a Raspberry Pi can afford (threat T-07, NFR-051).
const DefaultConcurrency = 2

// DefaultQueueWait is how long a hash waits for a free slot before the
// request is refused with ErrBusy.
const DefaultQueueWait = 10 * time.Second

// ErrBusy means every hashing slot stayed taken for the queue wait; the
// caller answers 503 unavailable, and the client retries.
var ErrBusy = errors.New("auth: too many password checks at once")

// ErrMalformedHash means a stored hash is not an Argon2id PHC string this
// package can read.
var ErrMalformedHash = errors.New("auth: malformed password hash")

// Hasher hashes and verifies passwords with Argon2id, at most a few at a
// time.
type Hasher struct {
	params Params
	slots  chan struct{}
	wait   time.Duration
	dummy  string // a hash of a random password, for unknown usernames
}

// NewHasher returns a hasher with the given parameters that runs at most
// concurrency hashes at once and waits at most wait for a slot. Zero values
// mean DefaultConcurrency and DefaultQueueWait.
func NewHasher(p Params, concurrency int, wait time.Duration) (*Hasher, error) {
	if concurrency <= 0 {
		concurrency = DefaultConcurrency
	}
	if wait <= 0 {
		wait = DefaultQueueWait
	}
	if err := p.check(); err != nil {
		return nil, err
	}
	h := &Hasher{params: p, slots: make(chan struct{}, concurrency), wait: wait}
	// The dummy hash has the current parameters, so checking a password of
	// an unknown user takes as long as checking a real one (threat T-05).
	dummy, err := h.hash(rand.Text())
	if err != nil {
		return nil, err
	}
	h.dummy = dummy
	return h, nil
}

func (p Params) check() error {
	switch {
	case p.Time < 1 || p.Time > maxTime:
		return fmt.Errorf("auth: Argon2id time must be 1 to %d, got %d", maxTime, p.Time)
	case p.Memory < 8*uint32(p.Threads) || p.Memory > maxMemoryKiB:
		return fmt.Errorf("auth: Argon2id memory must be %d KiB to 1 GiB, got %d KiB", 8*uint32(p.Threads), p.Memory)
	case p.Threads < 1:
		return errors.New("auth: Argon2id needs at least one thread")
	case p.SaltLen < 8 || p.SaltLen > maxSaltLen:
		return fmt.Errorf("auth: the salt must be 8 to %d bytes, got %d", maxSaltLen, p.SaltLen)
	case p.KeyLen < 16 || p.KeyLen > maxKeyLen:
		return fmt.Errorf("auth: the tag must be 16 to %d bytes, got %d", maxKeyLen, p.KeyLen)
	}
	return nil
}

// acquire takes a hashing slot, waiting at most the queue wait.
func (h *Hasher) acquire(ctx context.Context) (func(), error) {
	t := time.NewTimer(h.wait)
	defer t.Stop()
	select {
	case h.slots <- struct{}{}:
		return func() { <-h.slots }, nil
	case <-t.C:
		return nil, ErrBusy
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Hash returns the PHC string of a new hash of password.
func (h *Hasher) Hash(ctx context.Context, password string) (string, error) {
	release, err := h.acquire(ctx)
	if err != nil {
		return "", err
	}
	defer release()
	return h.hash(password)
}

func (h *Hasher) hash(password string) (string, error) {
	salt := make([]byte, h.params.SaltLen)
	_, _ = rand.Read(salt) // crypto/rand.Read never returns an error (Go 1.24+).
	p := h.params
	key := argon2.IDKey(normalizePassword(password), salt, p.Time, p.Memory, p.Threads, p.KeyLen)
	return encodePHC(p, salt, key), nil
}

// Verify reports whether password matches the stored hash, and whether the
// hash should be replaced because it was made with other parameters
// (ADR-0010: rehash on login when the parameters change).
func (h *Hasher) Verify(ctx context.Context, password, stored string) (ok, rehash bool, err error) {
	p, salt, key, err := decodePHC(stored)
	if err != nil {
		return false, false, err
	}
	release, err := h.acquire(ctx)
	if err != nil {
		return false, false, err
	}
	defer release()
	got := argon2.IDKey(normalizePassword(password), salt, p.Time, p.Memory, p.Threads, p.KeyLen)
	if subtle.ConstantTimeCompare(got, key) != 1 {
		return false, false, nil
	}
	cur := h.params
	rehash = p.Time != cur.Time || p.Memory != cur.Memory || p.Threads != cur.Threads ||
		p.SaltLen != cur.SaltLen || p.KeyLen != cur.KeyLen
	return true, rehash, nil
}

// VerifyDummy spends the time of one password check and always fails. It
// is used for unknown usernames, so they cannot be told apart by timing.
func (h *Hasher) VerifyDummy(ctx context.Context, password string) error {
	_, _, err := h.Verify(ctx, password, h.dummy)
	return err
}

// normalizePassword turns a password into the bytes that are hashed: its
// NFKC form, so the same password typed on another keyboard or system
// (composed or decomposed accents, fullwidth forms) still matches, as
// NIST SP 800-63B recommends for Unicode passwords.
func normalizePassword(password string) []byte {
	return []byte(norm.NFKC.String(password))
}

// encodePHC formats a hash in the PHC string format that Argon2's reference
// implementation uses: $argon2id$v=19$m=<KiB>,t=<n>,p=<n>$<salt>$<tag>,
// with unpadded standard base64.
func encodePHC(p Params, salt, key []byte) string {
	b64 := base64.RawStdEncoding
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, p.Memory, p.Time, p.Threads, b64.EncodeToString(salt), b64.EncodeToString(key))
}

// decodePHC reads a PHC string written by encodePHC, with bounds on every
// parameter.
func decodePHC(s string) (Params, []byte, []byte, error) {
	parts := strings.Split(s, "$")
	// "", "argon2id", "v=19", "m=..,t=..,p=..", salt, key
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" || parts[2] != "v="+strconv.Itoa(argon2.Version) {
		return Params{}, nil, nil, ErrMalformedHash
	}
	var p Params
	for _, kv := range strings.Split(parts[3], ",") {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			return Params{}, nil, nil, ErrMalformedHash
		}
		n, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return Params{}, nil, nil, ErrMalformedHash
		}
		switch k {
		case "m":
			p.Memory = uint32(n)
		case "t":
			p.Time = uint32(n)
		case "p":
			if n > 255 {
				return Params{}, nil, nil, ErrMalformedHash
			}
			p.Threads = uint8(n)
		default:
			return Params{}, nil, nil, ErrMalformedHash
		}
	}
	b64 := base64.RawStdEncoding
	salt, err := b64.DecodeString(parts[4])
	if err != nil {
		return Params{}, nil, nil, ErrMalformedHash
	}
	key, err := b64.DecodeString(parts[5])
	if err != nil {
		return Params{}, nil, nil, ErrMalformedHash
	}
	if len(salt) > maxSaltLen || len(key) > maxKeyLen {
		return Params{}, nil, nil, ErrMalformedHash
	}
	p.SaltLen, p.KeyLen = len(salt), uint32(len(key)) // #nosec G115 -- at most maxKeyLen, checked above
	if p.check() != nil {
		return Params{}, nil, nil, ErrMalformedHash
	}
	return p, salt, key, nil
}

// MinPasswordLength is the shortest password, in characters (decision D-6
// of the S03 stage document).
const MinPasswordLength = 12

// MaxPasswordBytes is the longest password, in bytes: enough for any
// passphrase, and a bound on the work of one hash.
const MaxPasswordBytes = 1024

// The password rules, reported as the rule of a weak_password problem.
const (
	RulePasswordTooShort     = "too_short"
	RulePasswordTooLong      = "too_long"
	RulePasswordNotUTF8      = "not_utf8"
	RulePasswordSameAsUser   = "same_as_username"
	RulePasswordControlChars = "control_character"
)

// PasswordError is a password that breaks a rule.
type PasswordError struct {
	Rule   string
	Detail string
}

func (e *PasswordError) Error() string { return "auth: " + e.Detail }

// CheckPassword applies the password rules: at least MinPasswordLength
// characters, at most MaxPasswordBytes bytes, valid UTF-8 without control
// characters, and not the username. There are no composition rules; the
// GUI suggests a passphrase instead (NIST SP 800-63B).
func CheckPassword(password, username string) error {
	switch {
	case !utf8.ValidString(password):
		return &PasswordError{RulePasswordNotUTF8, "the password is not valid text (UTF-8)"}
	case len(password) > MaxPasswordBytes:
		return &PasswordError{RulePasswordTooLong, fmt.Sprintf("the password is longer than %d bytes", MaxPasswordBytes)}
	case utf8.RuneCountInString(norm.NFKC.String(password)) < MinPasswordLength:
		return &PasswordError{RulePasswordTooShort, fmt.Sprintf("the password must have at least %d characters; a passphrase of a few words is easy to remember", MinPasswordLength)}
	case strings.EqualFold(strings.TrimSpace(password), strings.TrimSpace(username)):
		return &PasswordError{RulePasswordSameAsUser, "the password must not be the username"}
	}
	for _, r := range password {
		if r < 0x20 || r == 0x7f {
			return &PasswordError{RulePasswordControlChars, "the password must not contain control characters"}
		}
	}
	return nil
}
