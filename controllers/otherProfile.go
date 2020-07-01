package controllers

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"tupeuxcourrir_api/models"
	"tupeuxcourrir_api/orm"
)

func GetOtherProfile(ctx echo.Context) error {
	idTarget, err := strconv.Atoi(ctx.Param("id"))

	if err != nil {
		return err
	}

	sQB := orm.GetSelectQueryBuilder(models.NewUser()).
		Select([]string{"CreatedAt", "Pseudo", "PhotoPath"}).
		Where(orm.And(orm.H{"IdUser": idTarget}))
	err = sQB.ApplyQueryRow()

	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, sQB.EffectiveModel.(*models.User))
}

func MakeThreadWithOtherProfile(ctx echo.Context) error {
	user := ctx.Get("user").(*models.User)

	return ctx.JSON(http.StatusOK, user)
}
