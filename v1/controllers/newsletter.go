package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/rohan031/adgytec-api/helper"
	"github.com/rohan031/adgytec-api/v1/services"
)

func GetNewslettersEmail(w http.ResponseWriter, r *http.Request) {
	var payload services.JSONResponse

	payload.Error = false
	payload.Message = "no records yet"

	helper.EncodeJSON(w, http.StatusOK, payload)
}

func PostNewsletterEmail(w http.ResponseWriter, r *http.Request) {
	var payload services.Newsletter

	if err := helper.DecodeJSON(w, r, 1048576, &payload); err != nil {
		err = errors.New("500 internal server error")
		helper.ErrorResponse(w, err, http.StatusInternalServerError)
	}

	fmt.Println(payload.Email)
}
