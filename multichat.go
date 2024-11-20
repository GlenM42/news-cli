package main

import (
	"context"
	// "fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/muesli/termenv"
)

const (
	host = "localhost"
	port = "23234"
)

// app contains a wish server and the list of running programs.
type app struct {
	*ssh.Server
	progs []*tea.Program
}

type tab int

const (
	tabNews tab = iota
	tabChat
)

// send dispatches a message to all running programs.
func (a *app) send(msg tea.Msg) {
	for _, p := range a.progs {
		go p.Send(msg)
	}
}

func newApp() *app {
	a := new(app)
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
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
	return a
}

func (a *app) Start() {
	var err error
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting SSH server", "host", host, "port", port)
	go func() {
		if err = a.ListenAndServe(); err != nil {
			log.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := a.Shutdown(ctx); err != nil {
		log.Error("Could not stop server", "error", err)
	}
}

func (a *app) ProgramHandler(s ssh.Session) *tea.Program {
	model := initialModel() // Create the initial model (TUI state)
	model.app = a           // Link the app to the model
	model.id = s.User()     // Use the SSH username as the user ID

	p := tea.NewProgram(model, bubbletea.MakeOptions(s)...) // Create a new BubbleTea program
	a.progs = append(
		a.progs,
		p,
	) // add the program to the list of active programs

	return p // Return the program to Wish (so it can run the TUI for this session)
}

func main() {
	app := newApp()
	app.Start()
}

type (
	errMsg  error
	chatMsg struct {
		id   string
		text string
	}
)

type model struct {
	*app
	activeTab      tab
	viewport       viewport.Model
	messages       []string
	id             string
	textarea       textarea.Model
	senderStyle    lipgloss.Style
	unreadMessages bool
	err            error
}

// Styles
var (
	tabStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("69")).
			Padding(0, 1)

	activeTabStyle = tabStyle.Copy().
			Foreground(lipgloss.Color("205")).
			Underline(true)

	contentStyle = lipgloss.NewStyle().
			Padding(1, 2)
)

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Send a message in 280 charachters..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)
	vp.SetContent(`Welcome to the News tab!
Switch to Chat to interact with others by pressing Tab.
Type a message and press Enter to send.`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return model{
		textarea:    ta,
		messages:    []string{},
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		activeTab:   tabNews,
		err:         nil,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		// Adjust viewport and textarea sizes based on the new window size
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 20          // Leave room for the textarea and tabs
		m.textarea.SetWidth(msg.Width - 1)           // Set textarea width with padding
		m.textarea.SetHeight(3)                      // Keep textarea height fixed
		contentStyle = contentStyle.Width(msg.Width) // Update content style width

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyTab:
			// Switch between News and Chat tabs
			if m.activeTab == tabNews {
				// Switched TO chat tab
				m.activeTab = tabChat
				m.viewport.SetContent(strings.Join(m.messages, "\n"))
				m.viewport.GotoBottom()
				m.unreadMessages = false
			} else {
				m.activeTab = tabNews
				m.viewport.SetContent("Welcome to the News tab! No news here yet :-(")
			}
		case tea.KeyEnter:
			if m.activeTab == tabChat {
				m.app.send(chatMsg{
					id:   m.id,
					text: m.textarea.Value(),
				})
				m.textarea.Reset()
			}
		}

	case chatMsg:
		// Always update the chat history
		m.messages = append(m.messages, m.senderStyle.Render(msg.id)+": "+msg.text)

		// Update the viewport content only if the Chat tab is active
		if m.activeTab == tabChat {
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()
		} else {
			m.unreadMessages = true
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	var tabs string
	var chatTabLabel string

	if m.unreadMessages {
		chatTabLabel = "Chat*"
	} else {
		chatTabLabel = "Chat"
	}

	// Render tabs
	if m.activeTab == tabNews {
		tabs = lipgloss.JoinHorizontal(
			lipgloss.Top,
			activeTabStyle.Render("News"),
			tabStyle.Render(chatTabLabel),
		)
	} else {
		tabs = lipgloss.JoinHorizontal(
			lipgloss.Top,
			tabStyle.Render("News"),
			activeTabStyle.Render(chatTabLabel),
		)
	}

	content := contentStyle.Render(m.viewport.View())
	return lipgloss.JoinVertical(
		lipgloss.Left,
		tabs,
		content,
		m.textarea.View(),
	) + "\n\nPress Tab to switch tabs. Press Esc to quit"
}
