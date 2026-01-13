package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charly-vibes/fabbro/internal/config"
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
	if m.selected != -1 {
		t.Errorf("expected selected -1, got %d", m.selected)
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
	if m.selected != 0 {
		t.Errorf("expected selected 0, got %d", m.selected)
	}

	// Toggle selection off
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	m = newModel.(Model)
	if m.selected != -1 {
		t.Errorf("expected selected -1 after toggle, got %d", m.selected)
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
	if m.selected != -1 {
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
	m.selected = 0

	view := m.View()

	// Should have selection indicator
	if !strings.Contains(view, "●") {
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

func TestMax(t *testing.T) {
	if max(5, 3) != 5 {
		t.Error("max(5, 3) should be 5")
	}
	if max(2, 7) != 7 {
		t.Error("max(2, 7) should be 7")
	}
	if max(4, 4) != 4 {
		t.Error("max(4, 4) should be 4")
	}
}

func TestWriteCommand(t *testing.T) {
	sess := newTestSession("line1")
	m := New(sess)

	// w command should return quit
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	if cmd == nil {
		t.Error("expected quit command after w")
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

	// Add a comment annotation
	m.annotations = append(m.annotations, Annotation{Line: 0, Type: "comment", Text: "my comment"})

	// Call save
	m.save()

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

func TestPaletteRequiresSelection(t *testing.T) {
	sess := newTestSession("line1")
	m := New(sess)

	// SPC without selection should not open palette
	m = sendKeyType(m, tea.KeySpace)
	if m.mode != modeNormal {
		t.Error("SPC should not open palette without selection")
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
	if m.selected != 0 {
		t.Errorf("selection should remain after ESC, got %d", m.selected)
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

func TestPaletteViewShowsOverlay(t *testing.T) {
	sess := newTestSession("line1")
	m := New(sess)
	m.width = 80
	m.height = 20
	m.selected = 0
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
	m.annotations = []Annotation{
		{Line: 0, Type: "comment", Text: "a comment"},
		{Line: 1, Type: "delete", Text: "DELETE: remove"},
		{Line: 2, Type: "question", Text: "why?"},
		{Line: 3, Type: "expand", Text: "EXPAND: more"},
		{Line: 4, Type: "keep", Text: "KEEP: good"},
		{Line: 5, Type: "unclear", Text: "UNCLEAR: huh"},
	}

	m.save()

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
