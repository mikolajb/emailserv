package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/mikolajb/emailserv/internal/emailmanager"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	Message          string             `json:"message,omitempty"`
	ValidationErrors []*validationError `json:"validation_errors,omitempty"`
	Error            bool               `json:"error,omitempty"`
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

	jsonEncoder := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json")
	var message Message

	err := json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		h.logger.Debug("error while decoding message", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(Response{
			Message: "Invalid JSON format",
			Error:   true,
		})
		return
	}

	validationErrors := validate(&message)
	if len(validationErrors) > 0 {
		var validationFields []zapcore.Field
		for _, ve := range validationErrors {
			validationFields = append(validationFields, zap.Stringer("validation_error", ve))
		}
		h.logger.Debug("invalid message", validationFields...)
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(Response{
			Message:          "Request not valid",
			ValidationErrors: validationErrors,
			Error:            true,
		})
		return
	}

	err = h.emailManager.Send(ctx, message.Sender, message.Recipients, message.Subject)
	if err != nil {
		h.logger.Error("send error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(Response{
			Message: "Internal server error",
			Error:   true,
		})
	}
	w.WriteHeader(http.StatusOK)
	jsonEncoder.Encode(Response{Message: "sent"})
}

type validationError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

func (ve *validationError) String() string {
	return fmt.Sprintf("field %s is not valid: %s", ve.Field, ve.Error)
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
