package fields

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

func TestCleanName_Basic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"trims whitespace", "  Tiger Woods  ", "Tiger Woods"},
		{"normalizes multiple spaces", "Tiger  Woods", "Tiger Woods"},
		{"handles tabs and newlines", "Tiger\t\nWoods", "Tiger Woods"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCleanName_RemovesAmateur(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"removes (a)", "John Smith (a)", "John Smith"},
		{"removes (amateur)", "John Smith (amateur)", "John Smith"},
		{"removes asterisk", "John Smith*", "John Smith"},
		{"removes multiple markers", "John* Smith (a)", "John Smith"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCleanName_RemovesOWGR(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"removes trailing rank", "Tiger Woods 5", "Tiger Woods"},
		{"removes rank with hash", "Tiger Woods #1", "Tiger Woods"},
		{"removes large rank", "John Doe 125", "John Doe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCleanName_TooShort(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"single char", "A", ""},
		{"two chars", "AB", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCleanName_TooLong(t *testing.T) {
	longName := strings.Repeat("A", 101)
	result := cleanName(longName)
	assert.Equal(t, "", result)
}

func TestCleanName_SingleWord(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"single word", "Tiger", ""},
		{"single word with spaces", "  Tiger  ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCleanName_ValidNames(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal name", "Tiger Woods", "Tiger Woods"},
		{"three parts", "Jon Rahm Rodriguez", "Jon Rahm Rodriguez"},
		{"hyphenated", "Si Woo Kim", "Si Woo Kim"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsAmateur_WithA(t *testing.T) {
	html := `<div class="player"><span class="name">John Smith (a)</span></div>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	sel := doc.Find(".player")

	assert.True(t, isAmateur(sel))
}

func TestIsAmateur_WithAmateur(t *testing.T) {
	html := `<div class="player"><span class="status">amateur</span><span class="name">John Smith</span></div>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	sel := doc.Find(".player")

	assert.True(t, isAmateur(sel))
}

func TestIsAmateur_WithAm(t *testing.T) {
	html := `<div class="player"><span class="status">am.</span><span class="name">John Smith</span></div>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	sel := doc.Find(".player")

	assert.True(t, isAmateur(sel))
}

func TestIsAmateur_NotAmateur(t *testing.T) {
	html := `<div class="player"><span class="name">Tiger Woods</span></div>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	sel := doc.Find(".player")

	assert.False(t, isAmateur(sel))
}

func TestIsAmateur_CaseInsensitive(t *testing.T) {
	html := `<div class="player"><span class="status">AMATEUR</span></div>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	sel := doc.Find(".player")

	assert.True(t, isAmateur(sel))
}

func TestMapEntryStatus_Confirmed(t *testing.T) {
	result := mapEntryStatus("confirmed")
	assert.Equal(t, "confirmed", string(result))
}

func TestMapEntryStatus_Alternate(t *testing.T) {
	result := mapEntryStatus("alternate")
	assert.Equal(t, "alternate", string(result))
}

func TestMapEntryStatus_Withdrawn(t *testing.T) {
	result := mapEntryStatus("withdrawn")
	assert.Equal(t, "withdrawn", string(result))
}

func TestMapEntryStatus_Pending(t *testing.T) {
	result := mapEntryStatus("pending")
	assert.Equal(t, "pending", string(result))
}

func TestMapEntryStatus_CaseInsensitive(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"CONFIRMED", "confirmed"},
		{"Alternate", "alternate"},
		{"WITHDRAWN", "withdrawn"},
		{"Pending", "pending"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := mapEntryStatus(tt.input)
			assert.Equal(t, tt.expected, string(result))
		})
	}
}

func TestMapEntryStatus_Default(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"unknown status", "unknown"},
		{"empty string", ""},
		{"garbage", "xyz123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapEntryStatus(tt.input)
			assert.Equal(t, "confirmed", string(result))
		})
	}
}

func TestExtractPlayerName_FromNameClass(t *testing.T) {
	html := `<div class="player"><span class="player-name">Tiger Woods</span></div>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	sel := doc.Find(".player")

	result := extractPlayerName(sel)
	assert.Equal(t, "Tiger Woods", result)
}

func TestExtractPlayerName_FromH3(t *testing.T) {
	html := `<div class="player"><h3>Scottie Scheffler</h3></div>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	sel := doc.Find(".player")

	result := extractPlayerName(sel)
	assert.Equal(t, "Scottie Scheffler", result)
}

func TestExtractPlayerName_FromH4(t *testing.T) {
	html := `<div class="player"><h4>Rory McIlroy</h4></div>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	sel := doc.Find(".player")

	result := extractPlayerName(sel)
	assert.Equal(t, "Rory McIlroy", result)
}

func TestExtractPlayerName_FromAnchor(t *testing.T) {
	html := `<div class="player"><a href="/player/123">Jon Rahm</a></div>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	sel := doc.Find(".player")

	result := extractPlayerName(sel)
	assert.Equal(t, "Jon Rahm", result)
}

func TestExtractPlayerName_WithAmateurMarker(t *testing.T) {
	html := `<div class="player"><span class="name">John Smith (a)</span></div>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	sel := doc.Find(".player")

	result := extractPlayerName(sel)
	assert.Equal(t, "John Smith", result)
}

func TestExtractPlayerName_Empty(t *testing.T) {
	html := `<div class="player"></div>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	sel := doc.Find(".player")

	result := extractPlayerName(sel)
	assert.Equal(t, "", result)
}

func TestExtractPlayerName_InvalidName(t *testing.T) {
	html := `<div class="player"><span class="name">X</span></div>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
	sel := doc.Find(".player")

	result := extractPlayerName(sel)
	assert.Equal(t, "", result)
}
