package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charly-vibes/fabbro/internal/config"
	"github.com/charly-vibes/fabbro/internal/fem"
	"github.com/charly-vibes/fabbro/internal/session"
	tea "github.com/charmbracelet/bubbletea"
)

func newTestSession(content string) *session.Session {
	return &session.Session{
		ID:        "test-session",
		Content:   content,
		CreatedAt: time.Date(2026, 1, 11, 12, 0, 0, 0, time.UTC),
	}
}

func TestNew(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)

	if m.cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", m.cursor)
	}
	if m.selection.active {
		t.Error("expected selection to be inactive")
	}
	if m.mode != modeNormal {
		t.Errorf("expected modeNormal, got %d", m.mode)
	}
	if len(m.lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(m.lines))
	}
	if len(m.annotations) != 0 {
		t.Errorf("expected 0 annotations, got %d", len(m.annotations))
	}
}

func TestInit(t *testing.T) {
	sess := newTestSession("content")
	m := New(sess)
	cmd := m.Init()

	if cmd == nil {
		t.Error("expected EnterAltScreen command, got nil")
	}
}

func TestUpdateWindowSize(t *testing.T) {
	sess := newTestSession("content")
	m := New(sess)

	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	updated := newModel.(Model)

	if updated.width != 100 {
		t.Errorf("expected width 100, got %d", updated.width)
	}
	if updated.height != 50 {
		t.Errorf("expected height 50, got %d", updated.height)
	}
}

func TestNormalModeNavigation(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)

	// Move down with j
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newModel.(Model)
	if m.cursor != 1 {
		t.Errorf("expected cursor at 1 after j, got %d", m.cursor)
	}

	// Move down with arrow
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newModel.(Model)
	if m.cursor != 2 {
		t.Errorf("expected cursor at 2 after down, got %d", m.cursor)
	}

	// Can't go past end
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newModel.(Model)
	if m.cursor != 2 {
		t.Errorf("expected cursor to stay at 2, got %d", m.cursor)
	}

	// Move up with k
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = newModel.(Model)
	if m.cursor != 1 {
		t.Errorf("expected cursor at 1 after k, got %d", m.cursor)
	}

	// Move up with arrow
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = newModel.(Model)
	if m.cursor != 0 {
		t.Errorf("expected cursor at 0 after up, got %d", m.cursor)
	}

	// Can't go past start
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = newModel.(Model)
	if m.cursor != 0 {
		t.Errorf("expected cursor to stay at 0, got %d", m.cursor)
	}
}

func TestNormalModeSelection(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)

	// Select current line with v
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	m = newModel.(Model)
	if !m.selection.active || m.selection.anchor != 0 {
		t.Errorf("expected selection active at anchor 0, got active=%v anchor=%d", m.selection.active, m.selection.anchor)
	}

	// Toggle selection off
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	m = newModel.(Model)
	if m.selection.active {
		t.Error("expected selection inactive after toggle")
	}
}

func TestVCancelsMultiLineSelection(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3\nline4\nline5")
	m := New(sess)

	// Start selection at line 1
	m = sendKey(m, 'v')
	if !m.selection.active {
		t.Fatal("expected selection to be active")
	}

	// Extend selection to line 3
	m = sendKey(m, 'j')
	m = sendKey(m, 'j')
	if m.selection.cursor != 2 {
		t.Errorf("expected selection cursor at 2, got %d", m.selection.cursor)
	}

	// v should cancel the expanded selection (not start new one)
	m = sendKey(m, 'v')
	if m.selection.active {
		t.Error("v should cancel multi-line selection, but selection is still active")
	}
}

func TestNormalModeQuit(t *testing.T) {
	sess := newTestSession("content")
	m := New(sess)

	// Quit with Q (shift+q)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Q'}})
	if cmd == nil {
		t.Error("expected quit command from Q, got nil")
	}

	// Quit with ctrl+c
	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("expected quit command from ctrl+c, got nil")
	}
}

func TestLowercaseQDoesNotQuit(t *testing.T) {
	sess := newTestSession("content")
	m := New(sess)

	// q should NOT quit (reserved for question annotation)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd != nil {
		t.Error("lowercase q should not quit (reserved for question annotation)")
	}
}

func TestWriteKeySavesButDoesNotQuit(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	sess := &session.Session{
		ID:        "test-write-no-quit",
		Content:   "line1\nline2",
		CreatedAt: time.Date(2026, 1, 11, 12, 0, 0, 0, time.UTC),
	}
	m := New(sess)

	// Press w to save
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	m = newModel.(Model)

	// w should NOT quit
	if cmd != nil {
		t.Error("w should save but not quit; expected nil cmd")
	}

	// Verify file was saved
	sessionPath := filepath.Join(config.SessionsDir, "test-write-no-quit.fem")
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		t.Error("expected session file to be saved")
	}
}

func TestEnterCommentMode(t *testing.T) {
	sess := newTestSession("line1\nline2")
	m := New(sess)

	// Pressing c without selection does nothing
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m = newModel.(Model)
	if m.mode != modeNormal {
		t.Error("expected to stay in normal mode when no selection")
	}

	// Select a line first
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	m = newModel.(Model)

	// Now c enters input mode
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m = newModel.(Model)
	if m.mode != modeInput {
		t.Errorf("expected modeInput, got %d", m.mode)
	}
	if m.input != "" {
		t.Errorf("expected empty input, got %q", m.input)
	}
	if m.inputType != "comment" {
		t.Errorf("expected inputType 'comment', got %q", m.inputType)
	}
}

func TestAnnotationKeybindings(t *testing.T) {
	tests := []struct {
		key       rune
		wantType  string
	}{
		{'c', "comment"},
		{'d', "delete"},
		{'q', "question"},
		{'e', "expand"},
		{'u', "unclear"},
	}

	for _, tt := range tests {
		t.Run(tt.wantType, func(t *testing.T) {
			sess := newTestSession("line1")
			m := New(sess)

			// Select a line first
			m = sendKey(m, 'v')

			// Press the annotation key
			m = sendKey(m, tt.key)

			if m.mode != modeInput {
				t.Errorf("expected modeInput after pressing %c", tt.key)
			}
			if m.inputType != tt.wantType {
				t.Errorf("expected inputType %q, got %q", tt.wantType, m.inputType)
			}
		})
	}
}

func TestAnnotationKeybindingsRequireSelection(t *testing.T) {
	keys := []rune{'c', 'd', 'q', 'e', 'u'}

	for _, key := range keys {
		t.Run(string(key), func(t *testing.T) {
			sess := newTestSession("line1")
			m := New(sess)

			// Press key without selection
			m = sendKey(m, key)

			if m.mode != modeNormal {
				t.Errorf("key %c should not enter input mode without selection", key)
			}
		})
	}
}

func sendKey(m Model, key rune) Model {
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{key}})
	return newModel.(Model)
}

func sendKeyType(m Model, keyType tea.KeyType) Model {
	newModel, _ := m.Update(tea.KeyMsg{Type: keyType})
	return newModel.(Model)
}

func TestInputModeTyping(t *testing.T) {
	sess := newTestSession("line1")
	m := New(sess)

	// Select and enter comment mode
	m = sendKey(m, 'v')
	m = sendKey(m, 'c')

	// Type some text
	for _, r := range "hello" {
		m = sendKey(m, r)
	}
	if m.input != "hello" {
		t.Errorf("expected input 'hello', got %q", m.input)
	}

	// Backspace
	m = sendKeyType(m, tea.KeyBackspace)
	if m.input != "hell" {
		t.Errorf("expected input 'hell' after backspace, got %q", m.input)
	}
}

func TestInputModeSubmit(t *testing.T) {
	sess := newTestSession("line1")
	m := New(sess)

	// Select and enter comment mode
	m = sendKey(m, 'v')
	m = sendKey(m, 'c')

	// Type and submit
	for _, r := range "comment" {
		m = sendKey(m, r)
	}
	m = sendKeyType(m, tea.KeyEnter)

	if m.mode != modeNormal {
		t.Error("expected return to normal mode after enter")
	}
	if len(m.annotations) != 1 {
		t.Errorf("expected 1 annotation, got %d", len(m.annotations))
	}
	if m.annotations[0].Text != "comment" {
		t.Errorf("expected annotation text 'comment', got %q", m.annotations[0].Text)
	}
	if m.annotations[0].Type != "comment" {
		t.Errorf("expected annotation type 'comment', got %q", m.annotations[0].Type)
	}
	if m.selection.active {
		t.Error("expected selection cleared after submit")
	}
	if m.inputType != "" {
		t.Error("expected inputType cleared after submit")
	}
}

func TestInputModeSubmitAllTypes(t *testing.T) {
	tests := []struct {
		key      rune
		wantType string
	}{
		{'d', "delete"},
		{'q', "question"},
		{'e', "expand"},
		{'u', "unclear"},
	}

	for _, tt := range tests {
		t.Run(tt.wantType, func(t *testing.T) {
			sess := newTestSession("line1")
			m := New(sess)

			// Select and enter annotation mode
			m = sendKey(m, 'v')
			m = sendKey(m, tt.key)

			// Type and submit
			for _, r := range "test text" {
				m = sendKey(m, r)
			}
			m = sendKeyType(m, tea.KeyEnter)

			if len(m.annotations) != 1 {
				t.Fatalf("expected 1 annotation, got %d", len(m.annotations))
			}
			if m.annotations[0].Type != tt.wantType {
				t.Errorf("expected type %q, got %q", tt.wantType, m.annotations[0].Type)
			}
		})
	}
}

func TestInputModeEmptySubmit(t *testing.T) {
	sess := newTestSession("line1")
	m := New(sess)

	// Select and enter comment mode
	m = sendKey(m, 'v')
	m = sendKey(m, 'c')

	// Submit without typing
	m = sendKeyType(m, tea.KeyEnter)

	if len(m.annotations) != 0 {
		t.Error("expected no annotation for empty input")
	}
}

func TestInputModeEscape(t *testing.T) {
	sess := newTestSession("line1")
	m := New(sess)

	// Select and enter comment mode
	m = sendKey(m, 'v')
	m = sendKey(m, 'c')

	// Type something then escape
	for _, r := range "draft" {
		m = sendKey(m, r)
	}
	m = sendKeyType(m, tea.KeyEsc)

	if m.mode != modeNormal {
		t.Error("expected return to normal mode after escape")
	}
	if m.input != "" {
		t.Error("expected input cleared after escape")
	}
	if len(m.annotations) != 0 {
		t.Error("expected no annotation after escape")
	}
}

func TestView(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)
	m.width = 80
	m.height = 20

	view := m.View()

	// Should contain session ID
	if !strings.Contains(view, "test-session") {
		t.Error("view should contain session ID")
	}

	// Should contain lines
	if !strings.Contains(view, "line1") {
		t.Error("view should contain line content")
	}

	// Should contain help text
	if !strings.Contains(view, "[v]select") {
		t.Error("view should contain help text in normal mode")
	}
}

func TestViewInputMode(t *testing.T) {
	sess := newTestSession("line1")
	m := New(sess)
	m.width = 80
	m.height = 20
	m.mode = modeInput
	m.inputType = "comment"
	m.input = "typing"

	view := m.View()

	if !strings.Contains(view, "Comment: typing") {
		t.Error("view should show comment input in input mode")
	}
}

func TestViewWithSelection(t *testing.T) {
	sess := newTestSession("line1\nline2")
	m := New(sess)
	m.width = 80
	m.height = 20
	m.selection = selection{active: true, anchor: 0, cursor: 0}

	view := m.View()

	// Should have selection indicator (◆ for anchor)
	if !strings.Contains(view, "◆") {
		t.Error("view should show selection indicator")
	}
}

func TestViewScrolling(t *testing.T) {
	// Create content with many lines
	lines := make([]string, 50)
	for i := range lines {
		lines[i] = "line"
	}
	sess := newTestSession(strings.Join(lines, "\n"))
	m := New(sess)
	m.width = 80
	m.height = 20
	m.cursor = 40

	view := m.View()

	// Should have rendered something (scrolled view)
	if len(view) == 0 {
		t.Error("view should not be empty")
	}
}

func TestViewSmallHeight(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)
	m.width = 80
	m.height = 5 // Very small height triggers visibleLines < 5 branch

	view := m.View()

	// Should still render
	if len(view) == 0 {
		t.Error("view should not be empty with small height")
	}
}

func TestWriteCommand(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	sess := &session.Session{
		ID:        "test-write-cmd",
		Content:   "line1",
		CreatedAt: time.Date(2026, 1, 11, 12, 0, 0, 0, time.UTC),
	}
	m := New(sess)

	// w command should save but NOT quit
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	if cmd != nil {
		t.Error("w should save without quitting; expected nil cmd")
	}

	// Verify file was saved
	sessionPath := filepath.Join(config.SessionsDir, "test-write-cmd.fem")
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		t.Error("expected session file to be saved")
	}
}

func TestBackspaceOnEmptyInput(t *testing.T) {
	sess := newTestSession("line1")
	m := New(sess)

	// Enter input mode
	m = sendKey(m, 'v')
	m = sendKey(m, 'c')

	// Backspace on empty input should not panic
	m = sendKeyType(m, tea.KeyBackspace)
	if m.input != "" {
		t.Error("input should remain empty")
	}
}

func TestMultiCharKeyIgnored(t *testing.T) {
	sess := newTestSession("line1")
	m := New(sess)

	// Enter input mode
	m = sendKey(m, 'v')
	m = sendKey(m, 'c')

	// Multi-char key messages should be ignored
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a', 'b'}})
	m = newModel.(Model)
	if m.input != "" {
		t.Error("multi-char key should be ignored")
	}
}

func TestUnknownMessageType(t *testing.T) {
	sess := newTestSession("line1")
	m := New(sess)

	// Unknown message types should not crash
	type customMsg struct{}
	newModel, cmd := m.Update(customMsg{})
	if newModel == nil {
		t.Error("model should not be nil")
	}
	if cmd != nil {
		t.Error("command should be nil for unknown message")
	}
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	sess := &session.Session{
		ID:        "test-save-session",
		Content:   "line1\nline2",
		CreatedAt: time.Date(2026, 1, 11, 12, 0, 0, 0, time.UTC),
	}
	m := New(sess)

	// Add a comment annotation (1-based line numbers)
	m.annotations = append(m.annotations, fem.Annotation{StartLine: 1, EndLine: 1, Type: "comment", Text: "my comment"})

	// Call save
	if err := m.save(); err != nil {
		t.Fatalf("save() failed: %v", err)
	}

	// Check file was written
	sessionFile := filepath.Join(config.SessionsDir, "test-save-session.fem")
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		t.Fatalf("save() did not create file: %v", err)
	}

	content := string(data)

	// Should have frontmatter
	if !strings.HasPrefix(content, "---\n") {
		t.Error("saved file should start with frontmatter")
	}

	// Should have annotation in FEM format
	if !strings.Contains(content, "{>> my comment <<}") {
		t.Error("saved file should contain FEM annotation")
	}

	// Line without annotation should be unchanged
	if !strings.Contains(content, "line2") {
		t.Error("saved file should contain line2")
	}
}

func TestCtrlDScrollsDown(t *testing.T) {
	// Create 50 lines
	lines := make([]string, 50)
	for i := range lines {
		lines[i] = fmt.Sprintf("line%d", i+1)
	}
	sess := newTestSession(strings.Join(lines, "\n"))
	m := New(sess)
	m.height = 20 // viewport height

	// Ctrl+d should move cursor down by half page (8 lines with height 20)
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	m = newModel.(Model)

	// Half page = (height - 4) / 2 = 8
	if m.cursor != 8 {
		t.Errorf("expected cursor at 8 after Ctrl+d, got %d", m.cursor)
	}
}

func TestCtrlUScrollsUp(t *testing.T) {
	lines := make([]string, 50)
	for i := range lines {
		lines[i] = fmt.Sprintf("line%d", i+1)
	}
	sess := newTestSession(strings.Join(lines, "\n"))
	m := New(sess)
	m.height = 20
	m.cursor = 20

	// Ctrl+u should move cursor up by half page
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	m = newModel.(Model)

	if m.cursor != 12 {
		t.Errorf("expected cursor at 12 after Ctrl+u, got %d", m.cursor)
	}
}

func TestCtrlDClampsToEnd(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)
	m.height = 20
	m.cursor = 1

	// Ctrl+d should clamp to last line
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	m = newModel.(Model)

	if m.cursor != 2 {
		t.Errorf("expected cursor clamped to 2, got %d", m.cursor)
	}
}

func TestCtrlUClampsToStart(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)
	m.height = 20
	m.cursor = 1

	// Ctrl+u should clamp to first line
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	m = newModel.(Model)

	if m.cursor != 0 {
		t.Errorf("expected cursor clamped to 0, got %d", m.cursor)
	}
}

func TestGGJumpsToFirstLine(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3\nline4\nline5")
	m := New(sess)
	m.cursor = 4

	// First g sets pending
	m = sendKey(m, 'g')
	if !m.gPending {
		t.Error("expected gPending after first g")
	}

	// Second g jumps to first line
	m = sendKey(m, 'g')
	if m.cursor != 0 {
		t.Errorf("expected cursor at 0 after gg, got %d", m.cursor)
	}
	if m.gPending {
		t.Error("gPending should be cleared after gg")
	}
}

func TestGJumpsToLastLine(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3\nline4\nline5")
	m := New(sess)

	// G jumps to last line
	m = sendKey(m, 'G')
	if m.cursor != 4 {
		t.Errorf("expected cursor at 4 after G, got %d", m.cursor)
	}
}

func TestGPendingClearedByOtherKey(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)

	// g sets pending
	m = sendKey(m, 'g')
	if !m.gPending {
		t.Error("expected gPending after g")
	}

	// j clears pending and moves down
	m = sendKey(m, 'j')
	if m.gPending {
		t.Error("gPending should be cleared by other key")
	}
	if m.cursor != 1 {
		t.Errorf("expected cursor at 1 after j, got %d", m.cursor)
	}
}

func TestPaletteModeOpen(t *testing.T) {
	sess := newTestSession("line1\nline2")
	m := New(sess)

	// Select a line first
	m = sendKey(m, 'v')

	// SPC opens palette mode
	m = sendKeyType(m, tea.KeySpace)
	if m.mode != modePalette {
		t.Errorf("expected modePalette after SPC, got %d", m.mode)
	}
}

func TestPaletteAlwaysOpens(t *testing.T) {
	sess := newTestSession("line1")
	m := New(sess)

	// SPC without selection should still open palette
	m = sendKeyType(m, tea.KeySpace)
	if m.mode != modePalette {
		t.Error("SPC should open palette even without selection")
	}
}

func TestPaletteShowsMarkupOnlyWithSelection(t *testing.T) {
	sess := newTestSession("line1\nline2")
	m := New(sess)
	m.width = 80
	m.height = 24

	// Without selection: palette should show general commands only
	m = sendKeyType(m, tea.KeySpace)
	view := m.View()
	if strings.Contains(view, "[c]omment") {
		t.Error("palette without selection should not show markup commands")
	}
	if !strings.Contains(view, "[w]rite") {
		t.Error("palette should show general commands like [w]rite")
	}

	// With selection: palette should show markup commands
	m.mode = modeNormal
	m = sendKey(m, 'v') // start selection
	m = sendKeyType(m, tea.KeySpace)
	view = m.View()
	if !strings.Contains(view, "[c]omment") {
		t.Error("palette with selection should show markup commands")
	}
}

func TestPaletteAnnotationKeys(t *testing.T) {
	tests := []struct {
		key      rune
		wantType string
	}{
		{'c', "comment"},
		{'d', "delete"},
		{'q', "question"},
		{'e', "expand"},
		{'k', "keep"}, // k only available in palette!
		{'u', "unclear"},
	}

	for _, tt := range tests {
		t.Run(tt.wantType, func(t *testing.T) {
			sess := newTestSession("line1")
			m := New(sess)

			// Select, open palette
			m = sendKey(m, 'v')
			m = sendKeyType(m, tea.KeySpace)

			// Press annotation key in palette
			m = sendKey(m, tt.key)

			if m.mode != modeInput {
				t.Errorf("expected modeInput after pressing %c in palette", tt.key)
			}
			if m.inputType != tt.wantType {
				t.Errorf("expected inputType %q, got %q", tt.wantType, m.inputType)
			}
		})
	}
}

func TestPaletteEscDismisses(t *testing.T) {
	sess := newTestSession("line1")
	m := New(sess)

	// Select, open palette
	m = sendKey(m, 'v')
	m = sendKeyType(m, tea.KeySpace)

	// ESC dismisses palette
	m = sendKeyType(m, tea.KeyEscape)
	if m.mode != modeNormal {
		t.Errorf("expected modeNormal after ESC in palette, got %d", m.mode)
	}
	// Selection should remain
	if !m.selection.active {
		t.Error("selection should remain after ESC in palette")
	}
}

func TestPaletteUnknownKeyDismisses(t *testing.T) {
	sess := newTestSession("line1")
	m := New(sess)

	// Select, open palette
	m = sendKey(m, 'v')
	m = sendKeyType(m, tea.KeySpace)

	// Unknown key dismisses palette without action
	m = sendKey(m, 'x')
	if m.mode != modeNormal {
		t.Errorf("expected modeNormal after unknown key in palette, got %d", m.mode)
	}
}

func TestPaletteWriteCommand(t *testing.T) {
	sess := newTestSession("line1")
	m := New(sess)

	// Open palette (no selection needed for general commands)
	m = sendKeyType(m, tea.KeySpace)
	if m.mode != modePalette {
		t.Fatal("expected palette mode")
	}

	// Press w to save
	m = sendKey(m, 'w')
	if m.mode != modeNormal {
		t.Errorf("expected modeNormal after w in palette, got %d", m.mode)
	}
}

func TestPaletteQuitCommand(t *testing.T) {
	sess := newTestSession("line1")
	m := New(sess)

	// Open palette
	m = sendKeyType(m, tea.KeySpace)

	// Press Q to quit - should return tea.Quit command
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Q'}})
	m = newModel.(Model)

	if cmd == nil {
		t.Error("expected quit command from Q in palette")
	}
}

func TestPaletteMarkupIgnoredWithoutSelection(t *testing.T) {
	sess := newTestSession("line1")
	m := New(sess)

	// Open palette without selection
	m = sendKeyType(m, tea.KeySpace)

	// Pressing markup keys should do nothing (stay in palette)
	for _, key := range []rune{'c', 'd', 'e', 'k', 'u'} {
		m.mode = modePalette // reset
		m = sendKey(m, key)
		if m.mode == modeInput {
			t.Errorf("key %c should not enter input mode without selection", key)
		}
	}
}

func TestPaletteViewShowsOverlay(t *testing.T) {
	sess := newTestSession("line1")
	m := New(sess)
	m.width = 80
	m.height = 20
	m.selection = selection{active: true, anchor: 0, cursor: 0}
	m.mode = modePalette

	view := m.View()

	// Should show palette header
	if !strings.Contains(view, "Annotations") {
		t.Error("palette view should show 'Annotations' header")
	}
	// Should show all annotation options including [k]eep
	if !strings.Contains(view, "[k]") {
		t.Error("palette view should show [k]eep option")
	}
}

func TestMultiLineSelectionVThenJK(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3\nline4\nline5")
	m := New(sess)
	m.cursor = 2

	// v starts selection at current line
	m = sendKey(m, 'v')
	if !m.selection.active {
		t.Error("expected selection to be active after v")
	}
	if m.selection.anchor != 2 {
		t.Errorf("expected anchor at 2, got %d", m.selection.anchor)
	}

	// j extends selection down
	m = sendKey(m, 'j')
	if m.selection.cursor != 3 {
		t.Errorf("expected selection cursor at 3, got %d", m.selection.cursor)
	}

	// k moves selection cursor up
	m = sendKey(m, 'k')
	if m.selection.cursor != 2 {
		t.Errorf("expected selection cursor at 2, got %d", m.selection.cursor)
	}
}

func TestMultiLineSelectionLines(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3\nline4\nline5")
	m := New(sess)

	// Select from line 3 down to line 1 (anchor > cursor)
	m.selection = selection{active: true, anchor: 3, cursor: 1}
	start, end := m.selection.lines()
	if start != 1 || end != 3 {
		t.Errorf("expected lines (1, 3), got (%d, %d)", start, end)
	}

	// Select from line 1 down to line 3 (anchor < cursor)
	m.selection = selection{active: true, anchor: 1, cursor: 3}
	start, end = m.selection.lines()
	if start != 1 || end != 3 {
		t.Errorf("expected lines (1, 3), got (%d, %d)", start, end)
	}
}

func TestMultiLineSelectionEscClears(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)

	// Start selection
	m = sendKey(m, 'v')
	m = sendKey(m, 'j')

	// ESC clears selection
	m = sendKeyType(m, tea.KeyEscape)
	if m.selection.active {
		t.Error("expected selection to be cleared after ESC")
	}
}

func TestMultiLineSelectionAnnotation(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	sess := &session.Session{
		ID:        "test-multiline",
		Content:   "line1\nline2\nline3\nline4\nline5",
		CreatedAt: time.Date(2026, 1, 11, 12, 0, 0, 0, time.UTC),
	}
	m := New(sess)

	// Select lines 1-3
	m.cursor = 1
	m = sendKey(m, 'v')
	m = sendKey(m, 'j')
	m = sendKey(m, 'j') // now at line 3, selection spans 1-3

	// Add comment
	m = sendKey(m, 'c')
	// Type comment text
	for _, r := range "multi-line comment" {
		m = sendKey(m, r)
	}
	m = sendKeyType(m, tea.KeyEnter)

	// Should have 3 annotations (one per line)
	if len(m.annotations) != 3 {
		t.Errorf("expected 3 annotations for multi-line selection, got %d", len(m.annotations))
	}

	// All should be for lines 1, 2, 3
	for i, ann := range m.annotations {
		if ann.StartLine != i+1 {
			t.Errorf("expected annotation %d at line %d, got %d", i, i+1, ann.StartLine)
		}
		if ann.Type != "comment" {
			t.Errorf("expected type 'comment', got %q", ann.Type)
		}
	}
}

func TestMultiLineSelectionViewHighlight(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3\nline4\nline5")
	m := New(sess)
	m.width = 80
	m.height = 20
	m.selection = selection{active: true, anchor: 1, cursor: 3}
	m.cursor = 3

	view := m.View()

	// Should show selection indicator on multiple lines
	// Count occurrences of selection indicators (◆ anchor + ▌ extended)
	count := strings.Count(view, "◆") + strings.Count(view, "▌")
	if count != 3 {
		t.Errorf("expected 3 selection indicators, got %d", count)
	}
}

func TestSelectionShowsLineCount(t *testing.T) {
	sess := newTestSession("line 1\nline 2\nline 3\nline 4\nline 5")
	m := New(sess)

	// Start selection on line 2
	m.cursor = 1
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	m = newModel.(Model)

	// Extend to line 4
	m.cursor = 3
	m.selection.cursor = 3

	view := m.View()

	// Title should show line count
	if !strings.Contains(view, "[3 lines selected]") {
		t.Errorf("view should show line count, got:\n%s", view)
	}
}

func TestSelectionShowsAnchorIndicator(t *testing.T) {
	sess := newTestSession("line 1\nline 2\nline 3")
	m := New(sess)

	// Start selection on line 2
	m.cursor = 1
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	m = newModel.(Model)

	// Extend to line 3
	m.cursor = 2
	m.selection.cursor = 2

	view := m.View()

	// Should show anchor indicator (◆) on line 2 and extended indicator (▌) on line 3
	if !strings.Contains(view, "◆") {
		t.Errorf("view should show anchor indicator ◆, got:\n%s", view)
	}
	if !strings.Contains(view, "▌") {
		t.Errorf("view should show extended selection indicator ▌, got:\n%s", view)
	}
}

func TestSaveErrorHandling(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Create sessions directory, then make it read-only
	config.Init()
	os.Chmod(config.SessionsDir, 0444)
	defer os.Chmod(config.SessionsDir, 0755)

	sess := &session.Session{
		ID:        "test-save-error",
		Content:   "line1\nline2",
		CreatedAt: time.Date(2026, 1, 11, 12, 0, 0, 0, time.UTC),
	}
	m := New(sess)

	err := m.save()
	if err == nil {
		t.Error("expected error when saving to read-only directory")
	}
	if !strings.Contains(err.Error(), "failed to save session") {
		t.Errorf("expected error to contain 'failed to save session', got: %v", err)
	}
}

func TestViewportScrollsToKeepCursorVisible(t *testing.T) {
	// Create document with 50 lines
	var lines []string
	for i := 1; i <= 50; i++ {
		lines = append(lines, fmt.Sprintf("line %d", i))
	}
	content := strings.Join(lines, "\n")

	sess := newTestSession(content)
	m := New(sess)
	m.width = 80
	m.height = 20 // 16 visible lines (height - 4 for UI)

	// Position cursor at line 40 (last visible line when viewport is at bottom)
	m.cursor = 40

	// Start selection
	m = sendKey(m, 'v')

	// Move up 5 lines - cursor should be visible and selection should show
	for i := 0; i < 5; i++ {
		m = sendKey(m, 'k')
	}

	view := m.View()

	// The cursor line (35) should be visible in the output
	if !strings.Contains(view, "line 36") { // line 36 because 0-indexed line 35
		t.Errorf("cursor line should be visible after scrolling up, got:\n%s", view)
	}

	// The current cursor position should be marked with > and selection indicator
	if !strings.Contains(view, ">◆") && !strings.Contains(view, ">▌") {
		t.Errorf("current line should show cursor and selection indicator")
	}
}

func TestViewportScrollsUpWhenSelecting(t *testing.T) {
	// Bug scenario: start at line 0, scroll down to bottom of initial viewport,
	// start selection, then scroll up - viewport should follow cursor
	var lines []string
	for i := 1; i <= 50; i++ {
		lines = append(lines, fmt.Sprintf("line %d content here", i))
	}
	content := strings.Join(lines, "\n")

	sess := newTestSession(content)
	m := New(sess)
	m.width = 80
	m.height = 20 // 16 visible lines (height - 4 for UI)

	// Navigate to line 15 (last visible line in initial viewport: 0-15)
	for i := 0; i < 15; i++ {
		m = sendKey(m, 'j')
	}
	if m.cursor != 15 {
		t.Fatalf("expected cursor at 15, got %d", m.cursor)
	}

	// Start selection at line 15
	m = sendKey(m, 'v')

	// Move up 15 lines to get back to line 0
	for i := 0; i < 15; i++ {
		m = sendKey(m, 'k')
	}
	if m.cursor != 0 {
		t.Fatalf("expected cursor at 0, got %d", m.cursor)
	}

	view := m.View()

	// Line 1 (cursor at index 0) MUST be visible
	if !strings.Contains(view, "line 1 content here") {
		t.Errorf("cursor line (line 1) should be visible after scrolling up\ngot:\n%s", view)
	}

	// Cursor indicator should be visible on line 1 (cursor moved from anchor, so ▌)
	if !strings.Contains(view, ">▌   1 │ line 1") {
		t.Errorf("cursor indicator '>' should be on line 1 with selection indicator, got:\n%s", view)
	}
}

func TestSaveAllAnnotationTypes(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	sess := &session.Session{
		ID:        "test-save-types",
		Content:   "line1\nline2\nline3\nline4\nline5\nline6",
		CreatedAt: time.Date(2026, 1, 11, 12, 0, 0, 0, time.UTC),
	}
	m := New(sess)

	// Add one of each annotation type
	m.annotations = []fem.Annotation{
		{StartLine: 1, EndLine: 1, Type: "comment", Text: "a comment"},
		{StartLine: 2, EndLine: 2, Type: "delete", Text: "DELETE: remove"},
		{StartLine: 3, EndLine: 3, Type: "question", Text: "why?"},
		{StartLine: 4, EndLine: 4, Type: "expand", Text: "EXPAND: more"},
		{StartLine: 5, EndLine: 5, Type: "keep", Text: "KEEP: good"},
		{StartLine: 6, EndLine: 6, Type: "unclear", Text: "UNCLEAR: huh"},
	}

	if err := m.save(); err != nil {
		t.Fatalf("save() failed: %v", err)
	}

	sessionFile := filepath.Join(config.SessionsDir, "test-save-types.fem")
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		t.Fatalf("save() did not create file: %v", err)
	}

	content := string(data)

	// Check each annotation type has correct FEM markers
	expected := []string{
		"{>> a comment <<}",
		"{-- DELETE: remove --}",
		"{?? why? ??}",
		"{!! EXPAND: more !!}",
		"{== KEEP: good ==}",
		"{~~ UNCLEAR: huh ~~}",
	}

	for _, exp := range expected {
		if !strings.Contains(content, exp) {
			t.Errorf("saved file should contain %q", exp)
		}
	}
}
