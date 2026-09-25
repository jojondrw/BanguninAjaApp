package news

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	googleNewsRSSURL = "https://news.google.com/rss/search"
	maxArticles      = 5
)

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type rssFeed struct {
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Items []rssItem `xml:"item"`
}

type rssItem struct {
	Title     string    `xml:"title"`
	Link      string    `xml:"link"`
	PubDate   string    `xml:"pubDate"`
	Source    rssSource `xml:"source"`
}

type rssSource struct {
	Name string `xml:",chardata"`
}

func (c *Client) Fetch(ctx context.Context, region, district string) ([]NewsResponse, error) {
	parts := []string{
		strings.TrimSpace(region),
		strings.TrimSpace(district),
		"when:7d",
	}

	query := strings.Join(parts, " ")

	params := url.Values{}
	params.Set("q", query)
	params.Set("hl", "id")
	params.Set("gl", "ID")
	params.Set("ceid", "ID:id")

	reqURL := googleNewsRSSURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "BanguninAja/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google news returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var feed rssFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, err
	}

	limit := len(feed.Channel.Items)
	if limit > maxArticles {
		limit = maxArticles
	}

	result := make([]NewsResponse, 0, limit)

	for _, item := range feed.Channel.Items[:limit] {
		result = append(result, NewsResponse{
			Title:       strings.TrimSpace(item.Title),
			Source:      strings.TrimSpace(item.Source.Name),
			PublishedAt: strings.TrimSpace(item.PubDate),
			URL:         strings.TrimSpace(item.Link),
		})
	}

	return result, nil
}