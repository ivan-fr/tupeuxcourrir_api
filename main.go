package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
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
	defer db.Close()

	e := echo.New()
	e.Use(middleware.BodyLimit("3M"))

	e.Validator = &customValidator{validator: validator.New()}

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}, error=${error}\n",
	}))

	for _, r := range e.Routes() {
		log.Println(r.Name)
	}

	e.Static("/static", "public")
	routes.AuthRoutes(e.Group("/auth"))
	routes.CheckerRoutes(e.Group("/checker"))
	routes.ProfileRoutes(e.Group("/profile"))
	routes.OtherProfileRoutes(e.Group("/profiles"))
	routes.WsThreadRoutes(e.Group("/ws"))
	routes.SystemRoutes(e.Group("/system"))

	registeredRoutes, _ := json.MarshalIndent(e.Routes(), "", "  ")
	_ = ioutil.WriteFile("routes.json", registeredRoutes, 0644)

	e.Logger.Fatal(e.Start(":8080"))
}
