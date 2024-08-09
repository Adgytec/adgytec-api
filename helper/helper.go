package helper

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/rohan031/adgytec-api/v1/custom"
	"github.com/rohan031/adgytec-api/v1/services"
)

type Constraint interface {
	any
}

func DecodeJSON[T Constraint](w http.ResponseWriter, r *http.Request, maxBytes int) (T, error) {
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var payload T
	err := decoder.Decode(&payload)

	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			message := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			return payload, &custom.MalformedRequest{
				Status:  http.StatusBadRequest,
				Message: message,
			}

		case errors.Is(err, io.ErrUnexpectedEOF):
			message := "Request body contains badly-formed JSON"
			return payload, &custom.MalformedRequest{
				Status:  http.StatusBadRequest,
				Message: message,
			}

		case errors.As(err, &unmarshalTypeError):
			message := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return payload, &custom.MalformedRequest{
				Status:  http.StatusBadRequest,
				Message: message,
			}

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			message := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			return payload, &custom.MalformedRequest{
				Status:  http.StatusBadRequest,
				Message: message,
			}

		case errors.Is(err, io.EOF):
			message := "Request body must not be empty"
			return payload, &custom.MalformedRequest{
				Status:  http.StatusBadRequest,
				Message: message,
			}

		case err.Error() == "http: request body too large":
			message := "Request body must not be larger than 1MB"
			return payload, &custom.MalformedRequest{
				Status:  http.StatusRequestEntityTooLarge,
				Message: message,
			}

		default:
			log.Printf("Error decoding request body: %v\n", err)
			return payload, err
		}
	}
	err = decoder.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		message := "Request body must only contain a single JSON object"
		return payload, &custom.MalformedRequest{
			Status:  http.StatusBadRequest,
			Message: message,
		}
	}

	return payload, nil
}

func EncodeJSON[T any](w http.ResponseWriter, status int, data T) {
	jsonRes, err := json.MarshalIndent(data, "", "\t")

	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_, err = w.Write(jsonRes)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func ErrorResponse(w http.ResponseWriter, err error, status ...int) {
	statusCode := http.StatusBadRequest
	if len(status) >= 1 {
		statusCode = status[0]
	}

	var payload services.JSONResponse

	payload.Error = true
	payload.Message = err.Error()

	EncodeJSON(w, statusCode, payload)
}

func HandleError(w http.ResponseWriter, err error) {
	var mr *custom.MalformedRequest

	if errors.As(err, &mr) {
		ErrorResponse(w, mr, mr.Status)
	} else {
		err = errors.New(http.StatusText(http.StatusInternalServerError))

		ErrorResponse(w, err, http.StatusInternalServerError)
	}
}

func ParseMultipartForm(w http.ResponseWriter, r *http.Request, maxSize int) error {
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxSize))
	err := r.ParseMultipartForm(int64(maxSize))
	if err != nil {
		log.Println(err)
		if strings.Contains(err.Error(), "http: request body too large") {
			messgage := "request body too large. Limit 10MB"
			HandleError(w, &custom.MalformedRequest{Status: http.StatusRequestEntityTooLarge, Message: messgage})
			return err
		}

		if strings.Contains(err.Error(), "mime: no media type") {
			HandleError(w, &custom.MalformedRequest{
				Status:  http.StatusUnsupportedMediaType,
				Message: http.StatusText(http.StatusUnsupportedMediaType),
			})
			return err
		}

		if strings.Contains(err.Error(), "request Content-Type isn't multipart/form-data") {
			HandleError(w, &custom.MalformedRequest{
				Status:  http.StatusBadRequest,
				Message: "Request Content-Type isn't multipart/form-data",
			})
			return err
		}

		log.Printf("Error parsing multipart form data: %v\n", err)
		HandleError(w, err)
		return err
	}

	return nil
}
