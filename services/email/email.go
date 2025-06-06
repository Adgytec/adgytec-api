package email

import "os"

// emailService implements IEmailService interface
// username is used to authenticate the account in smtp server
// from is the email address used to send email,
// this will be alias for username or will be username
type emailService struct {
	username   string
	password   string
	from       string
	smtpServer string
	smtpPort   string
}

// CreateDefaultEmailService is a factory method to construct email service with defaults
//
// SMTP server used is configured using env variables
// user account used to authenticate is configured using env variables
func CreateDefaultEmailService(address string) (IEmailService, error) {
	return CreateEmailService(Config{
		Username:   os.Getenv("USERNAME"),
		Password:   os.Getenv("PASSWORD"),
		SmtpServer: os.Getenv("smtp_server"),
		SmtpPort:   os.Getenv("smtp_server"),
		From:       address,
	})
}

// CreateEmailService creates an email service with given Config
func CreateEmailService(emailConfig Config) (IEmailService, error) {
	err := validatorObj.Struct(emailConfig)
	if err != nil {
		return nil, err
	}

	return &emailService{
		username:   emailConfig.Username,
		password:   emailConfig.Password,
		smtpServer: emailConfig.SmtpServer,
		smtpPort:   emailConfig.SmtpPort,
		from:       emailConfig.From,
	}, nil
}
