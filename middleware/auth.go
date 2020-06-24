package middleware

import (
	"tupeuxcourrir_api/config"
	"tupeuxcourrir_api/models"
	"tupeuxcourrir_api/orm"
	"tupeuxcourrir_api/utils"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type ImplementJWTUser struct {
	addInitiatedThread bool
	addRoles           bool
	addReceivedThread  bool
}

func ImplementUserFromJWTWithConfig(configIJWTU *ImplementJWTUser) middleware.JWTSuccessHandler {
	return func(ctx echo.Context) {
		JWTContext := ctx.Get("JWTContext").(*jwt.Token)
		claims := JWTContext.Claims.(*utils.JwtCustomClaims)

		if claims.Subject == config.JwtLoginSubject {
			sQB := orm.GetSelectQueryBuilder(models.NewUser())

			if configIJWTU.addInitiatedThread {
				sQB = sQB.Consider("InitiatedThreads")
			}

			if configIJWTU.addReceivedThread {
				sQB = sQB.Consider("ReceiverThreads")
			}

			if configIJWTU.addRoles {
				sQB = sQB.Consider("Roles")
			}

			sQB = sQB.Where(orm.And(orm.H{"IdUser": claims.UserID}))

			mapUser, err := sQB.ApplyQueryRow()

			if err != nil {
				panic(err)
			}

			ctx.Set("user", mapUser)
		}
	}
}
