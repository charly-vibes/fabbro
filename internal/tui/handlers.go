package tui

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charly-vibes/fabbro/internal/fem"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.viewportTop < 0 {
			m.ensureCursorVisible()
		}
		return m, nil

	case clearMessageMsg:
		m.lastMessage = ""
		return m, nil

	case tea.KeyMsg:
		m.lastError = ""
		m.lastMessage = ""

		switch m.mode {
		case modeInput:
			return m.handleInputMode(msg)
		case modePalette:
			return m.handlePaletteMode(msg)
		case modeEditor:
			return m.handleEditorMode(msg)
		case modeQuitConfirm:
			return m.handleQuitConfirmMode(msg)
		default:
			return m.handleNormalMode(msg)
		}
	}
	return m, nil
}

func (m *Model) openInputMode(annType string) {
	// Match the box calculation: boxTotalWidth = width - 2, innerWidth = boxTotalWidth - 4
	boxTotalWidth := m.width - 2
	if boxTotalWidth < 24 {
		boxTotalWidth = 64
	}
	taWidth := boxTotalWidth - 4

	ta := textarea.New()
	ta.Focus()
	ta.Prompt = ""
	ta.CharLimit = 0
	ta.ShowLineNumbers = false
	ta.SetWidth(taWidth)
	ta.SetHeight(3)
	ta.KeyMap.InsertNewline.SetKeys("alt+enter", "ctrl+j")

	m.inputTA = &ta
	m.inputType = annType
	m.mode = modeInput
}

func (m Model) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.gPending {
		m.gPending = false
		if msg.String() == "g" {
			m.cursor = 0
			m.viewportTop = -1
			m.autoViewportTop = 0
			return m, nil
		}
	}

	if m.zPending {
		m.zPending = false
		visibleLines := m.height - 4
		if visibleLines < 5 {
			visibleLines = 10
		}
		switch msg.String() {
		case "z":
			m.viewportTop = m.cursor - visibleLines/2
		case "t":
			m.viewportTop = m.cursor
		case "b":
			m.viewportTop = m.cursor - visibleLines + 1
		}
		if m.viewportTop < 0 {
			m.viewportTop = 0
		}
		if m.viewportTop > len(m.lines)-1 {
			m.viewportTop = len(m.lines) - 1
		}
		return m, nil
	}

	switch msg.String() {
	case "ctrl+c":
		now := time.Now()
		if !m.lastCtrlC.IsZero() && now.Sub(m.lastCtrlC) < 2*time.Second {
			m.mode = modeQuitConfirm
			m.lastCtrlC = time.Time{}
			return m, nil
		}
		m.lastCtrlC = now
		m.lastMessage = "Press CTRL+C again to quit"
		return m, clearMessageAfter(2 * time.Second)

	case "j", "down":
		if m.cursor < len(m.lines)-1 {
			m.cursor++
			m.viewportTop = -1
			m.ensureCursorVisible()
			if m.selection.active {
				m.selection.cursor = m.cursor
			}
		}

	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
			m.viewportTop = -1
			m.ensureCursorVisible()
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
		m.viewportTop = -1
		m.ensureCursorVisible()

	case "ctrl+u":
		halfPage := (m.height - 4) / 2
		if halfPage < 1 {
			halfPage = 1
		}
		m.cursor -= halfPage
		if m.cursor < 0 {
			m.cursor = 0
		}
		m.viewportTop = -1
		m.ensureCursorVisible()

	case "g":
		m.gPending = true

	case "G":
		m.cursor = len(m.lines) - 1
		m.viewportTop = -1
		m.ensureCursorVisible()

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
			m.openInputMode("comment")
		}

	case "d":
		if m.selection.active {
			m.openInputMode("delete")
		}

	case "q":
		if m.selection.active {
			m.openInputMode("question")
		}

	case "e":
		if m.selection.active {
			m.openInputMode("expand")
		} else {
			m.tryEditAnnotation()
		}

	case "u":
		if m.selection.active {
			m.openInputMode("unclear")
		}

	case "r":
		if m.selection.active {
			m.openInputMode("change")
		}

	case "i":
		if m.selection.active {
			m.openEditor()
		}

	case " ":
		m.mode = modePalette

	case "w":
		if err := m.save(); err != nil {
			if errors.Is(err, ErrTutorSession) {
				m.lastMessage = "Tutorial sessions are not saved"
			} else {
				m.lastError = err.Error()
			}
			return m, clearMessageAfter(2 * time.Second)
		}
		m.dirty = false
		m.lastMessage = "Saved!"
		return m, clearMessageAfter(2 * time.Second)
	}
	return m, nil
}

func (m Model) handlePaletteMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.paletteKind == "annPick" {
		return m.handleAnnotationPicker(msg)
	}

	switch msg.String() {
	case "w":
		if err := m.save(); err != nil {
			if errors.Is(err, ErrTutorSession) {
				m.lastMessage = "Tutorial sessions are not saved"
			} else {
				m.lastError = err.Error()
			}
			m.mode = modeNormal
			return m, clearMessageAfter(2 * time.Second)
		}
		m.dirty = false
		m.lastMessage = "Saved!"
		m.mode = modeNormal
		return m, clearMessageAfter(2 * time.Second)
	case "Q":
		return m, tea.Quit
	case "c":
		if m.selection.active {
			m.openInputMode("comment")
		}
	case "d":
		if m.selection.active {
			m.openInputMode("delete")
		}
	case "q":
		if m.selection.active {
			m.openInputMode("question")
		}
	case "e":
		if m.selection.active {
			m.openInputMode("expand")
		}
	case "k":
		if m.selection.active {
			m.openInputMode("keep")
		}
	case "u":
		if m.selection.active {
			m.openInputMode("unclear")
		}
	case "r":
		if m.selection.active {
			m.openInputMode("change")
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
	if m.inputTA == nil {
		m.mode = modeNormal
		return m, nil
	}

	switch msg.Type {
	case tea.KeyEnter:
		inputValue := strings.TrimSpace(m.inputTA.Value())
		if inputValue != "" {
			start, end := m.selection.lines()
			text := encodeAnnText(inputValue)

			if m.inputType == "change" {
				startLine := start + 1
				endLine := end + 1
				var lineRef string
				if startLine == endLine {
					lineRef = fmt.Sprintf("[line %d] -> ", startLine)
				} else {
					lineRef = fmt.Sprintf("[lines %d-%d] -> ", startLine, endLine)
				}
				text = lineRef + text
			}

			m.annotations = append(m.annotations, fem.Annotation{
				StartLine: start + 1,
				EndLine:   end + 1,
				Type:      m.inputType,
				Text:      text,
			})
			m.dirty = true
		}
		m.mode = modeNormal
		m.inputTA = nil
		m.inputType = ""
		m.selection = selection{}
		return m, nil

	case tea.KeyEsc:
		m.mode = modeNormal
		m.inputTA = nil
		m.inputType = ""
		return m, nil
	}

	var cmd tea.Cmd
	*m.inputTA, cmd = m.inputTA.Update(msg)
	return m, cmd
}

func (m *Model) tryEditAnnotation() {
	cursorLine := m.cursor + 1
	indices := m.annotationsOnLine(cursorLine)

	if len(indices) == 0 {
		m.lastError = "No annotation on this line"
		return
	}

	if len(indices) == 1 {
		m.openEditorForAnnotation(indices[0])
		return
	}

	m.mode = modePalette
	m.paletteKind = "annPick"
	m.paletteItems = indices
	m.paletteCursor = 0
}

func (m *Model) openEditorForAnnotation(annIndex int) {
	ann := m.annotations[annIndex]
	content := decodeAnnText(ann.Text)

	// Match the box calculation: boxTotalWidth = width - 2, innerWidth = boxTotalWidth - 4
	boxTotalWidth := m.width - 2
	if boxTotalWidth < 24 {
		boxTotalWidth = 64
	}
	taWidth := boxTotalWidth - 4

	ta := textarea.New()
	ta.SetValue(content)
	ta.Focus()
	ta.Prompt = ""
	ta.CharLimit = 0
	ta.ShowLineNumbers = false
	ta.SetWidth(taWidth)
	ta.KeyMap.InsertNewline.SetKeys("alt+enter", "ctrl+j")

	m.editor = &editorState{
		ta:       ta,
		start:    ann.StartLine - 1,
		end:      ann.EndLine - 1,
		annIndex: annIndex,
	}
	m.mode = modeEditor
}

func (m *Model) openEditor() {
	start, end := m.selection.lines()
	content := strings.Join(m.lines[start:end+1], "\n")

	// Match the box calculation: boxTotalWidth = width - 2, innerWidth = boxTotalWidth - 4
	boxTotalWidth := m.width - 2
	if boxTotalWidth < 24 {
		boxTotalWidth = 64
	}
	taWidth := boxTotalWidth - 4

	ta := textarea.New()
	ta.SetValue(content)
	ta.Focus()
	ta.Prompt = ""
	ta.CharLimit = 0
	ta.ShowLineNumbers = false
	ta.SetWidth(taWidth)
	ta.KeyMap.InsertNewline.SetKeys("alt+enter", "ctrl+j")

	m.editor = &editorState{
		ta:       ta,
		start:    start,
		end:      end,
		annIndex: -1,
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
			m.editor = nil
			m.mode = modeNormal
			return m, nil
		}
		m.editor.escPending = true
		return m, nil

	case tea.KeyCtrlS:
		m.saveEditorContent()
		return m, nil

	case tea.KeyCtrlC:
		m.editor = nil
		m.mode = modeNormal
		return m, nil

	case tea.KeyEnter:
		m.saveEditorContent()
		return m, nil
	}

	m.editor.escPending = false

	var cmd tea.Cmd
	m.editor.ta, cmd = m.editor.ta.Update(msg)
	return m, cmd
}

func (m Model) handleQuitConfirmMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		return m, tea.Quit
	default:
		m.mode = modeNormal
		m.lastMessage = "Quit cancelled"
		return m, clearMessageAfter(2 * time.Second)
	}
}

func (m *Model) saveEditorContent() {
	if m.editor == nil {
		return
	}

	edited := m.editor.ta.Value()

	if m.editor.annIndex >= 0 {
		m.annotations[m.editor.annIndex].Text = encodeAnnText(edited)
		m.dirty = true
		m.editor = nil
		m.mode = modeNormal
		return
	}

	encoded := encodeAnnText(edited)

	startLine := m.editor.start + 1
	endLine := m.editor.end + 1
	var prefix string
	if startLine == endLine {
		prefix = fmt.Sprintf("[line %d] -> ", startLine)
	} else {
		prefix = fmt.Sprintf("[lines %d-%d] -> ", startLine, endLine)
	}
	text := prefix + encoded

	for line := m.editor.start; line <= m.editor.end; line++ {
		m.annotations = append(m.annotations, fem.Annotation{
			StartLine: line + 1,
			EndLine:   line + 1,
			Type:      "change",
			Text:      text,
		})
	}
	m.dirty = true

	m.editor = nil
	m.mode = modeNormal
	m.selection = selection{}
}
