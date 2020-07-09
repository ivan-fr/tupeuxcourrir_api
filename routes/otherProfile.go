package routes

import (
	"github.com/gorilla/mux"
	"tupeuxcourrir_api/config"
	"tupeuxcourrir_api/controllers"
	TPCMiddleware "tupeuxcourrir_api/middleware"
)

func OtherProfileRoutes(group *mux.Router) {
	group.HandleFunc("/:id", controllers.GetOtherProfile).Methods("GET")

	JwtConfig := TPCMiddleware.MyJWTUserConfig
	JwtConfig.SuccessHandler = TPCMiddleware.ImplementUserJwtSuccessHandler(&TPCMiddleware.ImplementJWTUser{Subject: config.JwtLoginSubject})
	group.HandleFunc("/:id/makeThread", controllers.MakeThreadWithOtherProfile).
		Methods("POST").
		Subrouter().
		Use(TPCMiddleware.JWTWithConfig(JwtConfig))
}
