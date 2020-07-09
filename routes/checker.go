package routes

import (
	"github.com/gorilla/mux"
	"tupeuxcourrir_api/config"
	"tupeuxcourrir_api/controllers"
	TPCMiddleware "tupeuxcourrir_api/middleware"
)

func CheckerRoutes(group *mux.Router) {
	JwtConfig := TPCMiddleware.MyJWTUserConfig
	JwtConfig.SuccessHandler = TPCMiddleware.ImplementUserJwtSuccessHandler(&TPCMiddleware.ImplementJWTUser{Subject: config.JwtCheckEmailSubject})
	group.HandleFunc("/checkMail", controllers.CheckMail).Subrouter().Use(TPCMiddleware.JWTWithConfig(JwtConfig))
}
