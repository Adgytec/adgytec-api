package controllers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rohan031/adgytec-api/helper"
	"github.com/rohan031/adgytec-api/v1/custom"
	"github.com/rohan031/adgytec-api/v1/services"
)

func PostPriseReports(w http.ResponseWriter, r *http.Request) {
	projectId := chi.URLParam(r, "projectId")
	region := chi.URLParam(r, "region")

	data, err := helper.DecodeJSON[map[string]interface{}](w, r, mb)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var contactUs services.PriseReports

	err = contactUs.PostPriseReports(projectId, region, data)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = "Successfully added prise report"

	helper.EncodeJSON(w, http.StatusCreated, payload)
}

func GetPriseReports(w http.ResponseWriter, r *http.Request) {
	projectId := chi.URLParam(r, "projectId")
	if len(projectId) == 0 {
		projectId = r.Context().Value(custom.ProjectId).(string)
	}
	region := chi.URLParam(r, "region")

	var contactUs services.PriseReports
	all, err := contactUs.GetPriseReports(projectId, region)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Data = all
	helper.EncodeJSON(w, http.StatusOK, payload)
}

func UpdatePriseReports(w http.ResponseWriter, r *http.Request) {
	reportId := chi.URLParam(r, "reportId")
	data, err := helper.DecodeJSON[map[string]interface{}](w, r, mb)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var contactUs services.PriseReports
	contactUs.Id = reportId

	err = contactUs.UpdatePriseReportById(data)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = "successfully updated the record"

	helper.EncodeJSON(w, http.StatusOK, payload)
}

func DeletePriseReports(w http.ResponseWriter, r *http.Request) {
	reportId := chi.URLParam(r, "reportId")

	var contactUs services.PriseReports
	contactUs.Id = reportId

	err := contactUs.DeletePriseReportById()
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = "successfully deleted the record"

	helper.EncodeJSON(w, http.StatusOK, payload)
}
