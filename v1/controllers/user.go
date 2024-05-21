package controllers

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"firebase.google.com/go/v4/auth"

	"github.com/go-chi/chi/v5"
	"github.com/rohan031/adgytec-api/firebase"
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
	payload.Message = "Successfully created user account."

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
	userId := chi.URLParam(r, "id")

	// success response
	var payload services.JSONResponse
	payload.Error = false
	payload.Message = "Successfully updated user details"

	// fetching user from firebase
	u, err := firebase.FirebaseClient.GetUser(ctx, userId)
	if err != nil {
		if auth.IsUserNotFound(err) {
			message := "No user found."
			err := &custom.MalformedRequest{Status: http.StatusNotFound, Message: message}
			helper.HandleError(w, err)
			return
		}

		log.Printf("Error getting user from firebase:%v\n", err)
		helper.HandleError(w, err)
		return
	}
	userRole := u.CustomClaims["role"]

	// decoding req body
	data, err := helper.DecodeJSON[services.User](w, r, mb)
	if err != nil {
		helper.HandleError(w, err)
		return
	}
	data.UserId = userId

	// validating request body parameters
	if !data.ValidateUpdateInput() {
		message := "The request body contains invalid input values."
		err = &custom.MalformedRequest{Status: http.StatusBadRequest, Message: message}
		helper.HandleError(w, err)
		return
	}

	// filling any missing data for update user function
	if data.Name == "" {
		data.Name = u.DisplayName
	}
	if data.Role == "" && userRole != nil {
		data.Role = userRole.(string)
	}

	// super admins can perform any action
	if myRole == "super_admin" {
		err = data.UpdateUser()
		if err != nil {
			helper.HandleError(w, err)
			return
		}
		helper.EncodeJSON(w, http.StatusOK, payload)
		return
	}

	// role is now admin or user and desired role is super admin
	// only super admin can grant this role
	if data.Role == "super_admin" {
		message := "Insufficient privileges to perform requested action."
		err := &custom.MalformedRequest{Status: http.StatusForbidden, Message: message}
		helper.HandleError(w, err)
		return
	}

	// if my role is user allow them to update name
	// in middleware we checked if they are trying to update their account
	// myid == userid because for role admin
	if myRole == "user" || myId == userId {
		err := data.UpdateUserName()
		if err != nil {
			helper.HandleError(w, err)
			return
		}
		helper.EncodeJSON(w, http.StatusOK, payload)
		return
	}

	// request owner role is admin
	// if user role is not user than its either admin or superadmin
	// and the req doesn't belong to update their resource
	if userRole != "user" {
		message := "Insufficient privileges to perform requested action."
		err := &custom.MalformedRequest{Status: http.StatusForbidden, Message: message}
		helper.HandleError(w, err)
		return
	}

	err = data.UpdateUser()
	if err != nil {
		helper.HandleError(w, err)
		return
	}
	helper.EncodeJSON(w, http.StatusOK, payload)
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

func GetUserById(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "id")

	userData := services.User{
		UserId: userId,
	}

	user, err := userData.GetUserById()
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse

	payload.Error = false
	payload.Data = user

	helper.EncodeJSON(w, http.StatusOK, payload)
}

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	cursor := r.URL.Query().Get("cursor")

	if cursor == "" {
		cursor = "0"
	}

	cursorInt, err := strconv.Atoi(cursor)
	if err != nil {
		var numError *strconv.NumError
		if errors.As(err, &numError) {
			message := "The provided cursor query parameter is invalid."
			err = &custom.MalformedRequest{Status: http.StatusBadRequest, Message: message}
			helper.HandleError(w, err)
			return
		}
		log.Printf("Error converting cursor string to int: %v\n", err)
		helper.HandleError(w, err)
	}

	var users services.User

	all, err := users.GetAllUsers(cursorInt)
	if err != nil {
		helper.HandleError(w, err)
		return
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Data = all

	helper.EncodeJSON(w, http.StatusOK, payload)
}
