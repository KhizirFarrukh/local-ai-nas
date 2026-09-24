package storage

import (
	"strings"
	"sync"
)

// Locks is a set of mutexes keyed by name (S01.6-T05). A key's mutex is
// created on first use and dropped when no one holds or waits for it, so
// the set stays as small as the work in progress. The files service locks
// a folder while it commits a new name into it, which makes its
// check-then-rename steps safe: the server is the only writer to the
// storage.
type Locks struct {
	mu   sync.Mutex
	keys map[string]*keyLock
}

type keyLock struct {
	mu   sync.Mutex
	refs int // holders and waiters
}

// NewLocks returns an empty lock set.
func NewLocks() *Locks {
	return &Locks{keys: map[string]*keyLock{}}
}

// Lock locks key, waiting while another caller holds it, and returns the
// function that unlocks it; call that exactly once. Keys are compared
// without case, so "Docs" and "docs" share a lock, as they share a folder
// on a case-insensitive disk.
func (l *Locks) Lock(key string) (unlock func()) {
	key = strings.ToLower(key)
	l.mu.Lock()
	k := l.keys[key]
	if k == nil {
		k = &keyLock{}
		l.keys[key] = k
	}
	k.refs++
	l.mu.Unlock()

	k.mu.Lock()
	return func() {
		k.mu.Unlock()
		l.mu.Lock()
		k.refs--
		if k.refs == 0 {
			delete(l.keys, key)
		}
		l.mu.Unlock()
	}
}

// Len returns the number of keys that are held or waited for.
func (l *Locks) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.keys)
}
