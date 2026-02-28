package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charly-vibes/fabbro/internal/config"
	"github.com/charly-vibes/fabbro/internal/fem"
	"github.com/charly-vibes/fabbro/internal/session"
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

	// First ctrl+c shows warning, does not quit
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	m = newModel.(Model)
	if cmd == nil {
		t.Error("expected message clear command from first ctrl+c, got nil")
	}
	if m.lastMessage != "Press CTRL+C again to quit" {
		t.Errorf("expected warning message, got %q", m.lastMessage)
	}

	// Second ctrl+c enters quit confirm mode
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	m = newModel.(Model)
	if m.mode != modeQuitConfirm {
		t.Errorf("expected modeQuitConfirm, got %v", m.mode)
	}

	// 'y' confirms quit
	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if cmd == nil {
		t.Error("expected quit command from 'y' in quit confirm mode, got nil")
	}
}

func TestQuitConfirmCancelled(t *testing.T) {
	sess := newTestSession("content")
	m := New(sess)
	m.mode = modeQuitConfirm

	// 'n' cancels quit
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = newModel.(Model)
	if m.mode != modeNormal {
		t.Errorf("expected modeNormal after cancel, got %v", m.mode)
	}
	if cmd == nil {
		t.Error("expected message clear command, got nil")
	}
}

func TestQuitConfirmShowsUnsavedWarning(t *testing.T) {
	sess := newTestSession("content")
	m := New(sess)
	m.mode = modeQuitConfirm
	m.dirty = true
	m.width = 80
	m.height = 20

	view := m.View()
	if !strings.Contains(view, "Unsaved changes") {
		t.Errorf("expected unsaved warning in view, got:\n%s", view)
	}
}

func TestQKeyDoesNotQuit(t *testing.T) {
	sess := newTestSession("content")
	m := New(sess)

	// Q should NOT quit anymore
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Q'}})
	if cmd != nil {
		t.Error("Q key should not quit")
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

	// w should NOT quit (cmd is for clearing message, not quitting)
	// Check that it's not a tea.Quit by verifying the model hasn't exited
	if m.lastMessage != "Saved!" {
		t.Error("expected 'Saved!' message after w")
	}

	// Verify file was saved
	sessionPath := filepath.Join(config.SessionsDir, "test-write-no-quit.fem")
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		t.Error("expected session file to be saved")
	}

	// Verify a command was returned (for clearing the message)
	if cmd == nil {
		t.Error("expected cmd to clear message after save")
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
	if m.inputTA == nil {
		t.Fatal("expected inputTA to be initialized")
	}
	if m.inputTA.Value() != "" {
		t.Errorf("expected empty input, got %q", m.inputTA.Value())
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
		{'r', "change"},
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
	keys := []rune{'c', 'd', 'q', 'e', 'u', 'r'}

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

	if m.inputTA == nil {
		t.Fatal("expected inputTA to be initialized")
	}

	// Type some text
	for _, r := range "hello" {
		m = sendKey(m, r)
	}
	if m.inputTA.Value() != "hello" {
		t.Errorf("expected input 'hello', got %q", m.inputTA.Value())
	}

	// Backspace
	m = sendKeyType(m, tea.KeyBackspace)
	if m.inputTA.Value() != "hell" {
		t.Errorf("expected input 'hell' after backspace, got %q", m.inputTA.Value())
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

func TestInputModeMultiline(t *testing.T) {
	sess := newTestSession("line1")
	m := New(sess)

	// Select and enter comment mode
	m = sendKey(m, 'v')
	m = sendKey(m, 'c')

	if m.inputTA == nil {
		t.Fatal("expected inputTA to be initialized")
	}

	// Type first line
	for _, r := range "first" {
		m = sendKey(m, r)
	}

	// Shift+Enter to insert newline
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\n'}, Alt: false})
	m = newM.(Model)

	// Type second line
	for _, r := range "second" {
		m = sendKey(m, r)
	}

	// Should still be in input mode
	if m.mode != modeInput {
		t.Error("expected to remain in input mode after Shift+Enter")
	}

	// Value should contain both lines
	val := m.inputTA.Value()
	if !strings.Contains(val, "first") || !strings.Contains(val, "second") {
		t.Errorf("expected multiline content, got %q", val)
	}

	// Submit with Enter
	m = sendKeyType(m, tea.KeyEnter)

	if m.mode != modeNormal {
		t.Error("expected return to normal mode after enter")
	}
	if len(m.annotations) != 1 {
		t.Errorf("expected 1 annotation, got %d", len(m.annotations))
	}
	// Newlines are encoded as \n in the annotation text
	if !strings.Contains(m.annotations[0].Text, "\\n") {
		t.Errorf("expected encoded newline in annotation text, got %q", m.annotations[0].Text)
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

func TestChangeAnnotationIncludesLineReference(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3\nline4\nline5")
	m := New(sess)

	// Select lines 2-4 (0-indexed: 1-3)
	m.cursor = 1
	m = sendKey(m, 'v')
	m = sendKey(m, 'j')
	m = sendKey(m, 'j')

	// Enter change mode and type replacement
	m = sendKey(m, 'r')
	for _, r := range "new content" {
		m = sendKey(m, r)
	}
	m = sendKeyType(m, tea.KeyEnter)

	if len(m.annotations) != 1 {
		t.Fatalf("expected 1 annotation for multi-line selection, got %d", len(m.annotations))
	}

	ann := m.annotations[0]
	if ann.Type != "change" {
		t.Errorf("expected type 'change', got %q", ann.Type)
	}
	if ann.StartLine != 2 || ann.EndLine != 4 {
		t.Errorf("expected StartLine=2, EndLine=4, got StartLine=%d, EndLine=%d", ann.StartLine, ann.EndLine)
	}
	if !strings.Contains(ann.Text, "->") {
		t.Errorf("expected text to contain '->', got %q", ann.Text)
	}
	if !strings.Contains(ann.Text, "new content") {
		t.Errorf("expected text to contain 'new content', got %q", ann.Text)
	}
	if !strings.Contains(ann.Text, "[lines 2-4]") {
		t.Errorf("expected text to contain '[lines 2-4]', got %q", ann.Text)
	}
}

func TestChangeAnnotationSingleLine(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)

	// Select single line (line 2, 0-indexed: 1)
	m.cursor = 1
	m = sendKey(m, 'v')

	// Enter change mode and type replacement
	m = sendKey(m, 'r')
	for _, r := range "replaced" {
		m = sendKey(m, r)
	}
	m = sendKeyType(m, tea.KeyEnter)

	if len(m.annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(m.annotations))
	}

	// Should use singular "line" for single line
	if !strings.Contains(m.annotations[0].Text, "[line 2]") {
		t.Errorf("expected text to contain '[line 2]', got %q", m.annotations[0].Text)
	}
	if !strings.Contains(m.annotations[0].Text, "-> replaced") {
		t.Errorf("expected text to contain '-> replaced', got %q", m.annotations[0].Text)
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
	if m.inputTA != nil {
		t.Error("expected inputTA cleared after escape")
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
	if !strings.Contains(view, "[v]sel") {
		t.Error("view should contain help text in normal mode")
	}
}

func TestViewInputMode(t *testing.T) {
	sess := newTestSession("line1")
	m := New(sess)
	m.width = 80
	m.height = 20

	// Start selection and enter comment mode
	m = sendKey(m, 'v')
	m = sendKey(m, 'c')

	// Type some text
	for _, r := range "typing" {
		m = sendKey(m, r)
	}

	view := m.View()

	if !strings.Contains(view, "Comment:") {
		t.Error("view should show Comment: prompt in input mode")
	}
	if !strings.Contains(view, "typing") {
		t.Error("view should show typed text in input mode")
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

	// w command should save and return a cmd (for clearing message)
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	m = newModel.(Model)

	// Verify success message is set
	if m.lastMessage != "Saved!" {
		t.Error("expected 'Saved!' message after w")
	}

	// Verify a command was returned (for clearing the message)
	if cmd == nil {
		t.Error("expected cmd to clear message after save")
	}

	// Verify file was saved
	sessionPath := filepath.Join(config.SessionsDir, "test-write-cmd.fem")
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		t.Error("expected session file to be saved")
	}
}

func TestWriteCommandShowsSavedMessage(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	sess := &session.Session{
		ID:        "test-save-msg",
		Content:   "line1",
		CreatedAt: time.Date(2026, 1, 11, 12, 0, 0, 0, time.UTC),
	}
	m := New(sess)
	m.width = 80
	m.height = 20

	// Press w to save
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	m = newModel.(Model)

	// View should contain success message
	view := m.View()
	if !strings.Contains(view, "✓ Saved!") {
		t.Errorf("expected view to contain '✓ Saved!', got:\n%s", view)
	}
}

func TestClearMessageClearsLastMessage(t *testing.T) {
	sess := newTestSession("line1")
	m := New(sess)
	m.lastMessage = "Test message"

	// Simulate clearMessageMsg
	newModel, _ := m.Update(clearMessageMsg{})
	m = newModel.(Model)

	if m.lastMessage != "" {
		t.Errorf("expected lastMessage to be cleared, got %q", m.lastMessage)
	}
}

func TestKeyPressClearsMessages(t *testing.T) {
	sess := newTestSession("line1\nline2")
	m := New(sess)
	m.lastMessage = "Some message"
	m.lastError = "Some error"

	// Any key press should clear messages
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newModel.(Model)

	if m.lastMessage != "" {
		t.Errorf("expected lastMessage to be cleared on key press, got %q", m.lastMessage)
	}
	if m.lastError != "" {
		t.Errorf("expected lastError to be cleared on key press, got %q", m.lastError)
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
	if m.inputTA == nil {
		t.Fatal("expected inputTA to be initialized")
	}
	if m.inputTA.Value() != "" {
		t.Error("input should remain empty")
	}
}

func TestMultiCharKeyHandled(t *testing.T) {
	sess := newTestSession("line1")
	m := New(sess)

	// Enter input mode
	m = sendKey(m, 'v')
	m = sendKey(m, 'c')

	// Multi-char key messages are now handled by textarea
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a', 'b'}})
	m = newModel.(Model)
	if m.inputTA == nil {
		t.Fatal("expected inputTA to be initialized")
	}
	// textarea handles multi-rune input
	if m.inputTA.Value() != "ab" {
		t.Errorf("expected 'ab', got %q", m.inputTA.Value())
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

func TestSavePreservesSourceFile(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	sess := &session.Session{
		ID:         "test-save-source",
		Content:    "line1\nline2",
		CreatedAt:  time.Date(2026, 1, 11, 12, 0, 0, 0, time.UTC),
		SourceFile: "plans/my-plan.md",
	}
	m := New(sess)

	if err := m.save(); err != nil {
		t.Fatalf("save() failed: %v", err)
	}

	sessionFile := filepath.Join(config.SessionsDir, "test-save-source.fem")
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		t.Fatalf("save() did not create file: %v", err)
	}

	content := string(data)

	// Should preserve source_file in frontmatter
	if !strings.Contains(content, "source_file: 'plans/my-plan.md'") {
		t.Errorf("saved file should contain source_file, got:\n%s", content)
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

func TestZZCentersCursor(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10")
	m := New(sess)
	m.height = 10 // visibleLines = 10 - 4 = 6
	m.cursor = 5  // cursor at line 6 (0-indexed)

	// z sets pending
	m = sendKey(m, 'z')
	if !m.zPending {
		t.Error("expected zPending after z")
	}

	// Second z centers viewport
	m = sendKey(m, 'z')
	if m.zPending {
		t.Error("zPending should be cleared after zz")
	}
	// visibleLines = 6, cursor = 5, viewportTop should be 5 - 6/2 = 2
	if m.viewportTop != 2 {
		t.Errorf("expected viewportTop at 2 after zz, got %d", m.viewportTop)
	}
}

func TestZTMovesCursorToTop(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10")
	m := New(sess)
	m.height = 10
	m.cursor = 5

	// zt moves cursor line to top of viewport
	m = sendKey(m, 'z')
	m = sendKey(m, 't')

	// viewportTop should equal cursor
	if m.viewportTop != 5 {
		t.Errorf("expected viewportTop at 5 after zt, got %d", m.viewportTop)
	}
}

func TestZBMovesCursorToBottom(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10")
	m := New(sess)
	m.height = 10 // visibleLines = 6
	m.cursor = 5

	// zb moves cursor line to bottom of viewport
	m = sendKey(m, 'z')
	m = sendKey(m, 'b')

	// visibleLines = 6, cursor = 5, viewportTop should be 5 - 6 + 1 = 0
	if m.viewportTop != 0 {
		t.Errorf("expected viewportTop at 0 after zb, got %d", m.viewportTop)
	}
}

func TestViewportResetOnCursorMove(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3\nline4\nline5")
	m := New(sess)
	m.height = 10
	m.cursor = 2
	m.viewportTop = 1 // explicitly set viewport

	// Moving cursor should reset viewportTop to auto-follow (-1)
	m = sendKey(m, 'j')

	if m.viewportTop != -1 {
		t.Errorf("expected viewportTop reset to -1 after cursor move, got %d", m.viewportTop)
	}
}

func TestZPendingClearedByOtherKey(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)

	// z sets pending
	m = sendKey(m, 'z')
	if !m.zPending {
		t.Error("expected zPending after z")
	}

	// 'x' (unrecognized for z-prefix) clears pending
	m = sendKey(m, 'x')
	if m.zPending {
		t.Error("zPending should be cleared by other key")
	}
}

func TestViewportClampsAtBoundaries(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)
	m.height = 10 // visibleLines = 6
	m.cursor = 0  // at top

	// zz at top should clamp viewportTop to 0
	m = sendKey(m, 'z')
	m = sendKey(m, 'z')

	// cursor=0, visibleLines=6, 0 - 6/2 = -3 -> clamped to 0
	if m.viewportTop != 0 {
		t.Errorf("expected viewportTop clamped to 0, got %d", m.viewportTop)
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

	// Select lines 2-4 (0-indexed cursor 1-3)
	m.cursor = 1
	m = sendKey(m, 'v')
	m = sendKey(m, 'j')
	m = sendKey(m, 'j') // now at cursor 3, selection spans 0-indexed 1-3

	// Add comment
	m = sendKey(m, 'c')
	// Type comment text
	for _, r := range "multi-line comment" {
		m = sendKey(m, r)
	}
	m = sendKeyType(m, tea.KeyEnter)

	// Should have 1 annotation spanning the range (not one per line)
	if len(m.annotations) != 1 {
		t.Errorf("expected 1 annotation for multi-line selection, got %d", len(m.annotations))
	}

	// Should span 1-indexed lines 2-4
	ann := m.annotations[0]
	if ann.StartLine != 2 {
		t.Errorf("expected StartLine 2, got %d", ann.StartLine)
	}
	if ann.EndLine != 4 {
		t.Errorf("expected EndLine 4, got %d", ann.EndLine)
	}
	if ann.Type != "comment" {
		t.Errorf("expected type 'comment', got %q", ann.Type)
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
	// Format: "> ◆" or "> ▌" (with rangeIndicator space between cursor and selection indicator)
	if !strings.Contains(view, "> ◆") && !strings.Contains(view, "> ▌") {
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
	// Note: content may have ANSI color codes, so check prefix separately
	// Format: "> ▌   1   │" (with rangeIndicator space and annotation indicator column)
	if !strings.Contains(view, "> ▌   1   │") {
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
		Content:   "line1\nline2\nline3\nline4\nline5\nline6\nline7",
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
		{StartLine: 7, EndLine: 7, Type: "change", Text: "[line 7] -> newcode"},
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
		"{++ [line 7] -> newcode ++}",
	}

	for _, exp := range expected {
		if !strings.Contains(content, exp) {
			t.Errorf("saved file should contain %q", exp)
		}
	}
}

func TestMultipleAnnotationsOnSameLine(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	sess := &session.Session{
		ID:        "test-multi-annotations",
		Content:   "line1\nline2\nline3",
		CreatedAt: time.Date(2026, 1, 11, 12, 0, 0, 0, time.UTC),
	}
	m := New(sess)

	// Add multiple annotations to the same line
	m.annotations = []fem.Annotation{
		{StartLine: 2, EndLine: 2, Type: "comment", Text: "first comment"},
		{StartLine: 2, EndLine: 2, Type: "question", Text: "why?"},
		{StartLine: 2, EndLine: 2, Type: "expand", Text: "EXPAND: more"},
	}

	if err := m.save(); err != nil {
		t.Fatalf("save() failed: %v", err)
	}

	sessionFile := filepath.Join(config.SessionsDir, "test-multi-annotations.fem")
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		t.Fatalf("save() did not create file: %v", err)
	}

	content := string(data)

	// All three annotations should be on line2
	if !strings.Contains(content, "line2 {>> first comment <<} {?? why? ??} {!! EXPAND: more !!}") {
		t.Errorf("expected all annotations on same line, got:\n%s", content)
	}
}

func TestOverlappingAnnotationsFromDifferentSelections(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3\nline4\nline5")
	m := New(sess)

	// Simulate first selection: lines 2-3, add comment
	m.cursor = 1
	m = sendKey(m, 'v')
	m = sendKey(m, 'j')
	m = sendKey(m, 'c')
	for _, r := range "first annotation" {
		m = sendKey(m, r)
	}
	m = sendKeyType(m, tea.KeyEnter)

	// Simulate second selection: lines 3-4 (overlaps on line 3), add question
	m.cursor = 2
	m = sendKey(m, 'v')
	m = sendKey(m, 'j')
	m = sendKey(m, 'q')
	for _, r := range "second annotation" {
		m = sendKey(m, r)
	}
	m = sendKeyType(m, tea.KeyEnter)

	// Should have 2 annotations: one spanning lines 2-3, one spanning lines 3-4
	if len(m.annotations) != 2 {
		t.Fatalf("expected 2 annotations (one per selection), got %d", len(m.annotations))
	}

	// First annotation: lines 2-3
	if m.annotations[0].StartLine != 2 || m.annotations[0].EndLine != 3 {
		t.Errorf("first annotation: expected lines 2-3, got %d-%d", m.annotations[0].StartLine, m.annotations[0].EndLine)
	}
	// Second annotation: lines 3-4 (overlaps with first on line 3)
	if m.annotations[1].StartLine != 3 || m.annotations[1].EndLine != 4 {
		t.Errorf("second annotation: expected lines 3-4, got %d-%d", m.annotations[1].StartLine, m.annotations[1].EndLine)
	}
}

func TestAnnotatedLinesShowIndicator(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3\nline4\nline5")
	m := New(sess)
	m.width = 80
	m.height = 20

	// Add annotations to lines 2 and 3 (1-indexed)
	m.annotations = []fem.Annotation{
		{StartLine: 2, EndLine: 2, Type: "comment", Text: "a comment"},
		{StartLine: 3, EndLine: 3, Type: "question", Text: "why?"},
		{StartLine: 3, EndLine: 3, Type: "expand", Text: "more"},
	}

	view := m.View()

	// Line 2 should show annotation indicator (1 annotation)
	if !strings.Contains(view, "2 ● │") {
		t.Errorf("line 2 should show annotation indicator ●, got:\n%s", view)
	}

	// Line 3 should show indicator (2 annotations)
	if !strings.Contains(view, "3 ● │") {
		t.Errorf("line 3 should show annotation indicator ●, got:\n%s", view)
	}

	// Lines without annotations should not show indicator
	if strings.Contains(view, "1 ● │") {
		t.Errorf("line 1 should not show annotation indicator")
	}
	if strings.Contains(view, "4 ● │") {
		t.Errorf("line 4 should not show annotation indicator")
	}
}

func TestLongLineWraps(t *testing.T) {
	longLine := strings.Repeat("x", 100)
	sess := newTestSession(longLine)
	m := New(sess)
	m.width = 40
	m.height = 20

	view := m.View()
	lines := strings.Split(view, "\n")

	// Find content lines (skip header)
	var contentLines []string
	for _, line := range lines {
		if strings.Contains(line, "│") && !strings.HasPrefix(line, "─") {
			contentLines = append(contentLines, line)
		}
	}

	// With 40 width and ~10 char prefix, content should wrap
	// Long line should produce multiple display lines
	if len(contentLines) < 2 {
		t.Errorf("expected long line to wrap into multiple lines, got %d content lines", len(contentLines))
	}

	// No line should exceed terminal width (ignoring ANSI escape codes)
	for i, line := range lines {
		visibleLen := visibleLength(line)
		if visibleLen > m.width {
			t.Errorf("line %d exceeds width %d: visible len=%d", i, m.width, visibleLen)
		}
	}
}

func TestSyntaxHighlightingEnabled(t *testing.T) {
	goCode := `package main

func main() {
	println("hello")
}`
	sess := newTestSession(goCode)
	m := NewWithFile(sess, "test.go")
	m.width = 80
	m.height = 20

	view := m.View()

	// Should contain ANSI escape codes for syntax highlighting
	if !strings.Contains(view, "\033[") {
		t.Error("expected syntax highlighting (ANSI codes) in view")
	}

	// Content should still be present
	if !strings.Contains(view, "package") {
		t.Error("expected 'package' in view")
	}
	if !strings.Contains(view, "main") {
		t.Error("expected 'main' in view")
	}
}

// --- Inline Editor Tests ---

func TestPressI_WithSelection_EntersEditorMode(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3\nline4")
	m := New(sess)

	// Select lines 1-2 (0-indexed)
	m.cursor = 1
	m = sendKey(m, 'v')
	m = sendKey(m, 'j') // now selection is 1-2

	// Press i to enter editor mode
	m = sendKey(m, 'i')

	if m.mode != modeEditor {
		t.Errorf("expected modeEditor, got %d", m.mode)
	}
	if m.editor == nil {
		t.Fatal("expected editor to be initialized")
	}
	if m.editor.start != 1 || m.editor.end != 2 {
		t.Errorf("expected editor range 1-2, got %d-%d", m.editor.start, m.editor.end)
	}

	// Textarea should contain the selected lines
	expectedContent := "line2\nline3"
	if m.editor.ta.Value() != expectedContent {
		t.Errorf("expected textarea content %q, got %q", expectedContent, m.editor.ta.Value())
	}
}

func TestPressI_WithoutSelection_DoesNothing(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)

	// No selection, press i
	m = sendKey(m, 'i')

	if m.mode != modeNormal {
		t.Errorf("expected modeNormal, got %d", m.mode)
	}
	if m.editor != nil {
		t.Error("expected editor to be nil when no selection")
	}
}

func TestEditor_Save_CreatesChangeAnnotation(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3\nline4")
	m := New(sess)

	// Select line 1 (0-indexed)
	m.cursor = 1
	m = sendKey(m, 'v')
	m = sendKey(m, 'i')

	// Simulate editing - set textarea value directly
	m.editor.ta.SetValue("modified line2")

	// Save with ctrl+s (more reliable than ctrl+enter)
	m = sendKeyType(m, tea.KeyCtrlS)

	if m.mode != modeNormal {
		t.Errorf("expected modeNormal after save, got %d", m.mode)
	}
	if m.editor != nil {
		t.Error("expected editor to be nil after save")
	}
	if len(m.annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(m.annotations))
	}

	ann := m.annotations[0]
	if ann.Type != "change" {
		t.Errorf("expected annotation type 'change', got %q", ann.Type)
	}
	if ann.StartLine != 2 { // 1-indexed
		t.Errorf("expected annotation on line 2, got %d", ann.StartLine)
	}
	// Text should contain the modified content
	if !strings.Contains(ann.Text, "modified line2") {
		t.Errorf("expected annotation text to contain 'modified line2', got %q", ann.Text)
	}
}

func TestEditor_Save_MultiLine_EncodesNewlines(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3\nline4")
	m := New(sess)

	// Select lines 1-2 (0-indexed)
	m.cursor = 1
	m = sendKey(m, 'v')
	m = sendKey(m, 'j')
	m = sendKey(m, 'i')

	// Edit to have newlines
	m.editor.ta.SetValue("new line2\nnew line3")

	// Save
	m = sendKeyType(m, tea.KeyCtrlS)

	if len(m.annotations) != 2 { // one per selected line
		t.Fatalf("expected 2 annotations (one per line), got %d", len(m.annotations))
	}

	// Check that newlines are encoded as \\n in the text
	ann := m.annotations[0]
	if !strings.Contains(ann.Text, "\\n") {
		t.Errorf("expected newlines to be encoded as \\\\n, got %q", ann.Text)
	}
}

func TestEditor_Cancel_EscTwice(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)

	// Select and enter editor
	m.cursor = 1
	m = sendKey(m, 'v')
	m = sendKey(m, 'i')

	if m.mode != modeEditor {
		t.Fatal("expected to be in editor mode")
	}

	// First Esc - should set escPending
	m = sendKeyType(m, tea.KeyEsc)
	if m.mode != modeEditor {
		t.Error("expected to still be in editor mode after first Esc")
	}
	if !m.editor.escPending {
		t.Error("expected escPending to be true after first Esc")
	}

	// Second Esc - should cancel and exit
	m = sendKeyType(m, tea.KeyEsc)
	if m.mode != modeNormal {
		t.Errorf("expected modeNormal after second Esc, got %d", m.mode)
	}
	if m.editor != nil {
		t.Error("expected editor to be nil after cancel")
	}
	if len(m.annotations) != 0 {
		t.Error("expected no annotations after cancel")
	}
}

func TestEditor_EscPending_ResetsOnTyping(t *testing.T) {
	sess := newTestSession("line1\nline2")
	m := New(sess)

	// Select and enter editor
	m.cursor = 0
	m = sendKey(m, 'v')
	m = sendKey(m, 'i')

	// First Esc
	m = sendKeyType(m, tea.KeyEsc)
	if !m.editor.escPending {
		t.Fatal("expected escPending after first Esc")
	}

	// Type something - should reset escPending
	m = sendKey(m, 'a')
	if m.editor.escPending {
		t.Error("expected escPending to be reset after typing")
	}
	if m.mode != modeEditor {
		t.Error("expected to still be in editor mode")
	}
}

func TestView_InEditorMode_ShowsEditorPanel(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)
	m.width = 80
	m.height = 30

	// Select and enter editor
	m.cursor = 1
	m = sendKey(m, 'v')
	m = sendKey(m, 'i')

	view := m.View()

	// Should show editor panel with instructions
	if !strings.Contains(view, "Edit") {
		t.Errorf("expected editor panel to show 'Edit', got:\n%s", view)
	}
}

// --- Edit Existing Annotation Tests ---

func TestPressE_WithoutSelection_OnAnnotatedLine_OpensEditor(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)

	// Add an annotation on line 2 (1-indexed)
	m.annotations = append(m.annotations, fem.Annotation{
		Type:      "comment",
		Text:      "existing comment",
		StartLine: 2,
		EndLine:   2,
	})

	// Move cursor to line 1 (0-indexed, which is line 2 in 1-indexed)
	m.cursor = 1

	// Press e without selection
	m = sendKey(m, 'e')

	if m.mode != modeEditor {
		t.Errorf("expected modeEditor, got %d", m.mode)
	}
	if m.editor == nil {
		t.Fatal("expected editor to be initialized")
	}
	if m.editor.annIndex != 0 {
		t.Errorf("expected annIndex 0, got %d", m.editor.annIndex)
	}
	// Editor should contain the annotation text (decoded)
	if m.editor.ta.Value() != "existing comment" {
		t.Errorf("expected textarea 'existing comment', got %q", m.editor.ta.Value())
	}
}

func TestPressE_WithoutSelection_NoAnnotation_ShowsError(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)

	// No annotations, cursor on line 1
	m.cursor = 1

	// Press e without selection
	m = sendKey(m, 'e')

	if m.mode != modeNormal {
		t.Errorf("expected to stay in modeNormal, got %d", m.mode)
	}
	if m.lastError != "No annotation on this line" {
		t.Errorf("expected error message, got %q", m.lastError)
	}
}

func TestPressE_WithSelection_CreatesExpandAnnotation(t *testing.T) {
	sess := newTestSession("line1\nline2")
	m := New(sess)

	// Select a line first
	m = sendKey(m, 'v')

	// Press e - should create expand annotation (existing behavior)
	m = sendKey(m, 'e')

	if m.mode != modeInput {
		t.Errorf("expected modeInput for expand annotation, got %d", m.mode)
	}
	if m.inputType != "expand" {
		t.Errorf("expected inputType 'expand', got %q", m.inputType)
	}
}

func TestEditAnnotation_Save_UpdatesExistingAnnotation(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)

	// Add an annotation
	m.annotations = append(m.annotations, fem.Annotation{
		Type:      "comment",
		Text:      "original",
		StartLine: 2,
		EndLine:   2,
	})

	// Move to annotated line and press e
	m.cursor = 1
	m = sendKey(m, 'e')

	// Modify the text
	m.editor.ta.SetValue("updated comment")

	// Save
	m = sendKeyType(m, tea.KeyCtrlS)

	if m.mode != modeNormal {
		t.Errorf("expected modeNormal after save, got %d", m.mode)
	}
	if len(m.annotations) != 1 {
		t.Fatalf("expected 1 annotation (updated), got %d", len(m.annotations))
	}
	if m.annotations[0].Text != "updated comment" {
		t.Errorf("expected annotation text 'updated comment', got %q", m.annotations[0].Text)
	}
}

func TestEditAnnotation_Cancel_KeepsOriginal(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)

	// Add an annotation
	m.annotations = append(m.annotations, fem.Annotation{
		Type:      "comment",
		Text:      "original",
		StartLine: 2,
		EndLine:   2,
	})

	// Move to annotated line and press e
	m.cursor = 1
	m = sendKey(m, 'e')

	// Modify the text
	m.editor.ta.SetValue("modified")

	// Cancel with Esc Esc
	m = sendKeyType(m, tea.KeyEsc)
	m = sendKeyType(m, tea.KeyEsc)

	if m.mode != modeNormal {
		t.Errorf("expected modeNormal after cancel, got %d", m.mode)
	}
	if m.annotations[0].Text != "original" {
		t.Errorf("expected annotation to remain 'original', got %q", m.annotations[0].Text)
	}
}

func TestPressE_MultipleAnnotations_OpensPicker(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)

	// Add two annotations on the same line
	m.annotations = append(m.annotations, fem.Annotation{
		Type:      "comment",
		Text:      "first comment",
		StartLine: 2,
		EndLine:   2,
	})
	m.annotations = append(m.annotations, fem.Annotation{
		Type:      "question",
		Text:      "a question",
		StartLine: 2,
		EndLine:   2,
	})

	// Move cursor to line 1 (0-indexed)
	m.cursor = 1

	// Press e without selection
	m = sendKey(m, 'e')

	if m.mode != modePalette {
		t.Errorf("expected modePalette for annotation picker, got %d", m.mode)
	}
	if m.paletteKind != "annPick" {
		t.Errorf("expected paletteKind 'annPick', got %q", m.paletteKind)
	}
	if len(m.paletteItems) != 2 {
		t.Errorf("expected 2 palette items, got %d", len(m.paletteItems))
	}
}

func TestAnnotationPicker_SelectAndEdit(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)

	// Add two annotations
	m.annotations = append(m.annotations, fem.Annotation{
		Type:      "comment",
		Text:      "first",
		StartLine: 2,
		EndLine:   2,
	})
	m.annotations = append(m.annotations, fem.Annotation{
		Type:      "question",
		Text:      "second",
		StartLine: 2,
		EndLine:   2,
	})

	m.cursor = 1
	m = sendKey(m, 'e') // opens picker

	// Move down to second item
	m = sendKey(m, 'j')
	if m.paletteCursor != 1 {
		t.Errorf("expected paletteCursor 1, got %d", m.paletteCursor)
	}

	// Select with enter
	m = sendKeyType(m, tea.KeyEnter)

	if m.mode != modeEditor {
		t.Errorf("expected modeEditor after selection, got %d", m.mode)
	}
	if m.editor.annIndex != 1 {
		t.Errorf("expected annIndex 1, got %d", m.editor.annIndex)
	}
	if m.editor.ta.Value() != "second" {
		t.Errorf("expected 'second' in editor, got %q", m.editor.ta.Value())
	}
}

func TestEnsureCursorVisible_ScrollsViewportWhenCursorMovesDown(t *testing.T) {
	lines := make([]string, 50)
	for i := range lines {
		lines[i] = fmt.Sprintf("line %d", i+1)
	}
	sess := newTestSession(strings.Join(lines, "\n"))
	m := New(sess)
	m.width = 80
	m.height = 20

	for i := 0; i < 30; i++ {
		m = sendKey(m, 'j')
	}

	if m.cursor != 30 {
		t.Errorf("expected cursor at 30, got %d", m.cursor)
	}
	if m.autoViewportTop == 0 {
		t.Error("expected autoViewportTop to have scrolled from 0")
	}
	if m.autoViewportTop > m.cursor {
		t.Errorf("autoViewportTop (%d) should not exceed cursor (%d)", m.autoViewportTop, m.cursor)
	}
}

func TestEnsureCursorVisible_ScrollsViewportWhenCursorMovesUp(t *testing.T) {
	lines := make([]string, 50)
	for i := range lines {
		lines[i] = fmt.Sprintf("line %d", i+1)
	}
	sess := newTestSession(strings.Join(lines, "\n"))
	m := New(sess)
	m.width = 80
	m.height = 20

	m.cursor = 40
	m.autoViewportTop = 35
	m.viewportTop = -1
	m.ensureCursorVisible()

	for i := 0; i < 30; i++ {
		m = sendKey(m, 'k')
	}

	if m.cursor != 10 {
		t.Errorf("expected cursor at 10, got %d", m.cursor)
	}
	if m.autoViewportTop > m.cursor {
		t.Errorf("autoViewportTop (%d) should be <= cursor (%d)", m.autoViewportTop, m.cursor)
	}
}

func TestGG_ResetsAutoViewportTop(t *testing.T) {
	lines := make([]string, 50)
	for i := range lines {
		lines[i] = fmt.Sprintf("line %d", i+1)
	}
	sess := newTestSession(strings.Join(lines, "\n"))
	m := New(sess)
	m.width = 80
	m.height = 20

	for i := 0; i < 30; i++ {
		m = sendKey(m, 'j')
	}

	m = sendKey(m, 'g')
	m = sendKey(m, 'g')

	if m.cursor != 0 {
		t.Errorf("expected cursor at 0 after gg, got %d", m.cursor)
	}
	if m.autoViewportTop != 0 {
		t.Errorf("expected autoViewportTop 0 after gg, got %d", m.autoViewportTop)
	}
}

func visibleLength(s string) int {
	inEscape := false
	count := 0
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		count++
	}
	return count
}

// --- Inline Editor Box Alignment Tests ---

func TestInlineEditorBox_AllLinesHaveSameWidth(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)
	m.width = 100
	m.height = 30

	// Select a line and enter editor mode
	m.cursor = 1
	m = sendKey(m, 'v')
	m = sendKey(m, 'i')

	if m.mode != modeEditor {
		t.Fatalf("expected modeEditor, got %d", m.mode)
	}

	view := m.View()
	lines := strings.Split(view, "\n")

	// Find the editor box lines (between ┌ and └)
	var boxLines []string
	inBox := false
	for _, line := range lines {
		if strings.HasPrefix(line, "┌") {
			inBox = true
		}
		if inBox {
			boxLines = append(boxLines, line)
		}
		if strings.HasPrefix(line, "└") {
			break
		}
	}

	if len(boxLines) < 3 {
		t.Fatalf("expected at least 3 box lines (header, content, footer), got %d", len(boxLines))
	}

	// All box lines should have the same visible width
	expectedWidth := lipgloss.Width(boxLines[0])
	for i, line := range boxLines {
		actualWidth := lipgloss.Width(line)
		if actualWidth != expectedWidth {
			t.Errorf("box line %d has width %d, expected %d\nline: %q", i, actualWidth, expectedWidth, line)
		}
	}
}

func TestInlineEditorBox_ContentLinesHaveRightBorder(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)
	m.width = 100
	m.height = 30

	// Select and enter editor mode
	m.cursor = 1
	m = sendKey(m, 'v')
	m = sendKey(m, 'i')

	view := m.View()
	lines := strings.Split(view, "\n")

	// Find content lines (between header and footer, starting with │)
	for _, line := range lines {
		if strings.HasPrefix(line, "│") && !strings.HasPrefix(line, "│ ...") {
			// Content line should end with " │"
			stripped := strings.TrimRight(line, " \t")
			if !strings.HasSuffix(stripped, "│") {
				t.Errorf("content line missing right border: %q", line)
			}
		}
	}
}

func TestInlineInputBox_AllLinesHaveSameWidth(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)
	m.width = 100
	m.height = 30

	// Select a line and enter comment mode
	m.cursor = 1
	m = sendKey(m, 'v')
	m = sendKey(m, 'c')

	if m.mode != modeInput {
		t.Fatalf("expected modeInput, got %d", m.mode)
	}

	view := m.View()
	lines := strings.Split(view, "\n")

	// Find the input box lines
	var boxLines []string
	inBox := false
	for _, line := range lines {
		if strings.HasPrefix(line, "┌") {
			inBox = true
		}
		if inBox {
			boxLines = append(boxLines, line)
		}
		if strings.HasPrefix(line, "└") {
			break
		}
	}

	if len(boxLines) < 3 {
		t.Fatalf("expected at least 3 box lines, got %d", len(boxLines))
	}

	// All box lines should have the same visible width
	expectedWidth := lipgloss.Width(boxLines[0])
	for i, line := range boxLines {
		actualWidth := lipgloss.Width(line)
		if actualWidth != expectedWidth {
			t.Errorf("box line %d has width %d, expected %d\nline: %q", i, actualWidth, expectedWidth, line)
		}
	}
}

func TestInlineEditorBox_WidthMatchesTerminalWidth(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)
	m.width = 120
	m.height = 30

	// Select and enter editor mode
	m.cursor = 1
	m = sendKey(m, 'v')
	m = sendKey(m, 'i')

	view := m.View()
	lines := strings.Split(view, "\n")

	// Find the header line
	for _, line := range lines {
		if strings.HasPrefix(line, "┌") {
			boxWidth := lipgloss.Width(line)
			// Box width should be close to terminal width (within margin)
			// We use width - 2 as the expected box width based on the implementation
			expectedBoxWidth := m.width - 2
			if boxWidth != expectedBoxWidth {
				t.Errorf("box width %d doesn't match expected %d (terminal width %d)", boxWidth, expectedBoxWidth, m.width)
			}
			break
		}
	}
}

// --- Search Tests ---

func TestSearchMode_SlashEntersSearchMode(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)
	m.width = 80
	m.height = 20

	m = sendKey(m, '/')

	if m.mode != modeSearch {
		t.Errorf("expected modeSearch, got %d", m.mode)
	}
}

func TestSearchMode_EscCancelsSearch(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)
	m.width = 80
	m.height = 20

	m = sendKey(m, '/')
	m = sendKeyEsc(m)

	if m.mode != modeNormal {
		t.Errorf("expected modeNormal after Esc, got %d", m.mode)
	}
}

func TestSearchMode_TypingBuildsQuery(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)
	m.width = 80
	m.height = 20

	m = sendKey(m, '/')
	m = sendKey(m, 'f')
	m = sendKey(m, 'o')
	m = sendKey(m, 'o')

	if m.search.query != "foo" {
		t.Errorf("expected query 'foo', got %q", m.search.query)
	}
}

func TestSearchMode_BackspaceRemovesCharacter(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)
	m.width = 80
	m.height = 20

	m = sendKey(m, '/')
	m = sendKey(m, 'f')
	m = sendKey(m, 'o')
	m = sendKeyBackspace(m)

	if m.search.query != "f" {
		t.Errorf("expected query 'f', got %q", m.search.query)
	}
}

func TestSearchMode_EnterPerformsSearch(t *testing.T) {
	sess := newTestSession("apple\nbanana\napricot")
	m := New(sess)
	m.width = 80
	m.height = 20

	m = sendKey(m, '/')
	m = sendKey(m, 'a')
	m = sendKey(m, 'p')
	m = sendKeyEnter(m)

	if m.mode != modeNormal {
		t.Errorf("expected modeNormal after Enter, got %d", m.mode)
	}
	if len(m.search.matches) != 2 {
		t.Errorf("expected 2 matches (apple, apricot), got %d", len(m.search.matches))
	}
	if m.cursor != 0 {
		t.Errorf("expected cursor at 0 (first match), got %d", m.cursor)
	}
}

func TestSearchMode_FuzzyMatching(t *testing.T) {
	sess := newTestSession("function foo() {}\nconst bar = 1\nfunc baz() {}")
	m := New(sess)
	m.width = 80
	m.height = 20

	m = sendKey(m, '/')
	m = sendKey(m, 'f')
	m = sendKey(m, 'o')
	m = sendKeyEnter(m)

	// "fo" should match "function foo() {}" (f...o) and "func baz() {}" would not match
	// Actually "function foo() {}" has f,o in sequence, "func baz() {}" has f but no o after
	// Let me reconsider: fuzzy match "fo" in "function foo() {}" - f at 0, o at 5 (functi'o'n)
	// In "func baz() {}" - f at 0, no o
	if len(m.search.matches) < 1 {
		t.Errorf("expected at least 1 fuzzy match for 'fo', got %d", len(m.search.matches))
	}
}

func TestSearchMode_NJumpsToNextMatch(t *testing.T) {
	sess := newTestSession("apple\nbanana\napricot\norange")
	m := New(sess)
	m.width = 80
	m.height = 20

	m = sendKey(m, '/')
	m = sendKey(m, 'a')
	m = sendKeyEnter(m)

	// All lines have 'a', so 4 matches
	if len(m.search.matches) != 4 {
		t.Errorf("expected 4 matches, got %d", len(m.search.matches))
	}

	initialCursor := m.cursor
	m = sendKey(m, 'n')

	if m.cursor == initialCursor {
		t.Error("expected cursor to move to next match")
	}
}

func TestSearchMode_NWrapsAround(t *testing.T) {
	sess := newTestSession("apple\nbanana\napricot")
	m := New(sess)
	m.width = 80
	m.height = 20

	m = sendKey(m, '/')
	m = sendKey(m, 'a')
	m = sendKeyEnter(m)

	// Jump through all matches
	for i := 0; i < len(m.search.matches); i++ {
		m = sendKey(m, 'n')
	}

	// Should wrap back to first match
	if m.cursor != m.search.matches[0] {
		t.Errorf("expected cursor to wrap to first match at %d, got %d", m.search.matches[0], m.cursor)
	}
}

func TestSearchMode_ShiftNJumpsToPreviousMatch(t *testing.T) {
	sess := newTestSession("apple\nbanana\napricot")
	m := New(sess)
	m.width = 80
	m.height = 20

	m = sendKey(m, '/')
	m = sendKey(m, 'a')
	m = sendKeyEnter(m)

	// Go to second match
	m = sendKey(m, 'n')
	secondMatchCursor := m.cursor

	// Go back with N
	m = sendKeyRune(m, 'N')

	if m.cursor >= secondMatchCursor {
		t.Errorf("expected cursor to go back from %d, got %d", secondMatchCursor, m.cursor)
	}
}

func TestSearchMode_NoMatchShowsError(t *testing.T) {
	sess := newTestSession("apple\nbanana\napricot")
	m := New(sess)
	m.width = 80
	m.height = 20

	m = sendKey(m, '/')
	m = sendKey(m, 'z')
	m = sendKey(m, 'z')
	m = sendKey(m, 'z')
	m = sendKeyEnter(m)

	if m.lastError != "No matches found" {
		t.Errorf("expected 'No matches found' error, got %q", m.lastError)
	}
}

func TestSearchMode_ViewShowsSearchPrompt(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)
	m.width = 80
	m.height = 20

	m = sendKey(m, '/')
	m = sendKey(m, 't')
	m = sendKey(m, 'e')
	m = sendKey(m, 's')
	m = sendKey(m, 't')

	view := m.View()
	if !strings.Contains(view, "/test") {
		t.Errorf("expected view to contain '/test', got:\n%s", view)
	}
}

func TestSearchMode_MatchIndicatorShown(t *testing.T) {
	sess := newTestSession("apple\nbanana\napricot")
	m := New(sess)
	m.width = 80
	m.height = 20

	m = sendKey(m, '/')
	m = sendKey(m, 'a')
	m = sendKey(m, 'p')
	m = sendKeyEnter(m)

	view := m.View()
	// Search match indicator is ◎
	if !strings.Contains(view, "◎") {
		t.Errorf("expected search match indicator '◎' in view, got:\n%s", view)
	}
}

func TestSearchMode_CaseInsensitive(t *testing.T) {
	sess := newTestSession("Apple\nBANANA\napricot")
	m := New(sess)
	m.width = 80
	m.height = 20

	m = sendKey(m, '/')
	m = sendKey(m, 'a')
	m = sendKeyEnter(m)

	// Should match all 3 lines (case insensitive)
	if len(m.search.matches) != 3 {
		t.Errorf("expected 3 case-insensitive matches, got %d", len(m.search.matches))
	}
}

func TestSearchMode_PJumpsToPreviousMatch(t *testing.T) {
	sess := newTestSession("apple\nbanana\napricot")
	m := New(sess)
	m.width = 80
	m.height = 20

	m = sendKey(m, '/')
	m = sendKey(m, 'a')
	m = sendKeyEnter(m)

	// Go to second match
	m = sendKey(m, 'n')
	secondMatchCursor := m.cursor

	// Go back with p (alternative to N)
	m = sendKey(m, 'p')

	if m.cursor >= secondMatchCursor {
		t.Errorf("expected cursor to go back from %d with 'p', got %d", secondMatchCursor, m.cursor)
	}
}

func TestSearchMode_EscClearsSearchInNormalMode(t *testing.T) {
	sess := newTestSession("apple\nbanana\napricot")
	m := New(sess)
	m.width = 80
	m.height = 20

	m = sendKey(m, '/')
	m = sendKey(m, 'a')
	m = sendKeyEnter(m)

	// Should have matches
	if len(m.search.matches) == 0 {
		t.Fatal("expected matches after search")
	}

	// Press Esc in normal mode to clear search
	m = sendKeyEsc(m)

	if len(m.search.matches) != 0 {
		t.Errorf("expected search to be cleared after Esc, got %d matches", len(m.search.matches))
	}
	if m.search.query != "" {
		t.Errorf("expected query to be cleared after Esc, got %q", m.search.query)
	}
}

func TestSearchMode_MatchCounterInView(t *testing.T) {
	sess := newTestSession("apple\nbanana\napricot\norange")
	m := New(sess)
	m.width = 120
	m.height = 20

	m = sendKey(m, '/')
	m = sendKey(m, 'a')
	m = sendKeyEnter(m)

	// All 4 lines have 'a'
	view := m.View()
	if !strings.Contains(view, "1/4") {
		t.Errorf("expected match counter '1/4' in view, got:\n%s", view)
	}

	// Jump to next and check counter updates
	m = sendKey(m, 'n')
	view = m.View()
	if !strings.Contains(view, "2/4") {
		t.Errorf("expected match counter '2/4' after n, got:\n%s", view)
	}
}

func TestSearchMode_HelpBarShowsSearchNav(t *testing.T) {
	sess := newTestSession("apple\nbanana\napricot")
	m := New(sess)
	m.width = 120
	m.height = 20

	m = sendKey(m, '/')
	m = sendKey(m, 'a')
	m = sendKeyEnter(m)

	view := m.View()
	if !strings.Contains(view, "[n]ext") {
		t.Errorf("expected '[n]ext' in help bar, got:\n%s", view)
	}
	if !strings.Contains(view, "[p]rev") {
		t.Errorf("expected '[p]rev' in help bar, got:\n%s", view)
	}
}

func sendKeyEsc(m Model) Model {
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	return newModel.(Model)
}

func sendKeyEnter(m Model) Model {
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	return newModel.(Model)
}

func sendKeyBackspace(m Model) Model {
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	return newModel.(Model)
}

func sendKeyRune(m Model, r rune) Model {
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	return newModel.(Model)
}

// --- Annotation Preview Tests ---

func TestAnnotationPreview_ShowsWhenCursorOnAnnotatedLine(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)
	m.width = 80
	m.height = 20

	m.annotations = []fem.Annotation{
		{Type: "comment", StartLine: 2, EndLine: 2, Text: "This needs review"},
	}

	// Move cursor to line 2 (index 1)
	m.cursor = 1

	view := m.View()

	if !strings.Contains(view, "comment [2-2]") {
		t.Errorf("expected preview to show 'comment [2-2]', got:\n%s", view)
	}
	if !strings.Contains(view, "This needs review") {
		t.Errorf("expected preview to show annotation text, got:\n%s", view)
	}
}

func TestAnnotationPreview_ShowsCountWhenMultipleAnnotations(t *testing.T) {
	sess := newTestSession("line1\nline2")
	m := New(sess)
	m.width = 80
	m.height = 20

	m.annotations = []fem.Annotation{
		{Type: "comment", StartLine: 2, EndLine: 2, Text: "First note"},
		{Type: "question", StartLine: 2, EndLine: 2, Text: "Is this correct?"},
	}

	// Move cursor to line 2 (index 1)
	m.cursor = 1

	view := m.View()

	if !strings.Contains(view, "1 of 2") {
		t.Errorf("expected preview to show '1 of 2', got:\n%s", view)
	}
}

func TestAnnotationPreview_DisappearsWhenCursorLeavesAnnotatedLine(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)
	m.width = 80
	m.height = 20

	m.annotations = []fem.Annotation{
		{Type: "comment", StartLine: 2, EndLine: 2, Text: "Some note"},
	}

	// Start on annotated line
	m.cursor = 1
	view := m.View()
	if !strings.Contains(view, "comment [2-2]") {
		t.Fatalf("expected preview on annotated line, got:\n%s", view)
	}

	// Move to non-annotated line
	m.cursor = 0
	view = m.View()

	if strings.Contains(view, "comment [2-2]") {
		t.Errorf("expected preview to disappear, but it's still there:\n%s", view)
	}
	// Should show normal help text instead
	if !strings.Contains(view, "[v]sel") {
		t.Errorf("expected normal help text when not on annotated line, got:\n%s", view)
	}
}

func TestAnnotationPreview_ShowsFirstAnnotationContent(t *testing.T) {
	sess := newTestSession("line1\nline2")
	m := New(sess)
	m.width = 80
	m.height = 20

	m.annotations = []fem.Annotation{
		{Type: "comment", StartLine: 2, EndLine: 2, Text: "First note"},
		{Type: "question", StartLine: 2, EndLine: 2, Text: "Second note"},
	}

	m.cursor = 1
	view := m.View()

	// Should show first annotation's text
	if !strings.Contains(view, "First note") {
		t.Errorf("expected first annotation text, got:\n%s", view)
	}
}

func TestAnnotationPreview_WrapsLongText(t *testing.T) {
	sess := newTestSession("line1\nline2")
	m := New(sess)
	m.width = 50
	m.height = 20

	longText := "This is a very long annotation that should wrap within the preview panel width properly"
	m.annotations = []fem.Annotation{
		{Type: "comment", StartLine: 2, EndLine: 2, Text: longText},
	}

	m.cursor = 1
	view := m.View()

	// Should contain the preview box structure
	if !strings.Contains(view, "┌─ comment") {
		t.Errorf("expected preview box header, got:\n%s", view)
	}
	if !strings.Contains(view, "└") {
		t.Errorf("expected preview box footer, got:\n%s", view)
	}
}

// --- Annotation Preview Tab Cycling Tests ---

func TestAnnotationPreview_TabCyclesAnnotations(t *testing.T) {
	sess := newTestSession("line1\nline2")
	m := New(sess)
	m.width = 80
	m.height = 20

	m.annotations = []fem.Annotation{
		{Type: "comment", StartLine: 2, EndLine: 2, Text: "First"},
		{Type: "question", StartLine: 2, EndLine: 2, Text: "Second"},
		{Type: "delete", StartLine: 2, EndLine: 2, Text: "Third"},
	}

	m.cursor = 1
	view := m.View()
	if !strings.Contains(view, "First") {
		t.Fatalf("expected first annotation, got:\n%s", view)
	}

	// Tab to second annotation
	m = sendKeyTab(m)
	view = m.View()
	if !strings.Contains(view, "Second") {
		t.Errorf("expected second annotation after Tab, got:\n%s", view)
	}
	if !strings.Contains(view, "2 of 3") {
		t.Errorf("expected '2 of 3' counter, got:\n%s", view)
	}

	// Tab to third annotation
	m = sendKeyTab(m)
	view = m.View()
	if !strings.Contains(view, "Third") {
		t.Errorf("expected third annotation after Tab, got:\n%s", view)
	}
	if !strings.Contains(view, "3 of 3") {
		t.Errorf("expected '3 of 3' counter, got:\n%s", view)
	}
}

func TestAnnotationPreview_TabWrapsAround(t *testing.T) {
	sess := newTestSession("line1\nline2")
	m := New(sess)
	m.width = 80
	m.height = 20

	m.annotations = []fem.Annotation{
		{Type: "comment", StartLine: 2, EndLine: 2, Text: "First"},
		{Type: "question", StartLine: 2, EndLine: 2, Text: "Second"},
	}

	m.cursor = 1

	// Tab twice to get to second, then tab again to wrap
	m = sendKeyTab(m)
	m = sendKeyTab(m)
	view := m.View()

	if !strings.Contains(view, "First") {
		t.Errorf("expected wrap to first annotation, got:\n%s", view)
	}
	if !strings.Contains(view, "1 of 2") {
		t.Errorf("expected '1 of 2' counter after wrap, got:\n%s", view)
	}
}

func TestAnnotationPreview_TabNoOpOnSingleAnnotation(t *testing.T) {
	sess := newTestSession("line1\nline2")
	m := New(sess)
	m.width = 80
	m.height = 20

	m.annotations = []fem.Annotation{
		{Type: "comment", StartLine: 2, EndLine: 2, Text: "Only one"},
	}

	m.cursor = 1

	m = sendKeyTab(m)
	view := m.View()

	// Should still show the same annotation, no count (only 1)
	if !strings.Contains(view, "Only one") {
		t.Errorf("expected annotation to remain, got:\n%s", view)
	}
}

func TestAnnotationPreview_MovingCursorResetsIndex(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)
	m.width = 80
	m.height = 20

	m.annotations = []fem.Annotation{
		{Type: "comment", StartLine: 2, EndLine: 2, Text: "First"},
		{Type: "question", StartLine: 2, EndLine: 2, Text: "Second"},
	}

	m.cursor = 1

	// Tab to second annotation
	m = sendKeyTab(m)
	view := m.View()
	if !strings.Contains(view, "Second") {
		t.Fatalf("expected second annotation, got:\n%s", view)
	}

	// Move cursor away and back
	m = sendKey(m, 'j') // move to line 3
	m = sendKey(m, 'k') // move back to line 2

	view = m.View()
	if !strings.Contains(view, "First") {
		t.Errorf("expected reset to first annotation after cursor move, got:\n%s", view)
	}
}

func TestAnnotationPreview_ShiftTabCyclesBackward(t *testing.T) {
	sess := newTestSession("line1\nline2")
	m := New(sess)
	m.width = 80
	m.height = 20

	m.annotations = []fem.Annotation{
		{Type: "comment", StartLine: 2, EndLine: 2, Text: "First"},
		{Type: "question", StartLine: 2, EndLine: 2, Text: "Second"},
		{Type: "delete", StartLine: 2, EndLine: 2, Text: "Third"},
	}

	m.cursor = 1

	// Shift+Tab should wrap to last annotation
	m = sendKeyShiftTab(m)
	view := m.View()
	if !strings.Contains(view, "Third") {
		t.Errorf("expected wrap to last annotation with Shift+Tab, got:\n%s", view)
	}
	if !strings.Contains(view, "3 of 3") {
		t.Errorf("expected '3 of 3' counter, got:\n%s", view)
	}

	// Shift+Tab again to go to second
	m = sendKeyShiftTab(m)
	view = m.View()
	if !strings.Contains(view, "Second") {
		t.Errorf("expected second annotation, got:\n%s", view)
	}
}

func sendKeyTab(m Model) Model {
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	return newModel.(Model)
}

func sendKeyShiftTab(m Model) Model {
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	return newModel.(Model)
}

// --- Help Menu Tests ---

func TestHelpMenu_QuestionMarkOpensHelp(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)
	m.width = 80
	m.height = 20

	m = sendKey(m, '?')

	if m.mode != modeHelp {
		t.Errorf("expected modeHelp, got %d", m.mode)
	}
}

func TestHelpMenu_AnyKeyCloses(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)
	m.width = 80
	m.height = 20

	m = sendKey(m, '?')
	if m.mode != modeHelp {
		t.Fatalf("expected modeHelp, got %d", m.mode)
	}

	// Any key should close it
	m = sendKey(m, 'j')
	if m.mode != modeNormal {
		t.Errorf("expected modeNormal after pressing any key, got %d", m.mode)
	}
}

func TestHelpMenu_EscCloses(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)
	m.width = 80
	m.height = 20

	m = sendKey(m, '?')
	m = sendKeyEsc(m)

	if m.mode != modeNormal {
		t.Errorf("expected modeNormal after Esc, got %d", m.mode)
	}
}

func TestHelpMenu_DisplaysKeybindings(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)
	m.width = 80
	m.height = 30

	m = sendKey(m, '?')
	view := m.View()

	// Should display key navigation bindings
	if !strings.Contains(view, "j/k") {
		t.Errorf("expected j/k navigation in help, got:\n%s", view)
	}
	// Should display selection binding
	if !strings.Contains(view, "v") && !strings.Contains(view, "selection") {
		t.Errorf("expected selection keybinding in help, got:\n%s", view)
	}
	// Should display save binding
	if !strings.Contains(view, "w") {
		t.Errorf("expected w (write/save) in help, got:\n%s", view)
	}
}

func TestHelpMenu_DisplaysVersion(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := NewWithVersion(sess, "", "1.2.3")
	m.width = 80
	m.height = 30

	m = sendKey(m, '?')
	view := m.View()

	if !strings.Contains(view, "1.2.3") {
		t.Errorf("expected version in help, got:\n%s", view)
	}
}

// --- Annotation Range Highlighting Tests ---

func TestAnnotationPreview_HighlightsAnnotationRange(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3\nline4\nline5")
	m := New(sess)
	m.width = 80
	m.height = 20

	// Annotation spans lines 2-4
	m.annotations = []fem.Annotation{
		{Type: "comment", StartLine: 2, EndLine: 4, Text: "Multi-line annotation"},
	}

	// Move cursor to line 2 (which has the annotation)
	m.cursor = 1 // 0-indexed, so line 2
	view := m.View()

	// Lines 2, 3, 4 should have highlighting indicator
	// Look for the annotation range highlight marker "▐"
	lines := strings.Split(view, "\n")
	var highlightedLines []int
	for i, line := range lines {
		if strings.Contains(line, "▐") {
			highlightedLines = append(highlightedLines, i)
		}
	}

	if len(highlightedLines) != 3 {
		t.Errorf("expected 3 highlighted lines for annotation range 2-4, got %d: %v\nview:\n%s",
			len(highlightedLines), highlightedLines, view)
	}
}

func TestAnnotationPreview_HighlightDisappearsWhenCursorLeavesAnnotatedLine(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3\nline4\nline5")
	m := New(sess)
	m.width = 80
	m.height = 20

	m.annotations = []fem.Annotation{
		{Type: "comment", StartLine: 2, EndLine: 3, Text: "Some note"},
	}

	// Cursor on annotated line - should show highlight
	m.cursor = 1
	view := m.View()
	if !strings.Contains(view, "▐") {
		t.Errorf("expected annotation range highlight when cursor on annotated line, got:\n%s", view)
	}

	// Move cursor away from annotated line
	m.cursor = 4 // line 5, no annotation
	m.resetPreviewIndex()
	view = m.View()
	if strings.Contains(view, "▐") {
		t.Errorf("expected no annotation range highlight when cursor leaves annotated line, got:\n%s", view)
	}
}

func TestAnnotationPreview_TabCyclingUpdatesHighlight(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3\nline4\nline5\nline6")
	m := New(sess)
	m.width = 80
	m.height = 20

	// Two annotations with different ranges on the same starting line
	m.annotations = []fem.Annotation{
		{Type: "comment", StartLine: 2, EndLine: 2, Text: "Single line"},
		{Type: "question", StartLine: 2, EndLine: 4, Text: "Multi-line"},
	}

	m.cursor = 1 // line 2
	view := m.View()

	// First annotation: only line 2 highlighted
	highlightCount := strings.Count(view, "▐")
	if highlightCount != 1 {
		t.Errorf("expected 1 highlighted line for first annotation, got %d\nview:\n%s", highlightCount, view)
	}

	// Tab to second annotation: lines 2-4 highlighted
	m = sendKeyTab(m)
	view = m.View()
	highlightCount = strings.Count(view, "▐")
	if highlightCount != 3 {
		t.Errorf("expected 3 highlighted lines for second annotation (lines 2-4), got %d\nview:\n%s", highlightCount, view)
	}
}

func TestAnnotationPreview_HighlightCoexistsWithSelection(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3\nline4\nline5")
	m := New(sess)
	m.width = 80
	m.height = 20

	m.annotations = []fem.Annotation{
		{Type: "comment", StartLine: 2, EndLine: 3, Text: "Note"},
	}

	// Start selection on line 2
	m.cursor = 1
	m = sendKey(m, 'v') // start selection
	m.cursor = 2        // extend to line 3

	view := m.View()

	// Should have both selection indicators and annotation highlight
	if !strings.Contains(view, "◆") && !strings.Contains(view, "▌") {
		t.Errorf("expected selection indicator, got:\n%s", view)
	}
	if !strings.Contains(view, "▐") {
		t.Errorf("expected annotation range highlight, got:\n%s", view)
	}
}

// --- Text Object Expansion Tests (ap, ab, as) ---

func TestTextObjectExpandToParagraph(t *testing.T) {
	content := "intro\n\npara line1\npara line2\npara line3\n\noutro"
	sess := newTestSession(content)
	m := New(sess)
	m.width = 80
	m.height = 24

	// Move to middle of paragraph (line index 3 = "para line2")
	m = sendKey(m, 'j') // line 1
	m = sendKey(m, 'j') // line 2
	m = sendKey(m, 'j') // line 3

	// Start selection, then expand with 'ap'
	m = sendKey(m, 'v')
	m = sendKey(m, 'a')
	m = sendKey(m, 'p')

	// Selection should span lines 2-4 (the paragraph)
	start, end := m.selection.lines()
	if start != 2 || end != 4 {
		t.Errorf("expected selection (2, 4), got (%d, %d)", start, end)
	}
}

func TestTextObjectExpandToCodeBlock(t *testing.T) {
	content := "before\n```go\nfunc main() {\n}\n```\nafter"
	sess := newTestSession(content)
	m := New(sess)
	m.width = 80
	m.height = 24

	// Move into code block (line index 2 = "func main() {")
	m = sendKey(m, 'j') // line 1 (```)
	m = sendKey(m, 'j') // line 2 (func)

	// Start selection, then expand with 'ab'
	m = sendKey(m, 'v')
	m = sendKey(m, 'a')
	m = sendKey(m, 'b')

	// Selection should span lines 1-4 (the code block with fences)
	start, end := m.selection.lines()
	if start != 1 || end != 4 {
		t.Errorf("expected selection (1, 4), got (%d, %d)", start, end)
	}
}

func TestTextObjectExpandToCodeBlockNotInBlock(t *testing.T) {
	content := "before\n```go\ncode\n```\nafter"
	sess := newTestSession(content)
	m := New(sess)
	m.width = 80
	m.height = 24

	// Stay on line 0 (not in code block)
	m = sendKey(m, 'v')
	m = sendKey(m, 'a')
	m = sendKey(m, 'b')

	// Selection should not change (still just line 0)
	start, end := m.selection.lines()
	if start != 0 || end != 0 {
		t.Errorf("expected selection (0, 0) when not in block, got (%d, %d)", start, end)
	}
}

func TestTextObjectExpandToSection(t *testing.T) {
	content := "# Heading One\ncontent1\ncontent2\n# Heading Two\nother"
	sess := newTestSession(content)
	m := New(sess)
	m.width = 80
	m.height = 24

	// Move to content1 (line index 1)
	m = sendKey(m, 'j')

	// Start selection, then expand with 'as'
	m = sendKey(m, 'v')
	m = sendKey(m, 'a')
	m = sendKey(m, 's')

	// Selection should span lines 0-2 (heading + content, before next heading)
	start, end := m.selection.lines()
	if start != 0 || end != 2 {
		t.Errorf("expected selection (0, 2), got (%d, %d)", start, end)
	}
}

func TestTextObjectAPendingClearsOnOtherKey(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)
	m.width = 80
	m.height = 24

	m = sendKey(m, 'v')
	m = sendKey(m, 'a')

	if !m.aPending {
		t.Error("expected aPending to be true after 'a'")
	}

	// Press unrelated key - should clear pending
	m = sendKey(m, 'x')

	if m.aPending {
		t.Error("expected aPending to be false after unrelated key")
	}
}

func TestTextObjectShrinkSelection(t *testing.T) {
	content := "line0\nline1\nline2\nline3\nline4"
	sess := newTestSession(content)
	m := New(sess)
	m.width = 80
	m.height = 24

	// Select lines 1-3
	m = sendKey(m, 'j') // line 1
	m = sendKey(m, 'v')
	m = sendKey(m, 'j') // line 2
	m = sendKey(m, 'j') // line 3

	start, end := m.selection.lines()
	if start != 1 || end != 3 {
		t.Fatalf("setup failed: expected (1, 3), got (%d, %d)", start, end)
	}

	// Shrink with '{'
	m = sendKey(m, '{')
	start, end = m.selection.lines()
	if start != 1 || end != 2 {
		t.Errorf("expected selection (1, 2) after shrink, got (%d, %d)", start, end)
	}
}

func TestTextObjectGrowSelection(t *testing.T) {
	content := "line0\nline1\nline2\nline3\nline4"
	sess := newTestSession(content)
	m := New(sess)
	m.width = 80
	m.height = 24

	// Select line 1 only
	m = sendKey(m, 'j') // line 1
	m = sendKey(m, 'v')

	start, end := m.selection.lines()
	if start != 1 || end != 1 {
		t.Fatalf("setup failed: expected (1, 1), got (%d, %d)", start, end)
	}

	// Grow with '}'
	m = sendKey(m, '}')
	start, end = m.selection.lines()
	if start != 1 || end != 2 {
		t.Errorf("expected selection (1, 2) after grow, got (%d, %d)", start, end)
	}
}

func TestTextObjectShrinkMinimumOneLine(t *testing.T) {
	sess := newTestSession("line0\nline1\nline2")
	m := New(sess)
	m.width = 80
	m.height = 24

	// Select just line 1
	m = sendKey(m, 'j')
	m = sendKey(m, 'v')

	// Try to shrink - should stay at 1 line
	m = sendKey(m, '{')
	start, end := m.selection.lines()
	if start != 1 || end != 1 {
		t.Errorf("expected selection (1, 1) minimum, got (%d, %d)", start, end)
	}
}

// --- Annotations List View Tests ---

func TestAnnotationsPanel_AKeyOpensPanel(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)
	m.width = 80
	m.height = 24

	m = sendKey(m, 'a')

	if m.mode != modeAnnotations {
		t.Errorf("expected modeAnnotations, got %d", m.mode)
	}
}

func TestAnnotationsPanel_AKeyWithSelectionDoesNotOpenPanel(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)
	m.width = 80
	m.height = 24

	m = sendKey(m, 'v') // start selection
	m = sendKey(m, 'a') // should trigger text object, not panel

	if m.mode == modeAnnotations {
		t.Error("a with selection should not open annotations panel")
	}
}

func TestAnnotationsPanel_EmptyState(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)
	m.width = 80
	m.height = 24
	m.mode = modeAnnotations

	view := m.View()
	if !strings.Contains(view, "No annotations yet") {
		t.Errorf("expected empty state message, got:\n%s", view)
	}
}

func TestAnnotationsPanel_ShowsAnnotations(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)
	m.width = 80
	m.height = 24
	m.annotations = []fem.Annotation{
		{StartLine: 1, EndLine: 1, Type: "comment", Text: "First note"},
		{StartLine: 3, EndLine: 3, Type: "question", Text: "Why?"},
	}
	m.mode = modeAnnotations

	view := m.View()
	if !strings.Contains(view, "Annotations (2)") {
		t.Errorf("expected header with count, got:\n%s", view)
	}
	if !strings.Contains(view, "comment") {
		t.Errorf("expected comment type in view")
	}
	if !strings.Contains(view, "question") {
		t.Errorf("expected question type in view")
	}
}

func TestAnnotationsPanel_NavigateWithJK(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)
	m.width = 80
	m.height = 24
	m.annotations = []fem.Annotation{
		{StartLine: 1, EndLine: 1, Type: "comment", Text: "A"},
		{StartLine: 2, EndLine: 2, Type: "delete", Text: "B"},
		{StartLine: 3, EndLine: 3, Type: "question", Text: "C"},
	}

	m = sendKey(m, 'a') // open panel
	if m.annotationsCursor != 0 {
		t.Errorf("expected cursor at 0, got %d", m.annotationsCursor)
	}

	m = sendKey(m, 'j')
	if m.annotationsCursor != 1 {
		t.Errorf("expected cursor at 1 after j, got %d", m.annotationsCursor)
	}

	m = sendKey(m, 'j')
	if m.annotationsCursor != 2 {
		t.Errorf("expected cursor at 2 after j, got %d", m.annotationsCursor)
	}

	// Can't go past end
	m = sendKey(m, 'j')
	if m.annotationsCursor != 2 {
		t.Errorf("expected cursor to stay at 2, got %d", m.annotationsCursor)
	}

	m = sendKey(m, 'k')
	if m.annotationsCursor != 1 {
		t.Errorf("expected cursor at 1 after k, got %d", m.annotationsCursor)
	}
}

func TestAnnotationsPanel_EnterJumpsToAnnotation(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3\nline4\nline5")
	m := New(sess)
	m.width = 80
	m.height = 24
	m.annotations = []fem.Annotation{
		{StartLine: 4, EndLine: 5, Type: "comment", Text: "Note"},
	}

	m = sendKey(m, 'a') // open panel
	m = sendKeyEnter(m) // jump to annotation

	if m.mode != modeNormal {
		t.Errorf("expected modeNormal after Enter, got %d", m.mode)
	}
	if m.cursor != 3 { // 0-indexed, line 4
		t.Errorf("expected cursor at 3 (line 4), got %d", m.cursor)
	}
}

func TestAnnotationsPanel_EscCloses(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3")
	m := New(sess)
	m.width = 80
	m.height = 24

	m = sendKey(m, 'a')
	if m.mode != modeAnnotations {
		t.Fatalf("expected modeAnnotations, got %d", m.mode)
	}

	m = sendKeyEsc(m)
	if m.mode != modeNormal {
		t.Errorf("expected modeNormal after Esc, got %d", m.mode)
	}
}

func TestAnnotationsPanel_SortedByLine(t *testing.T) {
	sess := newTestSession("line1\nline2\nline3\nline4\nline5")
	m := New(sess)
	m.width = 80
	m.height = 24
	// Add annotations out of order
	m.annotations = []fem.Annotation{
		{StartLine: 5, EndLine: 5, Type: "comment", Text: "Last"},
		{StartLine: 1, EndLine: 1, Type: "delete", Text: "First"},
		{StartLine: 3, EndLine: 3, Type: "question", Text: "Middle"},
	}

	m = sendKey(m, 'a') // open panel
	// Cursor at 0, jump with Enter should go to line 1 (first sorted annotation)
	m = sendKeyEnter(m)
	if m.cursor != 0 { // 0-indexed, line 1
		t.Errorf("expected cursor at 0 (line 1, first sorted), got %d", m.cursor)
	}
}

func TestTextObjectGrowBoundary(t *testing.T) {
	sess := newTestSession("line0\nline1\nline2")
	m := New(sess)
	m.width = 80
	m.height = 24

	// Select last line
	m = sendKey(m, 'j') // line 1
	m = sendKey(m, 'j') // line 2
	m = sendKey(m, 'v')

	// Try to grow past end - should stay at end
	m = sendKey(m, '}')
	start, end := m.selection.lines()
	if start != 2 || end != 2 {
		t.Errorf("expected selection (2, 2) at boundary, got (%d, %d)", start, end)
	}
}
