package emailmanager

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mikolajb/emailserv/internal/emailclient"
	"go.uber.org/zap/zaptest"
)

func TestEmailManager_Send(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	client1 := emailclient.NewMockEmailClient(mockCtrl)
	client2 := emailclient.NewMockEmailClient(mockCtrl)

	client1.EXPECT().ProviderName().Return("mock_client1")
	client2.EXPECT().ProviderName().Return("mock_client2")

	firstClientCall := client1.EXPECT().Send(gomock.Any(), "a", "b", "c")
	firstClientCall.Return(errors.New("some error"))
	client2.EXPECT().Send(gomock.Any(), "a", "b", "c").After(firstClientCall).Return(nil)

	em := EmailManager{
		Logger:       zaptest.NewLogger(t),
		EmailClients: []emailclient.EmailClient{client1, client2},
	}

	err := em.Send(context.TODO(), "a", "b", "c")
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
}
