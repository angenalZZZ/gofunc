/*
Package locker is a simple package to manage named ReadWrite mutexes. These
appear to be especially useful for synchronizing access to session based
information in web applications.

The common use case is to use the package level functions, which use a package
level set of locks (safe to use from multiple goroutines simultaneously).
However, you may also create a new separate set of locks.

All locks are implemented with read-write mutexes. To use them like a regular
mutex, simply ignore the RLock/RUnlock functions.
*/
package f

// github.com/BurntSushi/locker
// BUG: The locker here can grow without bound in long running
// programs. Since it's intended to be used in web applications, this is a
// major problem. Figure out a way to keep the locker lean.

import (
	"sync"
)

// Locker represents the set of named ReadWrite mutexes. It is safe to access
// from multiple goroutines simultaneously.
type Locker struct {
	locks   map[string]*sync.RWMutex
	locksRW *sync.RWMutex
}

var locker *Locker

func init() {
	locker = NewLocker()
}

func Lock(key string)    { locker.Lock(key) }
func Unlock(key string)  { locker.Unlock(key) }
func RLock(key string)   { locker.RLock(key) }
func RUnlock(key string) { locker.RUnlock(key) }
func DeLock(key string)  { locker.deleteLock(key) }

func NewLocker() *Locker {
	return &Locker{
		locks:   make(map[string]*sync.RWMutex),
		locksRW: new(sync.RWMutex),
	}
}

func (l *Locker) Lock(key string) {
	lk, ok := l.getLock(key)
	if !ok {
		lk = l.newLock(key)
	}
	lk.Lock()
}

func (l *Locker) Unlock(key string) {
	lk, ok := l.getLock(key)
	if ok {
		lk.Unlock()
	}
}

func (l *Locker) RLock(key string) {
	lk, ok := l.getLock(key)
	if !ok {
		lk = l.newLock(key)
	}
	lk.RLock()
}

func (l *Locker) RUnlock(key string) {
	lk, ok := l.getLock(key)
	if ok {
		lk.RUnlock()
	}
}

func (l *Locker) newLock(key string) *sync.RWMutex {
	l.locksRW.Lock()
	defer l.locksRW.Unlock()

	if lk, ok := l.locks[key]; ok {
		return lk
	}
	lk := new(sync.RWMutex)
	l.locks[key] = lk
	return lk
}

func (l *Locker) getLock(key string) (*sync.RWMutex, bool) {
	l.locksRW.RLock()
	defer l.locksRW.RUnlock()

	lock, ok := l.locks[key]
	return lock, ok
}

func (l *Locker) deleteLock(key string) {
	l.locksRW.Lock()
	defer l.locksRW.Unlock()

	if _, ok := l.locks[key]; ok {
		delete(l.locks, key)
	}
}
