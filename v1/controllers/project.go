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

	clientToken, err := project.CreateProject()
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = fmt.Sprintf("Successfully created new project: %s", project.ProjectName)
	payload.Data = &struct {
		ClientToken string `json:"clientToken"`
	}{
		ClientToken: clientToken,
	}

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

func GetAllProjects(w http.ResponseWriter, r *http.Request) {
	var projects services.Project

	all, err := projects.GetAllProjects()
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Data = all

	helper.EncodeJSON(w, http.StatusOK, payload)

}

func GetProjectById(w http.ResponseWriter, r *http.Request) {
	projectId := chi.URLParam(r, "projectId")
	var project services.Project
	project.Id = projectId

	p, err := project.GetProjectById()
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Data = p

	helper.EncodeJSON(w, http.StatusOK, payload)

}

func DeleteProjectById(w http.ResponseWriter, r *http.Request) {
	projectId := chi.URLParam(r, "projectId")
	var project services.Project
	project.Id = projectId

	err := project.DeleteProjectById()
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = fmt.Sprintf("Successfully deleted project with id: %v", projectId)

	helper.EncodeJSON(w, http.StatusOK, payload)
}

func DeleteProjectAndUser(w http.ResponseWriter, r *http.Request) {
	projectId := chi.URLParam(r, "projectId")

	user, err := helper.DecodeJSON[services.ProjectUserMap](w, r, mb)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	err = user.DeleteUserProjectMap(projectId)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = fmt.Sprintf("Successfuly removed user-id: %v from project-id: %s", user.UserId, projectId)

	helper.EncodeJSON(w, http.StatusOK, payload)

}

func DeleteProjectAndService(w http.ResponseWriter, r *http.Request) {
	projectId := chi.URLParam(r, "projectId")

	service, err := helper.DecodeJSON[services.ProjectServiceMap](w, r, mb)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	err = service.DeleteProjectServiceMap(projectId)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = fmt.Sprintf("Successfuly removed service-id: %v from project-id: %s", service.Services[0], projectId)

	helper.EncodeJSON(w, http.StatusOK, payload)
}

func GetAllServices(w http.ResponseWriter, r *http.Request) {
	var p services.Project

	all, err := p.GetAllServices()
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Data = all

	helper.EncodeJSON(w, http.StatusOK, payload)
}
