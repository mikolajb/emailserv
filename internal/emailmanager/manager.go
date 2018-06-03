package emailmanager

import (
	"context"
	"errors"
	"time"

	"github.com/mikolajb/emailserv/internal/emailclient"
	"go.uber.org/zap"
)

// EmailManager holds a state of an email manager
// It encapsulates multiple email clients
type EmailManager struct {
	Logger        *zap.Logger
	EmailClients  []emailclient.EmailClient
	ClientTimeout time.Duration
}

// Send sends an email using one of the available clients
func (em *EmailManager) Send(ctx context.Context, sender string, recipients []string, subject string, opts ...emailclient.EmailOption) error {
	logger := em.Logger.With(
		zap.String("sender", sender),
		zap.Strings("recipients", recipients),
		zap.String("subject", subject),
	)

	isSent := false

LoopOverClients:
	for _, ec := range em.EmailClients {
		iLogger := logger.With(zap.String("email_provider", ec.ProviderName()))
		clientCtx, cancel := context.WithTimeout(ctx, em.ClientTimeout)
		defer cancel()
		done := make(chan error)

		go func() {
			done <- ec.Send(clientCtx, sender, recipients, subject, opts...)
		}()

		select {
		case err := <-done:
			if err != nil {
				iLogger.Error("email client error", zap.Error(err))
			} else {
				iLogger.Debug("sent")
				isSent = true
				break LoopOverClients
			}
		case <-clientCtx.Done():
			logger.Error("client timeout")
		}
	}

	if !isSent {
		logger.Error("sending failed for all clients")
		return errors.New("sending emails failed for all clients")
	}

	return nil
}
