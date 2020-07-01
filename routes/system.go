package routes

import (
	"github.com/labstack/echo/v4"
	"tupeuxcourrir_api/controllers"
)

func SystemRoutes(group *echo.Group) {
	group.GET("/uri/:routeName", controllers.GetUri)
}
