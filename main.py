import os

import requests
from dotenv import load_dotenv

load_dotenv()

# Load the API key from an environment variable
API_KEY = os.getenv("NEWS_API_KEY")
BASE_URL = "https://newsapi.org/v2/top-headlines"

if not API_KEY:
    raise ValueError("Please set the NEWS_API_KEY environment variable")


def get_headlines(country="us"):
    """Fetch the latest news headlines."""
    response = requests.get(BASE_URL, params={"apiKey": API_KEY, "country": country})
    if response.status_code == 200:
        articles = response.json().get("articles", [])
        headlines = [
            (i + 1, article["title"], article["description"], article["url"])
            for i, article in enumerate(articles)
        ]
        return headlines


def display_headlines(headlines):
    """Display news headlines with numbering."""
    print("\nLatest News Headlines:\n")
    for index, title, _, _ in headlines:
        print(f"{index}. {title}")


def display_article_content(headlines, article_number):
    """Show content of a selected article by its number."""
    try:
        _, title, description, url = headlines[article_number - 1]
        print(f"\nTitle: {title}\nDescription: {description}\nURL: {url}\n")
    except IndexError:
        print("Invalid article number.")


def search_articles(query):
    """Search for articles by keyword."""
    response = requests.get(BASE_URL, params={"apiKey": API_KEY, "q": query})
    if response.status_code == 200:
        articles = response.json().get("articles", [])
        headlines = [
            (i + 1, article["title"], article["description"], article["url"])
            for i, article in enumerate(articles)
        ]
        return headlines
    else:
        print("Error searching for articles:", response.status_code)
        return []


def main():
    """Main function to interact with the news application."""
    while True:
        print("\nSelect an option:")
        print("1. Show latest news headlines")
        print("2. Show article content")
        print("3. Search by keyword")
        print("4. Exit")

        choice = input("Enter your choice: ")

        if choice == "1":
            headlines = get_headlines()
            display_headlines(headlines)
        elif choice == "2":
            if "headlines" not in locals():
                print("Please load headlines first (select option 1)")
            else:
                article_number = int(
                    input("Enter the article number to view content: ")
                )
                display_article_content(headlines, article_number)
        elif choice == "3":
            query = input("Enter a keyword to search: ")
            search_results = search_articles(query)
            if search_results:
                display_headlines(search_results)
            else:
                print("No results found for the query.")
        elif choice == "4":
            print("Exiting the application.")
            break
        else:
            print("Invalid choice, please try again.")


if __name__ == "__main__":
    main()
