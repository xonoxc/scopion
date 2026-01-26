package store

import (
	"sync"
)

/**
* HTTP Handler ─┐
* Background Job├──> AtomicStore ───> SQLite / DualWrite / Postgres
* Metrics       ┘
**/

/**
*this is go routine safe
*pointer to currently active store
*so that there are not race conditions
*or inconsistent reads/writes
 */

type StoreSnapshot struct {
	store        Storage
	currentState StorageState
}

type AtomicStore struct {
	mu       sync.Mutex
	currSnap StoreSnapshot
}

func NewAtomicStore(initialSnap StoreSnapshot) *AtomicStore {
	if initialSnap.store == nil {
		panic("AtomicStore: store must not be nil")
	}
	if initialSnap.currentState == "" {
		panic("AtomicStore: state must not be empty")
	}

	return &AtomicStore{
		currSnap: initialSnap,
	}
}

func (a *AtomicStore) GetCurrentStorageState() StorageState {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.currSnap.currentState
}

func (a *AtomicStore) GetCurrSnapShot() StoreSnapshot {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.currSnap
}

func (a *AtomicStore) Update(store Storage, currentState StorageState) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.currSnap = StoreSnapshot{
		store:        store,
		currentState: currentState,
	}
}
