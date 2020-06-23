package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
	"tupeuxcourrir_api/forms"
	"tupeuxcourrir_api/models"
	"tupeuxcourrir_api/orm"
	"tupeuxcourrir_api/utils"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

const jwtEditPasswordSubject = "forgotPassword"
const jwtLoginSubject = "login"

func SignUp(ctx echo.Context) error {
	var form forms.SignUpForm
	var user models.User

	if err := ctx.Bind(&form); err != nil {
		return err
	}

	if err := ctx.Validate(&form); err != nil {
		return ctx.JSON(http.StatusBadRequest, utils.JsonErrorPattern(err))
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(form.EncryptedPassword), bcrypt.MinCost)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, utils.JsonErrorPattern(err))
	}

	form.EncryptedPassword = string(hash)

	orm.BindForm(&user, &form)
	iQB := orm.GetInsertQueryBuilder(models.NewUser(), &user)

	if _, err := iQB.ApplyInsert(); err != nil {
		return ctx.JSON(http.StatusBadRequest, utils.JsonErrorPattern(err))
	}

	return ctx.JSON(http.StatusOK, orm.H{})
}

func Login(ctx echo.Context) error {
	var loginForm forms.LoginForm

	if err := ctx.Bind(&loginForm); err != nil {
		return err
	}

	if err := ctx.Validate(&loginForm); err != nil {
		return ctx.JSON(http.StatusBadRequest, utils.JsonErrorPattern(err))
	}

	sQB := orm.GetSelectQueryBuilder(models.NewUser()).
		Where(orm.And(orm.H{"Email": loginForm.Email}))

	mapUser, err := sQB.ApplyQueryRow()

	if err != nil {
		return ctx.JSON(http.StatusBadRequest, utils.JsonErrorPattern(err))
	}

	user := mapUser["User"].(*models.User)

	if err = bcrypt.CompareHashAndPassword([]byte(user.EncryptedPassword),
		[]byte(loginForm.EncryptedPassword)); err != nil {
		return ctx.JSON(http.StatusBadRequest, utils.JsonErrorPattern(err))
	}

	var expirationTime time.Time
	if loginForm.SaveConnection {
		expirationTime = time.Now().Add(1 * time.Hour)
	} else {
		expirationTime = time.Now().Add(5 * time.Hour)
	}

	claims := &utils.JwtCustomClaims{
		UserID: user.IdUser,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			Subject:   jwtLoginSubject,
		},
	}

	return ctx.JSON(http.StatusOK, orm.H{"token": claims.GetToken()})
}

func ForgotPassword(ctx echo.Context) error {
	var forgotPasswordForm forms.ForgotPasswordForm

	if err := ctx.Bind(&forgotPasswordForm); err != nil {
		return err
	}

	if err := ctx.Validate(&forgotPasswordForm); err != nil {
		return ctx.JSON(http.StatusBadRequest, utils.JsonErrorPattern(err))
	}

	sQB := orm.GetSelectQueryBuilder(models.NewUser()).
		Where(orm.And(orm.H{"Email": forgotPasswordForm.Email}))

	mapUser, err := sQB.ApplyQueryRow()

	if err != nil {
		return ctx.JSON(http.StatusBadRequest, utils.JsonErrorPattern(err))
	}

	user := mapUser["User"].(*models.User)

	var execute = true

	switch {
	case user.CheckedEmail == false:
		execute = false
	case user.SentChangePasswordMailAt.Valid:
		val, _ := user.SentChangePasswordMailAt.Value()
		predictionTime := val.(time.Time).Add(15 * time.Minute)
		nowTime := time.Now()

		execute = predictionTime.Before(nowTime)
	}

	if execute {
		expirationTime := time.Now().Add(15 * time.Minute)

		claims := &utils.JwtCustomClaims{
			UserID: user.IdUser,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expirationTime.Unix(),
				Subject:   jwtEditPasswordSubject,
			},
		}

		token := claims.GetToken()

		mailer := utils.NewMail([]string{user.Email}, "Change your password", "")
		err = mailer.ParseTemplate("htmlMail/changePassword.html",
			echo.Map{"fullName": fmt.Sprintf("%v %v", user.LastName, user.FirstName.String),
				"host": ctx.Request().Host, "token": token})

		if err != nil {
			log.Fatal(err)
		}

		err = mailer.SendEmail()

		if err != nil {
			return err
		}

		user.SentChangePasswordMailAt = sql.NullTime{Time: time.Now(), Valid: true}
		uQB := orm.GetUpdateQueryBuilder(user)
		_, errSub := uQB.ApplyUpdate()
		if errSub != nil {
			return errSub
		}

		return ctx.JSON(http.StatusOK, echo.Map{})
	}

	err = errors.New("we had already sent this mail type in last 15 minutes or your email wasn't validated")
	return ctx.JSON(http.StatusBadRequest, utils.JsonErrorPattern(err))
}

func EditPasswordFromLink(ctx echo.Context) error {
	JWTContext := ctx.Get("JWTContext").(*jwt.Token)
	claims := JWTContext.Claims.(*utils.JwtCustomClaims)

	if claims.Subject != jwtEditPasswordSubject {
		return errors.New("wrong jwt subject")
	}

	var form forms.EditPasswordForm

	if err := ctx.Bind(&form); err != nil {
		return err
	}

	if err := ctx.Validate(&form); err != nil {
		return ctx.JSON(http.StatusBadRequest, utils.JsonErrorPattern(err))
	}

	if form.EncryptedPassword != form.ConfirmPassword {
		return errors.New("the password aren't same")
	}

	sQB := orm.GetSelectQueryBuilder(models.NewUser()).
		Where(orm.And(orm.H{"IdUser": claims.UserID}))

	mapUser, err := sQB.ApplyQueryRow()

	if err != nil {
		return err
	}

	concernUser := mapUser["User"].(*models.User)

	hash, err := bcrypt.GenerateFromPassword([]byte(form.EncryptedPassword), bcrypt.MinCost)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, utils.JsonErrorPattern(err))
	}

	form.EncryptedPassword = string(hash)

	orm.BindForm(concernUser, &form)

	uQB := orm.GetUpdateQueryBuilder(concernUser)
	_, errSub := uQB.ApplyUpdate()
	if errSub != nil {
		return errSub
	}

	return ctx.JSON(http.StatusOK, orm.H{})
}
