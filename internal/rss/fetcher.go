//go:generate mockgen -package ${GOPACKAGE} -destination mock_fetcher.go -source fetcher.go
package rss

import (
	"encoding/xml"
	"errors"
	"net/http"

	"service-rss/internal/dto"
)

type Fetcher interface {
	Fetch(url string) (*dto.RssFeed, error)
}

type fetcher struct {
}

func NewFetcher() Fetcher {
	return &fetcher{}
}

func (f *fetcher) Fetch(url string) (*dto.RssFeed, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

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
