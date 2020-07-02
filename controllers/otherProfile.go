package controllers

import (
	"errors"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"tupeuxcourrir_api/models"
	"tupeuxcourrir_api/orm"
	"tupeuxcourrir_api/utils"
)

func GetOtherProfile(ctx echo.Context) error {
	idTarget, err := strconv.Atoi(ctx.Param("id"))

	if err != nil {
		return err
	}

	sQB := orm.GetSelectQueryBuilder(models.NewUser()).
		Select([]string{"CreatedAt", "Pseudo", "PhotoPath", "IdUser"}).
		Where(orm.And(orm.H{"IdUser": idTarget}))
	err = sQB.ApplyQueryRow()

	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, sQB.EffectiveModel.(*models.User))
}

func MakeThreadWithOtherProfile(ctx echo.Context) error {
	user := ctx.Get("user").(*models.User)

	idTarget, err := strconv.Atoi(ctx.Param("id"))

	if err != nil {
		return err
	}

	if idTarget == user.IdUser {
		err = errors.New("the receiver of your thread must have different ID from you")
		return ctx.JSON(http.StatusUnauthorized, utils.JsonErrorPattern(err))
	}

	sQB := orm.GetSelectQueryBuilder(models.NewUser()).
		Where(orm.And(orm.H{"IdUser": idTarget}))
	err = sQB.ApplyQueryRow()

	if err != nil {
		return err
	}

	targetUser := sQB.EffectiveModel.(*models.User)

	aNewThread := models.NewThread()
	aNewThread.InitiatorThreadIdUser = user.IdUser
	aNewThread.ReceiverThreadIdUser = targetUser.IdUser

	iQB := orm.GetInsertQueryBuilder(models.NewThread(), aNewThread)
	_, err = iQB.ApplyInsert()

	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, echo.Map{})
}
