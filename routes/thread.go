package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"tupeuxcourrir_api/config"
	"tupeuxcourrir_api/controllers"
	TPCMiddleware "tupeuxcourrir_api/middleware"
)

func WsRoutes(group *echo.Group) {
	JwtConfig := TPCMiddleware.JWTConfig
	JwtConfig.SuccessHandler = TPCMiddleware.ImplementUserFromJwt(config.JwtLoginSubject)
	group.GET("/thread", controllers.WsThread, middleware.JWTWithConfig(JwtConfig))
}
