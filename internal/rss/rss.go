package rss

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func FetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	// create a new request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("failed to create request: %w", err)
	}

	// apparently, closing the request body causes an error. maybe because
	// the request body is nil.
	// defer req.Body.Close()

	// set user-agent header to the one requesting the data, which is
	// the database name "gator"

	// this is a common practice to identify the program to the server
	req.Header.Set("User-Agent", "gator")

	// send/process the request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return &RSSFeed{}, err
	}
	defer res.Body.Close()

	// parse response body
	feed := &RSSFeed{}
	if err := xml.NewDecoder(res.Body).Decode(feed); err != nil {
		return &RSSFeed{}, fmt.Errorf("failed to decode feed: %w", err)
	}

	return feed, nil
}
