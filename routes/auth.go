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
	group.HandleFunc("/forgotPassword", controllers.ForgotPassword).Methods("POST")

	subgroup := group.PathPrefix("/jwt/editPassword").Subrouter()
	JwtConfig := TPCMiddleware.MyJWTUserConfig
	JwtConfig.SuccessHandler = TPCMiddleware.ImplementUserJwtSuccessHandler(&TPCMiddleware.ImplementJWTUser{Subject: config.JwtEditPasswordSubject})

	subgroup.Use(TPCMiddleware.JWTWithConfig(JwtConfig))
	subgroup.HandleFunc("", controllers.EditPasswordFromLink).
		Methods("POST")
}
