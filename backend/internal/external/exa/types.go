package exa

type SearchRequest struct {
	Query          string   `json:"query"`
	IncludeDomains []string `json:"includeDomains,omitempty"`
	IncludeText    []string `json:"includeText,omitempty"`
	NumResults     int      `json:"numResults"`
	Type           string   `json:"type"`
	Contents       Contents `json:"contents"`
}

type Contents struct {
	Text      bool   `json:"text"`
	LiveCrawl string `json:"livecrawl"`
}

type SearchResponse struct {
	Results []Result `json:"results"`
}

type Result struct {
	Title string  `json:"title"`
	URL   string  `json:"url"`
	Text  string  `json:"text"`
	Score float64 `json:"score"`
}
