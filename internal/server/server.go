package server

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	log "github.com/sirupsen/logrus"

	"service-rss/internal/handlers"
)

var (
	defaultServeTimeout = 10 * time.Second
)

type Server struct {
	server *http.Server
}

func New() *Server {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(defaultServeTimeout))
	router.Use(middleware.AllowContentType("application/json"))

	rssCreateHandler := handlers.NewRssCreateHandler()
	router.Post("/api/rss/create", rssCreateHandler.ServeHTTP)

	server := &http.Server{
		Addr:         ":80",
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return &Server{
		server: server,
	}
}

func (s *Server) Start() {
	go func() {
		err := s.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("failed to listen and serve")
		}
	}()
}

func (s *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), defaultServeTimeout)
	defer cancel()
	_ = s.server.Shutdown(ctx)
}
