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

type EarningsCandidate struct {
	Earnings int
	Context  string
}

type ParsedEntry struct {
	Name       string
	Earnings   int
	Candidates []EarningsCandidate
}
