package main

import (
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"tupeuxcourrir_api/db"
	"tupeuxcourrir_api/routes"
)

func main() {
	defer db.Close()

	r := mux.NewRouter()
	routes.AuthRoutes(r.PathPrefix("/auth").Subrouter())

	loggedRouter := handlers.LoggingHandler(os.Stdout, r)
	recoveryHandler := handlers.RecoveryHandler()(loggedRouter)
	log.Fatal(http.ListenAndServe(":1123", recoveryHandler))
}
