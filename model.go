package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const welcomeArtPath = "assets/welcome_art.txt"

// Styling for the welcome screen
var welcomeStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("69")).
	Background(lipgloss.Color("0")).
	Bold(true).
	Padding(1, 2).
	Align(lipgloss.Center)

// Styles for tabs and content
var (
	TabStyle                = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("69")).Padding(0, 1)
	ActiveTabStyle          = TabStyle.Copy().Foreground(lipgloss.Color("205")).Underline(true)
	ContentStyle            = lipgloss.NewStyle().Padding(1, 2)
	newsEnumeratorStyleOdd  = lipgloss.NewStyle().Foreground(lipgloss.Color("99")).Width(7).Align(lipgloss.Right).MarginRight(2)
	newsEnumeratorStyleEven = lipgloss.NewStyle().Foreground(lipgloss.Color("218")).Width(7).Align(lipgloss.Right).MarginRight(2)
	newsItemStyleOdd        = lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
	newsItemStyleEven       = lipgloss.NewStyle().Foreground(lipgloss.Color("218"))
	newsSelectedStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")).Bold(true)
	detailBoxStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")).Padding(2).Bold(true)
	linkStyle               = lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Background(lipgloss.Color("57")).Underline(true) // Style for links
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

	// News-related fields
	newsList      []newsItem
	selectedIndex int
	showDetail    bool
}

// initialModel initializes the TUI model with default settings.
func initialModel() model {
	// Load the welcome art from file
	welcomeArt, err := loadWelcomeArt()
	if err != nil {
		welcomeArt = "Welcome Art Not Found!"
	}

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
	vp.SetContent(strings.Join([]string{
		welcomeStyle.Render(welcomeArt),
		"Welcome to the News tab! Enjoy your stay.",
		"Press Tab to switch tabs and interact with others.",
	}, "\n\n"))

	return model{
		viewport:    vp,
		textarea:    ta,
		messages:    []string{},
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		activeTab:   tabNews,
	}
}

// Init is called when the program starts up.
func (m model) Init() tea.Cmd {
	return func() tea.Msg {
		m.loadFromCache()
		return nil
	}
}

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
		ContentStyle = ContentStyle.Width(msg.Width)

	case tea.KeyMsg:
		key := msg.String() // Get string representation of the key
		switch key {
		case "ctrl+c", "esc":
			// If detail is shown, hide it first
			if m.showDetail {
				m.showDetail = false
				return m, nil
			}
			return m, tea.Quit

		case "tab":
			// Switch tabs
			if m.activeTab == tabNews {
				m.activeTab = tabChat
				m.viewport.SetContent(strings.Join(m.messages, "\n"))
				m.viewport.GotoBottom()
				m.unreadMessages = false

				// Clean the buffer and focus the text area when switching to Chat tab
				m.textarea.Reset()
				m.textarea.Focus()
			} else {
				m.activeTab = tabNews
				m.loadFromCache()
				m.viewport.SetContent(m.renderNewsList())

				// Clean the buffer and blur the text area when switching to News tab
				m.textarea.Reset()
				m.textarea.Blur()
			}

		// News navigation
		case "j":
			if m.activeTab == tabNews {
				if m.selectedIndex < len(m.newsList)-1 {
					m.selectedIndex++
				}
				m.viewport.SetContent(m.renderNewsList())
			}
		case "k":
			if m.activeTab == tabNews {
				if m.selectedIndex > 0 {
					m.selectedIndex--
				}
				m.viewport.SetContent(m.renderNewsList())
			}
		case "d":
			if m.showDetail {
				m.showDetail = false
			} else if m.activeTab == tabNews {
				m.showDetail = true
			}

		case "enter":
			// Send chat message if we're in the Chat tab
			if m.activeTab == tabChat {
				m.app.send(chatMsg{
					id:   m.id,
					text: m.textarea.Value(),
				})
				m.textarea.Reset()
			}
		}

	case chatMsg:
		// Handle incoming chat messages
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

// renderNewsList returns a string containing the list of news with the current item highlighted.
func (m model) renderNewsList() string {
	var sb strings.Builder

	sb.WriteString("The news are brought by newsapi.org; \nnews-cli is not responsible for the content.\n\n")
	sb.WriteString("Use j/k to navigate, d to see details.\n\n")
	for i, news := range m.newsList {
		enumerator := romanNumeral(i + 1)

		// Alternate styles for odd and even rows
		style := newsItemStyleOdd
		enumeratorStyle := newsEnumeratorStyleOdd
		if i%2 == 1 {
			style = newsItemStyleEven
			enumeratorStyle = newsEnumeratorStyleEven
		}

		// Highlight the selected item
		if i == m.selectedIndex {
			sb.WriteString(newsSelectedStyle.Render(fmt.Sprintf("%s. %s", enumerator, news.Headline)) + "\n")
		} else {
			sb.WriteString(enumeratorStyle.Render(fmt.Sprintf("%s.", enumerator)) + " " + style.Render(news.Headline) + "\n")
		}
	}
	return sb.String()
}

// renderNewsDetail shows the description for the currently selected item.
func (m model) renderNewsDetail() string {
	// Check if newsList is empty or if the selectedIndex is out of bounds
	if len(m.newsList) == 0 || m.selectedIndex < 0 || m.selectedIndex >= len(m.newsList) {
		return "No news available to display."
	}

	news := m.newsList[m.selectedIndex]
	title := fmt.Sprintf("Detail for: %s", news.Headline)
	desc := wrapText(news.Description, 80) // Wrap description to 80 characters per line

	link := fmt.Sprintf("\n\nRead more: %s", linkStyle.Render(news.Link))

	detailView := fmt.Sprintf("%s\n\n%s%s", title, desc, link)
	return detailBoxStyle.Render(detailView) +
		"\n\nPress Esc to close detail."
}

// View renders the current state of the TUI.
func (m model) View() string {
	// If we are in the detail view, show it on top of everything else
	if m.activeTab == tabNews && m.showDetail {
		return m.renderNewsDetail()
	}

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
			ActiveTabStyle.Render("News"),
			TabStyle.Render(chatTabLabel),
		)
	} else {
		tabs = lipgloss.JoinHorizontal(
			lipgloss.Top,
			TabStyle.Render("News"),
			ActiveTabStyle.Render(chatTabLabel),
		)
	}

	content := ContentStyle.Render(m.viewport.View())

	// Combine tabs and content; include textarea only in Chat tab
	return lipgloss.JoinVertical(
		lipgloss.Left,
		tabs,
		content,
		m.textarea.View(),
	) + "\n\nPress Tab to switch tabs. Press Esc to quit"
}
