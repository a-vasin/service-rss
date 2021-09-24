package rss

import (
	"encoding/xml"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"service-rss/internal/database"
	"service-rss/internal/dto"
)

var (
	data = map[string]*dto.RssFeed{
		"https://one.com/": {
			Channel: &dto.RssFeedChannel{
				Ttl: 30,
				Items: []*dto.RssFeedItem{
					{
						Title:   "second",
						PubDate: "Mon, 02 Jan 2006 15:04:06 MST",
					},
					{
						Title: "fourth",
					},
				},
			},
		},
		"https://two.com/": {
			Channel: &dto.RssFeedChannel{
				Ttl: 25,
				Items: []*dto.RssFeedItem{
					{
						Title:   "third",
						PubDate: "Mon, 02 Jan 2006 15:04:05 MST",
					},
				},
			},
		},
		"https://three.com/": {
			Channel: &dto.RssFeedChannel{
				Ttl: 10,
				Items: []*dto.RssFeedItem{
					{
						Title:   "first",
						PubDate: "Mon, 02 Jan 2006 15:04:07 MST",
					},
				},
			},
		},
	}

	expectedOkFeed = &dto.RssFeed{
		XMLName: xml.Name{
			Local: "rss",
		},
		Channel: &dto.RssFeedChannel{
			Title:       "RSS Aggregator",
			Description: "Aggregated feed from different rss sources.",
			Ttl:         10,
			Items: []*dto.RssFeedItem{
				{
					Title:   "first",
					PubDate: "Mon, 02 Jan 2006 15:04:07 MST",
				},
				{
					Title:   "second",
					PubDate: "Mon, 02 Jan 2006 15:04:06 MST",
				},
				{
					Title:   "third",
					PubDate: "Mon, 02 Jan 2006 15:04:05 MST",
				},

				{
					Title: "fourth",
				},
			},
		},
	}

	expectedErrorFeed = &dto.RssFeed{
		XMLName: xml.Name{
			Local: "rss",
		},
		Channel: &dto.RssFeedChannel{
			Title:       "RSS Aggregator",
			Description: "Aggregated feed from different rss sources.",
			Ttl:         defaultTtl,
			Items:       []*dto.RssFeedItem{},
		},
	}
)

func TestBuilder_Build(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("base scenario", func(t *testing.T) {
		f := NewMockFetcher(ctrl)
		for url, feed := range data {
			f.EXPECT().Fetch(url).Return(feed, nil)
		}

		b := NewAggregator(f)

		rss := &database.Rss{
			Name:    "test",
			Sources: []string{"https://one.com/", "https://two.com/", "https://three.com/"},
		}
		feed := b.Aggregate(rss)

		buildDate := feed.Channel.LastBuildDate
		_, err := time.Parse(time.RFC1123, buildDate)
		assert.NoError(t, err)

		feed.Channel.LastBuildDate = ""

		assert.Equal(t, expectedOkFeed, feed)
	})

	t.Run("with error", func(t *testing.T) {
		f := NewMockFetcher(ctrl)
		f.EXPECT().Fetch(gomock.Any()).Return(nil, errors.New("error"))

		b := NewAggregator(f)

		rss := &database.Rss{
			Name:    "test",
			Sources: []string{"https://one.com/"},
		}
		feed := b.Aggregate(rss)

		feed.Channel.LastBuildDate = ""

		assert.Equal(t, expectedErrorFeed, feed)
	})
}
