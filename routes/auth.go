package routes

import (
	"tupeuxcourrir_api/controllers"
	"tupeuxcourrir_api/utils"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func AuthRoutes(group *echo.Group) {
	group.POST("/signUp", controllers.SignUp)
	group.POST("/login", controllers.Login)
	group.POST("/forgotPassword", controllers.ForgotPassword)
	group.POST("/editPassword", controllers.EditPasswordFromLink)
}

func JWTAuthRoutes(group *echo.Group) {
	group.Use(middleware.JWTWithConfig(utils.JWTConfig))
	group.POST("/editPassword", controllers.EditPasswordFromLink)
}
