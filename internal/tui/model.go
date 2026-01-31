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
	modeSearch
	modeHelp
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

type searchState struct {
	query   string     // current search query
	matches []int      // line indices (0-indexed) that match
	current int        // index into matches for current match
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
	viewportTop     int // explicit viewport start line (-1 means auto-follow cursor)
	autoViewportTop int // used only when viewportTop == -1 (auto-follow)
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
	search         searchState  // search state (query, matches, current position)
	previewIndex   int          // index into annotations on current line for preview cycling
	previewLine    int          // line number (1-indexed) for which previewIndex is valid
	version        string       // fabbro version for display in help
}

func New(sess *session.Session) Model {
	return NewWithFile(sess, "")
}

func NewWithFile(sess *session.Session, sourceFile string) Model {
	return NewWithAnnotations(sess, sourceFile, []fem.Annotation{})
}

func NewWithAnnotations(sess *session.Session, sourceFile string, annotations []fem.Annotation) Model {
	return NewWithAll(sess, sourceFile, annotations, "")
}

func NewWithVersion(sess *session.Session, sourceFile string, version string) Model {
	return NewWithAll(sess, sourceFile, []fem.Annotation{}, version)
}

func NewWithAll(sess *session.Session, sourceFile string, annotations []fem.Annotation, version string) Model {
	lines := strings.Split(sess.Content, "\n")
	return Model{
		session:         sess,
		lines:           lines,
		cursor:          0,
		selection:       selection{},
		mode:            modeNormal,
		annotations:     annotations,
		highlighter:     highlight.New(sourceFile, sess.Content),
		sourceFile:      sourceFile,
		viewportTop:     -1, // auto-follow cursor
		autoViewportTop: 0,
		version:         version,
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

// previewedAnnotation returns the currently previewed annotation, or nil if none.
// Returns the annotation being shown in the preview panel based on cursor position
// and Tab-cycling state.
func (m *Model) previewedAnnotation() *fem.Annotation {
	cursorLine := m.cursor + 1 // 1-indexed
	indices := m.annotationsOnLine(cursorLine)
	if len(indices) == 0 {
		return nil
	}

	previewIdx := m.previewIndex
	if m.previewLine != cursorLine {
		previewIdx = 0
	}
	if previewIdx >= len(indices) {
		previewIdx = 0
	}

	return &m.annotations[indices[previewIdx]]
}
