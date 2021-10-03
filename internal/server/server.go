package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/xeipuuv/gojsonschema"

	"service-rss/internal/auth"
	"service-rss/internal/config"
	"service-rss/internal/database"
	"service-rss/internal/handlers"
	"service-rss/internal/rss"
)

type Server struct {
	server *http.Server
	db     database.Database
}

func New(cfg *config.Config, db database.Database, aggregator rss.Aggregator) (*Server, error) {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.AllowContentType("application/json"))

	schema, err := loadJsonSchema("jsonschema/api/rss/create/request.json")
	if err != nil {
		return nil, err
	}

	authHandler := auth.NewGoogleAuthHandler(cfg)
	router.Get("/login", authHandler.Login)

	rssCreateHandler := handlers.NewRssCreateHandler(db, schema, authHandler)
	router.Post("/api/rss/create", rssCreateHandler.ServeHTTP)

	indexHandler := handlers.NewIndexHandler(db, authHandler)
	router.Get("/", indexHandler.ServeHTTP)

	rssGetHandler := handlers.NewRssGetHandler(db, aggregator)
	router.Get("/{email}/{name}", rssGetHandler.ServeHTTP)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.ServerPort),
		Handler:      router,
		ReadTimeout:  cfg.ServerReadTimeout,
		WriteTimeout: cfg.ServerWriteTimeout,
	}

	return &Server{
		server: server,
		db:     db,
	}, nil
}

func (s *Server) Start() {
	go func() {
		err := s.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("failed to listen and serve")
		}
	}()
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.server.WriteTimeout)
	defer cancel()
	return s.server.Shutdown(ctx)
}

func loadJsonSchema(path string) (*gojsonschema.Schema, error) {
	path = fmt.Sprintf("file://%s", path)
	schemaLoader := gojsonschema.NewReferenceLoader(path)
	return gojsonschema.NewSchema(schemaLoader)
}
