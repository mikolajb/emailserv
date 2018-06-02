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
	options := &emailOptions{}
	for _, fn := range opts {
		fn(options)
	}
	logger := sc.logger.With(
		zap.String("sender", sender),
		zap.Strings("recipients", recipients),
		zap.Strings("cc", options.ccRecipients),
		zap.Strings("bcc", options.bccRecipients),
		zap.String("subject", subject),
	)

	message := mail.NewV3Mail()
	message.From = mail.NewEmail("", sender)
	message.Subject = subject
	message.AddContent(mail.NewContent("plain/text", options.body))

	personalization := mail.NewPersonalization()
	personalization.AddTos(stringsToEmails(recipients...)...)
	if len(options.ccRecipients) > 0 {
		personalization.AddCCs(stringsToEmails(options.ccRecipients...)...)
	}
	if len(options.bccRecipients) > 0 {
		personalization.AddBCCs(stringsToEmails(options.bccRecipients...)...)
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

func stringsToEmails(addresses ...string) []*mail.Email {
	emails := make([]*mail.Email, len(addresses))
	for _, e := range addresses {
		emails = append(emails, mail.NewEmail(
			"", e,
		))
	}
	return emails
}
