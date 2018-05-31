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
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	client1 := emailclient.NewMockEmailClient(mockCtrl)
	client1.EXPECT().ProviderName().Return("mock_client1")
	em := &emailmanager.EmailManager{
		Logger:        zaptest.NewLogger(t),
		EmailClients:  []emailclient.EmailClient{client1},
		ClientTimeout: 100 * time.Millisecond,
	}

	handler := httpHandler{
		logger:       zaptest.NewLogger(t),
		emailManager: em,
	}

	cases := map[string]struct {
		message    *Message
		method     string
		path       string
		returnCode int
	}{
		"ok": {
			message: &Message{
				Sender:     "sender@example.com",
				Recipients: []string{"recipient@example.com"},
				Subject:    "subject",
			},
			method:     "POST",
			path:       "/email",
			returnCode: http.StatusOK,
		},
	}

	for hint, c := range cases {
		t.Run(hint, func(t *testing.T) {
			b := &bytes.Buffer{}
			json.NewEncoder(b).Encode(c.message)

			client1.EXPECT().
				Send(gomock.Any(), "recipient@example.com", "sender@example.com", "subject").
				DoAndReturn(func(ctx context.Context, recipiant, sender, subject string) error {
					return nil
				}).Times(1)

			req, err := http.NewRequest(c.method, c.path, b)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			if recorder.Code != c.returnCode {
				t.Errorf("expected return code %d but got %d", recorder.Code, c.returnCode)
			}
		})
	}
}

func Test_validate(t *testing.T) {
	m := &Message{
		Sender:       "abc",
		Recipients:   []string{"def", "def@abc.com"},
		CCRecipients: []string{"ghi@abc.com", "ghi"},
	}

	expected := map[string]string{
		"sender":        "not a valid email",
		"recipients":    "contains an invalid email address",
		"cc_recipients": "contains an invalid email address",
	}

	for _, ve := range validate(m) {
		message, ok := expected[ve.Field]
		if ok {
			if message != ve.Error {
				t.Errorf("expected message '%s' for field '%s' but got '%s'", message, ve.Field, ve.Error)
			}
			delete(expected, ve.Field)
		} else {
			t.Errorf("unexpected error '%s' for field '%s'", ve.Error, ve.Field)
		}
	}

	for field, error := range expected {
		t.Errorf("invalid field %s was accepted, expected error: %s", field, error)
	}
}
