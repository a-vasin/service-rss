package main

import (
	log "github.com/sirupsen/logrus"

	"service-rss/internal/config"
	"service-rss/internal/server"
	"service-rss/internal/signal"
)

func main() {
	_, err := config.Read()
	if err != nil {
		log.WithError(err).Fatal("failed to read config")
	}

	srv := server.New()
	srv.Start()
	log.Info("server was started")

	signalHandler := signal.NewHandler()
	signalHandler.Wait()
	log.Info("received termination signal")

	srv.Shutdown()
	log.Info("server was shut down successfully")
}
