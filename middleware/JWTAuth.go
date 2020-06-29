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
	AddInitiatedThread bool
	AddRoles           bool
	AddReceivedThread  bool
	GiveMeSQB          bool
	Subject            string
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
	return ImplementUserFromJWTWithConfig(&ImplementJWTUser{Subject: subject})
}

func ImplementUserFromJWTWithConfig(iJU *ImplementJWTUser) middleware.JWTSuccessHandler {
	return func(ctx echo.Context) {
		JWTContext := ctx.Get("JWTContext").(*jwt.Token)
		claims := JWTContext.Claims.(*JwtCustomClaims)

		if claims.Subject == iJU.Subject {
			user := models.NewUser()
			sQB := orm.GetSelectQueryBuilder(user)

			if iJU.AddInitiatedThread {
				sQB = sQB.Consider("InitiatedThreads")
			}

			if iJU.AddReceivedThread {
				sQB = sQB.Consider("ReceivedThreads")
			}

			if iJU.AddRoles {
				sQB = sQB.Consider("Roles")
			}

			sQB = sQB.Where(orm.And(orm.H{"IdUser": claims.UserID}))

			if iJU.GiveMeSQB {
				ctx.Set("uSQB", sQB)
			} else {
				var err error
				err = sQB.ApplyQueryRow()

				if err != nil {
					panic(err)
				}
				ctx.Set("user", user)
			}
		} else {
			if iJU.GiveMeSQB {
				ctx.Set("uSQB", nil)
			} else {
				ctx.Set("user", nil)
			}
		}
	}
}
