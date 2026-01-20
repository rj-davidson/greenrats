package scrapedo

import "context"

type ClientInterface interface {
	Scrape(ctx context.Context, targetURL string, opts ScrapeOptions) (*ScrapeResponse, error)
}

var _ ClientInterface = (*Client)(nil)
