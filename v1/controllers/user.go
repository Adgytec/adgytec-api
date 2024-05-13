package controllers

import (
	"net/http"

	"github.com/rohan031/adgytec-api/helper"
	"github.com/rohan031/adgytec-api/v1/services"
)

// creating new user
func PostUser(w http.ResponseWriter, r *http.Request) {
	data, err := helper.DecodeJSON[services.User](w, r, mb)

	if err != nil {
		helper.HandleError(w, err)
		return
	}

	if !data.ValidateInput() {
		err = &helper.MalformedRequest{Status: http.StatusBadRequest, Message: "Invalid input values"}
		helper.HandleError(w, err)
		return
	}

	password, err := data.CreateUser()
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var userDetails services.UserCreationDetails
	userDetails.Email = data.Email
	userDetails.Name = data.Name
	userDetails.Password = password

	templatePath := "./assets/template.html"
	to := []string{
		data.Email,
	}

	err = services.SendEmail(userDetails, templatePath, to)
	var payload services.JSONResponse

	payload.Error = false
	payload.Message = "Successfully created user account"

	status := http.StatusCreated
	if err != nil {
		payload.Message = "User account created. But can't send user credentials"
		status = http.StatusPartialContent
	}
	helper.EncodeJSON(w, status, payload)
}
