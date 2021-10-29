package rss

import (
	"encoding/xml"
	"math"
	"sort"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"service-rss/internal/database"
	"service-rss/internal/dto"
)

const (
	defaultTtl = 5 // rss ttl is in minutes according to specification
)

type Aggregator interface {
	Aggregate(rss *database.Rss) *dto.RssFeed
}

type aggregator struct {
	fetcher   Fetcher
	histogram *prometheus.HistogramVec
}

func NewAggregator(fetcher Fetcher) (Aggregator, error) {
	histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "aggregation_duration_seconds",
		Help:    "Histogram of aggregation time in seconds",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	}, []string{"status"})

	err := prometheus.Register(histogram)
	if err != nil {
		return nil, err
	}

	return &aggregator{
		fetcher:   fetcher,
		histogram: histogram,
	}, nil
}

func (a *aggregator) Aggregate(rss *database.Rss) *dto.RssFeed {
	start := time.Now()

	feed := a.aggregate(rss)

	status := "ok"
	if feed == nil {
		status = "error"
	}
	a.histogram.WithLabelValues(status).Observe(time.Since(start).Seconds())

	return feed
}

func (a *aggregator) aggregate(rss *database.Rss) *dto.RssFeed {
	if rss == nil {
		log.Error("empty rss")
		return nil
	}

	ttl := int64(math.MaxInt64)
	inputFeeds := make([]*dto.RssFeed, 0, len(rss.Sources))
	for _, rssUrl := range rss.Sources {
		feed, err := a.fetcher.Fetch(rssUrl)
		if err != nil {
			ttl = defaultTtl
			log.WithError(err).
				WithField("url", rssUrl).
				WithField("name", rss.Name).
				WithField("email", rss.Email).
				Warn("failed to get rss feed")
			continue
		}

		inputFeeds = append(inputFeeds, feed)
	}

	allItems := make([]*dto.RssFeedItem, 0, 5*len(inputFeeds))
	for _, feed := range inputFeeds {
		if feed.Channel.Ttl > 0 && feed.Channel.Ttl < ttl {
			ttl = feed.Channel.Ttl
		}
		allItems = append(allItems, feed.Channel.Items...)
	}

	if ttl == math.MaxInt64 {
		ttl = defaultTtl
	}

	sort.SliceStable(allItems, func(i, j int) bool {
		return getTimestamp(allItems[i].PubDate) > getTimestamp(allItems[j].PubDate)
	})

	return &dto.RssFeed{
		XMLName: xml.Name{
			Local: "rss",
		},
		Channel: &dto.RssFeedChannel{
			Title:         "RSS Aggregator",
			Description:   "Aggregated feed from different rss sources.",
			LastBuildDate: time.Now().Format(time.RFC1123),
			Ttl:           ttl,
			Items:         allItems,
		},
	}
}

func getTimestamp(pubDate string) int64 {
	t, err := time.Parse(time.RFC1123, pubDate)
	if err != nil {
		return 0
	}

	return t.Unix()
}
