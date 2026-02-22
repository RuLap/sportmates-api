package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/RuLap/sportmates-api/internal/app/mail/mailer"
	"github.com/RuLap/sportmates-api/internal/pkg/config"
	"github.com/RuLap/sportmates-api/internal/pkg/events"
	"github.com/RuLap/sportmates-api/internal/pkg/rabbitmq"
)

type MailService struct {
	log      *slog.Logger
	rabbitmq *rabbitmq.Service
	mailer   *mailer.Mailer
}

func NewMailService(log *slog.Logger, mqService *rabbitmq.Service, smtpConfig *config.SMTP) *MailService {
	mailer := mailer.NewMailer(
		smtpConfig.Host,
		smtpConfig.Port,
		smtpConfig.User,
		smtpConfig.Password,
		smtpConfig.FromName,
		smtpConfig.FromAddress,
	)

	return &MailService{
		log:      log,
		rabbitmq: mqService,
		mailer:   mailer,
	}
}

func (s *MailService) StartConsumer(ctx context.Context) error {
	if s.rabbitmq == nil {
		return fmt.Errorf("rabbitmq client is not initialized")
	}

	s.log.Info("starting mail service consumer")

	return s.rabbitmq.ConsumeEmailEvents(ctx, s.handleEmailEvent)
}

func (s *MailService) handleEmailEvent(event events.EmailEvent) error {
	switch event.Template {
	case "email_confirmation":
		return s.sendConfirmationEmail(event)
	case "password_reset":
		return s.sendPasswordResetEmail(event)
	case "welcome":
		return s.sendWelcomeEmail(event)
	default:
		s.log.Warn("unknown email template", "template", event.Template)
		return fmt.Errorf("unknown email template: %s", event.Template)
	}
}

func (s *MailService) sendConfirmationEmail(event events.EmailEvent) error {
	s.log.Info("sending confirmation email", "to", event.To)

	confirmationURL, _ := event.Data["confirmation_url"].(string)
	userEmail, _ := event.Data["user_email"].(string)

	msg := mailer.MailMessage{
		Email:   userEmail,
		Subject: "Подтвердите ваш email",
		Type:    "email_confirmation",
		Params: map[string]interface{}{
			"ConfirmationURL": confirmationURL,
			"UserEmail":       userEmail,
		},
	}

	if err := s.mailer.Send(msg); err != nil {
		s.log.Error("failed to send confirmation email", "error", err, "to", userEmail)
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.log.Info("confirmation email sent successfully", "to", userEmail)
	return nil
}

func (s *MailService) sendPasswordResetEmail(event events.EmailEvent) error {
	s.log.Info("sending password reset email", "to", event.To)

	resetURL, _ := event.Data["reset_url"].(string)
	userEmail, _ := event.Data["user_email"].(string)

	msg := mailer.MailMessage{
		Email:   userEmail,
		Subject: "Сброс пароля",
		Type:    "password_reset",
		Params: map[string]interface{}{
			"ResetURL":  resetURL,
			"UserEmail": userEmail,
		},
	}

	if err := s.mailer.Send(msg); err != nil {
		s.log.Error("failed to send password reset email", "error", err, "to", userEmail)
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s *MailService) sendWelcomeEmail(event events.EmailEvent) error {
	s.log.Info("sending welcome email", "to", event.To)

	userEmail, _ := event.Data["user_email"].(string)
	userName, _ := event.Data["user_name"].(string)

	msg := mailer.MailMessage{
		Email:   userEmail,
		Subject: "Добро пожаловать в Trackmus!",
		Type:    "welcome",
		Params: map[string]interface{}{
			"UserName":  userName,
			"UserEmail": userEmail,
		},
	}

	if err := s.mailer.Send(msg); err != nil {
		s.log.Error("failed to send welcome email", "error", err, "to", userEmail)
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
