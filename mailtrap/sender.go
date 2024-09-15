package mailtrap

import (
	"braces.dev/errtrace"
	"context"
	"github.com/vorobeyme/mailtrap-go/mailtrap"
	"go.uber.org/zap"
)

type Sender interface {
	Send(
		ctx context.Context,
		subject string,
		htmlContent string,
		recipients []mailtrap.EmailAddress,
	) error
}

type RealSender struct {
	config *ModuleConfig
	logger *zap.Logger
}

func NewSender(
	config *ModuleConfig,
	logger *zap.Logger,
) *RealSender {
	return &RealSender{
		config: config,
		logger: logger,
	}
}

func (s *RealSender) Send(
	ctx context.Context,
	subject string,
	htmlContent string,
	recipients []mailtrap.EmailAddress,
) error {
	message := &mailtrap.SendEmailRequest{
		From: mailtrap.EmailAddress{
			Email: s.config.FromEmail,
			Name:  s.config.FromName,
		},
		To: recipients,
		CustomVars: map[string]string{
			"user_id":  "1",
			"batch_id": "2",
		},
		Headers: map[string]string{
			"Content-Type": "text/html; charset=utf-8",
		},
		Subject: subject,
		HTML:    htmlContent,
	}
	if !s.config.SendEmails {
		s.logger.Info("email sending disabled", zap.Any("message", message))
		return nil
	}
	client, err := mailtrap.NewSendingClient(s.config.ApiKey)
	if err != nil {
		return errtrace.Wrap(err)
	}
	_, _, err = client.Send(message)
	return errtrace.Wrap(err)
}
