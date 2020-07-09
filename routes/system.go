package routes

import (
	"github.com/gorilla/mux"
	"tupeuxcourrir_api/controllers"
)

func SystemRoutes(mainRouter, group *mux.Router) {
	group.HandleFunc("/uri", controllers.GetUriBox(mainRouter))
}
