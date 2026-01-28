package tui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charly-vibes/fabbro/internal/config"
	"github.com/charly-vibes/fabbro/internal/fem"
	"github.com/charly-vibes/fabbro/internal/tutor"
)

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
	prefixLen := 12
	contentWidth := m.width - prefixLen
	if contentWidth < 10 {
		contentWidth = 40
	}

	annotatedLines := make(map[int]bool)
	for _, ann := range m.annotations {
		annotatedLines[ann.StartLine] = true
	}

	start := 0
	if len(m.lines) > 0 {
		if m.viewportTop >= 0 {
			start = m.viewportTop
			if start < 0 {
				start = 0
			}
			if start >= len(m.lines) {
				start = len(m.lines) - 1
			}
		} else {
			start = m.autoViewportTop
			if start < 0 {
				start = 0
			}
			if start >= len(m.lines) {
				start = len(m.lines) - 1
			}
		}
	}

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
		b.WriteString(fmt.Sprintf("┌─ %s (Ctrl+J newline, Enter submit) ─────────────────┐\n", prompt))
		if m.inputTA != nil {
			taView := m.inputTA.View()
			taLines := strings.Split(taView, "\n")
			innerWidth := width - 4
			maxLines := 4
			for i, line := range taLines {
				if i >= maxLines {
					b.WriteString("│ ...                                                │\n")
					break
				}
				lineRunes := []rune(line)
				if len(lineRunes) > innerWidth {
					lineRunes = lineRunes[:innerWidth]
				}
				padding := innerWidth - len(lineRunes)
				b.WriteString(fmt.Sprintf("│ %s%s │\n", string(lineRunes), strings.Repeat(" ", padding)))
			}
		}
		b.WriteString("└────────────────────────────────────────────────────┘\n")
	case modeEditor:
		b.WriteString("┌─ Edit (Enter save, Ctrl+J newline, Esc Esc cancel) ─────┐\n")
		if m.editor != nil {
			taView := m.editor.ta.View()
			taLines := strings.Split(taView, "\n")
			innerWidth := width - 4
			maxLines := 6
			for i, line := range taLines {
				if i >= maxLines {
					b.WriteString("│ ...                                                │\n")
					break
				}
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
			b.WriteString("│ [w]rite                                            │\n")
			if m.selection.active {
				b.WriteString("├─ Annotations ──────────────────────────────────────┤\n")
				b.WriteString("│ [c]omment  [d]elete  [q]uestion  [r]eplace         │\n")
				b.WriteString("│ [e]xpand   [k]eep    [u]nclear   [i]nline-edit     │\n")
			}
			b.WriteString("│                                  [ESC] cancel      │\n")
			b.WriteString("└────────────────────────────────────────────────────┘\n")
		}
	case modeQuitConfirm:
		if m.dirty {
			b.WriteString("⚠ Unsaved changes! Quit anyway? [y/n] ")
		} else {
			b.WriteString("Quit? [y/n] ")
		}
	default:
		b.WriteString("[v]sel [SPC]cmd [w]rite [^C^C]quit")
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

var ErrTutorSession = errors.New("tutorial sessions are not saved")

func (m Model) save() error {
	if m.session.ID == tutor.SessionID {
		return ErrTutorSession
	}

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

	var sourceFileLine string
	if m.session.SourceFile != "" {
		escaped := strings.ReplaceAll(m.session.SourceFile, "'", "''")
		sourceFileLine = fmt.Sprintf("source_file: '%s'\n", escaped)
	}

	fileContent := fmt.Sprintf(`---
session_id: %s
created_at: %s
%s---

%s`, m.session.ID, m.session.CreatedAt.Format(time.RFC3339), sourceFileLine, content)

	sessionsDir, err := config.GetSessionsDir()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}
	sessionPath := filepath.Join(sessionsDir, m.session.ID+".fem")
	if err := os.WriteFile(sessionPath, []byte(fileContent), 0600); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}
	return nil
}
