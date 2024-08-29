package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rohan031/adgytec-api/helper"
	"github.com/rohan031/adgytec-api/v1/custom"
	"github.com/rohan031/adgytec-api/v1/services"
)

func GetAlbumsByProjectId(w http.ResponseWriter, r *http.Request) {
	projectId := chi.URLParam(r, "projectId")
	cursor := r.URL.Query().Get("cursor")

	if len(cursor) == 0 {
		today := time.Now()
		cursor = today.Format(time.RFC3339)
	}

	var albums services.Album
	all, err := albums.GetAlbumsByProjectId(projectId, cursor)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Data = all

	helper.EncodeJSON(w, http.StatusOK, payload)
}

func GetAlbumsByProjectIdClient(w http.ResponseWriter, r *http.Request) {
	projectId := r.Context().Value(custom.ProjectId).(string)
	cursor := r.URL.Query().Get("cursor")

	if len(cursor) == 0 {
		today := time.Now()
		cursor = today.Format(time.RFC3339)
	}

	var albums services.Album
	all, err := albums.GetAlbumsByProjectId(projectId, cursor)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Data = all

	helper.EncodeJSON(w, http.StatusOK, payload)
}

func PostAlbum(w http.ResponseWriter, r *http.Request) {
	maxSize := 10 << 20 // 20mb
	err := helper.ParseMultipartForm(w, r, maxSize)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	projectId := chi.URLParam(r, "projectId")
	userId := r.Context().Value(custom.UserID).(string)

	requiredFields := []string{"name"}
	requiredFileFields := "cover"

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

	name := r.FormValue("name")
	var albumItem services.Album
	albumItem.Name = name

	err = albumItem.CreateAlbum(r, projectId, userId)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = "Successfully created the album"

	helper.EncodeJSON(w, http.StatusCreated, payload)
}

func PatchAlbumMetadataById(w http.ResponseWriter, r *http.Request) {
	albumId := chi.URLParam(r, "albumId")

	albumDetails, err := helper.DecodeJSON[services.Album](w, r, mb)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	albumDetails.Id = albumId
	err = albumDetails.PatchAlbumMetadataById()
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = "successfully updated album data"

	helper.EncodeJSON(w, http.StatusOK, payload)
}

func PatchAlbumCoverById(w http.ResponseWriter, r *http.Request) {
	maxSize := 10 << 20 // 10mb
	err := helper.ParseMultipartForm(w, r, maxSize)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	requiredFileFields := "cover"
	if _, ok := r.MultipartForm.File[requiredFileFields]; !ok {
		message := fmt.Sprintf("Missing required file: %s", requiredFileFields)
		helper.HandleError(w, &custom.MalformedRequest{
			Status:  http.StatusBadRequest,
			Message: message,
		})
		return
	}

	projectId := chi.URLParam(r, "projectId")
	albumId := chi.URLParam(r, "albumId")

	var album services.Album

	album.Id = albumId
	err = album.PatchAlbumCoverById(r, projectId)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = "successfully updated album cover image"

	helper.EncodeJSON(w, http.StatusOK, payload)

}

func DeleteAlbumById(w http.ResponseWriter, r *http.Request) {
	albumId := chi.URLParam(r, "albumId")
	projectId := chi.URLParam(r, "projectId")

	var album services.Album
	album.Id = albumId

	err := album.DeleteAlbumById(projectId)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = "successfully deleted the album"

	helper.EncodeJSON(w, http.StatusOK, payload)
}
