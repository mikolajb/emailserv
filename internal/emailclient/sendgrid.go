package emailclient

import (
	"context"
	"errors"

	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"go.uber.org/zap"
)

type SendgridClient struct {
	logger         *zap.Logger
	sendgridClient *sendgrid.Client
}

func NewSendgridClient(logger *zap.Logger) (*SendgridClient, error) {
	return &SendgridClient{
		logger:         logger,
		sendgridClient: sendgrid.NewSendClient(""),
	}, nil
}

func (sc *SendgridClient) ProviderName() string {
	return "sendgrid"
}

func (sc *SendgridClient) Send(ctx context.Context, to, from, subject string, opts ...EmailOption) error {
	logger := sc.logger.With(
		zap.String("to", to),
		zap.String("from", from),
		zap.String("subject", subject),
	)
	message := mail.NewSingleEmail(
		mail.NewEmail("", from),
		subject,
		mail.NewEmail("", to),
		" ",
		" ",
	)
	response, err := sc.sendgridClient.Send(message)
	if err != nil {
		logger.Error("sending error", zap.Error(err))
		return err
	}
	logger.Debug("request sent",
		zap.Int("status_code", response.StatusCode),
		zap.String("body", response.Body),
		zap.Reflect("headers", response.Headers),
	)
	if response.StatusCode/200 != 1 {
		return errors.New("unsuccesfull request")
	}

	return nil
}
