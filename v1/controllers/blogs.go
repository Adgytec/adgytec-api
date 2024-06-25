package controllers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rohan031/adgytec-api/helper"
	"github.com/rohan031/adgytec-api/v1/custom"
	"github.com/rohan031/adgytec-api/v1/services"
)

func GetUUID(w http.ResponseWriter, r *http.Request) {
	id := generateUUID()

	var payload services.JSONResponse
	payload.Error = false
	payload.Data = struct {
		UUID uuid.UUID `json:"uuid"`
	}{
		UUID: id,
	}

	helper.EncodeJSON(w, http.StatusOK, payload)
}

func PostImage(w http.ResponseWriter, r *http.Request) {
	projectId := chi.URLParam(r, "projectId")
	blogId := chi.URLParam(r, "blogId")
	maxSize := 10 << 20

	err := helper.ParseMultipartForm(w, r, maxSize)
	if err != nil {
		return
	}

	requiredFileFields := "image"

	if _, ok := r.MultipartForm.File[requiredFileFields]; !ok {
		message := fmt.Sprintf("Missing required file: %s", requiredFileFields)
		helper.HandleError(w, &custom.MalformedRequest{
			Status:  http.StatusBadRequest,
			Message: message,
		})
		return
	}

	var mediaDetails services.BlogMedia
	err = mediaDetails.UploadMedia(r, projectId, blogId)

	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Data = mediaDetails

	helper.EncodeJSON(w, http.StatusCreated, payload)
}

func DeleteMedia(w http.ResponseWriter, r *http.Request) {
	mediaDetails, err := helper.DecodeJSON[services.BlogMedia](w, r, mb)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	err = mediaDetails.DeleteMedia()
	if err != nil {
		helper.HandleError(w, err)
		return
	}
	var payload services.JSONResponse
	payload.Error = false
	payload.Message = "successfully deleted the media files"

	helper.EncodeJSON(w, http.StatusCreated, payload)

}
