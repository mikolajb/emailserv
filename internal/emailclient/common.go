package emailclient

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type EmailOption func(*emailOptions)

type EmailClient interface {
	Send(context.Context, string, []string, string, ...EmailOption) error
	ProviderName() string
}

type emailOptions struct {
	ccRecipients  []string
	bccRecipients []string
	body          string
}

func WithCCRecipient(recipient string) EmailOption {
	return func(o *emailOptions) {
		o.ccRecipients = append(o.ccRecipients, recipient)
	}
}

func WithCCRecipients(recipients []string) EmailOption {
	return func(o *emailOptions) {
		o.ccRecipients = recipients
	}
}

func WithBCCRecipient(recipient string) EmailOption {
	return func(o *emailOptions) {
		o.bccRecipients = append(o.bccRecipients, recipient)
	}
}

func WithBCCRecipients(recipients []string) EmailOption {
	return func(o *emailOptions) {
		o.bccRecipients = recipients
	}
}

func WithBody(body string) EmailOption {
	return func(o *emailOptions) {
		o.body = body
	}
}

func processOptions(opts ...EmailOption) *emailOptions {
	var result emailOptions
	for _, fn := range opts {
		fn(&result)
	}

	return &result
}

func loggerFields(sender string, recipients []string, subject string, options *emailOptions) []zapcore.Field {
	return []zapcore.Field{
		zap.String("sender", sender),
		zap.Strings("recipients", recipients),
		zap.Strings("cc", options.ccRecipients),
		zap.Strings("bcc", options.bccRecipients),
		zap.String("subject", subject),
	}
}
