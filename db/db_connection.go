package db

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type connection struct {
	Db  *sql.DB
	Err error
}

var singletonConnector *connection

func GetConnectionFromDB() *connection {
	if singletonConnector == nil {
		db, err := sql.Open("mysql", "root:YpEp5Kh7g.3/tupeuxcourrir_bdd?parseTime=true&loc=Europe%2FParis")

		if err != nil {
			panic(err)
		}

		singletonConnector = &connection{db, err}
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
