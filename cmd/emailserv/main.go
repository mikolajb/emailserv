package main

import (
	"net/http"
	"time"

	"github.com/mikolajb/emailserv/internal/emailclient"
	"github.com/mikolajb/emailserv/internal/emailmanager"
	"go.uber.org/zap"
)

func main() {
	var config configuration
	config.init()
	config.parse()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	ac := emailclient.NewAmazonClient(
		logger.Named("aws"),
		config.amazon.key,
		config.amazon.secret,
	)
	sc, _ := emailclient.NewSendgridClient(
		logger.Named("sendgrid"),
		config.sendgrid.key,
	)
	em := &emailmanager.EmailManager{
		Logger:       logger.Named("email-manager"),
		EmailClients: []emailclient.EmailClient{sc, ac},
	}

	handler := httpHandler{
		logger:       logger.Named("http-handler"),
		emailManager: em,
		timeout:      time.Second,
	}

	http.Handle("/email", handler)

	http.ListenAndServe(":8080", nil)
}
