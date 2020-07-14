package routes

import (
	"github.com/gorilla/mux"
	"tupeuxcourrir_api/config"
	"tupeuxcourrir_api/controllers"
	TPCMiddleware "tupeuxcourrir_api/middleware"
)

func AuthRoutes(group *mux.Router) {
	group.HandleFunc("/signUp", controllers.SignUp).Methods("POST")
	group.HandleFunc("/login", controllers.Login).Methods("POST")
	group.HandleFunc("/forgotPassword", controllers.SendForgotPassword).Methods("POST")

	JwtConfig := TPCMiddleware.MyJWTUserConfig
	JwtConfig.TokenLookup = "param:token"
	JwtConfig.SuccessHandler = TPCMiddleware.ImplementUserJwtSuccessHandler(
		&TPCMiddleware.ImplementJWTUser{Subject: config.JwtEditPasswordSubject})

	subgroup := group.PathPrefix("/jwt/editPassword/{token}").Subrouter()
	subgroup.Use(TPCMiddleware.JWTWithConfig(JwtConfig))
	subgroup.HandleFunc("", controllers.EditPassword).
		Methods("POST")
}
