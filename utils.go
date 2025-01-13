package main

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"
)

// wrapText wraps text to a given line width.
func wrapText(text string, width int) string {
	var sb strings.Builder
	var lineLen int

	for _, word := range strings.Fields(text) {
		wordLen := utf8.RuneCountInString(word)
		if lineLen+wordLen+1 > width {
			sb.WriteString("\n")
			lineLen = 0
		}
		sb.WriteString(word + " ")
		lineLen += wordLen + 1
	}
	return sb.String()
}

// romanNumeral converts an integer to its Roman numeral representation.
func romanNumeral(n int) string {
	romans := []struct {
		Value   int
		Numeral string
	}{
		{1000, "M"}, {900, "CM"}, {500, "D"}, {400, "CD"},
		{100, "C"}, {90, "XC"}, {50, "L"}, {40, "XL"},
		{10, "X"}, {9, "IX"}, {5, "V"}, {4, "IV"}, {1, "I"},
	}
	var result strings.Builder
	for _, roman := range romans {
		for n >= roman.Value {
			result.WriteString(roman.Numeral)
			n -= roman.Value
		}
	}
	return result.String()
}

// loadFromCache copies the cached news from app.newsCache into m.newsList.
func (m *model) loadFromCache() {
	m.app.newsCacheMux.RLock()
	defer m.app.newsCacheMux.RUnlock()

	m.newsList = make([]newsItem, len(m.app.newsCache))
	copy(m.newsList, m.app.newsCache)
}

// loadWelcomeArt reads the welcome art from a file.
func loadWelcomeArt() (string, error) {
	content, err := os.ReadFile(welcomeArtPath)
	if err != nil {
		return "", fmt.Errorf("failed to load welcome art: %w", err)
	}
	return string(content), nil
}

// getEnvOrDefault fetches an environment variable or returns a default value.
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
