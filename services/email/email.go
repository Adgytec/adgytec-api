package email

import "os"

// emailService implements IEmailService interface
// username is used to authenticate the account in smtp server
// from is the email address used to send email,
// this will be alias for username or will be username
type emailService struct {
	username   string
	password   string
	from       Email
	smtpServer string
	smtpPort   string
}

// CreateEmailService is a factory method to construct email service
func CreateEmailService(address Email) IEmailService {
	return &emailService{
		username:   os.Getenv("USERNAME"),
		password:   os.Getenv("PASSWORD"),
		smtpServer: os.Getenv("smtp_server"),
		smtpPort:   os.Getenv("smtp_server"),
		from:       address,
	}
}
