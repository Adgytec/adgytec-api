package services

import (
	"bytes"
	"html/template"
	"log"
	"net/smtp"
	"os"
)

type Constraint interface {
	any
}

func SendEmail[T Constraint](data T, templatePath string, to []string) error {
	smtpServer := "smtp.gmail.com"
	smtpPort := "587"

	from := os.Getenv("FROM")
	password := os.Getenv("PASS")

	// mail content
	subject := "Adgytec account creation"
	templateData := data

	t, err := template.ParseFiles(templatePath)
	if err != nil {
		log.Println("error trying to parse email template", err)
		return err
	}

	var body bytes.Buffer
	if err := t.Execute(&body, templateData); err != nil {
		log.Println("error trying to execute email template", err)
		return err
	}

	auth := smtp.PlainAuth("", from, password, smtpServer)

	mime := "MIME-version: 1.0;\nContent-Type: multipart/related; boundary=\"MIMEBOUNDARY\"\n\n"
	mime += "--MIMEBOUNDARY\n"
	mime += "Content-Type: text/html; charset=\"UTF-8\"\n\n"
	mime += body.String() + "\n"
	mime += "--MIMEBOUNDARY--"

	toHeader := "To: " + to[0] + "\r\n"
	subjectHeader := "Subject: " + subject + "\r\n"
	headers := toHeader + subjectHeader

	msg := []byte(headers + mime)

	err = smtp.SendMail(smtpServer+":"+smtpPort, auth, from, to, msg)
	if err != nil {
		log.Println("error sending mail", err)
		return err
	}

	return nil

}
