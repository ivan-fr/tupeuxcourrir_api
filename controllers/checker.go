package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"tupeuxcourrir_api/config"
	TPCMiddleware "tupeuxcourrir_api/middleware"
	"tupeuxcourrir_api/models"
	"tupeuxcourrir_api/orm"
	"tupeuxcourrir_api/utils"

	"github.com/dgrijalva/jwt-go"
)

func CheckMail(w http.ResponseWriter, r *http.Request) {
	JWTContext := r.Context().Value("JWTContext").(*jwt.Token)
	claims := JWTContext.Claims.(*TPCMiddleware.JwtUserCustomClaims)

	if claims.Subject != config.JwtCheckEmailSubject {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(utils.JsonErrorPattern(errors.New("wrong jwt subject")))
		return
	}

	sQB := orm.GetSelectQueryBuilder(models.NewUser()).
		Where(orm.And(orm.H{"IdUser": claims.UserID}))

	err := sQB.ApplyQueryRow()

	if err != nil {
		panic(err)
	}

	concernUser := sQB.EffectiveModel.(*models.User)

	if concernUser.CheckedEmail {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(nil)
		return
	}

	concernUser.CheckedEmail = true

	uQB := orm.GetUpdateQueryBuilder(concernUser)
	_, errSub := uQB.ApplyUpdate()
	if errSub != nil {
		panic(errSub)
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(nil)
}
