package data

import "github.com/jmoiron/sqlx"

var (
	// Dbo global db client
	Dbo *sqlx.DB
	// DbType global db type
	DbType string
	// DbConn global db connection string
	DbConn string
)
