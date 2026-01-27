package tui

import (
	"strings"
	"time"

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
	modeQuitConfirm
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

type Model struct {
	session        *session.Session
	lines          []string
	cursor         int
	selection      selection
	mode           mode
	inputType      string          // annotation type being entered: "comment", "delete", etc.
	inputTA        *textarea.Model // textarea for annotation input (text wrapping + multiline)
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
	lastCtrlC      time.Time    // timestamp of last CTRL+C press for double-tap quit
	dirty          bool         // true when there are unsaved changes
}

func New(sess *session.Session) Model {
	return NewWithFile(sess, "")
}

func NewWithFile(sess *session.Session, sourceFile string) Model {
	return NewWithAnnotations(sess, sourceFile, []fem.Annotation{})
}

func NewWithAnnotations(sess *session.Session, sourceFile string, annotations []fem.Annotation) Model {
	lines := strings.Split(sess.Content, "\n")
	return Model{
		session:     sess,
		lines:       lines,
		cursor:      0,
		selection:   selection{},
		mode:        modeNormal,
		annotations: annotations,
		highlighter: highlight.New(sourceFile, sess.Content),
		sourceFile:  sourceFile,
		viewportTop: -1, // auto-follow cursor
	}
}

func (m Model) Init() tea.Cmd {
	return tea.EnterAltScreen
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
