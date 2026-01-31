package tui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
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
	prefixLen := 13 // cursor(1) + rangeIndicator(1) + selIndicator(1) + space(1) + lineNum(3) + space(1) + annIndicator(1) + space(1) + │(1) + space(1) = 13
	contentWidth := m.width - prefixLen
	if contentWidth < 10 {
		contentWidth = 40
	}

	annotatedLines := make(map[int]bool)
	for _, ann := range m.annotations {
		annotatedLines[ann.StartLine] = true
	}

	// Get the currently previewed annotation for range highlighting
	previewedAnn := m.previewedAnnotation()

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

		// Annotation range highlight: show when previewing an annotation
		// and this line is within its range
		rangeIndicator := " "
		lineNum1Based := i + 1
		if previewedAnn != nil && lineNum1Based >= previewedAnn.StartLine && lineNum1Based <= previewedAnn.EndLine {
			rangeIndicator = "▐"
		}

		annIndicator := " "
		if annotatedLines[i+1] {
			annIndicator = "●"
		}

		searchIndicator := " "
		if m.isSearchMatch(i) {
			searchIndicator = "◎"
		}

		highlightedLine := m.highlighter.RenderLine(line)
		isCurrentMatch := m.isCurrentSearchMatch(i)
		if m.search.query != "" && m.isSearchMatch(i) {
			highlightedLine = m.highlightSearchMatches(highlightedLine, line, isCurrentMatch)
		}

		wrapped := wrapLine(line, contentWidth)
		for j, part := range wrapped {
			var displayPart string
			if j == 0 && len(wrapped) == 1 {
				displayPart = highlightedLine
			} else {
				displayPart = m.highlighter.RenderLine(part)
				if m.search.query != "" && m.isSearchMatch(i) {
					displayPart = m.highlightSearchMatches(displayPart, part, isCurrentMatch)
				}
			}
			indicator := annIndicator
			if searchIndicator != " " {
				indicator = searchIndicator
			}
			if j == 0 {
				b.WriteString(fmt.Sprintf("%s%s%s %s %s │ %s\n", cursor, rangeIndicator, selIndicator, lineNum, indicator, displayPart))
			} else {
				b.WriteString(fmt.Sprintf("          │ %s\n", displayPart))
			}
		}
	}

	b.WriteString(strings.Repeat("─", width))
	b.WriteString("\n")

	switch m.mode {
	case modeInput:
		prompt := fem.Prompts[m.inputType]
		// Box total width = width - 2 (leave small margin)
		// Inner content width = boxTotalWidth - 4 (for │ and spaces on each side)
		boxTotalWidth := width - 2
		if boxTotalWidth < 24 {
			boxTotalWidth = 64
		}
		innerWidth := boxTotalWidth - 4
		header := fmt.Sprintf("─ %s (Ctrl+J newline, Enter submit) ", prompt)
		headerPad := boxTotalWidth - len([]rune(header)) - 2 // -2 for ┌ and ┐
		if headerPad < 0 {
			headerPad = 0
		}
		b.WriteString(fmt.Sprintf("┌%s%s┐\n", header, strings.Repeat("─", headerPad)))
		if m.inputTA != nil {
			taView := m.inputTA.View()
			taLines := strings.Split(taView, "\n")
			maxLines := 4
			for i, line := range taLines {
				if i >= maxLines {
					b.WriteString(fmt.Sprintf("│ ...%s │\n", strings.Repeat(" ", innerWidth-4)))
					break
				}
				visibleWidth := lipgloss.Width(line)
				displayLine := line
				if visibleWidth > innerWidth {
					displayLine = ansi.Truncate(line, innerWidth, "")
					visibleWidth = innerWidth
				}
				padding := innerWidth - visibleWidth
				if padding < 0 {
					padding = 0
				}
				b.WriteString(fmt.Sprintf("│ %s%s │\n", displayLine, strings.Repeat(" ", padding)))
			}
		}
		b.WriteString(fmt.Sprintf("└%s┘\n", strings.Repeat("─", boxTotalWidth-2)))
	case modeEditor:
		// Box total width = width - 2 (leave small margin)
		// Inner content width = boxTotalWidth - 4 (for │ and spaces on each side)
		boxTotalWidth := width - 2
		if boxTotalWidth < 24 {
			boxTotalWidth = 64
		}
		innerWidth := boxTotalWidth - 4
		header := "─ Edit (Enter save, Ctrl+J newline, Esc Esc cancel) "
		headerPad := boxTotalWidth - len([]rune(header)) - 2 // -2 for ┌ and ┐
		if headerPad < 0 {
			headerPad = 0
		}
		b.WriteString(fmt.Sprintf("┌%s%s┐\n", header, strings.Repeat("─", headerPad)))
		if m.editor != nil {
			taView := m.editor.ta.View()
			taLines := strings.Split(taView, "\n")
			maxLines := 6
			for i, line := range taLines {
				if i >= maxLines {
					b.WriteString(fmt.Sprintf("│ ...%s │\n", strings.Repeat(" ", innerWidth-4)))
					break
				}
				visibleWidth := lipgloss.Width(line)
				displayLine := line
				if visibleWidth > innerWidth {
					displayLine = ansi.Truncate(line, innerWidth, "")
					visibleWidth = innerWidth
				}
				padding := innerWidth - visibleWidth
				if padding < 0 {
					padding = 0
				}
				b.WriteString(fmt.Sprintf("│ %s%s │\n", displayLine, strings.Repeat(" ", padding)))
			}
		}
		b.WriteString(fmt.Sprintf("└%s┘\n", strings.Repeat("─", boxTotalWidth-2)))
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
	case modeSearch:
		b.WriteString(fmt.Sprintf("/%s█", m.search.query))
		if len(m.search.matches) > 0 {
			b.WriteString(fmt.Sprintf(" [%d/%d]", m.search.current+1, len(m.search.matches)))
		}
	case modeHelp:
		b.WriteString(m.renderHelpPanel(width))
	default:
		// Check if cursor is on an annotated line
		cursorLine := m.cursor + 1 // 1-indexed
		annotationIndices := m.annotationsOnLine(cursorLine)
		if len(annotationIndices) > 0 && !m.selection.active {
			// Determine which annotation to show
			previewIdx := 0
			if m.previewLine == cursorLine && m.previewIndex < len(annotationIndices) {
				previewIdx = m.previewIndex
			}
			// Show annotation preview instead of help text
			b.WriteString(m.renderAnnotationPreviewAt(annotationIndices, previewIdx, width))
		} else {
			helpText := "[v]sel [SPC]cmd [/]search [w]rite [^C^C]quit [?]help"
			if m.selection.active {
				helpText += " │ [c]omment [d]elete [q]uestion [e]xpand [u]nclear [r]eplace [i]nline"
			}
			if len(m.search.matches) > 0 {
				helpText += fmt.Sprintf(" │ [n]ext [p]rev %d/%d", m.search.current+1, len(m.search.matches))
			}
			if len([]rune(helpText)) > width {
				helpText = string([]rune(helpText)[:width])
			}
			b.WriteString(helpText)
			b.WriteString("\n")
		}
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

func (m Model) isSearchMatch(lineIdx int) bool {
	for _, idx := range m.search.matches {
		if idx == lineIdx {
			return true
		}
	}
	return false
}

func (m Model) isCurrentSearchMatch(lineIdx int) bool {
	if len(m.search.matches) == 0 {
		return false
	}
	if m.search.current >= 0 && m.search.current < len(m.search.matches) {
		return m.search.matches[m.search.current] == lineIdx
	}
	return false
}

func (m Model) highlightSearchMatches(rendered, original string, isCurrent bool) string {
	if m.search.query == "" {
		return rendered
	}
	query := strings.ToLower(m.search.query)
	lowerOriginal := strings.ToLower(original)

	var matchStyle lipgloss.Style
	if isCurrent {
		matchStyle = lipgloss.NewStyle().Background(lipgloss.Color("208")).Foreground(lipgloss.Color("0"))
	} else {
		matchStyle = lipgloss.NewStyle().Background(lipgloss.Color("226")).Foreground(lipgloss.Color("0"))
	}

	matchPositions := fuzzyMatchPositions(lowerOriginal, query)
	if len(matchPositions) == 0 {
		return rendered
	}

	var result strings.Builder
	lastEnd := 0
	for _, pos := range matchPositions {
		if pos > lastEnd {
			result.WriteString(original[lastEnd:pos])
		}
		result.WriteString(matchStyle.Render(string(original[pos])))
		lastEnd = pos + 1
	}
	if lastEnd < len(original) {
		result.WriteString(original[lastEnd:])
	}
	return result.String()
}

func fuzzyMatchPositions(text, pattern string) []int {
	var positions []int
	pIdx := 0
	for i := 0; i < len(text) && pIdx < len(pattern); i++ {
		if text[i] == pattern[pIdx] {
			positions = append(positions, i)
			pIdx++
		}
	}
	if pIdx == len(pattern) {
		return positions
	}
	return nil
}

// renderAnnotationPreview renders a preview box for annotations on the current line.
// It displays the first annotation's content and shows a count if multiple exist.
func (m Model) renderAnnotationPreview(annotationIndices []int, width int) string {
	return m.renderAnnotationPreviewAt(annotationIndices, 0, width)
}

// renderAnnotationPreviewAt renders a preview box for the annotation at the given index.
func (m Model) renderAnnotationPreviewAt(annotationIndices []int, previewIdx int, width int) string {
	if len(annotationIndices) == 0 {
		return ""
	}
	if previewIdx < 0 || previewIdx >= len(annotationIndices) {
		previewIdx = 0
	}

	var b strings.Builder

	// Get the annotation at the specified index
	ann := m.annotations[annotationIndices[previewIdx]]

	// Box dimensions
	boxTotalWidth := width - 2
	if boxTotalWidth < 24 {
		boxTotalWidth = 64
	}
	innerWidth := boxTotalWidth - 4 // for │ and spaces on each side

	// Header: "─ type [start-end] ─────"
	header := fmt.Sprintf("─ %s [%d-%d] ", ann.Type, ann.StartLine, ann.EndLine)
	headerPad := boxTotalWidth - len([]rune(header)) - 2 // -2 for ┌ and ┐
	if headerPad < 0 {
		headerPad = 0
	}
	b.WriteString(fmt.Sprintf("┌%s%s┐\n", header, strings.Repeat("─", headerPad)))

	// Wrap annotation text to fit within inner width
	textLines := wrapText(ann.Text, innerWidth)
	maxLines := 3 // Limit preview to 3 lines
	for i, line := range textLines {
		if i >= maxLines {
			b.WriteString(fmt.Sprintf("│ ...%s │\n", strings.Repeat(" ", innerWidth-4)))
			break
		}
		lineRunes := []rune(line)
		padding := innerWidth - len(lineRunes)
		if padding < 0 {
			padding = 0
		}
		b.WriteString(fmt.Sprintf("│ %s%s │\n", line, strings.Repeat(" ", padding)))
	}

	// Footer: show count if multiple annotations
	var footer string
	if len(annotationIndices) > 1 {
		footer = fmt.Sprintf("─ (%d of %d annotations) ", previewIdx+1, len(annotationIndices))
	} else {
		footer = "─"
	}
	footerPad := boxTotalWidth - len([]rune(footer)) - 2 // -2 for └ and ┘
	if footerPad < 0 {
		footerPad = 0
	}
	b.WriteString(fmt.Sprintf("└%s%s┘\n", footer, strings.Repeat("─", footerPad)))

	return b.String()
}

// wrapText wraps text to fit within the given width.
func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	var lines []string
	runes := []rune(text)

	for len(runes) > 0 {
		if len(runes) <= width {
			lines = append(lines, string(runes))
			break
		}

		// Find a good break point (prefer space)
		breakAt := width
		for i := width - 1; i >= 0; i-- {
			if runes[i] == ' ' {
				breakAt = i + 1
				break
			}
		}

		lines = append(lines, string(runes[:breakAt]))
		runes = runes[breakAt:]
	}

	return lines
}

// renderHelpPanel renders the help menu overlay.
func (m Model) renderHelpPanel(width int) string {
	var b strings.Builder

	boxWidth := width - 4
	if boxWidth < 50 {
		boxWidth = 50
	}
	innerWidth := boxWidth - 4

	// Header
	header := "─ Help "
	versionStr := ""
	if m.version != "" {
		versionStr = fmt.Sprintf("(fabbro %s) ", m.version)
	}
	header += versionStr
	headerPad := boxWidth - len([]rune(header)) - 2
	if headerPad < 0 {
		headerPad = 0
	}
	b.WriteString(fmt.Sprintf("┌%s%s┐\n", header, strings.Repeat("─", headerPad)))

	// Helper to write a row
	writeRow := func(left, right string) {
		leftRunes := []rune(left)
		rightRunes := []rune(right)
		padding := innerWidth - len(leftRunes) - len(rightRunes)
		if padding < 1 {
			padding = 1
		}
		b.WriteString(fmt.Sprintf("│ %s%s%s │\n", left, strings.Repeat(" ", padding), right))
	}

	// Navigation section
	writeRow("NAVIGATION", "")
	writeRow("  j/k, ↑/↓", "move cursor")
	writeRow("  Ctrl+d/u", "scroll half page")
	writeRow("  gg / G", "jump to first/last line")
	writeRow("  zz/zt/zb", "center/top/bottom cursor")
	writeRow("", "")

	// Selection section
	writeRow("SELECTION", "")
	writeRow("  v", "toggle line selection")
	writeRow("  Esc", "clear selection")
	writeRow("", "")

	// Annotations section
	writeRow("ANNOTATIONS (with selection)", "")
	writeRow("  c", "comment")
	writeRow("  d", "delete")
	writeRow("  q", "question")
	writeRow("  e", "expand")
	writeRow("  u", "unclear")
	writeRow("  r", "replace/change")
	writeRow("  i", "inline edit")
	writeRow("", "")

	// General section
	writeRow("GENERAL", "")
	writeRow("  /", "search")
	writeRow("  n / N,p", "next/prev match")
	writeRow("  Space", "command palette")
	writeRow("  w", "save session")
	writeRow("  ?", "this help")
	writeRow("  Ctrl+C Ctrl+C", "quit")

	// Footer
	footer := "─ Press any key to close "
	footerPad := boxWidth - len([]rune(footer)) - 2
	if footerPad < 0 {
		footerPad = 0
	}
	b.WriteString(fmt.Sprintf("└%s%s┘\n", footer, strings.Repeat("─", footerPad)))

	return b.String()
}
