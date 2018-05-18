package emailclient

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"go.uber.org/zap"
)

type AmazonClient struct {
	logger    *zap.Logger
	sesClient *ses.SES
}

func NewAmazonClient(logger *zap.Logger) *AmazonClient {
	return &AmazonClient{
		logger: logger,
		sesClient: ses.New(session.Must(session.NewSession(
			&aws.Config{
				Logger: aws.LoggerFunc(func(args ...interface{}) {
					logger.Debug("abc", zap.Reflect("x", args))
				}),
				Credentials: credentials.NewStaticCredentials(
					"",
					"",
					"",
				),
				Region: aws.String("eu-west-1"),
			},
		))),
	}
}

func (ac *AmazonClient) Send(ctx context.Context, to, from, subject string, opts ...EmailOption) error {
	options := &emailOptions{}
	for _, fn := range opts {
		fn(options)
	}
	logger := ac.logger.With(
		zap.Strings("to", append(options.recipients, to)),
		zap.Strings("cc", options.ccRecipients),
		zap.String("from", from),
		zap.String("subject", subject),
	)

	logger.Debug("sending a message")
	recipients := []*string{aws.String(to)}
	ccRecipients := []*string{}

	for _, r := range options.recipients {
		recipients = append(recipients, aws.String(r))
	}
	for _, r := range options.ccRecipients {
		ccRecipients = append(ccRecipients, aws.String(r))
	}

	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: recipients,
			CcAddresses: ccRecipients,
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Text: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(options.body),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(from),
	}

	result, err := ac.sesClient.SendEmailWithContext(ctx, input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				fmt.Println(ses.ErrCodeMessageRejected, aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				fmt.Println(ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				fmt.Println(ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
			case ses.ErrCodeConfigurationSetSendingPausedException:
				fmt.Println(ses.ErrCodeConfigurationSetSendingPausedException, aerr.Error())
			case ses.ErrCodeAccountSendingPausedException:
				fmt.Println(ses.ErrCodeAccountSendingPausedException, aerr.Error())
			default:
				logger.Error("aws error", zap.Error(aerr))
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			logger.Error("unknown error", zap.Error(err))
		}
		return nil
	}

	fmt.Println(result)

	return nil
}
