package tgbot

import "sync"

// CommandLock prevents concurrent command execution per user.
// This is useful to avoid race conditions when a user sends
// multiple commands before the first one completes.
type CommandLock struct {
	mu    sync.Mutex
	locks map[int64]bool
}

// NewCommandLock creates a new CommandLock instance.
func NewCommandLock() *CommandLock {
	return &CommandLock{
		locks: make(map[int64]bool),
	}
}

// TryLock attempts to acquire lock for user.
// Returns true if lock was acquired, false if user already has a lock.
func (l *CommandLock) TryLock(userID int64) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.locks[userID] {
		return false
	}
	l.locks[userID] = true
	return true
}

// Unlock releases lock for user.
func (l *CommandLock) Unlock(userID int64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.locks, userID)
}

// IsLocked returns true if user has an active lock.
func (l *CommandLock) IsLocked(userID int64) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.locks[userID]
}
