package controllers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rohan031/adgytec-api/helper"
	"github.com/rohan031/adgytec-api/v1/services"
)

func PostCategoryByProjectId(w http.ResponseWriter, r *http.Request) {
	projectId := chi.URLParam(r, "projectId")

	category, err := helper.DecodeJSON[services.Category](w, r, mb)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	err = category.PostCategoryByProjectId(projectId)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = "Successfully created new category."

	helper.EncodeJSON(w, http.StatusCreated, payload)
}

func PatchCategoryById(w http.ResponseWriter, r *http.Request) {
	categoryId := chi.URLParam(r, "categoryId")

	category, err := helper.DecodeJSON[services.Category](w, r, mb)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	err = category.PatchCategoryById(categoryId)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = "Successfully updated the category."

	helper.EncodeJSON(w, http.StatusCreated, payload)
}

func GetCategoryByProjectId(w http.ResponseWriter, r *http.Request) {
	projectId := chi.URLParam(r, "projectId")

	var category services.Category

	categories, err := category.GetCategoryByProjectId(projectId)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Data = categories

	helper.EncodeJSON(w, http.StatusOK, payload)
}

func DeleteCategoryById(w http.ResponseWriter, r *http.Request) {
	categoryId := chi.URLParam(r, "categoryId")
	var category services.Category

	err := category.DeleteCategoryById(categoryId)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = "Successfully deleted the category."

	helper.EncodeJSON(w, http.StatusOK, payload)
}
