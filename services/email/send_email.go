package email

import (
	"bytes"
	"encoding/base64"
	"html/template"
	"net/smtp"
	"os"
	"strings"
)

func (s *emailService) SendEmail(emailData ISendEmail) error {
	// get template
	tmpl, err := template.ParseFiles(emailData.GetTemplateDir())
	if err != nil {
		return err
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, emailData.GetTemplateData()); err != nil {
		return err
	}

	// Prepare attachments
	var mimeBody strings.Builder
	boundary := "MIMEBOUNDARY"

	mimeBody.WriteString("MIME-version: 1.0;\nContent-Type: multipart/related; boundary=\"" + boundary + "\"\n\n")
	mimeBody.WriteString("--" + boundary + "\n")
	mimeBody.WriteString("Content-Type: text/html; charset=\"UTF-8\"\n\n")
	mimeBody.WriteString(body.String() + "\n")

	for _, att := range emailData.GetAttachments() {
		fileBytes, err := os.ReadFile(att.Path)
		if err != nil {
			return err
		}

		mimeBody.WriteString("--" + boundary + "\n")
		mimeBody.WriteString("Content-Type: " + att.ContentType + "\n")
		mimeBody.WriteString("Content-Transfer-Encoding: base64\n")

		mimeBody.WriteString("Content-ID: <" + att.ContentID + ">\n")

		mimeBody.WriteString("\n")
		mimeBody.WriteString(base64.StdEncoding.EncodeToString(fileBytes) + "\n")
	}
	mimeBody.WriteString("--" + boundary + "--")

	// Compose headers
	toHeader := "To: " + strings.Join(emailData.GetToSend(), ", ") + "\r\n"
	subjectHeader := "Subject: " + emailData.GetSubject() + "\r\n"
	headers := toHeader + subjectHeader

	msg := []byte(headers + mimeBody.String())
	auth := smtp.PlainAuth("", s.username, s.password, s.smtpServer)

	return smtp.SendMail(s.smtpServer+":"+s.smtpPort, auth, string(s.from), emailData.GetToSend(), msg)

}
