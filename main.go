package main

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"tupeuxcourrir_api/controllers"
	"tupeuxcourrir_api/db"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func main() {
	defer db.DeferClose()

	e := echo.New()

	e.Validator = &CustomValidator{validator: validator.New()}

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))
	e.Use(middleware.Recover())

	e.POST("/signUp", controllers.SignUp)
	e.POST("/login", controllers.Login)
	e.POST("/forgotPassword", controllers.ForgotPassword)

	e.Logger.Fatal(e.Start(":8080"))
}
