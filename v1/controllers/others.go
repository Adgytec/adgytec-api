package controllers

import (
	"net/http"

	"github.com/rohan031/adgytec-api/helper"
	"github.com/rohan031/adgytec-api/v1/services"
)

func PostContactUsJDK(w http.ResponseWriter, r *http.Request) {
	details, err := helper.DecodeJSON[services.JDKContact](w, r, mb)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	err = details.PostContactUsJDK()
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = "Sucessfully submitted form data."

	helper.EncodeJSON(w, http.StatusOK, payload)
}
