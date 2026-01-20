package scrapedo

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/go-resty/resty/v2"
)

const baseURL = "https://api.scrape.do/"

type Client struct {
	client *resty.Client
	token  string
	logger *slog.Logger
}

func New(apiKey string, logger *slog.Logger) *Client {
	client := resty.New()

	return &Client{
		client: client,
		token:  apiKey,
		logger: logger,
	}
}

func (c *Client) Scrape(ctx context.Context, targetURL string, opts ScrapeOptions) (*ScrapeResponse, error) {
	c.logger.Debug("scraping URL", "url", targetURL, "render", opts.Render)

	params := url.Values{}
	params.Set("token", c.token)
	params.Set("url", targetURL)
	if opts.Render {
		params.Set("render", "true")
	}

	requestURL := baseURL + "?" + params.Encode()

	resp, err := c.client.R().
		SetContext(ctx).
		Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape URL: %w", err)
	}

	c.logger.Debug("scrape complete", "status", resp.StatusCode(), "url", targetURL)
	return &ScrapeResponse{
		Content:    resp.String(),
		StatusCode: resp.StatusCode(),
		URL:        targetURL,
	}, nil
}
