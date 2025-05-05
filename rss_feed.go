package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"
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

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)

	if err != nil {

		return nil, fmt.Errorf("the request failed: %v\n", err)
	}

	req.Header.Set("User-Agent", "gator")

	client := http.Client{Timeout: 30 * time.Second}

	res, err := client.Do(req)

	if err != nil {
		fmt.Printf("the request failed for %s with the error: %s", feedURL, err)
		return nil, err
	}
	defer res.Body.Close()

	feed := RSSFeed{}

	data, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, fmt.Errorf("failed to read the body of the request: %v\n", err)
	}

	if err := xml.Unmarshal(data, &feed); err != nil {
		fmt.Printf("the xml file wasnt decoded properly: %s", err)
		return nil, err
	}

	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	for i, item := range feed.Channel.Item {
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
		feed.Channel.Item[i] = item
	}

	return &feed, nil

}
