package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"tupeuxcourrir_api/models"
	"tupeuxcourrir_api/orm"
	"tupeuxcourrir_api/utils"
)

func GetOtherProfile(w http.ResponseWriter, r *http.Request) {
	idTarget, err := strconv.Atoi(r.URL.Query().Get("id"))

	if err != nil {
		panic(err)
	}

	sQB := orm.GetSelectQueryBuilder(models.NewUser()).
		Select([]string{"CreatedAt", "Pseudo", "PhotoPath", "IdUser"}).
		Where(orm.And(orm.H{"IdUser": idTarget}))
	err = sQB.ApplyQueryRow()

	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(sQB.EffectiveModel.(*models.User))
}

func MakeThreadWithOtherProfile(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*models.User)

	idTarget, err := strconv.Atoi(r.URL.Query().Get("id"))

	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusUnauthorized)

	if idTarget == user.IdUser {
		_ = json.NewEncoder(w).Encode(
			utils.JsonErrorPattern(
				errors.New("the receiver of your thread must have different ID from you")))
		return
	}

	sQB := orm.GetSelectQueryBuilder(models.NewUser()).
		Where(orm.And(orm.H{"IdUser": idTarget}))
	err = sQB.ApplyQueryRow()

	if err != nil {
		panic(err)
	}

	targetUser := sQB.EffectiveModel.(*models.User)

	aNewThread := models.NewThread()
	aNewThread.InitiatorThreadIdUser = user.IdUser
	aNewThread.ReceiverThreadIdUser = targetUser.IdUser

	iQB := orm.GetInsertQueryBuilder(models.NewThread(), aNewThread)
	_, err = iQB.ApplyInsert()

	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(nil)
}
