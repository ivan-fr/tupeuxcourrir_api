package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"
	"tupeuxcourrir_api/config"
	TCPMiddleware "tupeuxcourrir_api/middleware"
	"tupeuxcourrir_api/models"
	"tupeuxcourrir_api/orm"
	"tupeuxcourrir_api/utils"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
)

func SendForValidateMail(ctx echo.Context) error {
	mapUser := ctx.Get("user").(orm.H)
	var err error

	if mapUser == nil {
		return errors.New("wrong jwt subject")
	}

	user := mapUser["User"].(*models.User)

	var execute = true

	switch {
	case user.CheckedEmail == true:
		execute = false
	case user.SentValidateMailAt.Valid:
		val, _ := user.SentValidateMailAt.Value()
		predictionTime := val.(time.Time).Add(15 * time.Minute)
		nowTime := time.Now()

		execute = predictionTime.Before(nowTime)
	}

	if execute {
		expirationTime := time.Now().Add(15 * time.Minute)

		newClaims := &TCPMiddleware.JwtCustomClaims{
			UserID: user.IdUser,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expirationTime.Unix(),
				Subject:   config.JwtCheckEmailSubject,
			},
		}

		token := newClaims.GetToken()

		mailer := utils.NewMail([]string{user.Email}, "Validate your email", "")
		err = mailer.ParseTemplate("htmlMail/checkMail.html",
			echo.Map{"fullName": fmt.Sprintf("%v %v", user.LastName, user.FirstName.String),
				"host": ctx.Request().Host, "token": token})

		if err != nil {
			return err
		}

		err = mailer.SendEmail()

		if err != nil {
			return err
		}

		user.SentValidateMailAt = sql.NullTime{Time: time.Now(), Valid: true}
		uQB := orm.GetUpdateQueryBuilder(user)
		_, errSub := uQB.ApplyUpdate()
		if errSub != nil {
			return errSub
		}

		return ctx.JSON(http.StatusOK, echo.Map{})
	}

	err = errors.New("we had already sent this mail type in last 15 minutes")
	return ctx.JSON(http.StatusBadRequest, utils.JsonErrorPattern(err))
}

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
