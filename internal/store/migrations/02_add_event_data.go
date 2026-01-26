package migrations

import "database/sql"

type AddEventDataColumn struct{}

func (m *AddEventDataColumn) ID() string {
	return "02_add_event_data_column"
}

func (m *AddEventDataColumn) UpPostgres(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE events
		ADD COLUMN IF NOT EXISTS data TEXT;
	`)
	return err
}

func (m *AddEventDataColumn) DownPostgres(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE events
		DROP COLUMN IF EXISTS data;
	`)
	return err
}

func (m *AddEventDataColumn) UpSqlite(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE events
		ADD COLUMN data TEXT;

	`)

	return err
}

func (m *AddEventDataColumn) DownSqlite(tx *sql.Tx) error {
	// SQLite does not support DROP COLUMN reliably.
	// We intentionally make this a no-op.
	return nil
}
