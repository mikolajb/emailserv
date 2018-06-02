package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/mikolajb/emailserv/internal/emailmanager"
	"go.uber.org/zap"
)

const (
	emailRegexp = "^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"
)

type Message struct {
	Sender        string   `json:"sender"`
	Recipients    []string `json:"recipients"`
	CCRecipients  []string `json:"cc_recipients"`
	BCCRecipients []string `json:"bcc_recipients"`
	Subject       string   `json:"subject"`
	Body          string   `json:"body"`
}

type Response struct {
	Message string `json:"message"`
	Error   bool
}

type httpHandler struct {
	logger       *zap.Logger
	emailManager *emailmanager.EmailManager
}

func (h httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		h.logger.Debug("received request with a wrong method", zap.String("method", r.Method))
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()

	var message Message

	err := json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		h.logger.Error("error while decoding message", zap.Error(err))
	}

	response := &Response{}

	validationErrors := validate(&message)
	if len(validationErrors) > 0 {
		response.Error = true
		w.WriteHeader(http.StatusBadRequest)
	}

	err = h.emailManager.Send(ctx, message.Sender, message.Recipients, message.Subject)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.logger.Error("send error", zap.Error(err))
	}
	w.WriteHeader(http.StatusOK)
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

	if len(message.Recipients) < 1 &&
		len(message.CCRecipients) < 1 &&
		len(message.BCCRecipients) < 1 {
		errors = append(errors, &validationError{
			Field: "recipients",
			Error: "at least one recipient has to be present",
		})
	}

	addresses := map[string][]string{
		"recipient":     message.Recipients,
		"cc_recipient":  message.CCRecipients,
		"bcc_recipient": message.BCCRecipients,
	}

	for addrType, addresses := range addresses {
		for i, r := range addresses {
			if !re.MatchString(r) {
				errors = append(errors, &validationError{
					Field: fmt.Sprintf(
						"%s[%d]",
						addrType,
						i,
					),
					Error: "invalid email address",
				})
			}
		}
	}

	return errors
}
