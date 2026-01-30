# TUI Mouse Support Plan

## Overview

Add mouse interaction to the TUI: click to position cursor, drag to select, right-click context menu.

**Spec**: `specs/03_tui_interaction.feature` - "Mouse Interaction (Phase 1.5)"

## Current State

- All interaction is keyboard-based
- Bubble Tea supports mouse events via `tea.EnableMouseCellMotion()`

## Desired End State

- Click positions cursor on clicked line
- Click-drag selects line range
- Right-click opens context menu matching SPC palette

---

## Phase 1: Enable Mouse and Click Positioning (~45 min)

### Changes Required

**File: internal/tui/tui.go**

Enable mouse in `Init()`:
```go
func (m model) Init() tea.Cmd {
    return tea.EnableMouseCellMotion()
}
```

Handle mouse events in `Update()`:
```go
case tea.MouseMsg:
    if msg.Type == tea.MouseLeft && msg.Action == tea.MouseActionPress {
        // Convert screen Y to document line
        line := m.viewport.YOffset + msg.Y - headerHeight
        if line >= 0 && line < len(m.lines) {
            m.cursor = line
        }
    }
```

### Challenges

- Account for viewport offset (scrolled content)
- Account for header/footer height
- Line number gutter offset

### Boundary Checking

```go
func (m Model) screenToLine(screenY int) (line int, valid bool) {
    // Header takes up headerHeight rows
    contentY := screenY - m.headerHeight
    if contentY < 0 {
        return 0, false  // Click in header
    }
    
    // Check footer boundary
    footerStart := m.height - m.footerHeight
    if screenY >= footerStart {
        return 0, false  // Click in footer
    }
    
    // Convert to document line
    line = m.viewport.YOffset + contentY
    if line >= len(m.lines) {
        return len(m.lines) - 1, false  // Below content, clamp to last line
    }
    
    return line, true
}
```

### Success Criteria

- [ ] Clicking on a line moves cursor to that line
- [ ] Clicking on header is ignored (no cursor move)
- [ ] Clicking on footer is ignored (no cursor move)
- [ ] Clicking below content clamps to last line (or ignored)
- [ ] Works correctly when scrolled

---

## Phase 2: Drag Selection (~1 hour)

### Changes Required

**File: internal/tui/tui.go**

Track mouse state:
```go
type model struct {
    // ...
    mouseSelecting bool
    mouseAnchor    int
}
```

Handle drag:
```go
case tea.MouseMsg:
    if msg.Type == tea.MouseLeft {
        line := m.screenToLine(msg.Y)
        switch msg.Action {
        case tea.MouseActionPress:
            m.mouseSelecting = true
            m.mouseAnchor = line
            m.selection.active = true
            m.selection.anchor = line
            m.selection.cursor = line
        case tea.MouseActionMotion:
            if m.mouseSelecting {
                m.selection.cursor = line
            }
        case tea.MouseActionRelease:
            m.mouseSelecting = false
        }
    }
```

### Success Criteria

- [ ] Click-drag selects range of lines
- [ ] Selection updates as mouse moves
- [ ] Release ends selection (stays selected)

---

## Phase 3: Right-Click Context Menu (~1 hour)

### Changes Required

**File: internal/tui/tui.go**

Add context menu mode:
```go
type model struct {
    // ...
    contextMenuOpen bool
    contextMenuX    int
    contextMenuY    int
}
```

Handle right-click:
```go
case tea.MouseMsg:
    if msg.Type == tea.MouseRight && msg.Action == tea.MouseActionPress {
        m.contextMenuOpen = true
        m.contextMenuX = msg.X
        m.contextMenuY = msg.Y
        // Position cursor on clicked line
        m.cursor = m.screenToLine(msg.Y)
    }
```

Render context menu in `View()`:
```
┌─────────────────┐
│ [c] Comment     │
│ [d] Delete      │
│ [e] Expand      │
│ [q] Question    │
│ [k] Keep        │
│ [u] Unclear     │
└─────────────────┘
```

Handle menu selection same as SPC palette.

### Success Criteria

- [ ] Right-click opens context menu
- [ ] Menu appears near click position
- [ ] Clicking menu item triggers annotation
- [ ] Click outside or ESC dismisses menu

---

## Summary

| Phase | Deliverable | Time |
|-------|-------------|------|
| 1 | Click to position cursor | 45m |
| 2 | Drag to select range | 1h |
| 3 | Right-click context menu | 1h |

**Total: ~2.75 hours**

## Dependencies

- Requires multi-line selection (post-tracer Phase 5)
- Context menu reuses SPC palette logic (post-tracer Phase 3)
