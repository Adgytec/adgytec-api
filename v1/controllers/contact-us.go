package controllers

import (
	"github.com/go-chi/chi/v5"
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

	var contactUs services.ContactUs

	err = contactUs.PostContactUs(projectId, data)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = "Successfully submitted user data"

	helper.EncodeJSON(w, http.StatusCreated, payload)
}

func GetContactUs(w http.ResponseWriter, r *http.Request) {
	projectId := chi.URLParam(r, "projectId")
	cursor := r.URL.Query().Get("cursor")

	if len(cursor) == 0 {
		cursor = getNow()
	}
	var contactUs services.ContactUs
	all, err := contactUs.GetContactUs(projectId, cursor)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Data = all

	helper.EncodeJSON(w, http.StatusOK, payload)
}

func DeleteContactUsItem(w http.ResponseWriter, r *http.Request) {
	contactId := chi.URLParam(r, "contactId")

	var contactUs services.ContactUs
	contactUs.Id = contactId

	err := contactUs.DeleteContactUsById()
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = "successfully deleted the record"

	helper.EncodeJSON(w, http.StatusOK, payload)
}
