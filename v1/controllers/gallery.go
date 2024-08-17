package controllers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rohan031/adgytec-api/helper"
	"github.com/rohan031/adgytec-api/v1/services"
)

func GetAlbumsByProjectId(w http.ResponseWriter, r *http.Request) {
	projectId := chi.URLParam(r, "projectId")

	var albums services.Album
	all, err := albums.GetAlbumsByProjectId(projectId)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Data = all

	helper.EncodeJSON(w, http.StatusOK, payload)
}
