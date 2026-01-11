package tui

import tea "github.com/charmbracelet/bubbletea"

type Model struct {
	content  []string
	cursor   int
	selected int
}

func New(content []string) Model {
	return Model{
		content:  content,
		cursor:   0,
		selected: -1,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	return ""
}
