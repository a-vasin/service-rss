package handlers

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"service-rss/internal/database"
	"service-rss/internal/rss"
)

type rssGetHandler struct {
	db               database.Database
	aggregator       rss.Aggregator
	cacheMissCounter prometheus.Counter
}

func NewRssGetHandler(db database.Database, aggregator rss.Aggregator) (http.Handler, error) {
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cache_miss_counter",
		Help: "Counter of cache misses",
	})

	err := prometheus.Register(counter)
	if err != nil {
		return nil, err
	}

	return &rssGetHandler{
		db:               db,
		aggregator:       aggregator,
		cacheMissCounter: counter,
	}, nil
}

func (h *rssGetHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	email := chi.URLParam(req, "email")
	name := chi.URLParam(req, "name")

	if len(email) == 0 || len(name) == 0 {
		writeBadRequest(writer, "email and name should be specified", "")
		return
	}

	rssCached, err := h.db.GetCachedRss(email, name)
	if err != nil {
		if err == sql.ErrNoRows {
			msg := fmt.Sprintf("email: %s, name: %s", email, name)
			writeNotFound(writer, "rss feed was not found", msg)
			return
		}

		writeInternalError(writer, "failed to get cached rss", err)
		return
	}

	rssFeedString := []byte(rssCached.RssFeed)
	// fallback in case rss has not been cached yet
	if len(rssFeedString) == 0 {
		h.cacheMissCounter.Inc()

		rssFeed := h.aggregator.Aggregate(&rssCached.Rss)

		rssFeedString, err = xml.Marshal(rssFeed)
		if err != nil {
			writeInternalError(writer, "failed to marshal rss feed", err)
			return
		}

		validUntil := rss.GetValidUntil(rssFeed)
		err = h.db.SaveCachedRss(rssCached.Rss.ID, string(rssFeedString), validUntil)
		if err != nil {
			log.WithError(err).Error("failed to save cached rss feed")
		}
	}

	writer.Header().Set("Content-Type", "application/xml")
	writer.WriteHeader(http.StatusOK)
	writer.Write(rssFeedString)
}
