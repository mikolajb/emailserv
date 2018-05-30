package main

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"time"

	"github.com/mikolajb/emailserv/internal/emailmanager"
	"go.uber.org/zap"
)

const (
	emailRegexp = "^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"
)

type Message struct {
	Sender       string   `json:"sender"`
	Recipients   []string `json:"recipients"`
	CCRecipients []string `json:"cc_recipients"`
	Subject      string   `json:"subject"`
	Body         string   `json:"body"`
}

type httpHandler struct {
	logger       *zap.Logger
	emailManager *emailmanager.EmailManager
	timeout      time.Duration
}

func (h httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	var message Message

	err := json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		h.logger.Error("error while decoding message", zap.Error(err))
	}

	err = h.emailManager.Send(ctx, message.Recipients[0], message.Sender, message.Subject)
	if err != nil {
		h.logger.Error("send error", zap.Error(err))
	}
}

type validationError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

func validate(message *Message) []*validationError {
	errors := []*validationError{}

	re := regexp.MustCompile(emailRegexp)

	if !re.MatchString(message.Sender) {
		errors = append(errors, &validationError{
			Field: "sender",
			Error: "not a valid email",
		})
	}
	for _, r := range message.Recipients {
		if !re.MatchString(r) {
			errors = append(errors, &validationError{
				Field: "recipients",
				Error: "contains an invalid email address",
			})
		}
	}
	for _, r := range message.CCRecipients {
		if !re.MatchString(r) {
			errors = append(errors, &validationError{
				Field: "cc_recipients",
				Error: "contains an invalid email address",
			})
		}
	}

	return errors
}
