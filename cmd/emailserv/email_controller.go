package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/mikolajb/emailserv/internal/emailclient"
	"github.com/mikolajb/emailserv/internal/emailmanager"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	emailRegexp = "^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"
)

// Message is an incoming message.
type Message struct {
	// Sender is email's "from" attribute.
	Sender string `json:"sender"`

	// Recipients are email's "to" attributes.
	Recipients []string `json:"recipients"`

	// CCRecipients...
	CCRecipients []string `json:"cc_recipients"`

	// BCCRecipients...
	BCCRecipients []string `json:"bcc_recipients"`

	// Subject is email's subject
	Subject string `json:"subject"`

	// Body is email's content
	Body string `json:"body"`
}

// Response holds service's response message.
type Response struct {
	// Message is a short information about a status.
	Message string `json:"message,omitempty"`

	// ValidationErrors is a list request's validation errors.
	ValidationErrors []*ValidationError `json:"validation_errors,omitempty"`

	// Error specifies if error occured.
	Error bool `json:"error,omitempty"`
}

// ValidationErrors holds an error of a particular field from the request
type ValidationError struct {
	// Field is a name of a field, e.g. "sender"
	// in case there is a list of fields, it will also contain an index,
	// e.g., recipients[1]
	Field string `json:"field"`

	// Error is a message explaining a validation error
	Error string `json:"error"`
}

// String builds a readable text from a ValidationError
func (ve *ValidationError) String() string {
	return fmt.Sprintf("field %s is not valid: %s", ve.Field, ve.Error)
}

type httpHandler struct {
	logger             *zap.Logger
	emailManager       *emailmanager.EmailManager
	authorizationToken string
}

// ServeHTTP is a main controller function
// it could be separated into two
// - one holding technical aspects of it
// - one holding application logic
func (h httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		h.logger.Debug("received request with a wrong method", zap.String("method", r.Method))
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Authorization") != h.authorizationToken {
		w.WriteHeader(http.StatusUnauthorized)
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

	err = h.emailManager.Send(
		ctx,
		message.Sender,
		message.Recipients,
		message.Subject,
		emailclient.WithBody(message.Body),
		emailclient.WithCCRecipients(message.CCRecipients),
		emailclient.WithBCCRecipients(message.BCCRecipients),
	)
	if err != nil {
		h.logger.Error("send error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(Response{
			Message: "Internal server error",
			Error:   true,
		})
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func validate(message *Message) []*ValidationError {
	errors := []*ValidationError{}

	re := regexp.MustCompile(emailRegexp)

	if !re.MatchString(message.Sender) {
		errors = append(errors, &ValidationError{
			Field: "sender",
			Error: "not a valid email",
		})
	}

	if len(message.Recipients) < 1 &&
		len(message.CCRecipients) < 1 &&
		len(message.BCCRecipients) < 1 {
		errors = append(errors, &ValidationError{
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
				errors = append(errors, &ValidationError{
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
