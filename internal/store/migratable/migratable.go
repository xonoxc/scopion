package migrateable

import "database/sql"

/*
* possible database dialects
**/
type DatabaseName string

const (
	POSTGRES DatabaseName = "postgres"
	SQLITE   DatabaseName = "sqlite3"
)

func (d DatabaseName) Valid() bool {
	switch d {
	case POSTGRES, SQLITE:
		return true
	default:
		return false
	}
}

/*
* database will implement this interface to provide the underlying sql.DB connection
**/
type Migratable interface {
	DB() *sql.DB
	Dialect() DatabaseName
}
