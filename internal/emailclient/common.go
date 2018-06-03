// Package specifies a common interface for all email clients, e.g. amazon sns,
package emailclient

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// EmailOption...
type EmailOption func(*emailOptions)

// EmailClient is an common interface used by all email clients
type EmailClient interface {
	Send(context.Context, string, []string, string, ...EmailOption) error
	ProviderName() string
}

type emailOptions struct {
	ccRecipients  []string
	bccRecipients []string
	body          string
}

// WithCCRecipient adds a cc recipient to the list of options
func WithCCRecipient(recipient string) EmailOption {
	return func(o *emailOptions) {
		o.ccRecipients = append(o.ccRecipients, recipient)
	}
}

// WithCCRecipients sets cc recipients
func WithCCRecipients(recipients []string) EmailOption {
	return func(o *emailOptions) {
		o.ccRecipients = recipients
	}
}

// WithBCCRecipient adds a bcc recipient to the list of options
func WithBCCRecipient(recipient string) EmailOption {
	return func(o *emailOptions) {
		o.bccRecipients = append(o.bccRecipients, recipient)
	}
}

// WithBCCRecipients sets cc recipients
func WithBCCRecipients(recipients []string) EmailOption {
	return func(o *emailOptions) {
		o.bccRecipients = recipients
	}
}

// WithBody sets body in email options
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
