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

	subgroup := group.Group("/jwt")
	JWTconfig := TPCMiddleware.JWTConfig
	JWTconfig.SuccessHandler = TPCMiddleware.ImplementUserFromJwt(config.JwtEditPasswordSubject)

	subgroup.POST("/editPassword", controllers.EditPasswordFromLink, middleware.JWTWithConfig(JWTconfig))
}
