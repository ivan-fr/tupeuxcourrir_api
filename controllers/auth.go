package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/schema"
	"log"
	"net/http"
	"time"
	"tupeuxcourrir_api/config"
	"tupeuxcourrir_api/forms"
	TPCMiddleware "tupeuxcourrir_api/middleware"
	"tupeuxcourrir_api/models"
	"tupeuxcourrir_api/orm"
	"tupeuxcourrir_api/utils"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func SignUp(w http.ResponseWriter, r *http.Request) {
	var form forms.SignUpForm
	var user models.User

	err := r.ParseForm()

	if err != nil {
		panic(err)
	}

	decoder := schema.NewDecoder()
	err = decoder.Decode(&form, r.PostForm)

	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusBadRequest)

	if err = validator.New().Struct(&form); err != nil {
		_ = json.NewEncoder(w).Encode(utils.JsonErrorPattern(err))
		return
	}

	var hash []byte
	hash, err = bcrypt.GenerateFromPassword([]byte(form.EncryptedPassword), bcrypt.MinCost)
	if err != nil {
		_ = json.NewEncoder(w).Encode(utils.JsonErrorPattern(err))
		return
	}

	form.EncryptedPassword = string(hash)

	orm.BindForm(&user, &form)
	iQB := orm.GetInsertQueryBuilder(models.NewUser(), &user)

	if _, err = iQB.ApplyInsert(); err != nil {
		_ = json.NewEncoder(w).Encode(utils.JsonErrorPattern(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(nil)
}

func Login(w http.ResponseWriter, r *http.Request) {
	var loginForm forms.LoginForm

	err := r.ParseForm()

	if err != nil {
		panic(err)
	}

	decoder := schema.NewDecoder()
	err = decoder.Decode(&loginForm, r.PostForm)

	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusBadRequest)
	if err = validator.New().Struct(&loginForm); err != nil {
		_ = json.NewEncoder(w).Encode(utils.JsonErrorPattern(err))
		return
	}

	sQB := orm.GetSelectQueryBuilder(models.NewUser()).
		Where(orm.And(orm.H{"Email": loginForm.Email}))

	err = sQB.ApplyQueryRow()

	if err != nil {
		_ = json.NewEncoder(w).Encode(utils.JsonErrorPattern(err))
		return
	}

	user := sQB.EffectiveModel.(*models.User)

	if err = bcrypt.CompareHashAndPassword([]byte(user.EncryptedPassword),
		[]byte(loginForm.EncryptedPassword)); err != nil {
		_ = json.NewEncoder(w).Encode(utils.JsonErrorPattern(err))
		return
	}

	var expirationTime time.Time
	if loginForm.SaveConnection {
		expirationTime = time.Now().Add(1 * time.Hour)
	} else {
		expirationTime = time.Now().Add(5 * time.Hour)
	}

	claims := &TPCMiddleware.JwtUserCustomClaims{
		UserID: user.IdUser,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			Subject:   config.JwtLoginSubject,
		},
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(orm.H{"token": claims.GetToken()})
}

func ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var forgotPasswordForm forms.ForgotPasswordForm

	err := r.ParseForm()

	if err != nil {
		panic(err)
	}

	decoder := schema.NewDecoder()
	err = decoder.Decode(&forgotPasswordForm, r.PostForm)

	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusBadRequest)
	if err = validator.New().Struct(&forgotPasswordForm); err != nil {
		_ = json.NewEncoder(w).Encode(utils.JsonErrorPattern(err))
		return
	}

	sQB := orm.GetSelectQueryBuilder(models.NewUser()).
		Where(orm.And(orm.H{"Email": forgotPasswordForm.Email}))

	err = sQB.ApplyQueryRow()

	if err != nil {
		panic(err)
	}

	user := sQB.EffectiveModel.(*models.User)

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

		claims := &TPCMiddleware.JwtUserCustomClaims{
			UserID: user.IdUser,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expirationTime.Unix(),
				Subject:   config.JwtEditPasswordSubject,
			},
		}

		token := claims.GetToken()

		mailer := utils.NewMail([]string{user.Email}, "Change your password", "")
		err = mailer.ParseTemplate("htmlMail/changePassword.html",
			echo.Map{"fullName": fmt.Sprintf("%v %v", user.LastName, user.FirstName),
				"host": r.Host, "token": token})

		if err != nil {
			log.Fatal(err)
		}

		err = mailer.SendEmail()

		if err != nil {
			panic(err)
		}

		user.SentChangePasswordMailAt = sql.NullTime{Time: time.Now(), Valid: true}
		uQB := orm.GetUpdateQueryBuilder(user)
		_, errSub := uQB.ApplyUpdate()
		if errSub != nil {
			panic(errSub)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(nil)
	}

	err = errors.New("we had already sent this mail type in last 15 minutes or your email wasn't validated")
	_ = json.NewEncoder(w).Encode(utils.JsonErrorPattern(err))
}

func EditPasswordFromLink(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	concernUser := r.Context().Value("user")

	if concernUser == nil {
		_ = json.NewEncoder(w).Encode("wrong token")
		return
	}

	concernUser = concernUser.(*models.User)

	var form forms.EditPasswordForm

	err := r.ParseForm()

	if err != nil {
		panic(err)
	}

	decoder := schema.NewDecoder()
	err = decoder.Decode(&form, r.PostForm)

	if err != nil {
		panic(err)
	}

	if err = validator.New().Struct(&form); err != nil {
		_ = json.NewEncoder(w).Encode(utils.JsonErrorPattern(err))
		return
	}

	if form.EncryptedPassword != form.ConfirmPassword {
		_ = json.NewEncoder(w).Encode(utils.JsonErrorPattern(errors.New("the password aren't same")))
		return
	}

	var hash []byte
	hash, err = bcrypt.GenerateFromPassword([]byte(form.EncryptedPassword), bcrypt.MinCost)
	if err != nil {
		_ = json.NewEncoder(w).Encode(utils.JsonErrorPattern(err))
		return
	}

	form.EncryptedPassword = string(hash)

	orm.BindForm(concernUser, &form)

	uQB := orm.GetUpdateQueryBuilder(concernUser)
	_, errSub := uQB.ApplyUpdate()
	if errSub != nil {
		panic(errSub)
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(nil)
}
