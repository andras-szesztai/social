package mailer

import (
	"bytes"
	"fmt"
	"log"
	"text/template"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridMailer struct {
	fromEmail string
	apiKey    string
	client    *sendgrid.Client
}

func NewSendGridMailer(fromEmail, apiKey string) *SendGridMailer {
	client := sendgrid.NewSendClient(apiKey)
	return &SendGridMailer{
		fromEmail: fromEmail,
		apiKey:    apiKey,
		client:    client,
	}
}

func (m *SendGridMailer) Send(templateFile, username, email string, data any, isSandbox bool) error {
	from := mail.NewEmail(fromName, m.fromEmail)
	to := mail.NewEmail(username, email)

	tmpl, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	subject := new(bytes.Buffer)
	htmlContent := new(bytes.Buffer)

	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	err = tmpl.ExecuteTemplate(htmlContent, "content", data)
	if err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	message := mail.NewSingleEmail(from, subject.String(), to, "", htmlContent.String())

	message.SetMailSettings(&mail.MailSettings{
		SandboxMode: &mail.Setting{
			Enable: &isSandbox,
		},
	})

	for i := 0; i < maxRetries; i++ {
		res, err := m.client.Send(message)
		if err != nil {
			log.Printf("Failed to send email to %s: %v", email, err)
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}

		log.Printf("Email sent to %s: %v", email, res.StatusCode)
		return nil
	}

	return fmt.Errorf("failed to send email to %s after %d retries", email, maxRetries)
}
