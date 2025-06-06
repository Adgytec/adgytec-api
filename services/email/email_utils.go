package email

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

// Email defines a email address used to send emails
type Email string

// address used to send all types of email
const (
	ADMIN   Email = "admin@adgytec.in"
	TEAM    Email = "team@adgytec.in"
	NOREPLY Email = "no-reply@adgytec.in"
	AUTH    Email = "auth@adgytec.in"
)

func (e *Email) IsValid() bool {
	switch *e {
	case ADMIN, TEAM, NOREPLY, AUTH:
		return true
	default:
		return false
	}
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
