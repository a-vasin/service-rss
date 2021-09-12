package rss

import (
	"encoding/xml"
	"math"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"

	"service-rss/internal/dto"
)

const (
	defaultTtl = 5 // rss ttl is in minutes according to specification
)

type Builder interface {
	Build(name string, sources []string) *dto.RssFeed
}

type builder struct {
	fetcher Fetcher
}

func NewBuilder(fetcher Fetcher) Builder {
	return &builder{
		fetcher: fetcher,
	}
}

func (b *builder) Build(name string, sources []string) *dto.RssFeed {
	ttl := int64(math.MaxInt64)
	inputFeeds := make([]*dto.RssFeed, 0, len(sources))
	for _, rssUrl := range sources {
		feed, err := b.fetcher.Fetch(rssUrl)
		if err != nil {
			ttl = defaultTtl
			log.WithError(err).WithField("url", rssUrl).WithField("name", name).Warn("failed to get rss feed")
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
