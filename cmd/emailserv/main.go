package main

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mikolajb/emailserv/internal/emailclient"
	"github.com/mikolajb/emailserv/internal/emailmanager"
	"go.uber.org/zap"
)

func main() {
	var config configuration
	config.init()
	config.parse()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	go func() {
		sig := <-sigs
		logger.Info("signal received, closing service", zap.Stringer("signal", sig))
		done <- true
	}()

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
		Logger:        logger.Named("email-manager"),
		EmailClients:  []emailclient.EmailClient{sc, ac},
		ClientTimeout: time.Duration(config.clientTimeout) * time.Millisecond,
	}

	handler := httpHandler{
		logger:       logger.Named("http-handler"),
		emailManager: em,
	}

	http.Handle("/email", handler)

	logger.Info("starting")

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		logger.Fatal("cannot start listener", zap.Error(err))
	}
	go func() {
		err = http.Serve(listener, nil)
		if err != nil {
			logger.Debug("http serve error (it always returns an error)", zap.Error(err))
		}
	}()

	<-done
	err = listener.Close()
	if err != nil {
		logger.Error("cannot close listener", zap.Error(err))
	}
	logger.Info("bye")
}
