package services

import (
	"net/http"

	"github.com/rohan031/adgytec-api/v1/custom"
	"github.com/rohan031/adgytec-api/v1/validation"
)

type JDKContact struct {
	Name  string
	Email string
	Phone string
	Query string
}

func (jdk *JDKContact) PostContactUsJDK() error {
	err := jdk.validateInput()
	if err != nil {
		message := "Invalid form details"
		return &custom.MalformedRequest{Status: http.StatusBadRequest, Message: message}
	}

	templatePath := "./assets/template-jdk.html"
	to := []string{
		"info@jdkshipping.com",
	}

	subject := "JDK-Shipping Form Submission"
	return SendEmail(jdk, templatePath, to, subject, 1)
}

func (jdk *JDKContact) validateInput() error {
	if !validation.ValidateName(jdk.Name) {
		return &custom.MalformedRequest{Status: http.StatusBadRequest,
			Message: "Invalid name",
		}
	}

	if !validation.ValidatePhone(jdk.Phone) {
		return &custom.MalformedRequest{Status: http.StatusBadRequest,
			Message: "Invalid phone number",
		}
	}

	if !validation.ValidateEmail(jdk.Email) {
		return &custom.MalformedRequest{Status: http.StatusBadRequest,
			Message: "Invalid email",
		}
	}

	if len(jdk.Query) > 0 && len(jdk.Query) < 10 {
		return &custom.MalformedRequest{Status: http.StatusBadRequest,
			Message: "Invalid query",
		}
	}

	return nil
}
