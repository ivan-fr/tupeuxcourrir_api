package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/labstack/gommon/random"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
	"tupeuxcourrir_api/config"
	TCPMiddleware "tupeuxcourrir_api/middleware"
	"tupeuxcourrir_api/models"
	"tupeuxcourrir_api/orm"
	"tupeuxcourrir_api/utils"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
)

func PutPhoto(ctx echo.Context) error {
	photoFile, err := ctx.FormFile("photoFile")

	if err != nil {
		return err
	}

	contentType := photoFile.Header.Get("Content-Type")

	if contentType != "image/jpeg" && contentType != "image/png" {
		err = errors.New("only accept jpeg & png, current : " + contentType)
		return ctx.JSON(http.StatusBadRequest, utils.JsonErrorPattern(err))
	}

	var src multipart.File
	src, err = photoFile.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	mapUser := ctx.Get("user").(orm.H)
	user := mapUser["User"].(*models.User)

	if mapUser == nil {
		return errors.New("wrong jwt subject")
	}

	user.PhotoPath.String = random.String(5) + string(user.IdUser)
	user.PhotoPath.Valid = true

	photoFile.Filename = user.PhotoPath.String

	// Destination
	var dst *os.File
	dst, err = os.Create(photoFile.Filename)

	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	uQB := orm.GetUpdateQueryBuilder(user)
	_, errSub := uQB.ApplyUpdate()
	if errSub != nil {
		return errSub
	}

	return ctx.JSON(http.StatusOK, echo.Map{})
}

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

	err = errors.New("we had already sent this mail type in last 15 minutes or your email is already checked")
	return ctx.JSON(http.StatusBadRequest, utils.JsonErrorPattern(err))
}
