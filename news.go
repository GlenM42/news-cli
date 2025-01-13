package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// NewsAPIResponse represents the top-level JSON response from NewsAPI
type NewsAPIResponse struct {
	Status       string        `json:"status"`
	TotalResults int           `json:"totalResults"`
	Articles     []NewsArticle `json:"articles"`
}

// NewsArticle represents a single news item in the NewsAPI response
type NewsArticle struct {
	Source      NewsSource `json:"source"`
	Author      string     `json:"author"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	URL         string     `json:"url"`
	URLToImage  string     `json:"urlToImage"`
	PublishedAt string     `json:"publishedAt"`
	Content     string     `json:"content"`
}

// NewsSource holds article source information
type NewsSource struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// LoadNews retrieves the top headlines from NewsAPI
func LoadNews(country string) ([]NewsArticle, error) {
	// Get the API key from the environment variables
	apiKey := os.Getenv("NEWS_API_KEY")
	if apiKey == "" {
		return nil, errors.New("NEWS_API_KEY is not set in .env file")
	}

	// Build the request URL (top-headlines endpoint)
	url := fmt.Sprintf("https://newsapi.org/v2/top-headlines?country=%s&apiKey=%s", country, apiKey)

	// Create an HTTP client with a timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Prepare the request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch news: status code %d", resp.StatusCode)
	}

	// Read and parse JSON response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var newsResp NewsAPIResponse
	if err = json.Unmarshal(body, &newsResp); err != nil {
		return nil, err
	}

	// Check the status field returned by News API
	if newsResp.Status != "ok" {
		return nil, fmt.Errorf("NewsAPI returned status: %s", newsResp.Status)
	}

	// Return the articles
	return newsResp.Articles, nil
}
