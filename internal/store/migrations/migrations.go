package migrations

import "database/sql"

type Migration interface {
	/*
	* returns the unique identifier of the migration
	 */
	ID() string

	/*
	* returns the SQL statement to apply the migration for Postgres
	**/
	UpPostgres(tx *sql.Tx) error

	/*
		* returns the SQL statement to apply the migration for Sqlite
	    ***/
	UpSqlite(tx *sql.Tx) error
}

/*
* method to get all migrations in order
***/
func GetAll() []Migration {
	return []Migration{
		&CreateEventsTable{},
		&AddEventDataColumn{},
	}
}
