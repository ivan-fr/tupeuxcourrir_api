package routes

import (
	"github.com/gorilla/mux"
	"tupeuxcourrir_api/config"
	"tupeuxcourrir_api/controllers"
	TPCMiddleware "tupeuxcourrir_api/middleware"
)

func CheckerRoutes(group *mux.Router) {
	JwtConfig := TPCMiddleware.MyJWTUserConfig
	JwtConfig.TokenLookup = "param:token"
	JwtConfig.SuccessHandler = TPCMiddleware.ImplementUserJwtSuccessHandler(&TPCMiddleware.ImplementJWTUser{Subject: config.JwtCheckEmailSubject})
	group.HandleFunc("/checkMail/{token}", controllers.CheckMail).
		Methods("POST").
		Subrouter().
		Use(TPCMiddleware.JWTWithConfig(JwtConfig))
}
