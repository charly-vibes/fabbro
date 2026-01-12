package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charly-vibes/fabbro/internal/config"
	"github.com/charly-vibes/fabbro/internal/session"
	tea "github.com/charmbracelet/bubbletea"
)

type mode int

const (
	modeNormal mode = iota
	modeInput
)

type Annotation struct {
	Line int
	Text string
}

type Model struct {
	session     *session.Session
	lines       []string
	cursor      int
	selected    int
	mode        mode
	input       string
	annotations []Annotation
	width       int
	height      int
}

func New(sess *session.Session) Model {
	lines := strings.Split(sess.Content, "\n")
	return Model{
		session:     sess,
		lines:       lines,
		cursor:      0,
		selected:    -1,
		mode:        modeNormal,
		annotations: []Annotation{},
	}
}

func (m Model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.mode == modeInput {
			return m.handleInputMode(msg)
		}
		return m.handleNormalMode(msg)
	}
	return m, nil
}

func (m Model) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "Q", "ctrl+c":
		return m, tea.Quit

	case "j", "down":
		if m.cursor < len(m.lines)-1 {
			m.cursor++
		}

	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}

	case "v":
		if m.selected == m.cursor {
			m.selected = -1
		} else {
			m.selected = m.cursor
		}

	case "c":
		if m.selected >= 0 {
			m.mode = modeInput
			m.input = ""
		}

	case "w":
		m.save()
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.input != "" {
			m.annotations = append(m.annotations, Annotation{
				Line: m.selected,
				Text: m.input,
			})
		}
		m.mode = modeNormal
		m.input = ""
		m.selected = -1

	case "esc":
		m.mode = modeNormal
		m.input = ""

	case "backspace":
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}

	default:
		if len(msg.String()) == 1 {
			m.input += msg.String()
		}
	}
	return m, nil
}

func (m Model) save() {
	annotationsByLine := make(map[int]string)
	for _, a := range m.annotations {
		annotationsByLine[a.Line] = a.Text
	}

	var result []string
	for i, line := range m.lines {
		if comment, ok := annotationsByLine[i]; ok {
			result = append(result, line+" {>> "+comment+" <<}")
		} else {
			result = append(result, line)
		}
	}

	content := strings.Join(result, "\n")
	fileContent := fmt.Sprintf(`---
session_id: %s
created_at: %s
---

%s`, m.session.ID, m.session.CreatedAt.Format("2006-01-02T15:04:05Z07:00"), content)

	sessionPath := filepath.Join(config.SessionsDir, m.session.ID+".fem")
	os.WriteFile(sessionPath, []byte(fileContent), 0644)
}

func (m Model) View() string {
	var b strings.Builder

	title := fmt.Sprintf("─── Review: %s ", m.session.ID)
	b.WriteString(title)
	b.WriteString(strings.Repeat("─", max(0, 50-len(title))))
	b.WriteString("\n")

	visibleLines := m.height - 4
	if visibleLines < 5 {
		visibleLines = 10
	}

	start := 0
	if m.cursor >= visibleLines {
		start = m.cursor - visibleLines + 1
	}
	end := start + visibleLines
	if end > len(m.lines) {
		end = len(m.lines)
	}

	for i := start; i < end; i++ {
		lineNum := fmt.Sprintf("%3d", i+1)
		line := m.lines[i]

		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		selection := " "
		if i == m.selected {
			selection = "●"
		}

		b.WriteString(fmt.Sprintf("%s%s %s │ %s\n", cursor, selection, lineNum, line))
	}

	b.WriteString(strings.Repeat("─", 50))
	b.WriteString("\n")

	if m.mode == modeInput {
		b.WriteString(fmt.Sprintf("Comment: %s_\n", m.input))
	} else {
		b.WriteString("[v]select [c]omment [w]rite [Q]uit\n")
	}

	return b.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
