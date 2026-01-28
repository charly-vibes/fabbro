package tui

func (m *Model) ensureCursorVisible() {
	if m.viewportTop >= 0 || len(m.lines) == 0 {
		return
	}

	visibleRows := m.height - 4
	if visibleRows < 5 {
		visibleRows = 10
	}

	prefixLen := 12
	contentWidth := m.width - prefixLen
	if contentWidth < 10 {
		contentWidth = 40
	}

	scrolloff := 2
	if visibleRows/4 < scrolloff {
		scrolloff = visibleRows / 4
	}
	if scrolloff < 0 {
		scrolloff = 0
	}

	if m.autoViewportTop < 0 {
		m.autoViewportTop = 0
	}
	if m.autoViewportTop >= len(m.lines) {
		m.autoViewportTop = len(m.lines) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(m.lines) {
		m.cursor = len(m.lines) - 1
	}

	rowsAbove := 0
	for i := m.autoViewportTop; i < m.cursor && rowsAbove <= visibleRows+scrolloff; i++ {
		rowsAbove += len(wrapLine(m.lines[i], contentWidth))
	}
	cursorHeight := len(wrapLine(m.lines[m.cursor], contentWidth))
	if cursorHeight < 1 {
		cursorHeight = 1
	}

	for rowsAbove < scrolloff && m.autoViewportTop > 0 {
		m.autoViewportTop--
		rowsAbove += len(wrapLine(m.lines[m.autoViewportTop], contentWidth))
	}

	maxAbove := visibleRows - scrolloff - cursorHeight
	if maxAbove < 0 {
		maxAbove = 0
	}
	for rowsAbove > maxAbove && m.autoViewportTop < m.cursor {
		rowsAbove -= len(wrapLine(m.lines[m.autoViewportTop], contentWidth))
		m.autoViewportTop++
	}
}
