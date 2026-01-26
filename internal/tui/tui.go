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
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type mode int

const (
	modeNormal mode = iota
	modeInput
	modePalette
	modeEditor
)

type editorState struct {
	ta         textarea.Model
	start, end int  // 0-indexed line range being edited
	escPending bool // true after first Esc press
	annIndex   int  // -1 if new annotation, >=0 if editing existing
}

type clearMessageMsg struct{}

func clearMessageAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return clearMessageMsg{}
	})
}

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

func (m *Model) annotationsOnLine(lineNum int) []int {
	var indices []int
	for i, ann := range m.annotations {
		if ann.StartLine <= lineNum && lineNum <= ann.EndLine {
			indices = append(indices, i)
		}
	}
	return indices
}

func decodeAnnText(s string) string {
	return strings.ReplaceAll(s, "\\n", "\n")
}

func encodeAnnText(s string) string {
	return strings.ReplaceAll(s, "\n", "\\n")
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
	zPending       bool   // waiting for second key in z commands (zz, zt, zb)
	viewportTop    int    // explicit viewport start line (-1 means auto-follow cursor)
	lastError      string // last error message to display
	lastMessage    string // last success message to display
	highlighter    *highlight.Highlighter
	sourceFile     string
	editor         *editorState // non-nil when in editor mode
	paletteKind    string       // "commands" or "annPick"
	paletteItems   []int        // annotation indices for picker
	paletteCursor  int          // current selection in picker
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
		viewportTop: -1, // auto-follow cursor
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

	case clearMessageMsg:
		m.lastMessage = ""
		return m, nil

	case tea.KeyMsg:
		// Clear messages on any key press
		m.lastError = ""
		m.lastMessage = ""

		switch m.mode {
		case modeInput:
			return m.handleInputMode(msg)
		case modePalette:
			return m.handlePaletteMode(msg)
		case modeEditor:
			return m.handleEditorMode(msg)
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
			m.viewportTop = -1 // reset to auto-follow
			return m, nil
		}
		// Fall through to handle the key normally
	}

	// Handle z-pending state (viewport centering)
	if m.zPending {
		m.zPending = false
		visibleLines := m.height - 4
		if visibleLines < 5 {
			visibleLines = 10
		}
		switch msg.String() {
		case "z": // zz - center cursor in viewport
			m.viewportTop = m.cursor - visibleLines/2
		case "t": // zt - cursor at top of viewport
			m.viewportTop = m.cursor
		case "b": // zb - cursor at bottom of viewport
			m.viewportTop = m.cursor - visibleLines + 1
		}
		// Clamp viewportTop to valid range
		if m.viewportTop < 0 {
			m.viewportTop = 0
		}
		if m.viewportTop > len(m.lines)-1 {
			m.viewportTop = len(m.lines) - 1
		}
		return m, nil
	}

	switch msg.String() {
	case "Q", "ctrl+c":
		return m, tea.Quit

	case "j", "down":
		if m.cursor < len(m.lines)-1 {
			m.cursor++
			m.viewportTop = -1 // reset to auto-follow on cursor movement
			if m.selection.active {
				m.selection.cursor = m.cursor
			}
		}

	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
			m.viewportTop = -1 // reset to auto-follow on cursor movement
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
		m.viewportTop = -1 // reset to auto-follow on cursor movement

	case "ctrl+u":
		halfPage := (m.height - 4) / 2
		if halfPage < 1 {
			halfPage = 1
		}
		m.cursor -= halfPage
		if m.cursor < 0 {
			m.cursor = 0
		}
		m.viewportTop = -1 // reset to auto-follow on cursor movement

	case "g":
		m.gPending = true

	case "G":
		m.cursor = len(m.lines) - 1
		m.viewportTop = -1 // reset to auto-follow

	case "z":
		m.zPending = true

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
		} else {
			m.tryEditAnnotation()
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

	case "i":
		if m.selection.active {
			m.openEditor()
		}

	case " ":
		m.mode = modePalette

	case "w":
		if err := m.save(); err != nil {
			m.lastError = err.Error()
			return m, nil
		}
		m.lastMessage = "Saved!"
		return m, clearMessageAfter(2 * time.Second)
	}
	return m, nil
}

func (m Model) handlePaletteMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Annotation picker mode
	if m.paletteKind == "annPick" {
		return m.handleAnnotationPicker(msg)
	}

	// Standard command palette
	switch msg.String() {
	case "w":
		if err := m.save(); err != nil {
			m.lastError = err.Error()
			m.mode = modeNormal
			return m, nil
		}
		m.lastMessage = "Saved!"
		m.mode = modeNormal
		return m, clearMessageAfter(2 * time.Second)
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
	case "i":
		if m.selection.active {
			m.openEditor()
		}
	default:
		m.mode = modeNormal
	}
	return m, nil
}

func (m Model) handleAnnotationPicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.paletteCursor < len(m.paletteItems)-1 {
			m.paletteCursor++
		}
	case "k", "up":
		if m.paletteCursor > 0 {
			m.paletteCursor--
		}
	case "enter":
		if len(m.paletteItems) > 0 {
			annIndex := m.paletteItems[m.paletteCursor]
			m.paletteKind = ""
			m.paletteItems = nil
			m.paletteCursor = 0
			m.openEditorForAnnotation(annIndex)
		}
	case "esc":
		m.mode = modeNormal
		m.paletteKind = ""
		m.paletteItems = nil
		m.paletteCursor = 0
	default:
		m.mode = modeNormal
		m.paletteKind = ""
		m.paletteItems = nil
		m.paletteCursor = 0
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

func (m *Model) tryEditAnnotation() {
	cursorLine := m.cursor + 1 // 1-indexed
	indices := m.annotationsOnLine(cursorLine)

	if len(indices) == 0 {
		m.lastError = "No annotation on this line"
		return
	}

	if len(indices) == 1 {
		m.openEditorForAnnotation(indices[0])
		return
	}

	// Multiple annotations: open picker
	m.mode = modePalette
	m.paletteKind = "annPick"
	m.paletteItems = indices
	m.paletteCursor = 0
}

func (m *Model) openEditorForAnnotation(annIndex int) {
	ann := m.annotations[annIndex]
	content := decodeAnnText(ann.Text)

	ta := textarea.New()
	ta.SetValue(content)
	ta.Focus()
	ta.Prompt = ""
	ta.CharLimit = 0
	ta.ShowLineNumbers = false

	m.editor = &editorState{
		ta:       ta,
		start:    ann.StartLine - 1, // 0-indexed
		end:      ann.EndLine - 1,
		annIndex: annIndex,
	}
	m.mode = modeEditor
}

func (m *Model) openEditor() {
	start, end := m.selection.lines()
	content := strings.Join(m.lines[start:end+1], "\n")

	ta := textarea.New()
	ta.SetValue(content)
	ta.Focus()
	ta.Prompt = ""
	ta.CharLimit = 0
	ta.ShowLineNumbers = false

	m.editor = &editorState{
		ta:       ta,
		start:    start,
		end:      end,
		annIndex: -1, // new annotation
	}
	m.mode = modeEditor
}

func (m Model) handleEditorMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.editor == nil {
		m.mode = modeNormal
		return m, nil
	}

	switch msg.Type {
	case tea.KeyEsc:
		if m.editor.escPending {
			// Second Esc - cancel and exit
			m.editor = nil
			m.mode = modeNormal
			return m, nil
		}
		// First Esc - set pending
		m.editor.escPending = true
		return m, nil

	case tea.KeyCtrlS:
		// Save the edited content as a change annotation
		m.saveEditorContent()
		return m, nil

	case tea.KeyCtrlC:
		// Immediate cancel
		m.editor = nil
		m.mode = modeNormal
		return m, nil
	}

	// Reset escPending on any other key
	m.editor.escPending = false

	// Delegate to textarea
	var cmd tea.Cmd
	m.editor.ta, cmd = m.editor.ta.Update(msg)
	return m, cmd
}

func (m *Model) saveEditorContent() {
	if m.editor == nil {
		return
	}

	edited := m.editor.ta.Value()

	// Editing existing annotation
	if m.editor.annIndex >= 0 {
		m.annotations[m.editor.annIndex].Text = encodeAnnText(edited)
		m.editor = nil
		m.mode = modeNormal
		return
	}

	// New annotation: encode newlines and format as change annotation
	encoded := encodeAnnText(edited)

	// Format the change prefix
	startLine := m.editor.start + 1 // 1-indexed
	endLine := m.editor.end + 1
	var prefix string
	if startLine == endLine {
		prefix = fmt.Sprintf("[line %d] -> ", startLine)
	} else {
		prefix = fmt.Sprintf("[lines %d-%d] -> ", startLine, endLine)
	}
	text := prefix + encoded

	// Add annotation for each line in the range
	for line := m.editor.start; line <= m.editor.end; line++ {
		m.annotations = append(m.annotations, fem.Annotation{
			StartLine: line + 1, // 1-indexed
			EndLine:   line + 1,
			Type:      "change",
			Text:      text,
		})
	}

	// Clear editor and selection
	m.editor = nil
	m.mode = modeNormal
	m.selection = selection{}
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

	selStart, selEnd := m.selection.lines()
	prefixLen := 12 // ">◆ 123 ● │ " is ~12 chars
	contentWidth := m.width - prefixLen
	if contentWidth < 10 {
		contentWidth = 40 // sensible fallback
	}

	// Build annotation lookup for quick access
	annotatedLines := make(map[int]bool)
	for _, ann := range m.annotations {
		annotatedLines[ann.StartLine] = true
	}

	// Calculate start and end accounting for wrapped lines consuming multiple screen rows
	start := 0
	if m.viewportTop >= 0 {
		// Use explicit viewport position (set by zz, zt, zb)
		start = m.viewportTop
		if start < 0 {
			start = 0
		}
		if start >= len(m.lines) {
			start = len(m.lines) - 1
		}
	} else if m.cursor > 0 {
		// Auto-follow: Find start such that cursor is visible within visibleLines screen rows
		screenRows := 0
		for i := m.cursor; i >= 0; i-- {
			wrapped := wrapLine(m.lines[i], contentWidth)
			if screenRows+len(wrapped) > visibleLines {
				start = i + 1
				break
			}
			screenRows += len(wrapped)
		}
	}

	// Calculate end based on screen rows
	screenRows := 0
	end := start
	for end < len(m.lines) && screenRows < visibleLines {
		wrapped := wrapLine(m.lines[end], contentWidth)
		screenRows += len(wrapped)
		end++
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

		// Annotation indicator (1-indexed line numbers)
		annIndicator := " "
		if annotatedLines[i+1] {
			annIndicator = "●"
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
				b.WriteString(fmt.Sprintf("%s%s %s %s │ %s\n", cursor, selIndicator, lineNum, annIndicator, displayPart))
			} else {
				b.WriteString(fmt.Sprintf("         │ %s\n", displayPart))
			}
		}
	}

	b.WriteString(strings.Repeat("─", width))
	b.WriteString("\n")

	switch m.mode {
	case modeInput:
		prompt := fem.Prompts[m.inputType]
		b.WriteString(fmt.Sprintf("%s %s_\n", prompt, m.input))
	case modeEditor:
		b.WriteString("┌─ Edit selection (Ctrl+S save, Esc Esc cancel) ─────┐\n")
		if m.editor != nil {
			// Render textarea content in the panel
			taView := m.editor.ta.View()
			taLines := strings.Split(taView, "\n")
			innerWidth := width - 4 // account for "│ " and " │"
			maxLines := 6           // limit editor height
			for i, line := range taLines {
				if i >= maxLines {
					b.WriteString("│ ...                                                │\n")
					break
				}
				// Pad or truncate line to fit
				lineRunes := []rune(line)
				if len(lineRunes) > innerWidth {
					lineRunes = lineRunes[:innerWidth]
				}
				padding := innerWidth - len(lineRunes)
				b.WriteString(fmt.Sprintf("│ %s%s │\n", string(lineRunes), strings.Repeat(" ", padding)))
			}
		}
		b.WriteString("└────────────────────────────────────────────────────┘\n")
	case modePalette:
		if m.paletteKind == "annPick" {
			b.WriteString("┌─ Select annotation to edit ───────────────────────┐\n")
			for i, idx := range m.paletteItems {
				ann := m.annotations[idx]
				cursor := " "
				if i == m.paletteCursor {
					cursor = ">"
				}
				preview := ann.Text
				if len(preview) > 30 {
					preview = preview[:27] + "..."
				}
				b.WriteString(fmt.Sprintf("│%s %-10s [%d-%d] %s\n", cursor, ann.Type, ann.StartLine, ann.EndLine, preview))
			}
			b.WriteString("│                    j/k move, Enter select, Esc cancel │\n")
			b.WriteString("└────────────────────────────────────────────────────┘\n")
		} else {
			b.WriteString("┌─ Commands ─────────────────────────────────────────┐\n")
			b.WriteString("│ [w]rite    [Q]uit                                  │\n")
			if m.selection.active {
				b.WriteString("├─ Annotations ──────────────────────────────────────┤\n")
				b.WriteString("│ [c]omment  [d]elete  [q]uestion  [r]eplace         │\n")
				b.WriteString("│ [e]xpand   [k]eep    [u]nclear   [i]nline-edit     │\n")
			}
			b.WriteString("│                                  [ESC] cancel      │\n")
			b.WriteString("└────────────────────────────────────────────────────┘\n")
		}
	default:
		b.WriteString("[v]select [SPC]palette [w]rite [Q]uit")
		if m.selection.active {
			b.WriteString(" │ [c]omment [d]elete [q]uestion [e]xpand [u]nclear [r]eplace [i]nline")
		}
		b.WriteString("\n")
	}

	if m.lastError != "" {
		b.WriteString(fmt.Sprintf("Error: %s\n", m.lastError))
	}
	if m.lastMessage != "" {
		b.WriteString(fmt.Sprintf("✓ %s\n", m.lastMessage))
	}

	return b.String()
}
