package emailmanager

import (
	"context"

	"github.com/mikolajb/emailserv/internal/emailclient"
	"go.uber.org/zap"
)

type EmailManager struct {
	Logger       *zap.Logger
	EmailClients []emailclient.EmailClient
}

func (em *EmailManager) Send(ctx context.Context, to, from, subject string, opts ...emailclient.EmailOption) error {
	logger := em.Logger.With(
		zap.String("to", to),
		zap.String("from", from),
		zap.String("subject", subject),
	)

	for _, ec := range em.EmailClients {
		err := ec.Send(ctx, to, from, subject, opts...)
		if err == nil {
			logger.Debug("sent", zap.String("email_provider", ec.ProviderName()))
			break
		}
		logger.Error("email client error", zap.Error(err), zap.String("email_provider", ec.ProviderName()))
	}
	return nil
}
