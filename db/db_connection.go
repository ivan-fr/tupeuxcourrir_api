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
		db, err := sql.Open("mysql", "root:Koko32145.3@/tupeuxcourrir_bdd?parseTime=true&loc=Europe%2FParis")

		if err != nil {
			panic(err)
		}

		singletonConnector = &Connection{db, err}
	}

	return singletonConnector
}

func DeferClose() {
	if singletonConnector != nil {
		connection := GetConnectionFromDB()
		if err := connection.Db.Close(); err != nil {
			fmt.Println(err)
		}
	}
}
