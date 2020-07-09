package routes

import (
	"tupeuxcourrir_api/config"
	"tupeuxcourrir_api/controllers"
	TPCMiddleware "tupeuxcourrir_api/middleware"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func ProfileRoutes(group *echo.Group) {
	JwtConfig := TPCMiddleware.JWTConfig
	JwtConfig.SuccessHandler = TPCMiddleware.ImplementUserFromJwtSuccessHandler(config.JwtLoginSubject)

	group.GET("", controllers.GetProfile, middleware.JWTWithConfig(JwtConfig))
	group.POST("/sendForValidateMail", controllers.SendForValidateMail, middleware.JWTWithConfig(JwtConfig))
	group.PUT("/putPhoto", controllers.PutPhoto, middleware.JWTWithConfig(JwtConfig))
	group.PUT("/putAddress", controllers.PutAddress, middleware.JWTWithConfig(JwtConfig))

	JwtConfig.SuccessHandler = TPCMiddleware.ImplementUserFromJWTSuccessHandler(
		&TPCMiddleware.ImplementJWTUser{AddInitiatedThread: true,
			AddReceivedThread: true,
			Subject:           config.JwtLoginSubject,
			GiveMeSQB:         true})

	group.GET("/threads", controllers.GetThreads, middleware.JWTWithConfig(JwtConfig))
}
