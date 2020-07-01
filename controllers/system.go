package controllers

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
	"tupeuxcourrir_api/utils"
)

func GetUri(ctx echo.Context) error {
	params := strings.Split(ctx.QueryParam("params"), ",")
	paramsInterface := make([]interface{}, len(params))

	for i, v := range params {
		paramsInterface[i] = v
	}

	return ctx.JSON(http.StatusOK, utils.JsonOkPattern(ctx.Echo().Reverse(ctx.QueryParam("routeName"), paramsInterface...)))
}
