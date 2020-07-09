package routes

import (
	"github.com/gorilla/mux"
	"tupeuxcourrir_api/config"
	"tupeuxcourrir_api/controllers"
	TPCMiddleware "tupeuxcourrir_api/middleware"
)

func WsThreadRoutes(group *mux.Router) {
	JwtConfig := TPCMiddleware.MyJWTUserConfig
	JwtConfig.SuccessHandler = TPCMiddleware.ImplementUserJwtSuccessHandler(&TPCMiddleware.ImplementJWTUser{Subject: config.JwtLoginSubject})
	group.HandleFunc("/thread", controllers.WsThread).
		Methods("GET").
		Subrouter().
		Use(TPCMiddleware.JWTWithConfig(JwtConfig))
}
