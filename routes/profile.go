package routes

import (
	"github.com/gorilla/mux"
	"tupeuxcourrir_api/config"
	"tupeuxcourrir_api/controllers"
	TPCMiddleware "tupeuxcourrir_api/middleware"
)

func ProfileRoutes(group *mux.Router) {
	JwtConfig := TPCMiddleware.MyJWTUserConfig
	JwtConfig.SuccessHandler = TPCMiddleware.ImplementUserJwtSuccessHandler(&TPCMiddleware.ImplementJWTUser{Subject: config.JwtLoginSubject})

	group.HandleFunc("", controllers.GetProfile).
		Subrouter().
		Use(TPCMiddleware.JWTWithConfig(JwtConfig))

	group.HandleFunc("/sendForValidateMail", controllers.SendForValidateMail).
		Subrouter().
		Use(TPCMiddleware.JWTWithConfig(JwtConfig))

	group.HandleFunc("/putPhoto", controllers.PutPhoto).
		Subrouter().
		Use(TPCMiddleware.JWTWithConfig(JwtConfig))

	group.HandleFunc("/putAddress", controllers.PutAddress).
		Subrouter().
		Use(TPCMiddleware.JWTWithConfig(JwtConfig))

	JwtConfig2 := TPCMiddleware.MyJWTUserConfig
	JwtConfig2.SuccessHandler = TPCMiddleware.ImplementUserJwtSuccessHandler(
		&TPCMiddleware.ImplementJWTUser{AddInitiatedThread: true,
			AddReceivedThread: true,
			Subject:           config.JwtLoginSubject,
			GiveMeSQB:         true})

	group.HandleFunc("/threads", controllers.GetThreads).
		Subrouter().
		Use(TPCMiddleware.JWTWithConfig(JwtConfig2))
}
