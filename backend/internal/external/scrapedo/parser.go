package scrapedo

import (
	"regexp"
	"strconv"
	"strings"
)

const (
	minEntriesForSuccess = 10
	minEarningsForValid  = 1000
)

var (
	namePattern     = `([A-Z][a-z]+(?:[-'][A-Z]?[a-z]+)*(?:\s+(?:[A-Z]\.?\s*)?[A-Z][a-z]+(?:[-'][A-Z]?[a-z]+)*)+)`
	earningsPattern = `\$\s*([\d,]+(?:\.\d{2})?)`

	nameEarningsPattern = regexp.MustCompile(namePattern + `[^$]{0,500}` + earningsPattern)

	tableRowPattern = regexp.MustCompile(`(?i)(?:T?\d+|CUT|WD|DQ)\s+` + namePattern + `[^$]{0,500}` + earningsPattern)

	invalidNamePattern = regexp.MustCompile(
		`(?i)\b(Corp|Inc|Ltd|LLC|GmbH|Plc|Intl)$`,
	)
)

func ParseLeaderboard(content string) *ParseResult {
	entries := tryTableRowPattern(content)
	if len(entries) >= minEntriesForSuccess {
		return &ParseResult{
			Success: true,
			Entries: entries,
		}
	}

	entries = tryNameEarningsPattern(content)
	if len(entries) >= minEntriesForSuccess {
		return &ParseResult{
			Success: true,
			Entries: entries,
		}
	}

	return &ParseResult{
		Success: false,
		Entries: entries,
	}
}

func tryTableRowPattern(content string) []ParsedEntry {
	matches := tableRowPattern.FindAllStringSubmatch(content, -1)
	candidates := make(map[string][]EarningsCandidate)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		name := cleanName(match[1])
		earnings := parseEarnings(match[2])
		if name == "" || !isValidGolferName(name) || earnings < minEarningsForValid {
			continue
		}

		context := extractContext(content, match[0], 500)
		candidates[name] = append(candidates[name], EarningsCandidate{
			Earnings: earnings,
			Context:  context,
		})
	}

	return candidatesToEntries(candidates)
}

func tryNameEarningsPattern(content string) []ParsedEntry {
	matches := nameEarningsPattern.FindAllStringSubmatch(content, -1)
	candidates := make(map[string][]EarningsCandidate)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		name := cleanName(match[1])
		earnings := parseEarnings(match[2])
		if name == "" || !isValidGolferName(name) || earnings < minEarningsForValid {
			continue
		}

		context := extractContext(content, match[0], 500)
		candidates[name] = append(candidates[name], EarningsCandidate{
			Earnings: earnings,
			Context:  context,
		})
	}

	return candidatesToEntries(candidates)
}

func cleanName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "  ", " ")
	return name
}

func isValidGolferName(name string) bool {
	parts := strings.Fields(strings.TrimSpace(name))
	if len(parts) < 2 || len(parts) > 4 {
		return false
	}
	return !invalidNamePattern.MatchString(name)
}

func parseEarnings(s string) int {
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "$", "")
	s = strings.TrimSpace(s)

	if idx := strings.Index(s, "."); idx != -1 {
		s = s[:idx]
	}

	val, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return val
}

func extractContext(content, matchStr string, contextLen int) string {
	idx := strings.Index(content, matchStr)
	if idx == -1 {
		return matchStr
	}
	start := max(0, idx-contextLen)
	end := min(len(content), idx+len(matchStr)+contextLen)
	return content[start:end]
}

func candidatesToEntries(candidates map[string][]EarningsCandidate) []ParsedEntry {
	entries := make([]ParsedEntry, 0, len(candidates))
	for name, cands := range candidates {
		if len(cands) == 1 {
			entries = append(entries, ParsedEntry{
				Name:     name,
				Earnings: cands[0].Earnings,
			})
		} else {
			uniqueCands := deduplicateCandidates(cands)
			if len(uniqueCands) == 1 {
				entries = append(entries, ParsedEntry{
					Name:     name,
					Earnings: uniqueCands[0].Earnings,
				})
			} else {
				entries = append(entries, ParsedEntry{
					Name:       name,
					Earnings:   0,
					Candidates: uniqueCands,
				})
			}
		}
	}
	return entries
}

func deduplicateCandidates(cands []EarningsCandidate) []EarningsCandidate {
	seen := make(map[int]bool)
	unique := make([]EarningsCandidate, 0, len(cands))
	for _, c := range cands {
		if !seen[c.Earnings] {
			seen[c.Earnings] = true
			unique = append(unique, c)
		}
	}
	return unique
}
