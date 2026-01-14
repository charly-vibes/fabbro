package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charly-vibes/fabbro/internal/config"
	"github.com/charly-vibes/fabbro/internal/fem"
	"github.com/charly-vibes/fabbro/internal/highlight"
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

func wrapLine(s string, width int) []string {
	if width <= 0 || len(s) <= width {
		return []string{s}
	}

	var result []string
	runes := []rune(s)
	for len(runes) > width {
		result = append(result, string(runes[:width]))
		runes = runes[width:]
	}
	if len(runes) > 0 {
		result = append(result, string(runes))
	}
	return result
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
	highlighter    *highlight.Highlighter
	sourceFile     string
}

func New(sess *session.Session) Model {
	return NewWithFile(sess, "")
}

func NewWithFile(sess *session.Session, sourceFile string) Model {
	lines := strings.Split(sess.Content, "\n")
	return Model{
		session:     sess,
		lines:       lines,
		cursor:      0,
		selection:   selection{},
		mode:        modeNormal,
		annotations: []fem.Annotation{},
		highlighter: highlight.New(sourceFile, sess.Content),
		sourceFile:  sourceFile,
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

	case "r":
		if m.selection.active {
			m.mode = modeInput
			m.input = ""
			m.inputType = "change"
		}

	case " ":
		m.mode = modePalette

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
	case "w":
		if err := m.save(); err != nil {
			m.lastError = err.Error()
		} else {
			m.lastError = ""
		}
		m.mode = modeNormal
	case "Q":
		return m, tea.Quit
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
	case "k":
		if m.selection.active {
			m.mode = modeInput
			m.input = ""
			m.inputType = "keep"
		}
	case "u":
		if m.selection.active {
			m.mode = modeInput
			m.input = ""
			m.inputType = "unclear"
		}
	case "r":
		if m.selection.active {
			m.mode = modeInput
			m.input = ""
			m.inputType = "change"
		}
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
			text := m.input

			// For change annotations, prefix with line reference
			if m.inputType == "change" {
				startLine := start + 1 // 1-indexed for display
				endLine := end + 1
				var lineRef string
				if startLine == endLine {
					lineRef = fmt.Sprintf("[line %d] -> ", startLine)
				} else {
					lineRef = fmt.Sprintf("[lines %d-%d] -> ", startLine, endLine)
				}
				text = lineRef + m.input
			}

			for line := start; line <= end; line++ {
				m.annotations = append(m.annotations, fem.Annotation{
					StartLine: line + 1, // 1-indexed for storage
					EndLine:   line + 1,
					Type:      m.inputType,
					Text:      text,
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
	if err := os.WriteFile(sessionPath, []byte(fileContent), 0600); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}
	return nil
}

func (m Model) View() string {
	var b strings.Builder

	width := m.width
	if width < 20 {
		width = 50
	}

	title := fmt.Sprintf("─── Review: %s ", m.session.ID)
	if m.selection.active {
		selStart, selEnd := m.selection.lines()
		lineCount := selEnd - selStart + 1
		title += fmt.Sprintf("[%d lines selected] ", lineCount)
	}
	titleRunes := []rune(title)
	if len(titleRunes) > width {
		titleRunes = titleRunes[:width]
	}
	b.WriteString(string(titleRunes))
	remaining := width - len(titleRunes)
	if remaining > 0 {
		b.WriteString(strings.Repeat("─", remaining))
	}
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
	prefixLen := 10 // ">◆ 123 │ " is ~10 chars
	contentWidth := m.width - prefixLen
	if contentWidth < 10 {
		contentWidth = 40 // sensible fallback
	}

	for i := start; i < end; i++ {
		lineNum := fmt.Sprintf("%3d", i+1)
		line := m.lines[i]

		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		selIndicator := " "
		if m.selection.active && i >= selStart && i <= selEnd {
			if i == m.selection.anchor {
				selIndicator = "◆"
			} else {
				selIndicator = "▌"
			}
		}

		highlightedLine := m.highlighter.RenderLine(line)

		wrapped := wrapLine(line, contentWidth)
		for j, part := range wrapped {
			var displayPart string
			if j == 0 && len(wrapped) == 1 {
				displayPart = highlightedLine
			} else {
				displayPart = m.highlighter.RenderLine(part)
			}
			if j == 0 {
				b.WriteString(fmt.Sprintf("%s%s %s │ %s\n", cursor, selIndicator, lineNum, displayPart))
			} else {
				b.WriteString(fmt.Sprintf("       │ %s\n", displayPart))
			}
		}
	}

	b.WriteString(strings.Repeat("─", width))
	b.WriteString("\n")

	switch m.mode {
	case modeInput:
		prompt := fem.Prompts[m.inputType]
		b.WriteString(fmt.Sprintf("%s %s_\n", prompt, m.input))
	case modePalette:
		b.WriteString("┌─ Commands ─────────────────────────────────────────┐\n")
		b.WriteString("│ [w]rite    [Q]uit                                  │\n")
		if m.selection.active {
			b.WriteString("├─ Annotations ──────────────────────────────────────┤\n")
			b.WriteString("│ [c]omment  [d]elete  [q]uestion  [r]eplace         │\n")
			b.WriteString("│ [e]xpand   [k]eep    [u]nclear                     │\n")
		}
		b.WriteString("│                                  [ESC] cancel      │\n")
		b.WriteString("└────────────────────────────────────────────────────┘\n")
	default:
		b.WriteString("[v]select [SPC]palette [w]rite [Q]uit")
		if m.selection.active {
			b.WriteString(" │ [c]omment [d]elete [q]uestion [e]xpand [u]nclear [r]eplace")
		}
		b.WriteString("\n")
	}

	if m.lastError != "" {
		b.WriteString(fmt.Sprintf("Error: %s\n", m.lastError))
	}

	return b.String()
}
