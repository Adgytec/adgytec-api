package email

import "github.com/go-playground/validator/v10"

// validatorObj used to validate structs inside package email
var validatorObj = validator.New()

// IEmailService is an interface that allows sending emails
type IEmailService interface {
	SendEmail(emailData ISendEmail) error
}

// Attachment defines fields required to add attachments to an email template
type Attachment struct {
	Path        string
	ContentType string
	ContentID   string
}

// ISendEmail gives required fields data to send email
type ISendEmail interface {
	GetToSend() []string
	GetSubject() string
	GetTemplateDir() string
	// GetTemplateData defines data to be added on template and also contain fields for each attachment
	// with its content-id as its value
	GetTemplateData() any
	GetAttachments() []Attachment
}

// Config defines email config required to create email service
//
// Username: username used to authenticate with the smtp server
// From: email address used to send email, it can be alias for Username or can be Username itself
type Config struct {
	SmtpServer string `validate:"required,fqdn"`
	SmtpPort   string `validate:"required,number"`
	Username   string `validate:"required,email"`
	Password   string `validate:"required"`
	From       string `validate:"required,email"`
}

// address used to send email for different purposes
const (
	AUTH            = "Adgytec Auth <auth@adgytec.in>"
	INVITE          = "Adgytec Invites <invite@adgytec.in>"
	BILLING         = "Adgytec Billing <billing@adgytec.in>"
	CONTRACTS       = "Adgytec Contracts <contracts@adgytec.in>"
	BIRTHDAY_WISHES = "Adgytec Wishes <wishes@adgytec.in>"
	USER_MANAGEMENT = "Adgytec Accounts <accounts@adgytec.in>"
)
