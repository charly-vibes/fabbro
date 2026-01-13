# Tech Debt Cleanup Implementation Plan

## Overview

Address technical debt identified during code review (Rule of 5 methodology). Focus on DRY violations, silent error handling, and deprecated patterns.

## Related

- Code Review: Iterative review session 2026-01-12
- Specs: All specs remain valid; no behavioral changes

## Current State

1. **Duplicate Annotation struct**: `tui.Annotation` and `fem.Annotation` are identical
2. **Duplicate FEM markers**: `tui.markers` and `fem.patterns` define same markup
3. **Silent save failure**: `tui.save()` ignores `os.WriteFile` error
4. **Custom max()**: Redundant helper; Go 1.22 has builtin

## Desired End State

- Single source of truth for Annotation type and FEM markers
- User notified when save fails
- Cleaner code using Go 1.22+ builtins
- Coverage remains ≥93%
- All existing tests pass

## Out of Scope

- New features (search, help panel, etc.)
- Spec changes
- Gherkin executable tests (separate plan)

---

## Phase 1: Consolidate Annotation Type (~20 min)

### Problem

Two identical Annotation structs exist:

```go
// internal/fem/parser.go
type Annotation struct {
    Type string `json:"type"`
    Text string `json:"text"`
    Line int    `json:"line"`
}

// internal/tui/tui.go
type Annotation struct {
    Line int
    Type string
    Text string
}
```

### Changes Required

**File: internal/tui/tui.go**
- Remove local `Annotation` struct definition (lines 22-26)
- Import and use `fem.Annotation` instead
- Update all references from `Annotation` to `fem.Annotation`

**File: internal/tui/tui_test.go**
- Update test code to use `fem.Annotation`

### Success Criteria

#### Automated:
- [ ] `just test` passes
- [ ] `just ci` passes

#### Manual:
- [ ] `grep -r "type Annotation struct" internal/` shows only one definition

---

## Phase 2: Consolidate FEM Markers (~20 min)

### Problem

FEM markers are defined in two places:

```go
// internal/fem/parser.go - for parsing
var patterns = map[string]*regexp.Regexp{...}

// internal/tui/tui.go - for writing
var markers = map[string][2]string{...}
```

### Changes Required

**File: internal/fem/markers.go** (new file)
- Create shared marker definitions that both parser and TUI can use

```go
package fem

// AnnotationTypes lists all valid annotation types
var AnnotationTypes = []string{"comment", "delete", "question", "expand", "keep", "unclear"}

// Markers maps annotation type to opening and closing delimiters
var Markers = map[string][2]string{
    "comment":  {"{>> ", " <<}"},
    "delete":   {"{-- ", " --}"},
    "question": {"{?? ", " ??}"},
    "expand":   {"{!! ", " !!}"},
    "keep":     {"{== ", " ==}"},
    "unclear":  {"{~~ ", " ~~}"},
}
```

**File: internal/fem/parser.go**
- Derive regex patterns from `Markers` map (or keep separate, reference shared source)

**File: internal/tui/tui.go**
- Remove local `markers` map (lines 41-48)
- Import and use `fem.Markers`

### Success Criteria

#### Automated:
- [ ] `just test` passes
- [ ] `just ci` passes

#### Manual:
- [ ] `grep -r "markers\|Markers" internal/` shows single definition in fem package

---

## Phase 3: Handle save() Error (~15 min)

### Problem

`tui.save()` silently ignores WriteFile errors - user could lose work:

```go
// internal/tui/tui.go:316
os.WriteFile(sessionPath, []byte(fileContent), 0644)  // error ignored!
```

### Design Decision

Since TUI is exiting after save, we have options:
1. **Return error from save()** - requires changing quit flow
2. **Store error in model, display before quit** - complex
3. **Panic on save error** - too aggressive
4. **Log to stderr before exit** - simple, effective

Recommendation: **Option 4** - Print error to stderr. User sees the failure.

### Changes Required

**File: internal/tui/tui.go**
- Change `save()` signature to `save() error`
- Handle error in WriteFile call
- Update `handleNormalMode` "w" case to handle error (print to stderr, still quit)

```go
func (m Model) save() error {
    // ... existing code ...
    if err := os.WriteFile(sessionPath, []byte(fileContent), 0644); err != nil {
        return fmt.Errorf("failed to save session: %w", err)
    }
    return nil
}
```

In handleNormalMode:
```go
case "w":
    if err := m.save(); err != nil {
        fmt.Fprintln(os.Stderr, err)
    }
    return m, tea.Quit
```

**File: internal/tui/tui_test.go**
- Add test for save() returning error on invalid path

### Success Criteria

#### Automated:
- [ ] `just test` passes
- [ ] Test exists for save error handling

#### Manual:
- [ ] Create read-only directory, attempt save, see error printed

---

## Phase 4: Remove Custom max() (~5 min)

### Problem

```go
// internal/tui/tui.go:378-383
func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}
```

Go 1.22 has builtin `max()`.

### Changes Required

**File: internal/tui/tui.go**
- Delete custom `max()` function (lines 378-383)
- No import changes needed (builtin)

**File: internal/tui/tui_test.go**
- Remove `TestMax` test (if exists)

### Success Criteria

#### Automated:
- [ ] `just test` passes
- [ ] `just ci` passes

#### Manual:
- [ ] `grep "func max" internal/` returns nothing

---

## Phase 5: Final Verification (~10 min)

### Changes Required

None - verification only.

### Success Criteria

#### Automated:
- [ ] `just ci` passes with coverage ≥93%
- [ ] `go vet ./...` passes

#### Manual:
- [ ] Manual smoke test: `echo "test" | just run review --stdin`
- [ ] Verify annotations save correctly

---

## Summary

| Phase | Description | Time | Risk |
|-------|-------------|------|------|
| 1 | Consolidate Annotation type | 20m | Low |
| 2 | Consolidate FEM markers | 20m | Low |
| 3 | Handle save() error | 15m | Low |
| 4 | Remove custom max() | 5m | None |
| 5 | Final verification | 10m | None |

**Total: ~70 minutes**

## Testing Strategy

All changes are refactors - existing tests should continue to pass. Additional tests:
- Test save() error handling (Phase 3)

## Commit Strategy

Following conventional commits:
1. `refactor(tui): use fem.Annotation instead of duplicate struct`
2. `refactor(fem): extract shared marker definitions`
3. `fix(tui): handle save error and report to user`
4. `refactor(tui): remove redundant max() helper (Go 1.22 builtin)`

Each phase = one commit for easy review/revert.
