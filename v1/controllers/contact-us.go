package controllers

import (
	"github.com/rohan031/adgytec-api/helper"
	"github.com/rohan031/adgytec-api/v1/custom"
	"github.com/rohan031/adgytec-api/v1/services"
	"net/http"
)

func PostContactUs(w http.ResponseWriter, r *http.Request) {
	projectId := r.Context().Value(custom.ProjectId).(string)

	data, err := helper.DecodeJSON[map[string]interface{}](w, r, mb)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	contactUs := services.ContactUs{
		ProjectId: projectId,
	}

	err = contactUs.PostContactUs(data)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = "Successfully submitted user data"

	helper.EncodeJSON(w, http.StatusCreated, payload)
}
