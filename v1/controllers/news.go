package controllers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rohan031/adgytec-api/helper"
	"github.com/rohan031/adgytec-api/v1/custom"
	"github.com/rohan031/adgytec-api/v1/services"
)

func PostNews(w http.ResponseWriter, r *http.Request) {
	projectId := chi.URLParam(r, "projectId")
	maxSize := 10 << 20
	err := helper.ParseMultipartForm(w, r, maxSize)
	if err != nil {
		return
	}

	requiredFields := []string{"title", "text", "link"}
	requiredFileFields := "image"

	for _, field := range requiredFields {
		if _, ok := r.MultipartForm.Value[field]; !ok {
			message := fmt.Sprintf("Missing required field: %s", field)
			helper.HandleError(w, &custom.MalformedRequest{
				Status:  http.StatusBadRequest,
				Message: message,
			})
			return
		}
	}

	if _, ok := r.MultipartForm.File[requiredFileFields]; !ok {
		message := fmt.Sprintf("Missing required file: %s", requiredFileFields)
		helper.HandleError(w, &custom.MalformedRequest{
			Status:  http.StatusBadRequest,
			Message: message,
		})
		return
	}

	title := r.FormValue("title")
	text := r.FormValue("text")
	link := r.FormValue("link")
	newsDetails := &services.News{
		Title: title,
		Text:  text,
		Link:  link,
	}

	err = newsDetails.CreateNewsItem(r, projectId)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = "Successfully created news item."

	helper.EncodeJSON(w, http.StatusCreated, payload)
}

func GetAllNewsClient(w http.ResponseWriter, r *http.Request) {
	projectId := r.Context().Value(custom.ProjectId).(string)

	var news services.News
	all, err := news.GetAllNewsByProjectId(projectId, 5)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Data = all

	helper.EncodeJSON(w, http.StatusOK, payload)
}

func GetNews(w http.ResponseWriter, r *http.Request) {
	projectId := chi.URLParam(r, "projectId")

	var news services.News
	all, err := news.GetAllNewsByProjectId(projectId, 100)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Data = all

	helper.EncodeJSON(w, http.StatusOK, payload)
}

func DeleteNews(w http.ResponseWriter, r *http.Request) {
	serviceId := chi.URLParam(r, "serviceId")

	var news services.News
	news.Id = serviceId

	err := news.DeleteNews()
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = "Successfully delete news item"

	helper.EncodeJSON(w, http.StatusOK, payload)
}
