package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charly-vibes/fabbro/internal/config"
	"github.com/charly-vibes/fabbro/internal/fem"
	"github.com/charly-vibes/fabbro/internal/session"
	tea "github.com/charmbracelet/bubbletea"
)

type mode int

const (
	modeNormal mode = iota
	modeInput
	modePalette
)

type selection struct {
	active bool
	anchor int // where selection started
	cursor int // current end of selection
}

func (s selection) lines() (start, end int) {
	if s.anchor < s.cursor {
		return s.anchor, s.cursor
	}
	return s.cursor, s.anchor
}

var inputPrompts = map[string]string{
	"comment":  "Comment:",
	"delete":   "Reason for deletion:",
	"question": "Question:",
	"expand":   "What to expand:",
	"keep":     "Reason to keep:",
	"unclear":  "What's unclear:",
}

type Model struct {
	session        *session.Session
	lines          []string
	cursor         int
	selection      selection
	mode           mode
	input          string
	inputType      string // annotation type being entered: "comment", "delete", etc.
	annotations    []fem.Annotation
	width          int
	height         int
	gPending       bool   // waiting for second 'g' in gg command
	lastError      string // last error message to display
}

func New(sess *session.Session) Model {
	lines := strings.Split(sess.Content, "\n")
	return Model{
		session:     sess,
		lines:       lines,
		cursor:      0,
		selection:   selection{},
		mode:        modeNormal,
		annotations: []fem.Annotation{},
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
		switch m.mode {
		case modeInput:
			return m.handleInputMode(msg)
		case modePalette:
			return m.handlePaletteMode(msg)
		default:
			return m.handleNormalMode(msg)
		}
	}
	return m, nil
}

func (m Model) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle g-pending state first
	if m.gPending {
		m.gPending = false
		if msg.String() == "g" {
			m.cursor = 0
			return m, nil
		}
		// Fall through to handle the key normally
	}

	switch msg.String() {
	case "Q", "ctrl+c":
		return m, tea.Quit

	case "j", "down":
		if m.cursor < len(m.lines)-1 {
			m.cursor++
			if m.selection.active {
				m.selection.cursor = m.cursor
			}
		}

	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
			if m.selection.active {
				m.selection.cursor = m.cursor
			}
		}

	case "ctrl+d":
		halfPage := (m.height - 4) / 2
		if halfPage < 1 {
			halfPage = 1
		}
		m.cursor += halfPage
		if m.cursor > len(m.lines)-1 {
			m.cursor = len(m.lines) - 1
		}

	case "ctrl+u":
		halfPage := (m.height - 4) / 2
		if halfPage < 1 {
			halfPage = 1
		}
		m.cursor -= halfPage
		if m.cursor < 0 {
			m.cursor = 0
		}

	case "g":
		m.gPending = true

	case "G":
		m.cursor = len(m.lines) - 1

	case "esc":
		m.selection = selection{}

	case "v":
		if m.selection.active {
			m.selection = selection{}
		} else {
			m.selection = selection{active: true, anchor: m.cursor, cursor: m.cursor}
		}

	case "c":
		if m.selection.active {
			m.mode = modeInput
			m.input = ""
			m.inputType = "comment"
		}

	case "d":
		if m.selection.active {
			m.mode = modeInput
			m.input = ""
			m.inputType = "delete"
		}

	case "q":
		if m.selection.active {
			m.mode = modeInput
			m.input = ""
			m.inputType = "question"
		}

	case "e":
		if m.selection.active {
			m.mode = modeInput
			m.input = ""
			m.inputType = "expand"
		}

	case "u":
		if m.selection.active {
			m.mode = modeInput
			m.input = ""
			m.inputType = "unclear"
		}

	case " ":
		if m.selection.active {
			m.mode = modePalette
		}

	case "w":
		if err := m.save(); err != nil {
			m.lastError = err.Error()
		} else {
			m.lastError = ""
		}
	}
	return m, nil
}

func (m Model) handlePaletteMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "c":
		m.mode = modeInput
		m.input = ""
		m.inputType = "comment"
	case "d":
		m.mode = modeInput
		m.input = ""
		m.inputType = "delete"
	case "q":
		m.mode = modeInput
		m.input = ""
		m.inputType = "question"
	case "e":
		m.mode = modeInput
		m.input = ""
		m.inputType = "expand"
	case "k":
		m.mode = modeInput
		m.input = ""
		m.inputType = "keep"
	case "u":
		m.mode = modeInput
		m.input = ""
		m.inputType = "unclear"
	default:
		m.mode = modeNormal
	}
	return m, nil
}

func (m Model) handleInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.input != "" {
			start, end := m.selection.lines()
			for line := start; line <= end; line++ {
				m.annotations = append(m.annotations, fem.Annotation{
					StartLine: line,
					EndLine:   line,
					Type:      m.inputType,
					Text:      m.input,
				})
			}
		}
		m.mode = modeNormal
		m.input = ""
		m.inputType = ""
		m.selection = selection{}

	case "esc":
		m.mode = modeNormal
		m.input = ""
		m.inputType = ""

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

func (m Model) save() error {
	annotationsByLine := make(map[int][]fem.Annotation)
	for _, a := range m.annotations {
		annotationsByLine[a.StartLine] = append(annotationsByLine[a.StartLine], a)
	}

	var result []string
	for i, line := range m.lines {
		lineNum := i + 1
		if anns, ok := annotationsByLine[lineNum]; ok {
			for _, ann := range anns {
				marker := fem.Markers[ann.Type]
				line = line + " " + marker[0] + ann.Text + marker[1]
			}
		}
		result = append(result, line)
	}

	content := strings.Join(result, "\n")
	fileContent := fmt.Sprintf(`---
session_id: %s
created_at: %s
---

%s`, m.session.ID, m.session.CreatedAt.Format(time.RFC3339), content)

	sessionPath := filepath.Join(config.SessionsDir, m.session.ID+".fem")
	if err := os.WriteFile(sessionPath, []byte(fileContent), 0644); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}
	return nil
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

	selStart, selEnd := m.selection.lines()
	for i := start; i < end; i++ {
		lineNum := fmt.Sprintf("%3d", i+1)
		line := m.lines[i]

		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		selIndicator := " "
		if m.selection.active && i >= selStart && i <= selEnd {
			selIndicator = "●"
		}

		b.WriteString(fmt.Sprintf("%s%s %s │ %s\n", cursor, selIndicator, lineNum, line))
	}

	b.WriteString(strings.Repeat("─", 50))
	b.WriteString("\n")

	switch m.mode {
	case modeInput:
		prompt := inputPrompts[m.inputType]
		b.WriteString(fmt.Sprintf("%s %s_\n", prompt, m.input))
	case modePalette:
		b.WriteString("┌─ Annotations ──────────────────────────────────────┐\n")
		b.WriteString("│ [c]omment  [d]elete  [q]uestion                    │\n")
		b.WriteString("│ [e]xpand   [k]eep    [u]nclear   [ESC] cancel      │\n")
		b.WriteString("└────────────────────────────────────────────────────┘\n")
	default:
		b.WriteString("[v]select [SPC]palette [w]rite [Q]uit")
		if m.selection.active {
			b.WriteString(" │ [c]omment [d]elete [q]uestion [e]xpand [u]nclear")
		}
		b.WriteString("\n")
	}

	if m.lastError != "" {
		b.WriteString(fmt.Sprintf("Error: %s\n", m.lastError))
	}

	return b.String()
}
