package orchestrator

import (
	"fmt"

	"github.com/xonoxc/scopion/internal/app/appcontext"
	"github.com/xonoxc/scopion/internal/store"
	"github.com/xonoxc/scopion/internal/store/dualwrite"
	"github.com/xonoxc/scopion/internal/store/migrations"
	"github.com/xonoxc/scopion/internal/store/postgres"

	migrateable "github.com/xonoxc/scopion/internal/store/migratable"
)

/*
* Orchestrator
* is responsible for migrations
* and handling the switching process
**/
type Orchestrator struct {
	app      *appcontext.AtomicAppState
	migrator *migrations.Migrator
}

func New(appState *appcontext.AtomicAppState, migrator *migrations.Migrator) *Orchestrator {
	return &Orchestrator{
		migrator: migrator,
		app:      appState,
	}
}

func (o *Orchestrator) MigrateTo(targetState store.StorageState) error {
	currentState := o.app.Snapshot()

	storageState := currentState.StorageState

	if storageState == targetState {
		return nil
	}

	switch storageState {
	case store.SINGLE_PRIMARY:
		if targetState != store.DUAL_WRITE {
			return fmt.Errorf("illegal transition: %s → %s", storageState, targetState)
		}
		return o.switchToDualWrite()

	case store.DUAL_WRITE:
		if targetState != store.SINGLE_SECONDARY {
			return fmt.Errorf("illegal transition: %s → %s", storageState, targetState)
		}
		return o.promoteSecondary()

	default:
		panic("unknown system state")

	}
}

func (o *Orchestrator) switchToDualWrite() error {
	snap := o.app.Snapshot()
	primary := snap.Store

	secondaryStore, err := postgres.New(o.migrator.Dsn)
	if err != nil {
		return err
	}

	if err := o.migrator.Migrate(migrateable.POSTGRES, migrations.GetAll()); err != nil {
		return err
	}

	dw := dualwrite.New(primary, secondaryStore)
	o.app.Set(dw, store.DUAL_WRITE)

	return nil
}

func (o *Orchestrator) promoteSecondary() error {
	snap := o.app.Snapshot()

	dw, ok := snap.Store.(*dualwrite.DualWriteStore)
	if !ok {
		return fmt.Errorf("expected dual write store")
	}

	o.app.Set(dw.Secondary(), store.SINGLE_SECONDARY)
	return nil
}
