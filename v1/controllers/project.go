package controllers

import (
	"fmt"
	"net/http"

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
