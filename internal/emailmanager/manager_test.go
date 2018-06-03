package emailmanager

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/mikolajb/emailserv/internal/emailclient"
	"go.uber.org/zap/zaptest"
)

func TestEmailManager_Send(t *testing.T) {
	cases := map[string]struct {
		delay time.Duration
		err   error
	}{
		"OK": {
			delay: 10 * time.Millisecond,
		},
		"timeout": {
			delay: time.Second,
			err:   errors.New("sending emails failed for all clients"),
		},
	}

	for hint, c := range cases {
		t.Run(hint, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			client1 := emailclient.NewMockEmailClient(mockCtrl)
			client2 := emailclient.NewMockEmailClient(mockCtrl)

			em := EmailManager{
				Logger:        zaptest.NewLogger(t),
				EmailClients:  []emailclient.EmailClient{client1, client2},
				ClientTimeout: 100 * time.Millisecond,
			}

			client1.EXPECT().ProviderName().Return("mock_client1")
			client2.EXPECT().ProviderName().Return("mock_client2")
			firstClientCall := client1.EXPECT().Send(gomock.Any(), "a", []string{"b"}, "c")
			firstClientCall.Return(errors.New("some error")).Times(1)

			secondClientCall := client2.EXPECT().Send(gomock.Any(), "a", []string{"b"}, "c")
			secondClientCall.After(firstClientCall)
			secondClientCall.DoAndReturn(func(ctx context.Context, sender string, recipients []string, subject string) error {
				if c.delay != 0 {
					select {
					case <-time.After(c.delay):
					case <-ctx.Done():
						// sometime context is too fast
						<-time.After(10 * time.Millisecond)
					}
				}
				return nil
			}).Times(1)

			err := em.Send(context.Background(), "a", []string{"b"}, "c")
			if c.err != nil && err != nil {
				if c.err.Error() != err.Error() {
					t.Errorf("expected error '%s' but got '%s'", c.err.Error(), err.Error())
				}
			} else if !(c.err == nil && err == nil) {
				if err != nil {
					t.Errorf("unexpected error: %s", err.Error())
				}
				if c.err != nil {
					t.Errorf("expected error '%s' but got nothing", c.err.Error())
				}
			}
		})
	}
}
