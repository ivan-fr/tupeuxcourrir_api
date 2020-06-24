package main

import (
	"tupeuxcourrir_api/db"
	"tupeuxcourrir_api/routes"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type customValidator struct {
	validator *validator.Validate
}

func (cv *customValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func main() {
	defer db.DeferClose()

	e := echo.New()

	e.Validator = &customValidator{validator: validator.New()}

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}, error=${error}\n",
	}))

	routes.AuthRoutes(e.Group("/auth"))
	routes.JWTCheckerRoutes(e.Group("/checker"))
	routes.JWTProfileRoutes(e.Group("/profile"))

	e.Logger.Fatal(e.Start(":8080"))
}
