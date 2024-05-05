package controllers

import (
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
	payload, err := helper.DecodeJSON[services.Newsletter](w, r, 1048576)

	if err != nil {
		helper.HandleError(w, err)
		return
	}

	fmt.Println(payload.Email)
}
