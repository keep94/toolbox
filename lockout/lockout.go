// Package lockout locks accounts after consecutive login failures.
package lockout

import (
	"sync"
)

// Lockout locks out accounts after consecutive login failures.
// A nil Lockout pointer means no account lock out.
type Lockout struct {
	failures int
	lock     sync.Mutex
	counts   map[string]int
}

// New creates a New lockout instance. failures is the number of consecutive
// failures causing lockout. New panics if failures is less than 1.
// To disable lockout, use a nil pointer instead of calling New.
func New(failures int) *Lockout {
	if failures < 1 {
		panic("Failures must be at least 1")
	}
	return &Lockout{
		failures: failures,
		counts:   make(map[string]int),
	}
}

// Success indicates login success for given account and clears the number of
// consecutive failures for that account if account is not already locked.
func (l *Lockout) Success(userName string) {
	if l == nil {
		return
	}
	l.lock.Lock()
	defer l.lock.Unlock()
	// once locked, it stays locked
	if l.counts[userName] >= l.failures {
		return
	}
	delete(l.counts, userName)
}

// Failure indicates a login failure for given account. Failure returns true
// if that account is being locked because failure limit has just been
// reached.
func (l *Lockout) Failure(userName string) bool {
	if l == nil {
		return false
	}
	l.lock.Lock()
	defer l.lock.Unlock()
	l.counts[userName]++
	return l.counts[userName] == l.failures
}

// Locked returns true if given account is locked.
func (l *Lockout) Locked(userName string) bool {
	if l == nil {
		return false
	}
	l.lock.Lock()
	defer l.lock.Unlock()
	return l.counts[userName] >= l.failures
}
