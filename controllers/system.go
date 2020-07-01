package controllers

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"tupeuxcourrir_api/utils"
)

func GetUri(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, utils.JsonOkPattern(ctx.Echo().Reverse(ctx.Param("routeName"))))
}
