# TUI Selection Expansion Plan

## Overview

Add vim-surround style text object selection: expand to paragraph (`ap`), code block (`ab`), section (`as`), and shrink/grow by line (`{`, `}`).

**Spec**: `specs/03_tui_interaction.feature` - "Vim-surround Style Context Expansion"

## Current State

- `v` starts line selection at current cursor
- `j`/`k` extends selection up/down
- No text object awareness

## Desired End State

- `ap` expands selection to paragraph (blank-line delimited)
- `ab` expands to code block (fenced ``` markers)
- `as` expands to section (heading to next heading)
- `{` shrinks selection by one line from end
- `}` grows selection by one line

---

## Phase 1: Text Object Detection (~1 hour)

### Changes Required

**File: internal/tui/textobjects.go** (new)

```go
package tui

// FindParagraph returns start/end lines of paragraph containing line
func FindParagraph(lines []string, line int) (start, end int)

// FindCodeBlock returns start/end of fenced code block containing line
// Returns -1, -1 if line is not in a code block
func FindCodeBlock(lines []string, line int) (start, end int)

// FindSection returns start/end of markdown section containing line
// Section runs from heading to next heading of same or higher level
func FindSection(lines []string, line int) (start, end int)
```

### Implementation Details

**Paragraph detection**:
- Walk up from line until blank line or start of file
- Walk down until blank line or end of file

**Code block detection**:
- Scan for ``` markers
- Track open/close state
- Return block boundaries if line is inside

**Section detection**:
- Find heading above current line (# to ######)
- Find next heading of same or higher level
- Return section boundaries

### Success Criteria

- [ ] Unit tests for each text object type
- [ ] Edge cases: start/end of file, nested blocks

---

## Phase 2: Selection Mode Expansion (~45 min)

### Changes Required

**File: internal/tui/tui.go**

Add `aPending bool` to model (similar to `gPending` for `gg`).

In selection mode, handle:
- `a` → set `aPending = true`
- If `aPending`:
  - `p` → expand to paragraph
  - `b` → expand to code block
  - `s` → expand to section
  - Any other key → clear pending

### Success Criteria

- [ ] `v` then `ap` selects paragraph
- [ ] `v` then `ab` selects code block (or does nothing if not in block)
- [ ] `v` then `as` selects section
- [ ] Status bar updates to show expanded selection

---

## Phase 3: Shrink/Grow Selection (~30 min)

### Changes Required

**File: internal/tui/tui.go**

In selection mode:
- `{` → shrink selection (move end toward anchor, min 1 line)
- `}` → grow selection (move end away from anchor)

### Behavior Details

- `{` reduces selection: if cursor > anchor, move cursor up; else move anchor down
- `}` expands selection: if cursor > anchor, move cursor down; else move anchor up
- Minimum selection is 1 line (can't shrink below that)

### Success Criteria

- [ ] `{` shrinks selection by one line
- [ ] `}` grows selection by one line
- [ ] Cannot shrink below 1 line

---

## Summary

| Phase | Deliverable | Time |
|-------|-------------|------|
| 1 | Text object detection functions | 1h |
| 2 | `ap`, `ab`, `as` keybindings | 45m |
| 3 | `{`, `}` shrink/grow | 30m |

**Total: ~2.25 hours**

## Dependencies

Phase 2 and 3 depend on multi-line selection being implemented (from post-tracer-features.md Phase 5).
