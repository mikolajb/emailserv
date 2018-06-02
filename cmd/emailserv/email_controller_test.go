package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

	clients := []emailclient.EmailClient{nil}

	em := &emailmanager.EmailManager{
		Logger:        zaptest.NewLogger(t),
		EmailClients:  clients,
		ClientTimeout: 100 * time.Millisecond,
	}

	token := "abc"

	handler := httpHandler{
		logger:             zaptest.NewLogger(t),
		emailManager:       em,
		authorizationToken: token,
	}

	cases := map[string]struct {
		message    *Message
		method     string
		path       string
		returnCode int
		token      string
	}{
		"ok": {
			message: &Message{
				Sender:     sender,
				Recipients: recipients,
				Subject:    subject,
			},
			method:     "POST",
			path:       "/email",
			returnCode: http.StatusCreated,
			token:      token,
		},
		"unauthorized": {
			message: &Message{
				Sender:     sender,
				Recipients: recipients,
				Subject:    subject,
			},
			method:     "POST",
			path:       "/email",
			returnCode: http.StatusUnauthorized,
			token:      "xyz",
		},
	}

	for hint, c := range cases {
		t.Run(hint, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			client1 := emailclient.NewMockEmailClient(mockCtrl)
			if c.returnCode == http.StatusCreated {
				client1.EXPECT().ProviderName().Return("mock_client1").Times(1)
				client1.EXPECT().
					Send(gomock.Any(), sender, recipients, subject).
					DoAndReturn(func(ctx context.Context, sender string, recipiants []string, subject string) error {
						return nil
					}).Times(1)
			}
			clients[0] = client1

			b := &bytes.Buffer{}
			json.NewEncoder(b).Encode(c.message)
			req, err := http.NewRequest(c.method, c.path, b)
			req.Header.Add("Authorization", c.token)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			if recorder.Code != c.returnCode {
				t.Errorf("expected return code %d but got %d", c.returnCode, recorder.Code)
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
