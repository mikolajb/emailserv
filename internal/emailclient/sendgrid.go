package emailclient

import (
	"context"
	"fmt"

	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"go.uber.org/zap"
)

type SendgridClient struct {
	logger         *zap.Logger
	sendgridClient *sendgrid.Client
}

func NewSendgridClient(logger *zap.Logger, key string) (*SendgridClient, error) {
	return &SendgridClient{
		logger:         logger,
		sendgridClient: sendgrid.NewSendClient(key),
	}, nil
}

func (sc *SendgridClient) ProviderName() string {
	return "sendgrid"
}

func (sc *SendgridClient) Send(ctx context.Context, sender string, recipients []string, subject string, opts ...EmailOption) error {
	options := processOptions(opts...)
	logger := sc.logger.With(
		loggerFields(sender, recipients, subject, options)...,
	)

	message := mail.NewV3Mail()
	message.SetFrom(mail.NewEmail("", sender))
	message.Subject = subject
	message.AddContent(
		mail.NewContent("text/plain", options.body),
		mail.NewContent("text/html", options.body),
	)

	personalization := mail.NewPersonalization()
	for _, r := range recipients {
		personalization.AddTos(mail.NewEmail("", r))
	}
	for _, r := range options.ccRecipients {
		personalization.AddCCs(mail.NewEmail("", r))
	}
	for _, r := range options.bccRecipients {
		personalization.AddBCCs(mail.NewEmail("", r))
	}
	message.AddPersonalizations(personalization)

	response, err := sc.sendgridClient.Send(message)
	if err != nil {
		logger.Error("sending error", zap.Error(err))
		return fmt.Errorf("sending error: %s", err.Error())
	}
	logger.Debug("request sent",
		zap.Int("status_code", response.StatusCode),
		zap.String("body", response.Body),
		zap.Reflect("headers", response.Headers),
	)
	if response.StatusCode/200 != 1 {
		return fmt.Errorf("unsuccessful request, status code: %d", response.StatusCode)
	}

	return nil
}
