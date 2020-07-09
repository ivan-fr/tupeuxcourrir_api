package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/schema"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
	"tupeuxcourrir_api/config"
	"tupeuxcourrir_api/forms"
	TPCMiddleware "tupeuxcourrir_api/middleware"
	"tupeuxcourrir_api/models"
	"tupeuxcourrir_api/orm"
	"tupeuxcourrir_api/utils"

	"github.com/dgrijalva/jwt-go"
)

func GetProfile(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(user)
}

func GetThreads(w http.ResponseWriter, r *http.Request) {
	uSQB := r.Context().Value("uSQB").(*orm.SelectQueryBuilder)
	err := uSQB.ApplyQuery()

	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(uSQB.EffectiveModel.(*models.User))
}

func PutAddress(w http.ResponseWriter, r *http.Request) {
	var form forms.PutAddressForm

	user := r.Context().Value("user").(*models.User)
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

	orm.BindForm(user, &form)
	uQB := orm.GetUpdateQueryBuilder(user)

	if _, err := uQB.ApplyUpdate(); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(nil)
}

func PutPhoto(w http.ResponseWriter, r *http.Request) {
	photoFile, photoFileHeader, err := r.FormFile("photoFile")

	if err != nil {
		panic(err)
	}

	contentType := photoFileHeader.Header.Get("Content-Type")

	if contentType != "image/jpeg" && contentType != "image/png" {
		w.WriteHeader(http.StatusBadRequest)
		err = errors.New("only accept jpeg & png, current : " + contentType)
		_ = json.NewEncoder(w).Encode(utils.JsonErrorPattern(err))
		return
	}

	defer func() {
		_ = photoFile.Close()
	}()

	user := r.Context().Value("user").(*models.User)

	if user.PhotoPath.Valid {
		err = os.Remove("public/uploads/" + user.PhotoPath.String)

		if err != nil {
			panic(err)
		}
	}

	splitDot := strings.Split(photoFileHeader.Filename, ".")
	photoFileHeader.Filename = string(user.IdUser) + "." + splitDot[len(splitDot)-1]

	user.PhotoPath.String = photoFileHeader.Filename
	user.PhotoPath.Valid = true

	// Destination
	var dst *os.File
	dst, err = os.Create("public/uploads/" + photoFileHeader.Filename)

	if err != nil {
		panic(err)
	}
	defer func() {
		_ = dst.Close()
	}()

	// Copy
	if _, err = io.Copy(dst, photoFile); err != nil {
		panic(err)
	}

	uQB := orm.GetUpdateQueryBuilder(user)
	_, errSub := uQB.ApplyUpdate()
	if errSub != nil {
		panic(errSub)
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(nil)
}

func SendForValidateMail(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)

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

		newClaims := &TPCMiddleware.JwtUserCustomClaims{
			UserID: user.IdUser,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expirationTime.Unix(),
				Subject:   config.JwtCheckEmailSubject,
			},
		}

		token := newClaims.GetToken()

		mailer := utils.NewMail([]string{user.Email}, "Validate your email", "")
		err := mailer.ParseTemplate("htmlMail/checkMail.html",
			orm.H{"fullName": fmt.Sprintf("%v %v", user.LastName, user.FirstName),
				"host": r.Host, "token": token})

		if err != nil {
			panic(err)
		}

		err = mailer.SendEmail()

		if err != nil {
			panic(err)
		}

		user.SentValidateMailAt = sql.NullTime{Time: time.Now(), Valid: true}
		uQB := orm.GetUpdateQueryBuilder(user)
		_, errSub := uQB.ApplyUpdate()
		if errSub != nil {
			panic(errSub)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(nil)
	}

	err := errors.New("we had already sent this mail type in last 15 minutes or your email is already checked")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(utils.JsonErrorPattern(err))
}
