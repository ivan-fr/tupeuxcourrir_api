package main

import (
	"github.com/gin-gonic/gin"
	"tupeuxcourrir_api/controllers"
	"tupeuxcourrir_api/db"
)

func main() {
	defer db.DeferClose()

	router := gin.Default()

	router.POST("/signUp", controllers.SignUp)
	router.POST("/login", controllers.Login)
	router.POST("/forgotPassword", controllers.ForgotPassword)

	_ = router.Run()
}
