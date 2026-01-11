package tui

import (
	"strings"

	"github.com/charly-vibes/fabbro/internal/session"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	session  *session.Session
	lines    []string
	cursor   int
	selected int
}

func New(sess *session.Session) Model {
	lines := strings.Split(sess.Content, "\n")
	return Model{
		session:  sess,
		lines:    lines,
		cursor:   0,
		selected: -1,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	return "TUI placeholder - press q to quit"
}
