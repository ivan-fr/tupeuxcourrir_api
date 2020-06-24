package controllers

import (
	"errors"
	"net/http"
	"tupeuxcourrir_api/config"
	TCPMiddleware "tupeuxcourrir_api/middleware"
	"tupeuxcourrir_api/models"
	"tupeuxcourrir_api/orm"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
)

func CheckMail(ctx echo.Context) error {
	JWTContext := ctx.Get("JWTContext").(*jwt.Token)
	claims := JWTContext.Claims.(*TCPMiddleware.JwtCustomClaims)

	if claims.Subject != config.JwtCheckEmailSubject {
		return errors.New("wrong jwt subject")
	}

	sQB := orm.GetSelectQueryBuilder(models.NewUser()).
		Where(orm.And(orm.H{"IdUser": claims.UserID}))

	mapUser, err := sQB.ApplyQueryRow()

	if err != nil {
		return err
	}

	concernUser := mapUser["User"].(*models.User)

	if concernUser.CheckedEmail {
		return ctx.JSON(http.StatusUnauthorized, echo.Map{})
	}

	concernUser.CheckedEmail = true

	uQB := orm.GetUpdateQueryBuilder(concernUser)
	_, errSub := uQB.ApplyUpdate()
	if errSub != nil {
		return errSub
	}

	return ctx.JSON(http.StatusOK, echo.Map{})
}
