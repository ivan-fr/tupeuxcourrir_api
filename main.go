package main

import (
	"github.com/gin-gonic/gin"
	"tupeuxcourrir_api/db"
)

func main() {
	defer db.DeferClose()

	server := gin.Default()
	_ = server.Run(":8080")
}
