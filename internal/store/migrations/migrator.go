package migrations

import (
	"database/sql"
	"fmt"

	migrateable "github.com/xonoxc/scopion/internal/store/migratable"
)

type Migrator struct {
	migrations []Migration
}

func NewMigrator(migrations []Migration) *Migrator {
	return &Migrator{
		migrations: migrations,
	}
}

func (m *Migrator) Migrate(conn *sql.DB, dialect migrateable.DatabaseName) error {
	return withTransaction(conn, func(tx *sql.Tx) error {
		return m.runMigrations(dialect, tx)
	})
}

func (m *Migrator) runMigrations(dialect migrateable.DatabaseName, tx *sql.Tx) error {
	for _, migr := range m.migrations {
		switch dialect {
		case migrateable.POSTGRES:
			if err := migr.UpPostgres(tx); err != nil {
				return fmt.Errorf("migration %s failed: %w", migr.ID(), err)
			}
		case migrateable.SQLITE:
			if err := migr.UpSqlite(tx); err != nil {
				return fmt.Errorf("migration %s failed: %w", migr.ID(), err)
			}
		}
	}
	return nil
}

func (m *Migrator) AquireMigrationLock(tx *sql.DB) error {
	_, err := tx.Exec("SELECT pg_advisory_lock(424242)")
	if err != nil {
		return fmt.Errorf("failed to acquire migration lock: %w", err)
	}

	return nil
}

func (m *Migrator) ReleaseMigrationLock(tx *sql.DB) error {
	_, err := tx.Exec("SELECT pg_advisory_unlock(424242)")
	if err != nil {
		return fmt.Errorf("failed to acquire migration lock: %w", err)
	}

	return nil
}

func withTransaction(db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}
