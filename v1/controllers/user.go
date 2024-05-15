package controllers

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rohan031/adgytec-api/helper"
	"github.com/rohan031/adgytec-api/v1/custom"
	"github.com/rohan031/adgytec-api/v1/services"
	"github.com/rohan031/adgytec-api/v1/validation"
)

// creating new user
func PostUser(w http.ResponseWriter, r *http.Request) {
	myRole := r.Context().Value(custom.UserRole).(string)

	// decoding request body
	data, err := helper.DecodeJSON[services.User](w, r, mb)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	// validating request body parameters
	if !data.ValidateInput() {
		message := "The request body contains invalid input values."
		err = &custom.MalformedRequest{Status: http.StatusBadRequest, Message: message}
		helper.HandleError(w, err)
		return
	}

	// validate required permission
	if !validation.AuthorizeRole(myRole, data.Role) {
		message := "Insufficient privileges to create a user account with the specified role."
		err = &custom.MalformedRequest{Status: http.StatusForbidden, Message: message}
		helper.HandleError(w, err)
		return
	}

	// creating user and generating password
	password, err := data.CreateUser()
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var userDetails services.UserCreationDetails
	userDetails.Email = data.Email
	userDetails.Name = data.Name
	userDetails.Password = password

	// path to email template
	templatePath := "./assets/template.html"
	to := []string{
		data.Email, // user to send email
	}

	// sending user credentials via email
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

// update user details
func PatchUser(w http.ResponseWriter, r *http.Request) {
	myId := r.Context().Value(custom.UserID).(string)
	myRole := r.Context().Value(custom.UserRole).(string)
	idParam := chi.URLParam(r, "id")

	if myRole != "super_admin" || myId != idParam {
		// fetch role for the given id from db
		// if role != user => return
		return
	}

	log.Println(idParam)
	log.Println(myRole)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	userToDeleteId := chi.URLParam(r, "id")

	userData := services.User{
		UserId: userToDeleteId,
	}

	err := userData.DeleteUser()
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = "The user has been successfully deleted."

	helper.EncodeJSON(w, http.StatusOK, payload)
}
