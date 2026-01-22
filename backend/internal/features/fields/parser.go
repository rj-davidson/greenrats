package fields

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/rj-davidson/greenrats/internal/external/pgatour"
	"github.com/rj-davidson/greenrats/internal/external/scrapedo"
)

type MajorURLs struct {
	Masters   string
	USOpen    string
	OpenChamp string
	PGA       string
}

var DefaultMajorURLs = MajorURLs{
	Masters:   "https://www.masters.com/en_US/players/index.html",
	USOpen:    "https://www.usopen.com/players.html",
	OpenChamp: "https://www.theopen.com/the-field",
	PGA:       "https://www.pgachampionship.com/field",
}

type MajorScraper struct {
	scraper scrapedo.ClientInterface
	logger  *slog.Logger
}

func NewMajorScraper(scraper scrapedo.ClientInterface, logger *slog.Logger) *MajorScraper {
	return &MajorScraper{
		scraper: scraper,
		logger:  logger,
	}
}

func (s *MajorScraper) ScrapeField(ctx context.Context, tournamentName, url string) ([]pgatour.FieldEntry, error) {
	s.logger.Info("scraping major tournament field", "tournament", tournamentName, "url", url)

	resp, err := s.scraper.Scrape(ctx, url, scrapedo.ScrapeOptions{Render: true})
	if err != nil {
		return nil, fmt.Errorf("failed to scrape %s: %w", url, err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("scrape returned status %d for %s", resp.StatusCode, url)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.Content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	name := strings.ToLower(tournamentName)
	switch {
	case strings.Contains(name, "masters"):
		return s.parseMastersField(doc)
	case strings.Contains(name, "u.s. open") || strings.Contains(name, "us open"):
		return s.parseUSOpenField(doc)
	case strings.Contains(name, "open championship") || strings.Contains(name, "the open"):
		return s.parseOpenChampionshipField(doc)
	case strings.Contains(name, "pga championship"):
		return s.parsePGAChampionshipField(doc)
	default:
		return s.parseGenericField(doc)
	}
}

func (s *MajorScraper) parseMastersField(doc *goquery.Document) ([]pgatour.FieldEntry, error) {
	var entries []pgatour.FieldEntry

	doc.Find(".player-list-item, .player-card, [class*='player']").Each(func(_ int, sel *goquery.Selection) {
		name := extractPlayerName(sel)
		if name == "" {
			return
		}

		entry := pgatour.FieldEntry{
			DisplayName: name,
		}

		if country, exists := sel.Attr("data-country"); exists {
			entry.CountryCode = country
		}

		entry.IsAmateur = isAmateur(sel)

		entries = append(entries, entry)
	})

	s.logger.Debug("parsed Masters field", "count", len(entries))
	return entries, nil
}

func (s *MajorScraper) parseUSOpenField(doc *goquery.Document) ([]pgatour.FieldEntry, error) {
	var entries []pgatour.FieldEntry

	doc.Find(".player-item, .field-player, [class*='player']").Each(func(_ int, sel *goquery.Selection) {
		name := extractPlayerName(sel)
		if name == "" {
			return
		}

		entry := pgatour.FieldEntry{
			DisplayName: name,
		}

		entry.IsAmateur = isAmateur(sel)

		entries = append(entries, entry)
	})

	s.logger.Debug("parsed US Open field", "count", len(entries))
	return entries, nil
}

func (s *MajorScraper) parseOpenChampionshipField(doc *goquery.Document) ([]pgatour.FieldEntry, error) {
	var entries []pgatour.FieldEntry

	doc.Find(".player, .field-player, [class*='player']").Each(func(_ int, sel *goquery.Selection) {
		name := extractPlayerName(sel)
		if name == "" {
			return
		}

		entry := pgatour.FieldEntry{
			DisplayName: name,
		}

		if country := sel.Find(".flag, .country, [class*='country']").Text(); country != "" {
			entry.CountryCode = strings.TrimSpace(country)
		}

		entry.IsAmateur = isAmateur(sel)

		entries = append(entries, entry)
	})

	s.logger.Debug("parsed Open Championship field", "count", len(entries))
	return entries, nil
}

func (s *MajorScraper) parsePGAChampionshipField(doc *goquery.Document) ([]pgatour.FieldEntry, error) {
	var entries []pgatour.FieldEntry

	doc.Find(".player-card, .field-player, [class*='player']").Each(func(_ int, sel *goquery.Selection) {
		name := extractPlayerName(sel)
		if name == "" {
			return
		}

		entry := pgatour.FieldEntry{
			DisplayName: name,
			IsAmateur:   isAmateur(sel),
		}

		entries = append(entries, entry)
	})

	s.logger.Debug("parsed PGA Championship field", "count", len(entries))
	return entries, nil
}

func (s *MajorScraper) parseGenericField(doc *goquery.Document) ([]pgatour.FieldEntry, error) {
	var entries []pgatour.FieldEntry

	doc.Find("[class*='player'], [class*='golfer'], [class*='field'] li, [class*='field'] tr").Each(func(_ int, sel *goquery.Selection) {
		name := extractPlayerName(sel)
		if name == "" {
			return
		}

		entry := pgatour.FieldEntry{
			DisplayName: name,
			IsAmateur:   isAmateur(sel),
		}

		entries = append(entries, entry)
	})

	s.logger.Debug("parsed generic field", "count", len(entries))
	return entries, nil
}

func extractPlayerName(sel *goquery.Selection) string {
	nameSelectors := []string{
		".player-name",
		".name",
		"h3",
		"h4",
		".first-name + .last-name",
		"[class*='name']",
		"a",
	}

	for _, selector := range nameSelectors {
		if name := sel.Find(selector).First().Text(); name != "" {
			return cleanName(name)
		}
	}

	text := sel.Text()
	if text != "" {
		return cleanName(text)
	}

	return ""
}

var (
	nameCleanupPattern = regexp.MustCompile(`\s+`)
	amateurPattern     = regexp.MustCompile(`\(a\)|\(amateur\)|\*`)
	owgrPattern        = regexp.MustCompile(`#?\d+$`)
)

func cleanName(name string) string {
	name = strings.TrimSpace(name)
	name = amateurPattern.ReplaceAllString(name, "")
	name = owgrPattern.ReplaceAllString(name, "")
	name = nameCleanupPattern.ReplaceAllString(name, " ")
	name = strings.TrimSpace(name)

	if len(name) < 3 || len(name) > 100 {
		return ""
	}

	words := strings.Fields(name)
	if len(words) < 2 {
		return ""
	}

	return name
}

func isAmateur(sel *goquery.Selection) bool {
	html, _ := sel.Html()
	text := sel.Text()
	combined := strings.ToLower(html + text)

	return strings.Contains(combined, "(a)") ||
		strings.Contains(combined, "amateur") ||
		strings.Contains(combined, "am.")
}
