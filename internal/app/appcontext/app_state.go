package appcontext

import (
	"sync"

	"github.com/xonoxc/scopion/internal/store"
)

type ApplicationState struct {
	Store        store.Storage
	StorageState store.StorageState
}

type AtomicAppState struct {
	mu   sync.Mutex
	snap ApplicationState
}

func NewAtomicAppState(store store.Storage, state store.StorageState) *AtomicAppState {
	if store == nil || state == "" {
		panic("invalid initial app state")
	}

	return &AtomicAppState{
		snap: ApplicationState{
			Store:        store,
			StorageState: state,
		},
	}
}

func (a *AtomicAppState) Snapshot() ApplicationState {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.snap
}

func (a *AtomicAppState) Set(store store.Storage, state store.StorageState) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.snap = ApplicationState{
		Store:        store,
		StorageState: state,
	}
}
