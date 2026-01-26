package migrations

import "database/sql"

type CreateEventsTable struct{}

func (m *CreateEventsTable) ID() string {
	return "001_create_event_table"
}

func (m *CreateEventsTable) UpSqlite(tx *sql.Tx) error {
	query := `
	CREATE TABLE IF NOT EXISTS events (
		id TEXT PRIMARY KEY,
		timestamp DATETIME NOT NULL,
		level TEXT NOT NULL,
		service TEXT NOT NULL,
		name TEXT NOT NULL,
		trace_id TEXT NOT NULL
	);
	`
	_, err := tx.Exec(query)
	return err
}

func (m *CreateEventsTable) UpPostgres(tx *sql.Tx) error {
	query := `
	CREATE TABLE IF NOT EXISTS events (
		id TEXT PRIMARY KEY,
		timestamp TIMESTAMPTZ NOT NULL,
		level TEXT NOT NULL,
		service TEXT NOT NULL,
		name TEXT NOT NULL,
		trace_id TEXT NOT NULL
	);
	`
	_, err := tx.Exec(query)
	return err
}

func (m *CreateEventsTable) DownSqlite(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS events;`)
	return err
}
