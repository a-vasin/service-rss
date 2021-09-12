package dto

import "encoding/xml"

type ErrorResponse struct {
	Error string `json:"error"`
	Value string `json:"value,omitempty"`
}

type RssFeed struct {
	XMLName xml.Name        `xml:"rss"`
	Channel *RssFeedChannel `xml:"channel"`
}

type RssFeedChannel struct {
	Title         string         `xml:"title"`
	Link          string         `xml:"link"`
	Description   string         `xml:"description"`
	LastBuildDate string         `xml:"lastBuildDate,omitempty"`
	Ttl           int64          `xml:"ttl,omitempty"`
	Items         []*RssFeedItem `xml:"item"`
}

type RssFeedItem struct {
	Title       string   `xml:"title,omitempty"`
	Link        string   `xml:"link,omitempty"`
	Description string   `xml:"description,omitempty"`
	Author      string   `xml:"author,omitempty"`
	Category    []string `xml:"category,omitempty"`
	Comments    string   `xml:"comments,omitempty"`
	Enclosure   string   `xml:"enclosure,omitempty"`
	Guid        string   `xml:"guid,omitempty"`
	PubDate     string   `xml:"pubDate,omitempty"`
	Source      string   `xml:"source,omitempty"`
}
