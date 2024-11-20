package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const baseURL = "https://newsapi.org/v2/top-headlines"

var apiKey string

// Article represents the structure of a news article
type Article struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

// NewsAPIResponse represents the structure of the response from the News API
type NewsAPIResponse struct {
	Status   string    `json:"status"`
	Articles []Article `json:"articles"`
}

func getHeadlines() []Article {
	return fetchArticles("country=us")
}

func searchArticles(query string) []Article {
	return fetchArticles(fmt.Sprintf("q=%s", query))
}

func fetchArticles(query string) []Article {
	url := fmt.Sprintf("%s?%s&apiKey=%s", baseURL, query, apiKey)
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Println("Error fetching news:", err)
		return nil
	}
	defer resp.Body.Close()

	var result NewsAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("Error decoding response:", err)
		return nil
	}

	if result.Status != "ok" {
		fmt.Println("Error fetching news:", result.Status)
		return nil
	}

	return result.Articles
}

func displayHeadlines(headlines []Article) {
	fmt.Println("\nLatest News Headlines:")
	for i, article := range headlines {
		if article.Description != "" {
			fmt.Printf("%d*. %s\n", i+1, article.Title)
		} else {
			fmt.Printf("%d. %s\n", i+1, article.Title)
		}
	}
}

func displayArticleContent(headlines []Article, articleNumber int) {
	if articleNumber < 1 || articleNumber > len(headlines) {
		fmt.Println("Invalid article number.")
		return
	}

	article := headlines[articleNumber-1]
	fmt.Printf(
		"\nTitle: %s\nDescription: %s\nURL: %s\n",
		article.Title,
		article.Description,
		article.URL,
	)
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	apiKey = os.Getenv("NEWS_API_KEY")
	if apiKey == "" {
		log.Fatal("Please set the NEWS_API_KEY environment variable")
	}

	for {
		fmt.Println("\nSelect an option:")
		fmt.Println("1. Show latest news headlines")
		fmt.Println("2. Show article content")
		fmt.Println("3. Search by keyword")
		fmt.Println("4. Exit")

		var choice string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Enter your choice").
					Value(&choice).
					Validate(func(input string) error {
						if input != "1" && input != "2" && input != "3" && input != "4" {
							return errors.New(
								"Invalid choice, please select a number between 1 and 4",
							)
						}
						return nil
					}),
			),
		)

		if err := form.Run(); err != nil {
			log.Fatal(err)
		}

		switch choice {
		case "1":
			headlines := getHeadlines()
			displayHeadlines(headlines)
		case "2":
			var articleNumber string
			if headlines := getHeadlines(); len(headlines) > 0 {
				huh.NewForm(
					huh.NewGroup(
						huh.NewInput().
							Title("Enter the article number to view content").
							Value(&articleNumber).
							Validate(func(input string) error {
								_, err := strconv.Atoi(input)
								if err != nil {
									return errors.New("Please enter a valid article number")
								}
								return nil
							}),
					),
				).Run()
				num, _ := strconv.Atoi(articleNumber)
				displayArticleContent(headlines, num)
			} else {
				fmt.Println("No headlines loaded. Please load headlines first (select option 1).")
			}
		case "3":
			var query string
			huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Enter a keyword to search").
						Value(&query).
						Validate(func(input string) error {
							if input == "" {
								return errors.New("Please enter a keyword")
							}
							return nil
						}),
				),
			).Run()
			searchResults := searchArticles(query)
			if len(searchResults) > 0 {
				displayHeadlines(searchResults)
			} else {
				fmt.Println("No results found for the query.")
			}
		case "4":
			fmt.Println("Exiting the application.")
			return
		}
	}
}
