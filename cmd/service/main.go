package main

import (
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"

	"service-rss/internal/config"
	"service-rss/internal/database"
	"service-rss/internal/rss"
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
	defer db.Shutdown()

	fetcher, err := rss.NewFetcher()
	if err != nil {
		log.WithError(err).Fatal("failed to init fetcher")
	}

	aggregator, err := rss.NewAggregator(fetcher)
	if err != nil {
		log.WithError(err).Fatal("failed to init aggregator")
	}

	cacher := rss.NewCacher(cfg, db, aggregator)
	go cacher.Start()
	defer cacher.Shutdown()

	srv, err := server.New(cfg, db, aggregator)
	if err != nil {
		log.WithError(err).Fatal("failed to init server")
	}
	defer srv.Shutdown()

	srv.Start()
	log.Info("server was started")

	signalHandler := signal.NewHandler()
	signalHandler.Wait()
	log.Info("received termination signal")
}
