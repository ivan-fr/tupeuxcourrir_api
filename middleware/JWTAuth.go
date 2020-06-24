package middleware

import (
	"tupeuxcourrir_api/config"
	"tupeuxcourrir_api/models"
	"tupeuxcourrir_api/orm"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type ImplementJWTUser struct {
	addInitiatedThread bool
	addRoles           bool
	addReceivedThread  bool
	subject            string
}

type JwtCustomClaims struct {
	UserID int `json:"id"`
	jwt.StandardClaims
}

func (jCC *JwtCustomClaims) GetToken() string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jCC)

	// Generate encoded token and send it as response.
	stringToken, err := token.SignedString([]byte(config.JWTSecret))

	if err != nil {
		panic(err)
	}

	return stringToken
}

var JWTConfig = middleware.JWTConfig{
	Claims:     &JwtCustomClaims{},
	SigningKey: []byte(config.JWTSecret),
	ContextKey: "JWTContext",
}

func ImplementUserFromJwt(subject string) middleware.JWTSuccessHandler {
	return ImplementUserFromJWTWithConfig(&ImplementJWTUser{subject: subject})
}

func ImplementUserFromJWTWithConfig(iJU *ImplementJWTUser) middleware.JWTSuccessHandler {
	return func(ctx echo.Context) {
		JWTContext := ctx.Get("JWTContext").(*jwt.Token)
		claims := JWTContext.Claims.(*JwtCustomClaims)
		var mapUser orm.H

		if claims.Subject == iJU.subject {
			sQB := orm.GetSelectQueryBuilder(models.NewUser())

			if iJU.addInitiatedThread {
				sQB = sQB.Consider("InitiatedThreads")
			}

			if iJU.addReceivedThread {
				sQB = sQB.Consider("ReceiverThreads")
			}

			if iJU.addRoles {
				sQB = sQB.Consider("Roles")
			}

			sQB = sQB.Where(orm.And(orm.H{"IdUser": claims.UserID}))

			var err error
			mapUser, err = sQB.ApplyQueryRow()

			if err != nil {
				panic(err)
			}
		}

		ctx.Set("user", mapUser)
	}
}
