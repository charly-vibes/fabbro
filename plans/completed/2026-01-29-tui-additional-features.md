# TUI Additional Features Plan

## Overview

Implement remaining TUI features: help panel, annotations list view, edit annotation range, and exit confirmation prompt.

**Spec**: `specs/03_tui_interaction.feature`

## Current State

- `?` not implemented (help panel)
- No annotations list view
- No way to edit annotation range (only text)
- `Q`/`Ctrl+C` exits immediately without confirmation

## Desired End State

- `?` shows help panel with all keybindings
- `a` or `SPC a` shows annotations list with navigation
- `R` allows editing annotation range
- Exit prompts for unsaved changes (optional future)

---

## Phase 1: Help Panel (~45 min)

**Beads ticket**: fabbro-4qo (already exists)

**Spec scenario**: "Viewing help"

### Changes Required

**File: internal/tui/tui.go**

Add `helpOpen bool` to model.

Handle `?` key:
```go
case "?":
    m.helpOpen = !m.helpOpen
```

Render help overlay:
```
┌─ Help ─────────────────────────────────────────────────────┐
│                                                            │
│ NAVIGATION                                                 │
│   j/↓      Move down          k/↑      Move up             │
│   gg       Go to top          G        Go to bottom        │
│   Ctrl+d   Page down          Ctrl+u   Page up             │
│   /        Search             n/N      Next/prev match     │
│   zz       Center cursor                                   │
│                                                            │
│ SELECTION                                                  │
│   v        Start/toggle selection                          │
│   Esc      Clear selection                                 │
│                                                            │
│ ANNOTATIONS (with selection or SPC menu)                   │
│   c        Comment            d        Delete              │
│   e        Expand             q        Question            │
│   k        Keep               u        Unclear             │
│   r        Replace/Change                                  │
│                                                            │
│ EDITING                                                    │
│   i        Inline edit        e        Edit annotation     │
│   R        Edit range                                      │
│                                                            │
│ FILE                                                       │
│   w        Save and exit      Q        Quit (no save)      │
│   SPC      Command palette    ?        This help           │
│                                                            │
│                        Press any key to close              │
└────────────────────────────────────────────────────────────┘
```

### Success Criteria

- [ ] `?` opens help overlay
- [ ] Any key closes help
- [ ] All keybindings documented

---

## Phase 2: Annotations List View (~1.5 hours)

**Spec scenarios**: "Viewing all annotations in session", "Jumping to annotation from list"

### Changes Required

**File: internal/tui/tui.go**

Add annotations panel mode:
```go
type model struct {
    annotationsPanelOpen bool
    annotationsCursor    int
}
```

Handle `a` key or add to SPC palette.

Render annotations panel:
```
┌─ Annotations (6) ──────────────────────────────────────────┐
│   LINE   TYPE       PREVIEW                                │
│ > 10-15  comment    This section needs more examples       │
│   20     question   Why not use dependency injection?      │
│   42-50  delete     Too verbose                            │
│   60     keep       [no text]                              │
│   72     expand     Add error handling examples            │
│   85     unclear    What does this mean?                   │
│                                                            │
│   j/k: navigate   Enter: jump to line   Esc: close         │
└────────────────────────────────────────────────────────────┘
```

Navigation:
- `j`/`k` moves selection in list
- `Enter` jumps to annotation line and closes panel
- `Esc` closes panel without jumping

### Success Criteria

- [ ] `a` opens annotations panel
- [ ] List shows all annotations with type and preview
- [ ] Can navigate and jump to annotation
- [ ] Empty state shows "No annotations yet"

---

## Phase 3: Edit Annotation Range (~45 min)

**Spec scenario**: "Editing annotation range"

### Changes Required

**File: internal/tui/tui.go**

Handle `R` key (uppercase) when cursor is on annotated line:
1. Find annotation on current line
2. If multiple, show picker (existing logic)
3. Activate selection mode with annotation's current range
4. `j`/`k` adjusts range
5. `Enter` confirms new range
6. `Esc` cancels

```go
case "R":
    annotations := m.annotationsOnLine(m.cursor)
    if len(annotations) == 0 {
        m.setError("No annotation on this line")
        return m, nil
    }
    // Start range edit mode
    m.rangeEditMode = true
    m.rangeEditAnnotation = annotations[0] // or show picker
    m.selection.active = true
    m.selection.anchor = annotations[0].StartLine
    m.selection.cursor = annotations[0].EndLine
```

### Success Criteria

- [ ] `R` starts range edit on current annotation
- [ ] Selection shows current annotation range
- [ ] `j`/`k` adjusts range
- [ ] `Enter` saves new range
- [ ] Annotation picker shown when multiple on line

---

## Phase 4: Exit Confirmation (Optional) (~30 min)

**Spec scenario**: "Exiting with confirmation prompt"

### Changes Required

This is marked `@planned` but noted as "Future". Implementation:

**File: internal/tui/tui.go**

Track unsaved changes:
```go
type model struct {
    dirty bool  // true when annotations added/modified since last save
}
```

On `Ctrl+C` when dirty:
```
┌─ Unsaved Changes ──────────────────────────────────────────┐
│                                                            │
│   You have unsaved annotations.                            │
│                                                            │
│   [y] Save and exit   [n] Exit without saving   [c] Cancel │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

### Success Criteria

- [ ] `Ctrl+C` with unsaved changes shows prompt
- [ ] `y` saves and exits
- [ ] `n` exits without saving
- [ ] `c` cancels and returns to TUI

---

## Summary

| Phase | Deliverable | Time |
|-------|-------------|------|
| 1 | Help panel (`?`) | 45m |
| 2 | Annotations list view (`a`) | 1.5h |
| 3 | Edit annotation range (`R`) | 45m |
| 4 | Exit confirmation (optional) | 30m |

**Total: ~3.5 hours**

## Dependencies

- Phase 3 depends on annotation picker being implemented (already exists for `e` key)
