package utils

import (
	"fmt"
	"sync"
)

// tran ..
type tran struct {
	locker *sync.Mutex
	holder int
}

// newTran ..
func newTran() *tran {
	return &tran{
		locker: new(sync.Mutex),
		holder: 0,
	}
}

// Transaction ..
type Transaction struct {
	mu      *sync.Mutex
	mapping map[string]*tran
}

// NewTransaction returns a new Transaction.
func NewTransaction() *Transaction {
	return &Transaction{
		mu:      new(sync.Mutex),
		mapping: make(map[string]*tran),
	}
}

// Lock locks the transaction that corresponding to identifier.
// If target transaction not exist, then it will creates a new one.
func (t *Transaction) Lock(identifier string) {
	var tran *tran
	var ok bool

	func() {
		t.mu.Lock()
		defer t.mu.Unlock()

		tran, ok = t.mapping[identifier]
		if !ok {
			tran = newTran()
			t.mapping[identifier] = tran
		}
		tran.holder++
	}()

	tran.locker.Lock()
}

// TryLock tries to lock the transaction that corresponding to
// identifier, and reports whether it succeeded.
//
// Note that while correct uses of TryLock do exist, they are
// rare, and use of TryLock is often a sign of a deeper problem
// in a particular use of mutexes.
func (t *Transaction) TryLock(identifier string) bool {
	var tran *tran
	var ok bool

	t.mu.Lock()
	defer t.mu.Unlock()

	tran, ok = t.mapping[identifier]
	if !ok {
		tran = newTran()
		t.mapping[identifier] = tran
	}

	if tran.locker.TryLock() {
		tran.holder++
		return true
	}
	return false
}

// Unlock unlocks the transaction that corresponding to identifier.
// If target transaction not exist, or the underlying locker is
// already unlocked, then this func will be panic.
func (t *Transaction) Unlock(identifier string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	tran, ok := t.mapping[identifier]
	if !ok {
		// We should panic here because when here is not ok,
		// it means somewhere may have some internal error,
		// and the lock states is not completely.
		panic(fmt.Sprintf("Unlock: Transaction %#v not found", identifier))
	}

	tran.holder--
	if tran.holder == 0 {
		delete(t.mapping, identifier)
	}
	tran.locker.Unlock()
}
