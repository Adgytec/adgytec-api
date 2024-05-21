package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rohan031/adgytec-api/v1/controllers"
	"github.com/rohan031/adgytec-api/v1/services"
)

func TestGetAllUsers(t *testing.T) {
	// testing /users
	url := "/v1/users"
	req := httptest.NewRequest(http.MethodGet, url, http.NoBody)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(controllers.GetAllUsers)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("GetAllUsers returned wrong status code: got %v want %v",
			status, http.StatusUnauthorized)
	}

	var payload services.JSONResponse
	payload.Error = false
	payload.Message = "The user has been successfully deleted."
	expectedBody, _ := json.MarshalIndent(payload, "", "\t")

	if rr.Body.String() != string(expectedBody) {
		t.Errorf("GetAllUsers returned unexpected body: got %v want %v",
			rr.Body.String(), string(expectedBody))
	}
}
