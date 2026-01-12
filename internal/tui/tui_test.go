package tui

import (
	"strings"
	"testing"
	"time"

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

	// Quit with q
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("expected quit command, got nil")
	}

	// Quit with ctrl+c
	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("expected quit command from ctrl+c, got nil")
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
	if m.selected != -1 {
		t.Error("expected selection cleared after submit")
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
