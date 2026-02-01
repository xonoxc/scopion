package orchestrator

import (
	"fmt"

	"github.com/xonoxc/scopion/internal/app/appcontext"
	"github.com/xonoxc/scopion/internal/store"
	"github.com/xonoxc/scopion/internal/store/dbprovisioner"
	"github.com/xonoxc/scopion/internal/store/dualwrite"
	"github.com/xonoxc/scopion/internal/store/migrations"
	"github.com/xonoxc/scopion/internal/store/postgres"
)

/*
* Orchestrator
* is responsible for migrations
* and handling the switching process
**/
type Orchestrator struct {
	app *appcontext.AtomicAppState
}

func New(appState *appcontext.AtomicAppState) *Orchestrator {
	return &Orchestrator{
		app: appState,
	}
}

func (o *Orchestrator) MigrateTo(targetState store.StorageState, secondaryDSN string) error {
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
		return o.switchToDualWrite(secondaryDSN)

	case store.DUAL_WRITE:
		if targetState != store.SINGLE_SECONDARY {
			return fmt.Errorf("illegal transition: %s → %s", storageState, targetState)
		}
		return o.promoteSecondary()

	default:
		panic("unknown system state")

	}
}

func (o *Orchestrator) switchToDualWrite(dsn string) error {
	snap := o.app.Snapshot()
	primaryStore := snap.Store

	migrator := migrations.NewMigrator(migrations.GetAll())
	dbProvisioner := dbprovisioner.New(migrator)

	dbConn, err := dbProvisioner.Provision("postgres", dsn)
	if err != nil {
		return err
	}

	dw := dualwrite.New(primaryStore, postgres.NewWithDB(dbConn))
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
