package routes

import (
	"tupeuxcourrir_api/config"
	"tupeuxcourrir_api/controllers"
	TPCMiddleware "tupeuxcourrir_api/middleware"

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
	JWTconfig := TPCMiddleware.JWTConfig
	JWTconfig.SuccessHandler = TPCMiddleware.ImplementUserFromJwt(config.JwtEditPasswordSubject)

	group.Use(middleware.JWTWithConfig(JWTconfig))
	group.POST("/editPassword", controllers.EditPasswordFromLink)
}

func JWTCheckerRoutes(group *echo.Group) {
	JWTconfig := TPCMiddleware.JWTConfig
	JWTconfig.SuccessHandler = TPCMiddleware.ImplementUserFromJwt(config.JwtLoginSubject)

	group.Use(middleware.JWTWithConfig(JWTconfig))
	group.POST("/sendForValidateMail", controllers.SendForValidateMail)
	group.POST("/checkMail", controllers.CheckMail)
}
