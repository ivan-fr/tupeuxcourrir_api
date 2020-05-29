package db

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type Connection struct {
	Db  *sql.DB
	Err error
}

var singletonConnector *Connection

func GetConnectionFromDB() *Connection {
	if singletonConnector == nil {
		db, err := sql.Open("mysql", "root:Koko32145.3@tcp(localhost)/tupeuxcourrir_bdd")
		singletonConnector = &Connection{db, err}
		return &Connection{db, err}
	}

	return singletonConnector
}
