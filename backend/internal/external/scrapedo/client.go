package scrapedo

import (
	"context"
	"fmt"
	"net/url"

	"github.com/go-resty/resty/v2"
)

const baseURL = "https://api.scrape.do/"

type Client struct {
	client *resty.Client
	token  string
}

func New(apiKey string) *Client {
	client := resty.New()

	return &Client{
		client: client,
		token:  apiKey,
	}
}

func (c *Client) Scrape(ctx context.Context, targetURL string, opts ScrapeOptions) (*ScrapeResponse, error) {
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

	return &ScrapeResponse{
		Content:    resp.String(),
		StatusCode: resp.StatusCode(),
		URL:        targetURL,
	}, nil
}
