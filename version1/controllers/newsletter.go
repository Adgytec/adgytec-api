package controllers

import (
	"net/http"

	"github.com/rohan031/adgytec-api/version1/helper"
	"github.com/rohan031/adgytec-api/version1/services"
)

func GetNewslettersEmail(w http.ResponseWriter, r *http.Request) {
	var payload services.JSONResponse

	payload.Error = false
	payload.Message = "no records yet"

	helper.EncodeJSON(w, http.StatusOK, payload)
}
