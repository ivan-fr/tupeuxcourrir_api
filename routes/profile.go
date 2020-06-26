package routes

import (
	"tupeuxcourrir_api/config"
	"tupeuxcourrir_api/controllers"
	TPCMiddleware "tupeuxcourrir_api/middleware"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func JWTProfileRoutes(group *echo.Group) {
	JwtConfig := TPCMiddleware.JWTConfig
	JwtConfig.SuccessHandler = TPCMiddleware.ImplementUserFromJwt(config.JwtLoginSubject)

	group.POST("/sendForValidateMail", controllers.SendForValidateMail, middleware.JWTWithConfig(JwtConfig))
	group.PUT("/putPhoto", controllers.PutPhoto, middleware.JWTWithConfig(JwtConfig))
	group.PUT("/putAddress", controllers.PutAddress, middleware.JWTWithConfig(JwtConfig))
}
