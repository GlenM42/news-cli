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
)

// Constants for SSH server host and port.
const (
	host = "localhost"
	port = "23234"
)

// app wraps the Wish server and tracks active Bubble Tea programs.
type app struct {
	*ssh.Server
	progs []*tea.Program
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
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
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
