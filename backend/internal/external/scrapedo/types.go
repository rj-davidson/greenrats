package scrapedo

type ScrapeOptions struct {
	Render bool
}

type ScrapeResponse struct {
	Content    string
	StatusCode int
	URL        string
}

type ParseResult struct {
	Success bool
	Entries []ParsedEntry
}

type ParsedEntry struct {
	Name     string
	Earnings int
}
