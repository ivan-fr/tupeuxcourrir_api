package routes

import (
	"tupeuxcourrir_api/config"
	"tupeuxcourrir_api/controllers"
	TPCMiddleware "tupeuxcourrir_api/middleware"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func JWTCheckerRoutes(group *echo.Group) {
	JWTconfig := TPCMiddleware.JWTConfig
	JWTconfig.SuccessHandler = TPCMiddleware.ImplementUserFromJwt(config.JwtCheckEmailSubject)

	group.Use(middleware.JWTWithConfig(JWTconfig))
	group.POST("/checkMail", controllers.CheckMail)
}
