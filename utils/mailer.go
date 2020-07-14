package utils

import (
	"bytes"
	"errors"
	"html/template"
	"net/smtp"
	"tupeuxcourrir_api/config"
)

var auth = smtp.PlainAuth("", config.EmailUsername, config.EmailPassword, config.EmailHost)

type mail struct {
	from    string
	to      []string
	subject string
	body    string
}

func NewMail(to []string, subject, body string) *mail {
	return &mail{
		to:      to,
		subject: subject,
		body:    body,
	}
}

func (m *mail) SendEmail() error {
	if m.subject == "" || m.body == "" {
		return errors.New("subject or body are empty")
	}

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subject := "Subject: " + m.subject + "\n"
	msg := []byte(subject + mime + "\n" + m.body)
	addr := "smtp.gmail.com:587"

	if err := smtp.SendMail(addr, auth, "ivan.besevic_fr@yahoo.com", m.to, msg); err != nil {
		return err
	}
	return nil
}

func (m *mail) ParseTemplate(templateFileName string, data interface{}) error {
	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return err
	}
	m.body = buf.String()
	return nil
}
