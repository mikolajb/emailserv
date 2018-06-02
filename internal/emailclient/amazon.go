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

func NewAmazonClient(logger *zap.Logger, keyID, secretKey string) (*AmazonClient, error) {
	return &AmazonClient{
		logger: logger,
		sesClient: ses.New(session.Must(session.NewSession(
			&aws.Config{
				Logger: aws.LoggerFunc(func(args ...interface{}) {
					logger.Debug("abc", zap.Reflect("x", args))
				}),
				Credentials: credentials.NewStaticCredentials(
					keyID,
					secretKey,
					"",
				),
				Region: aws.String("eu-west-1"),
			},
		))),
	}, nil
}

func (ac *AmazonClient) ProviderName() string {
	return "aws"
}

func (ac *AmazonClient) Send(ctx context.Context, sender string, recipients []string, subject string, opts ...EmailOption) error {
	options := processOptions(opts...)
	logger := ac.logger.With(
		loggerFields(sender, recipients, subject, options)...,
	)

	logger.Debug("sending a message")

	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses:  aws.StringSlice(recipients),
			CcAddresses:  aws.StringSlice(options.ccRecipients),
			BccAddresses: aws.StringSlice(options.bccRecipients),
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
		Source: aws.String(sender),
	}

	result, err := ac.sesClient.SendEmailWithContext(ctx, input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				logger.Error("message cannot be sent", zap.Error(aerr))
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				logger.Error("sender's domain is not verified", zap.Error(aerr))
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				logger.Error("configuration set does not exist", zap.Error(aerr))
			case ses.ErrCodeConfigurationSetSendingPausedException:
				logger.Error("sending email is paused for a given configuration", zap.Error(aerr))
			case ses.ErrCodeAccountSendingPausedException:
				logger.Error("sending email is paused for a given SNS account", zap.Error(aerr))
			default:
				logger.Error("unknown aws error", zap.Error(aerr))
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			logger.Error("unknown error", zap.Error(err))
		}
		return fmt.Errorf("message cannot be sent: %s", err.Error())
	}

	logger.Debug("message is sent", zap.String("aws_message_id", *result.MessageId))

	return nil
}
