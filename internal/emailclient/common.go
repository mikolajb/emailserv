package emailclient

import "context"

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

func WithBCCRecipient(recipient string) EmailOption {
	return func(o *emailOptions) {
		o.bccRecipients = append(o.bccRecipients, recipient)
	}
}

func WithBody(body string) EmailOption {
	return func(o *emailOptions) {
		o.body = body
	}
}
