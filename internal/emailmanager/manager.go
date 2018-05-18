package emailmanager

import (
	"context"

	"go.uber.org/zap"

	"github.com/mikolajb/emailserv/internal/emailclient"
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
			break
		}
		logger.Error("email client error", zap.Error(err))
	}
	return nil
}
