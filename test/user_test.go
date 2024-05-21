package test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/rohan031/adgytec-api/v1/services"
)

func TestPostUser(t *testing.T) {
	tests := []test{
		{
			name:            "successfull user creation",
			expectedStatus:  http.StatusCreated,
			expectedMessage: "Successfully created user account.",
			reqBody:         services.User{Name: "aryan dhoundiyal", Email: "aryanluvmomdad@gmail.com", Role: "user"},
		}, {
			name:            "duplicate user creation",
			expectedStatus:  http.StatusConflict,
			expectedMessage: "The email address provided is already associated with an existing user account.",
			reqBody:         services.User{Name: "test user", Email: "vermarohan031@gmail.com", Role: "user"},
		},
	}

	url := baseUrl + "/v1/user"
	client := http.Client{
		Timeout: 30 * time.Second,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody io.Reader
			bodyBytes, _ := json.Marshal(tt.reqBody)
			reqBody = bytes.NewBuffer(bodyBytes)

			req, err := http.NewRequest(http.MethodPost, url, reqBody)
			if err != nil {
				t.Fatalf("Error creating post request to user endpoint: %v", err)
			}
			req.Header.Add("Authorization", "Bearer "+idToken)
			req.Header.Add("Content-Type", "application/json")

			res, err := client.Do(req)
			if err != nil {
				t.Fatalf("client: error making HTTP request: %v", err)
			}
			defer res.Body.Close()

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("PostUser returned unexpected status code: got %v want %v", res.StatusCode, tt.expectedStatus)
			}

			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("client: error reading response body: %v", err)
			}

			var payload services.JSONResponse
			json.Unmarshal(body, &payload)
			if payload.Message != tt.expectedMessage {
				t.Errorf("PostUser returned unexpected body: got %v want %v", payload.Message, tt.expectedMessage)
			}
		})
	}

}
