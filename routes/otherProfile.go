package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"tupeuxcourrir_api/config"
	"tupeuxcourrir_api/controllers"
	TPCMiddleware "tupeuxcourrir_api/middleware"
)

func OtherProfileRoutes(group *echo.Group) {
	group.GET("/:id", controllers.GetOtherProfile)

	JwtConfig := TPCMiddleware.JWTConfig
	JwtConfig.SuccessHandler = TPCMiddleware.ImplementUserFromJwt(config.JwtLoginSubject)

	group.POST("/:id/makeThread", controllers.MakeThreadWithOtherProfile, middleware.JWTWithConfig(JwtConfig))
}
