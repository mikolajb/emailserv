package emailclient

import "context"

type EmailOption func(*emailOptions)

type EmailClient interface {
	Send(context.Context, string, string, string, ...EmailOption) error
	ProviderName() string
}

type emailOptions struct {
	recipients   []string
	ccRecipients []string
	body         string
}

func WithRecipient(recipient string) EmailOption {
	return func(o *emailOptions) {
		o.recipients = append(o.recipients, recipient)
	}
}

func WithCCRecipient(recipient string) EmailOption {
	return func(o *emailOptions) {
		o.ccRecipients = append(o.ccRecipients, recipient)
	}
}

func WithBody(body string) EmailOption {
	return func(o *emailOptions) {
		o.body = body
	}
}
