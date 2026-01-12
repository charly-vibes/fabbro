# Post-Tracer Features Implementation Plan

## Overview

With the tracer bullet complete (init → review → annotate → apply --json), this plan adds the features needed to make fabbro useful for real review workflows.

**Goal**: Transform the skeleton into a usable tool by adding remaining annotation types, better navigation, and the SPC command palette.

## Related

- Tracer Bullet: `plans/2026-01-11-tracer-bullet.md`
- Testing Strategy: `plans/2026-01-11-testing-strategy.md`
- Specs: `specs/03_tui_interaction.feature`, `specs/06_fem_markup.feature`
- Design: `research/2026-01-09-fabbro-design-document.md`

## Current State

After tracer bullet:
- ✅ `fabbro init` creates `.fabbro/sessions/`
- ✅ `fabbro review --stdin` creates session, opens TUI
- ✅ TUI: j/k navigation, v selection, c comment, w save, q quit
- ✅ FEM parser extracts `{>> comment <<}` annotations
- ✅ `fabbro apply <id> --json` outputs structured JSON
- ⚠️ Coverage at 69.7% (target: 98%)

## Desired End State

After this plan:
- All 6 annotation types work (comment, delete, expand, question, keep, unclear)
- SPC command palette for discoverability
- Page navigation (Ctrl+d/u, gg, G)
- Multi-line selection (v + j/k)
- Coverage ≥85%

## Out of Scope (Future Plans)

- Block delete syntax (`{-- --}...{--/--}` spanning lines) - inline delete only for now
- Search (`/`)
- Mouse support
- Selection expansion (ap, ab, as)
- Templates
- Config file
- `fabbro resume`
- Annotations panel view

---

## Phase 1: Fix Quit Keybinding Conflict (~15 min)

**Spec**: `specs/03_tui_interaction.feature` - "Adding a question annotation"

### Problem

`q` is used for both quit and question annotation. This blocks adding question annotations.

### Solution

Change quit to `Q` (shift+q). Keep `q` available for question annotation.

### Changes Required

**File: internal/tui/tui.go**
- Change quit from `q` to `Q` (shift+q)
- Update help text in View()

### Success Criteria

#### Automated:
- [ ] Test: `Q` quits without saving
- [ ] Test: `q` no longer quits

#### Manual:
- [ ] `Q` exits TUI
- [ ] `q` does nothing (until Phase 2 adds question annotation)

---

## Phase 2: Remaining Annotation Types (~2 hours)

**Spec**: `specs/06_fem_markup.feature` - All annotation type scenarios

### 2.1 Parser: Add All FEM Patterns

**File: internal/fem/parser.go**

Add regex patterns for:
| Type | Pattern | Example |
|------|---------|---------|
| delete | `{-- ... --}` | `{-- DELETE: Too verbose --}` |
| question | `{?? ... ??}` | `{?? Why this approach? ??}` |
| expand | `{!! ... !!}` | `{!! EXPAND: Add examples !!}` |
| keep | `{== ... ==}` | `{== KEEP: Good section ==}` |
| unclear | `{~~ ... ~~}` | `{~~ UNCLEAR: Confusing ~~}` |

**Implementation approach:**
```go
var patterns = map[string]*regexp.Regexp{
    "comment":  regexp.MustCompile(`\{>>\s*(.*?)\s*<<\}`),
    "delete":   regexp.MustCompile(`\{--\s*(.*?)\s*--\}`),
    "question": regexp.MustCompile(`\{\?\?\s*(.*?)\s*\?\?\}`),
    "expand":   regexp.MustCompile(`\{!!\s*(.*?)\s*!!\}`),
    "keep":     regexp.MustCompile(`\{==\s*(.*?)\s*==\}`),
    "unclear":  regexp.MustCompile(`\{~~\s*(.*?)\s*~~\}`),
}
```

### 2.2 TUI: Add Annotation Keybindings

**File: internal/tui/tui.go**

In `handleNormalMode`, add:
- `d` → delete annotation (prompt: "Reason for deletion:")
- `q` → question annotation (prompt: "Question:")  
- `e` → expand annotation (prompt: "What to expand:")
- `k` → keep annotation (no prompt, just mark)
- `u` → unclear annotation (prompt: "What's unclear:")

**Note**: Phase 1 already changed quit to `Q`, so `q` is available.

### Success Criteria

#### Automated:
- [ ] `go test ./...` passes
- [ ] Parser tests for all 6 annotation types
- [ ] TUI tests for all keybindings

#### Manual:
- [ ] Each annotation type can be added in TUI
- [ ] `fabbro apply --json` shows correct type for each

---

## Phase 3: SPC Command Palette (~1.5 hours)

**Spec**: `specs/03_tui_interaction.feature` - "Helix-style SPC Menu (Discoverability)"

### Changes Required

**File: internal/tui/tui.go**

Add new mode: `"palette"`

When SPC pressed:
```
─── Review: abc123 ───────────────────────────────────
  1 │ First line of content
  2 │ Second line                          ← SELECTED
  3 │ Third line
──────────────────────────────────────────────────────
┌─ Annotations ──────────────────────────────────────┐
│ [c] comment   [d] delete   [e] expand              │
│ [q] question  [k] keep     [u] unclear             │
└────────────────────────────────────────────────────┘
```

**Implementation:**
- `mode: "palette"` shows overlay
- Any annotation key (c/d/e/q/k/u) triggers that action
- ESC dismisses palette
- Any other key dismisses and is ignored

### Success Criteria

#### Automated:
- [ ] `go test ./...` passes
- [ ] Test: SPC opens palette mode
- [ ] Test: pressing 'c' in palette triggers comment

#### Manual:
- [ ] SPC shows command palette overlay
- [ ] Each key triggers correct annotation
- [ ] ESC dismisses without action

---

## Phase 4: Improved Navigation (~1 hour)

**Spec**: `specs/03_tui_interaction.feature` - "Navigation" scenarios

### Changes Required

**File: internal/tui/tui.go**

Add to `handleNormalMode`:
- `Ctrl+d` → scroll down half page
- `Ctrl+u` → scroll up half page
- `g` → start "g-pending" mode
- `gg` (g then g) → jump to first line
- `G` → jump to last line

**Implementation notes:**
- Track `gPending bool` in model
- If `g` pressed and `gPending`, go to line 1, clear pending
- If `g` pressed and not pending, set pending
- Any other key clears pending

### Success Criteria

#### Automated:
- [ ] `go test ./...` passes
- [ ] Test: Ctrl+d moves cursor down by half viewport
- [ ] Test: gg moves to line 0
- [ ] Test: G moves to last line

#### Manual:
- [ ] Navigation feels responsive on 100+ line files

---

## Phase 5: Multi-line Selection (~1 hour)

**Spec**: `specs/03_tui_interaction.feature` - "Text Selection" scenarios

### Changes Required

**File: internal/tui/tui.go**

Current: `selected int` (single line, -1 if none)

Change to:
```go
type selection struct {
    active bool
    anchor int  // where selection started
    cursor int  // current end of selection
}

func (s selection) lines() (start, end int) {
    if s.anchor < s.cursor {
        return s.anchor, s.cursor
    }
    return s.cursor, s.anchor
}
```

**Behavior:**
- `v` toggles selection mode
- In selection mode, j/k extends selection from anchor
- ESC clears selection
- Any annotation key operates on selected range

**View changes:**
- Highlight all lines in selection range
- Status bar shows "N lines selected"

### Success Criteria

#### Automated:
- [ ] `go test ./...` passes
- [ ] Test: v starts selection at current line
- [ ] Test: j/k in selection extends range
- [ ] Test: annotation applies to full range

#### Manual:
- [ ] Can select lines 5-15 and add comment to all

---

## Phase 6: Coverage Push to 85% (~1.5 hours)

**Ref**: `plans/2026-01-11-testing-strategy.md` - "Coverage by Design" section

### Strategy

Current coverage by package:
- cmd/fabbro: ~0% (main only)
- internal/config: 100%
- internal/fem: 100%
- internal/session: 100%
- internal/tui: 100%

The 69.7% total is due to cmd/fabbro main.go not being tested.

**Fix:**
1. Refactor main.go to use `realMain()` pattern (per testing-strategy.md)
2. Add tests for CLI command parsing
3. Add tests for error paths

**File: cmd/fabbro/main.go**
```go
func main() {
    os.Exit(realMain(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func realMain(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
    // ... all logic here
}
```

### Success Criteria

#### Automated:
- [ ] `just check-coverage` passes (≥85%)
- [ ] `go test ./...` passes

---

## Summary

| Phase | Deliverable | Time |
|-------|-------------|------|
| 1 | Fix q/Q quit conflict | 15m |
| 2 | All 6 annotation types | 2h |
| 3 | SPC command palette | 1.5h |
| 4 | Page navigation (Ctrl+d/u, gg, G) | 1h |
| 5 | Multi-line selection | 1h |
| 6 | Coverage ≥85% | 1.5h |

**Total: ~7-8 hours**

---

## Dependency Graph

```
Phase 1 (quit fix) ──► Phase 2 (annotations) ──► Phase 3 (palette)
                                              │
                                              └─► Phase 5 (selection)

Phase 4 (navigation) ──► independent (can run parallel with Phase 2)

Phase 6 (coverage) ────► after all features complete
```

**Parallelizable:** Phase 2 + Phase 4 can run in parallel after Phase 1.

---

## Next Steps After This Plan

1. Search (`/`) with match highlighting
2. Annotations list view (`a` key)
3. Mouse support (Phase 1.5 in specs)
4. Selection expansion (ap, ab, as)
5. `fabbro resume <session-id>`
6. Config file and templates
