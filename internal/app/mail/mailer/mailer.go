package mailer

import (
	"bytes"
	"fmt"
	"net/smtp"
	"path/filepath"
	"text/template"
)

type MailMessage struct {
	Email   string
	Subject string
	Type    string
	Params  map[string]interface{}
}

type Mailer struct {
	SMTPHost    string
	SMTPPort    string
	User        string
	Password    string
	FromName    string
	FromAddress string
	TemplateDir string
}

func NewMailer(host, port, user, password, fromName, fromAddress string) *Mailer {
	return &Mailer{
		SMTPHost:    host,
		SMTPPort:    port,
		User:        user,
		Password:    password,
		FromName:    fromName,
		FromAddress: fromAddress,
		TemplateDir: "internal/app/mail/mailer/templates/",
	}
}

func (m *Mailer) Send(msg MailMessage) error {
	templatePath := filepath.Join(m.TemplateDir, msg.Type+".html")

	template, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", templatePath, err)
	}

	msgHeader := []byte("From: " + m.FromName + " <" + m.FromAddress + ">\r\n" +
		"To: " + msg.Email + "\r\n" +
		"Subject: " + msg.Subject + "\r\n" +
		"MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n")

	var body bytes.Buffer
	if err := template.Execute(&body, msg.Params); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	bodyWithHeader := append(msgHeader, body.Bytes()...)

	auth := smtp.PlainAuth("", m.User, m.Password, m.SMTPHost)
	addr := m.SMTPHost + ":" + m.SMTPPort

	if err := smtp.SendMail(addr, auth, m.FromAddress, []string{msg.Email}, bodyWithHeader); err != nil {
		return fmt.Errorf("failed to send email via SMTP: %w", err)
	}

	return nil
}
