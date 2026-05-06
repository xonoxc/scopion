package migrations

import "database/sql"

type CreateSpansTable struct{}

func (m *CreateSpansTable) ID() string {
	return "003_create_spans_table"
}

func (m *CreateSpansTable) UpSqlite(tx *sql.Tx) error {
	query := `
	CREATE TABLE IF NOT EXISTS spans (
		trace_id TEXT NOT NULL,
		span_id TEXT NOT NULL,
		parent_span_id TEXT,
		name TEXT NOT NULL,
		service TEXT NOT NULL,
		start_time DATETIME NOT NULL,
		end_time DATETIME NOT NULL,
		status TEXT NOT NULL,
		PRIMARY KEY (trace_id, span_id)
	);
	`
	_, err := tx.Exec(query)
	return err
}

func (m *CreateSpansTable) UpPostgres(tx *sql.Tx) error {
	query := `
	CREATE TABLE IF NOT EXISTS spans (
		trace_id TEXT NOT NULL,
		span_id TEXT NOT NULL,
		parent_span_id TEXT,
		name TEXT NOT NULL,
		service TEXT NOT NULL,
		start_time TIMESTAMPTZ NOT NULL,
		end_time TIMESTAMPTZ NOT NULL,
		status TEXT NOT NULL,
		PRIMARY KEY (trace_id, span_id)
	);
	`
	_, err := tx.Exec(query)
	return err
}
