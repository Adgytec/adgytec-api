package controllers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/rohan031/adgytec-api/helper"
	"github.com/rohan031/adgytec-api/v1/services"
)

func PostProject(w http.ResponseWriter, r *http.Request) {
	project, err := helper.DecodeJSON[services.Project](w, r, mb)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	err = project.CreateProject()
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = fmt.Sprintf("Successfully created new project: %s", project.ProjectName)

	helper.EncodeJSON(w, http.StatusCreated, payload)
}

func PostProjectAndServices(w http.ResponseWriter, r *http.Request) {
	projectId := chi.URLParam(r, "projectId")

	s, err := helper.DecodeJSON[services.ProjectServiceMap](w, r, mb)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	err = s.CreateProjectServiceMap(projectId)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = fmt.Sprintf("Successfuly added services to project-id: %s", projectId)

	helper.EncodeJSON(w, http.StatusCreated, payload)
}

func PostProjectAndUser(w http.ResponseWriter, r *http.Request) {
	projectId := chi.URLParam(r, "projectId")

	user, err := helper.DecodeJSON[services.ProjectUserMap](w, r, mb)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	err = user.CreateUserProjectMap(projectId)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = fmt.Sprintf("Successfuly added user-id: %v to project-id: %s", user.UserId, projectId)

	helper.EncodeJSON(w, http.StatusCreated, payload)
}
