package test

import (
	"encoding/json"
	"log"
)

const baseUrl = "http://localhost:8080"
const idToken = "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJuYW1lIjoicm9oYW4gdmVybWEiLCJyb2xlIjoic3VwZXJfYWRtaW4iLCJlbWFpbCI6InJvaGFudmVybWEwMzFAZ21haWwuY29tIiwiZW1haWxfdmVyaWZpZWQiOmZhbHNlLCJhdXRoX3RpbWUiOjE3MTYyODI0NjgsInVzZXJfaWQiOiJKRlY5dmFkMWNQNFk0UFp6YmtrdnFZZlRIVU5hIiwiZmlyZWJhc2UiOnsiaWRlbnRpdGllcyI6eyJlbWFpbCI6WyJyb2hhbnZlcm1hMDMxQGdtYWlsLmNvbSJdfSwic2lnbl9pbl9wcm92aWRlciI6InBhc3N3b3JkIn0sImlhdCI6MTcxNjI4MjQ2OCwiZXhwIjoxNzE2Mjg2MDY4LCJhdWQiOiJhZGd5dGVjLWFmYTMzIiwiaXNzIjoiaHR0cHM6Ly9zZWN1cmV0b2tlbi5nb29nbGUuY29tL2FkZ3l0ZWMtYWZhMzMiLCJzdWIiOiJKRlY5dmFkMWNQNFk0UFp6YmtrdnFZZlRIVU5hIn0."

func encodejson(data any) string {
	jsonRes, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		log.Fatalf("Erorr encoding json: %v", err)
	}
	return string(jsonRes)
}

type test struct {
	name            string
	expectedStatus  int
	expectedMessage string
	reqBody         interface{}
}
