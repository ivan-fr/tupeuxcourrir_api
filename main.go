package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"tupeuxcourrir_api/db"
)

func deferClose(connection *db.Connection) {
	if err := connection.Db.Close(); err != nil {
		fmt.Println(err)
	}
}

func main() {
	var connection = db.GetConnectionFromDB()

	if connection.Err != nil {
		fmt.Println("Failed connect to databse: ", connection.Err)
		return
	}

	defer deferClose(connection)

	server := gin.Default()
	_ = server.Run(":8080")
}
