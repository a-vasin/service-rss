package main

import (
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"

	"service-rss/internal/config"
	"service-rss/internal/database"
	"service-rss/internal/server"
	"service-rss/internal/signal"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.WithError(err).Fatal("failed to read config")
	}

	db, err := database.New(cfg)
	if err != nil {
		log.WithError(err).Fatal("failed to establish db connection")
	}

	srv, err := server.New(cfg, db)
	if err != nil {
		log.WithError(err).Fatal("failed to init server")
	}

	srv.Start()
	log.Info("server was started")

	signalHandler := signal.NewHandler()
	signalHandler.Wait()
	log.Info("received termination signal")

	_ = srv.Shutdown()
	log.Info("server was shutdown")

	_ = db.Shutdown()
	log.Info("db was shutdown")
}
