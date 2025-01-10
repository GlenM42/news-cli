package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// errMsg and chatMsg are used to communicate in the Bubble Tea model.
type (
	errMsg  error
	chatMsg struct {
		id   string
		text string
	}
)

// tab indicates which tab is active.
type tab int

const (
	tabNews tab = iota
	tabChat
)

// model holds the state of our TUI, including references to the SSH app, the user's ID, etc.
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

// initialModel initializes the TUI model with default settings.
func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Send a message in 280 characters..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 280
	ta.SetWidth(30)
	ta.SetHeight(3)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle() // Remove cursor line styling
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	vp := viewport.New(30, 5)
	vp.SetContent(`Welcome to the News tab!
Switch to Chat to interact with others by pressing Tab.
Type a message and press Enter to send.`)

	return model{
		textarea:    ta,
		messages:    []string{},
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		activeTab:   tabNews,
		err:         nil,
	}
}

// Init is called when the program starts up.
func (m model) Init() tea.Cmd {
	return textarea.Blink
}

// Update processes incoming messages and handles UI logic.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 20
		m.textarea.SetWidth(msg.Width - 1)
		m.textarea.SetHeight(3)
		contentStyle = contentStyle.Width(msg.Width)

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyTab:
			// Switch between tabs
			if m.activeTab == tabNews {
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
				// Broadcast the chat message to all programs
				m.app.send(chatMsg{
					id:   m.id,
					text: m.textarea.Value(),
				})
				m.textarea.Reset()
			}
		}

	case chatMsg:
		// Update the chat history
		m.messages = append(m.messages, m.senderStyle.Render(msg.id)+": "+msg.text)
		if m.activeTab == tabChat {
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()
		} else {
			m.unreadMessages = true
		}

	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	return m, tea.Batch(tiCmd, vpCmd)
}

// View renders the current state of the TUI.
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
