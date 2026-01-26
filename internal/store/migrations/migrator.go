package migrations

import (
	"database/sql"
	"fmt"

	migrateable "github.com/xonoxc/scopion/internal/store/migratable"
)

type Migrator struct {
	Dsn string
}

func New(connStr string) *Migrator {
	return &Migrator{
		Dsn: connStr,
	}
}

func (m *Migrator) Migrate(dialect migrateable.DatabaseName, migrations []Migration) error {
	conn, err := connByDialect(dialect, m.Dsn)
	if err != nil {
		return err
	}
	defer conn.Close()

	return withTransaction(conn, func(tx *sql.Tx) error {
		return runMigrations(dialect, tx, migrations)
	})
}

func connByDialect(dialect migrateable.DatabaseName, dsn string) (*sql.DB, error) {
	driver, err := driverFor(dialect)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("migration open err: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("migration ping err: %w", err)
	}

	return db, nil
}

func runMigrations(dialect migrateable.DatabaseName, tx *sql.Tx, migrations []Migration) error {
	for _, migr := range migrations {
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

func driverFor(d migrateable.DatabaseName) (string, error) {
	if !d.Valid() {
		return "", fmt.Errorf("unsupported dialect: %s", d)
	}
	return string(d), nil
}
