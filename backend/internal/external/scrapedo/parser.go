package scrapedo

import (
	"regexp"
	"strconv"
	"strings"
)

const minEntriesForSuccess = 10

var (
	namePattern     = `([A-Z][a-z]+(?:[-'][A-Z]?[a-z]+)*(?:\s+(?:[A-Z]\.?\s*)?[A-Z][a-z]+(?:[-'][A-Z]?[a-z]+)*)+)`
	earningsPattern = `\$\s*([\d,]+(?:\.\d{2})?)`

	nameEarningsPattern = regexp.MustCompile(namePattern + `[^$]*?` + earningsPattern)

	tableRowPattern = regexp.MustCompile(`(?i)(?:T?\d+|CUT|WD|DQ)\s+` + namePattern + `[^$]*?` + earningsPattern)
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
	entries := make([]ParsedEntry, 0, len(matches))
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		name := cleanName(match[1])
		earnings := parseEarnings(match[2])
		if name == "" || earnings <= 0 {
			continue
		}
		if seen[name] {
			continue
		}
		seen[name] = true
		entries = append(entries, ParsedEntry{
			Name:     name,
			Earnings: earnings,
		})
	}

	return entries
}

func tryNameEarningsPattern(content string) []ParsedEntry {
	matches := nameEarningsPattern.FindAllStringSubmatch(content, -1)
	entries := make([]ParsedEntry, 0, len(matches))
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		name := cleanName(match[1])
		earnings := parseEarnings(match[2])
		if name == "" || earnings <= 0 {
			continue
		}
		if seen[name] {
			continue
		}
		seen[name] = true
		entries = append(entries, ParsedEntry{
			Name:     name,
			Earnings: earnings,
		})
	}

	return entries
}

func cleanName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "  ", " ")
	return name
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
