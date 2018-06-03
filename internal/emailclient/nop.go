package emailclient

import (
	"context"

	"go.uber.org/zap"
)

// NopClient holds a state of a client
type NopClient struct {
	logger *zap.Logger
}

// NewNopClient returns a new NOP client
// It is useful for tests
func NewNopClient(logger *zap.Logger) *NopClient {
	return &NopClient{
		logger: logger,
	}
}

// ProviderName returns "nop"
func (nc *NopClient) ProviderName() string {
	return "nop"
}

// Send logs message content and does nothing
func (nc *NopClient) Send(ctx context.Context, sender string, recipients []string, subject string, opts ...EmailOption) error {
	options := processOptions(opts...)
	logger := nc.logger.With(
		loggerFields(sender, recipients, subject, options)...,
	)

	logger.Debug("logging a message in NOP client")

	return nil
}
