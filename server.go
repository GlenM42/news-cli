package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/muesli/termenv"
	"net"
	"sync"
	"time"
)

// newsItem is used to store a headline and description in the app's cache.
type newsItem struct {
	Headline    string
	Description string
	Link        string
}

// app wraps the Wish server and tracks active Bubble Tea programs.
type app struct {
	*ssh.Server
	progs []*tea.Program

	newsCache    []newsItem   // The cached news
	newsCacheMux sync.RWMutex // Protects access to newsCache
}

// send broadcasts a message to all running Bubble Tea programs.
func (a *app) send(msg tea.Msg) {
	for _, p := range a.progs {
		go p.Send(msg)
	}
}

// newApp initializes a new SSH server with the desired middleware and returns an app struct.
func newApp() *app {
	a := new(app)

	// Fetch SSH address and port from environment variables
	host := getEnvOrDefault("SSH_HOST", "localhost")
	port := getEnvOrDefault("SSH_PORT", "23234")

	address := net.JoinHostPort(host, port)

	s, err := wish.NewServer(
		wish.WithAddress(address),
		wish.WithHostKeyPath(".ssh/id_ed25519"),

		// Public-Key Authentication
		wish.WithPublicKeyAuth(publicKeyAuth),

		// Keyboard-Interactive Authentication
		wish.WithKeyboardInteractiveAuth(keyboardInteractiveAuth),

		// Attach middlewares
		wish.WithMiddleware(
			bubbletea.MiddlewareWithProgramHandler(a.ProgramHandler, termenv.ANSI256),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("Could not start server", "error", err)
	}
	a.Server = s

	// Start the background goroutine that updates the news cache every 15 minutes
	go a.startNewsCacheUpdater("us")

	return a
}

// ProgramHandler is called whenever a new SSH session starts. It creates a new Bubble Tea program.
func (a *app) ProgramHandler(s ssh.Session) *tea.Program {
	m := initialModel()
	m.app = a
	m.id = s.User()

	// Create a new Bubble Tea program
	p := tea.NewProgram(m, bubbletea.MakeOptions(s)...)

	// Add this program to our active programs list
	a.progs = append(a.progs, p)
	return p
}

// startNewsCacheUpdater periodically fetches news from NewsAPI and updates the cache.
// This function runs in a separate goroutine. With a 15-minute interval, you'll make up to
// 96 requests per day, staying within most free-tier limits.
func (a *app) startNewsCacheUpdater(country string) {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	// Initial fetch right away
	a.updateNewsCache(country)

	for range ticker.C {
		a.updateNewsCache(country)
	}
}

// updateNewsCache calls LoadNews(...) and saves the result to our local cache.
func (a *app) updateNewsCache(country string) {
	articles, err := LoadNews(country) // This should be defined in news.go or similar
	if err != nil {
		log.Error("Failed to fetch news from NewsAPI", "error", err)
		return
	}

	// Convert []NewsArticle to []newsItem
	newItems := make([]newsItem, 0, len(articles))
	for _, art := range articles {
		newItems = append(newItems, newsItem{
			Headline:    art.Title,
			Description: art.Description,
			Link:        art.URL,
		})
	}

	// Lock and replace the cache
	a.newsCacheMux.Lock()
	a.newsCache = newItems
	a.newsCacheMux.Unlock()

	log.Info("News cache updated", "count", len(newItems))
}
