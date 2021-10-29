//go:generate mockgen -package ${GOPACKAGE} -destination mock_fetcher.go -source fetcher.go
package rss

import (
	"encoding/xml"
	"errors"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"service-rss/internal/dto"
)

type Fetcher interface {
	Fetch(url string) (*dto.RssFeed, error)
}

type fetcher struct {
	histogram *prometheus.HistogramVec
}

func NewFetcher() (Fetcher, error) {
	histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "fetch_duration_seconds",
		Help:    "Histogram of fetch time in seconds",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	}, []string{"status"})

	err := prometheus.Register(histogram)
	if err != nil {
		return nil, err
	}

	return &fetcher{
		histogram: histogram,
	}, nil
}

func (f *fetcher) Fetch(url string) (*dto.RssFeed, error) {
	start := time.Now()

	feed, err := f.fetch(url)

	status := "ok"
	if err != nil {
		status = "error"
	}
	f.histogram.WithLabelValues(status).Observe(time.Since(start).Seconds())

	return feed, err
}

func (f *fetcher) fetch(url string) (*dto.RssFeed, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	feed := &dto.RssFeed{}
	err = xml.NewDecoder(resp.Body).Decode(feed)
	if err != nil {
		return nil, err
	}

	if feed == nil || feed.Channel == nil || len(feed.Channel.Items) == 0 {
		return nil, errors.New("malformed rss feed")
	}

	return feed, nil
}
