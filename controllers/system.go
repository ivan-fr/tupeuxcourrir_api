package controllers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
	"tupeuxcourrir_api/utils"
)

func GetUriBox(mainRouter *mux.Router) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		params := strings.Split(r.URL.Query().Get("params"), ",")

		url, _ := mainRouter.Get(r.URL.Query().Get("routeNames")).URL(params...)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(
			utils.JsonOkPattern(url.String()))
	}
}
