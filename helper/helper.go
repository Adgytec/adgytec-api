package helper

import (
	"encoding/json"
	"net/http"

	"github.com/rohan031/adgytec-api/v1/services"
)

func DecodeJSON(w http.ResponseWriter, r *http.Request, maxBytes int, payload interface{}) error {
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(payload); err != nil {
		return err
	}

	return nil
}

func EncodeJSON(w http.ResponseWriter, status int, data interface{}) error {
	jsonRes, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_, err = w.Write(jsonRes)
	if err != nil {
		return err
	}

	return nil
}

func ErrorResponse(w http.ResponseWriter, err error, status ...int) {
	statusCode := http.StatusBadRequest
	if len(status) > 1 {
		statusCode = status[0]
	}

	var payload services.JSONResponse

	payload.Error = true
	payload.Message = err.Error()

	EncodeJSON(w, statusCode, payload)
}
