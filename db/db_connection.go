package db

import (
	"database/sql"
	"fmt"
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

		if err != nil {
			panic(err)
		}

		singletonConnector = &Connection{db, err}
		return &Connection{db, err}
	}

	return singletonConnector
}

func DeferClose() {
	connection := GetConnectionFromDB()
	if err := connection.Db.Close(); err != nil {
		fmt.Println(err)
	}
}
