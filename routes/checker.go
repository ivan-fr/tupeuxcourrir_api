package routes

import (
	"tupeuxcourrir_api/config"
	"tupeuxcourrir_api/controllers"
	TPCMiddleware "tupeuxcourrir_api/middleware"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func CheckerRoutes(group *echo.Group) {
	JwtConfig := TPCMiddleware.JWTConfig
	JwtConfig.SuccessHandler = TPCMiddleware.ImplementUserJwtSuccessHandler(config.JwtCheckEmailSubject)

	group.POST("/checkMail", controllers.CheckMail, middleware.JWTWithConfig(JwtConfig))
}
