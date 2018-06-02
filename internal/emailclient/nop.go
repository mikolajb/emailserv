package emailclient

import (
	"context"

	"go.uber.org/zap"
)

type NopClient struct {
	logger *zap.Logger
}

func NewNopClient(logger *zap.Logger) *NopClient {
	return &NopClient{
		logger: logger,
	}
}

func (nc *NopClient) ProviderName() string {
	return "nop"
}

func (nc *NopClient) Send(ctx context.Context, sender string, recipients []string, subject string, opts ...EmailOption) error {
	options := processOptions(opts...)
	logger := nc.logger.With(
		loggerFields(sender, recipients, subject, options)...,
	)

	logger.Debug("logging a message in NOP client")

	return nil
}
