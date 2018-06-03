package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/mikolajb/emailserv/internal/emailclient"
	"github.com/mikolajb/emailserv/internal/emailmanager"
	"go.uber.org/zap/zaptest"
)

func TestEmailControllerHandler(t *testing.T) {
	sender := "sender@example.com"
	recipients := []string{"recipient@example.com"}
	subject := "subject"
	body := "some body"
	token := "abc"
	bccRecipients := []string{"def@abc.com"}
	message := &bytes.Buffer{}
	json.NewEncoder(message).Encode(&Message{
		Sender:        sender,
		Recipients:    recipients,
		Subject:       subject,
		Body:          body,
		BCCRecipients: bccRecipients,
	})
	invalidMessage := &bytes.Buffer{}
	json.NewEncoder(invalidMessage).Encode(&Message{
		Sender:     "abc",
		Recipients: []string{"def"},
	})

	cases := map[string]struct {
		message       *bytes.Buffer
		method        string
		returnCode    int
		returnMessage string
		token         string
		clientDelay   time.Duration
		clientError   error
	}{
		"ok": {
			returnCode: http.StatusCreated,
		},
		"invalid-json": {
			message:       bytes.NewBufferString("abc"),
			returnCode:    http.StatusBadRequest,
			returnMessage: "Invalid JSON format",
		},
		"invalid-message": {
			message:       invalidMessage,
			returnCode:    http.StatusBadRequest,
			returnMessage: "Request not valid",
		},
		"bad-method": {
			method:     "GET",
			returnCode: http.StatusMethodNotAllowed,
		},
		"unauthorized": {
			returnCode: http.StatusUnauthorized,
			token:      "xyz",
		},
		"client-timeout": {
			returnCode:  http.StatusInternalServerError,
			clientDelay: time.Second,
		},
		"client-error": {
			returnCode:  http.StatusInternalServerError,
			clientError: errors.New("some error"),
		},
	}

	for hint, c := range cases {
		t.Run(hint, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			client1 := emailclient.NewMockEmailClient(mockCtrl)
			if c.returnCode == http.StatusCreated || c.returnCode == http.StatusInternalServerError {
				client1.EXPECT().ProviderName().Return("mock_client1").Times(1)
				client1.EXPECT().
					Send(gomock.Any(), sender, recipients, subject, gomock.Any()).
					DoAndReturn(func(ctx context.Context, senderX string, recipientsX []string, subjectX string, opts ...emailclient.EmailOption) error {
						if sender != senderX {
							t.Errorf("expected '%s' sender but got '%s'", sender, senderX)
						}
						if !reflect.DeepEqual(recipients, recipientsX) {
							t.Errorf("expected '%v' recipients but got '%v'", recipients, recipientsX)
						}
						if subject != subjectX {
							t.Errorf("expected '%s' subject but got '%s'", subject, subjectX)
						}
						select {
						case <-time.After(c.clientDelay):
							return c.clientError
						case <-ctx.Done():
							// sometime context is too fast
							<-time.After(10 * time.Millisecond)
						}
						return nil
					}).Times(1)
			}

			em := &emailmanager.EmailManager{
				Logger:        zaptest.NewLogger(t),
				EmailClients:  []emailclient.EmailClient{client1},
				ClientTimeout: 100 * time.Millisecond,
			}
			handler := httpHandler{
				logger:             zaptest.NewLogger(t),
				emailManager:       em,
				authorizationToken: token,
			}

			method := "POST"
			if c.method != "" {
				method = c.method
			}

			requestMessage := c.message
			if requestMessage == nil {
				requestMessage = bytes.NewBufferString(message.String())
			}
			req, err := http.NewRequest(method, "/email", requestMessage)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			if c.token == "" {
				req.Header.Add("Authorization", token)
			} else {
				req.Header.Add("Authorization", c.token)
			}
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			if recorder.Code != c.returnCode {
				t.Errorf("expected return code %d but got %d", c.returnCode, recorder.Code)
			}

			if c.returnMessage != "" {
				if recorder.Body != nil {
					var response Response
					err := json.NewDecoder(recorder.Body).Decode(&response)
					if err != nil {
						t.Errorf("cannot decode response body")
					}
					if response.Message != c.returnMessage {
						t.Errorf(
							"expected '%s' in a response message, got '%s'",
							c.returnMessage,
							response.Message,
						)
					}
				} else {
					t.Errorf("expected body in a response")
				}
			}
		})
	}
}

func Test_validate(t *testing.T) {
	cases := map[string]struct {
		message *Message
		errors  map[string]string
	}{
		"OK": {
			message: &Message{
				Sender:     "abc@abc.com",
				Recipients: []string{"def@abc.com"},
			},
		},
		"OK-only-BCC": {
			message: &Message{
				Sender:        "abc@abc.com",
				BCCRecipients: []string{"def@abc.com"},
			},
		},
		"sender-invalid": {
			message: &Message{
				Sender:     "abc",
				Recipients: []string{"def@abc.com"},
			},
			errors: map[string]string{
				"sender": "not a valid email",
			},
		},
		"recipients-missing": {
			message: &Message{
				Sender: "abc@abc.com",
			},
			errors: map[string]string{
				"recipients": "at least one recipient has to be present",
			},
		},
		"recipient[1]-invalid": {
			message: &Message{
				Sender:     "abc@abc.com",
				Recipients: []string{"def@abc.com", "abc"},
			},
			errors: map[string]string{
				"recipient[1]": "invalid email address",
			},
		},
		"cc_recipient[0]-invalid": {
			message: &Message{
				Sender:       "abc@abc.com",
				Recipients:   []string{"def@abc.com", "abc@abc.com"},
				CCRecipients: []string{"abc", "def@abc.com"},
			},
			errors: map[string]string{
				"cc_recipient[0]": "invalid email address",
			},
		},
		"bcc_recipient[1]-invalid": {
			message: &Message{
				Sender:        "abc@abc.com",
				Recipients:    []string{"def@abc.com", "abc@abc.com"},
				CCRecipients:  []string{"abc@abc.com", "def@abc.com"},
				BCCRecipients: []string{"abc@abc.com", "def", "ghi@abc.com"},
			},
			errors: map[string]string{
				"bcc_recipient[1]": "invalid email address",
			},
		},
	}

	for hint, c := range cases {
		t.Run(hint, func(t *testing.T) {
			for _, ve := range validate(c.message) {
				message, ok := c.errors[ve.Field]
				if ok {
					if message != ve.Error {
						t.Errorf("expected message '%s' for field '%s' but got '%s'", message, ve.Field, ve.Error)
					}
					delete(c.errors, ve.Field)
				} else {
					t.Errorf("unexpected error '%s' for field '%s'", ve.Error, ve.Field)
				}

			}
			for field, error := range c.errors {
				t.Errorf("invalid field %s wasn't detected, expected error: %s", field, error)
			}

		})
	}
}
