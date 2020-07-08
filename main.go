package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"tupeuxcourrir_api/db"
	"tupeuxcourrir_api/routes"
)

func main() {
	defer db.Close()

	r := mux.NewRouter()
	routes.AuthRoutes(r.PathPrefix("/lol").Subrouter())

	http.Handle("/", r)
}
