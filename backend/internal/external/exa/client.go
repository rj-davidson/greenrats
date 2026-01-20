package exa

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-resty/resty/v2"
)

const (
	baseURL    = "https://api.exa.ai"
	numResults = 3
)

type Client struct {
	client *resty.Client
	logger *slog.Logger
}

func New(apiKey string, logger *slog.Logger) *Client {
	client := resty.New().
		SetBaseURL(baseURL).
		SetHeader("x-api-key", apiKey).
		SetHeader("Content-Type", "application/json")

	return &Client{client: client, logger: logger}
}

func (c *Client) SearchEarnings(ctx context.Context, tournamentName string, year int) (*SearchResponse, error) {
	c.logger.Info("searching earnings", "tournament", tournamentName, "year", year)

	query := fmt.Sprintf("%d %s earnings", year, tournamentName)

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

	c.logger.Debug("exa search complete", "results", len(response.Results))
	return &response, nil
}
