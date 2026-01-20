package exa

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
)

const (
	baseURL    = "https://api.exa.ai"
	numResults = 3
)

type Client struct {
	client *resty.Client
}

func New(apiKey string) *Client {
	client := resty.New().
		SetBaseURL(baseURL).
		SetHeader("x-api-key", apiKey).
		SetHeader("Content-Type", "application/json")

	return &Client{client: client}
}

func (c *Client) SearchEarnings(ctx context.Context, tournamentName string, year int) (*SearchResponse, error) {
	query := fmt.Sprintf("%d %s leaderboard earnings results", year, tournamentName)

	req := SearchRequest{
		Query:          query,
		IncludeDomains: []string{"pgatour.com", "golf.com"},
		IncludeText:    []string{"earnings"},
		NumResults:     numResults,
		Type:           "auto",
		Contents: Contents{
			Text:      true,
			LiveCrawl: "preferred",
		},
	}

	var response SearchResponse
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&response).
		Post("/search")
	if err != nil {
		return nil, fmt.Errorf("failed to search exa: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("exa API error: %s - %s", resp.Status(), resp.String())
	}

	return &response, nil
}
