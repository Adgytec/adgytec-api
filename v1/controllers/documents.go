package controllers

import (
	"github.com/go-chi/chi/v5"
	"github.com/rohan031/adgytec-api/helper"
	"github.com/rohan031/adgytec-api/v1/custom"
	"github.com/rohan031/adgytec-api/v1/services"
	"net/http"
)

func GetDocumentCoverByProjectId(w http.ResponseWriter, r *http.Request) {
	projectId := chi.URLParam(r, "projectId")
	cursor := r.URL.Query().Get("cursor")

	if len(cursor) == 0 {
		cursor = getNow()
	}

	var documentCover services.DocumentCover
	all, err := documentCover.GetDocumentCoverByProjectId(projectId, cursor)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Data = all

	helper.EncodeJSON(w, http.StatusOK, payload)
}

func GetDocumentCoverByProjectIdClient(w http.ResponseWriter, r *http.Request) {
	projectId := r.Context().Value(custom.ProjectId).(string)
	cursor := r.URL.Query().Get("cursor")

	if len(cursor) == 0 {
		cursor = getNow()
	}

	var documentCover services.DocumentCover
	all, err := documentCover.GetDocumentCoverByProjectId(projectId, cursor)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Data = all

	helper.EncodeJSON(w, http.StatusOK, payload)
}

func PostDocumentCover(w http.ResponseWriter, r *http.Request) {
	projectId := chi.URLParam(r, "projectId")
	userId := r.Context().Value(custom.UserID).(string)

	coverDetails, err := helper.DecodeJSON[services.DocumentCover](w, r, mb)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	err = coverDetails.PostDocumentCoverByProjectId(projectId, userId)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = "successfully created document cover"

	helper.EncodeJSON(w, http.StatusOK, payload)
}

func PatchDocumentCoverById(w http.ResponseWriter, r *http.Request) {
	coverId := chi.URLParam(r, "cover")

	coverDetails, err := helper.DecodeJSON[services.DocumentCover](w, r, mb)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	coverDetails.Id = coverId
	err = coverDetails.PatchDocumentCoverById()
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = "successfully updated document cover"

	helper.EncodeJSON(w, http.StatusOK, payload)
}

func DeleteDocumentCoverById(w http.ResponseWriter, r *http.Request) {
	coverId := chi.URLParam(r, "cover")
	projectId := chi.URLParam(r, "projectId")

	var documentCover services.DocumentCover
	documentCover.Id = coverId

	err := documentCover.DeleteDocumentCoverById(projectId)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = "successfully deleted the document cover"

	helper.EncodeJSON(w, http.StatusOK, payload)
}
