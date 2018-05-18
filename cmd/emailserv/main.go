package main

import (
	"context"
	"time"

	"github.com/mikolajb/emailserv/internal/emailclient"
	"github.com/mikolajb/emailserv/internal/emailmanager"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	ac := emailclient.NewAmazonClient(logger.Named("aws"))
	em := &emailmanager.EmailManager{
		Logger:       logger.Named("email-manager"),
		EmailClients: []emailclient.EmailClient{ac},
	}
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	em.Send(ctx, "mikolajb@gmail.com", "m@mikolajb.xyz", "test", emailclient.WithBody("abc"))
}
