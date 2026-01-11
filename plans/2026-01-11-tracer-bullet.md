# Tracer Bullet Implementation Plan

## Overview

A tracer bullet implementation of fabbro that proves the end-to-end architecture works: from piped input → session creation → TUI annotation → FEM parsing → JSON output for LLM consumption.

**Goal**: Validate the core loop works, not polish. Ship the skeleton, then flesh it out.

## Related

- Design: `research/2026-01-09-fabbro-design-document.md`
- Evaluation: `research/2026-01-11-implementation-evaluation.md`
- Debate: `debates/2026-01-11-tui-vs-editor.md`
- Specs: `specs/01_initialization.feature`, `specs/02_review_session.feature`, `specs/06_fem_markup.feature`

## Current State

- No source code exists
- All specs written
- Design decisions made (Go CLI, bubbletea TUI, FEM markup)

## Desired End State

After this tracer bullet, you can run:

```bash
echo "Hello world" | fabbro review --stdin
# → TUI opens, you can add ONE type of annotation (comment)
# → Save and exit

fabbro apply <session-id> --json
# → Outputs JSON with annotations
```

This proves:
1. CLI scaffolding works (cobra)
2. Session storage works (.fabbro/sessions/*.fem)
3. TUI framework works (bubbletea)
4. FEM parser works (even with just one annotation type)
5. JSON output works (structured for LLM)

## Out of Scope

- Multiple annotation types (only `{>> comment <<}` for tracer)
- Templates
- Config file parsing
- `fabbro resume`
- Mouse support
- SPC command palette (use single keybind)
- Selection expansion (ap, ab, as)
- Search
- Any polish or styling

## Tracer Bullet Path

```
stdin → CLI → Session File → TUI → FEM in file → Parser → JSON out
         ↓           ↓          ↓                    ↓
       cobra     .fabbro/   bubbletea             regex
                 sessions/    basic               parse
```

---

## Phase 1: Project Setup (~30 min)

### Changes Required

**Create project structure:**
```
cmd/fabbro/
├── main.go           # Entry point, cobra root command
internal/
├── config/
│   └── config.go     # Minimal: just check .fabbro exists
├── session/
│   └── session.go    # Create/load session files
├── fem/
│   └── parser.go     # Parse FEM annotations
└── tui/
    └── tui.go        # Basic bubbletea app
go.mod
go.sum
```

**File: go.mod**
```go
module github.com/charly-vibes/fabbro

go 1.22

require (
    github.com/charmbracelet/bubbletea v1.3.2
    github.com/charmbracelet/lipgloss v1.0.0
    github.com/spf13/cobra v1.8.1
)
```

### Success Criteria

#### Automated:
- [x] `go mod tidy` succeeds
- [x] `go build ./...` succeeds

#### Manual:
- [x] Directory structure matches plan

---

## Phase 2: `fabbro init` Command (~45 min)

### Changes Required

**File: cmd/fabbro/main.go**
- Create root cobra command
- Add `init` subcommand
- Create `.fabbro/sessions/` directory

**File: internal/config/config.go**
- `IsInitialized() bool` - check if `.fabbro` exists
- `Init() error` - create `.fabbro/sessions/`

**Minimum implementation:**
- No config.yaml (defer)
- No templates directory (defer)
- No .gitignore (defer)

### Success Criteria

#### Automated:
- [x] `go test ./...` passes (write one test)

#### Manual:
- [x] `go run ./cmd/fabbro init` creates `.fabbro/sessions/`
- [x] Running again prints "already initialized"

---

## Phase 3: `fabbro review --stdin` Command (~1 hour)

### Changes Required

**File: cmd/fabbro/main.go**
- Add `review` subcommand with `--stdin` flag
- Check initialization, error if not
- Read stdin, create session, launch TUI

**File: internal/session/session.go**
- `Session` struct: `ID`, `Content`, `CreatedAt`
- `Create(content string) (*Session, error)` - generate ID, write .fem file
- `Load(id string) (*Session, error)` - read .fem file

**Session file format (minimal):**
```
---
session_id: abc123
created_at: 2026-01-11T10:00:00Z
---

Original content here...
```

### Success Criteria

#### Automated:
- [x] `go test ./...` passes

#### Manual:
- [x] `echo "test" | go run ./cmd/fabbro review --stdin` creates session file
- [x] Session file exists in `.fabbro/sessions/`

---

## Phase 4: Minimal TUI (~2 hours)

### Changes Required

**File: internal/tui/tui.go**

Bubbletea model with:
- Viewport showing content with line numbers
- `j/k` navigation
- `v` to toggle line selection (single line only for tracer)
- `c` to add comment (opens text input at bottom)
- `w` to save and exit
- `q` to quit without saving

**Minimal UI:**
```
─── Review: abc123 ───────────────────────────────────
  1 │ First line of content
  2 │ Second line                          ← SELECTED
  3 │ Third line
──────────────────────────────────────────────────────
[v]select [c]omment [w]rite [q]uit
```

When `c` pressed with line selected:
```
─── Review: abc123 ───────────────────────────────────
  1 │ First line of content
  2 │ Second line                          ← SELECTED
  3 │ Third line
──────────────────────────────────────────────────────
Comment: _                                    [Enter]
```

**Model state:**
- `content []string` - lines
- `cursor int` - current line
- `selected int` - selected line (-1 if none)
- `annotations []Annotation` - collected annotations
- `mode string` - "normal" | "input"
- `input string` - current input text

**On save:**
- Insert FEM comment markup at selected line
- Write back to session file

### Success Criteria

#### Automated:
- [x] `go test ./...` passes

#### Manual:
- [ ] TUI displays content with line numbers
- [ ] Can navigate with j/k
- [ ] Can select line with v
- [ ] Can add comment with c → type → Enter
- [ ] Can save with w
- [ ] Session file now contains FEM markup

---

## Phase 5: FEM Parser (~1 hour)

### Changes Required

**File: internal/fem/parser.go**

```go
type Annotation struct {
    Type    string // "comment" for tracer
    Text    string
    Line    int
}

func Parse(content string) ([]Annotation, string, error)
```

**Parsing strategy (regex-based for tracer):**
- Find `{>> ... <<}` patterns
- Extract text between markers
- Determine line number from position
- Return clean content (markers stripped)

### Success Criteria

#### Automated:
- [x] Test: `{>> hello <<}` extracts annotation with text "hello"
- [x] Test: multiple annotations on different lines

#### Manual:
- [ ] Can parse the session file created by TUI

---

## Phase 6: `fabbro apply --json` Command (~45 min)

### Changes Required

**File: cmd/fabbro/main.go**
- Add `apply` subcommand with `--json` flag
- Load session, parse FEM, output JSON

**JSON output format:**
```json
{
  "session_id": "abc123",
  "annotations": [
    {
      "type": "comment",
      "text": "This needs clarification",
      "line": 2
    }
  ]
}
```

### Success Criteria

#### Automated:
- [x] `go test ./...` passes

#### Manual:
- [x] `fabbro apply abc123 --json` outputs valid JSON
- [x] Annotations appear in output

---

## Phase 7: Integration Test (~30 min)

### Changes Required

**File: cmd/fabbro/main_test.go**

End-to-end test (may need to mock TUI for automation):

```go
func TestTracerBullet(t *testing.T) {
    // 1. Init
    // 2. Create session with known content
    // 3. Manually add FEM markup to session file
    // 4. Run apply --json
    // 5. Verify JSON output
}
```

### Success Criteria

#### Automated:
- [x] Integration test passes
- [x] `go build ./cmd/fabbro` produces working binary

#### Manual:
- [ ] Full flow works: init → review → annotate → apply --json

---

## Testing Strategy

**Tracer bullet philosophy: minimal tests, maximum coverage of the path.**

1. Unit test for FEM parser (this is the riskiest part)
2. Integration test for init → apply flow (skip TUI interaction)
3. Manual testing for TUI (bubbletea testing is complex, defer for now)

---

## Summary

| Phase | Component | Deliverable | Time |
|-------|-----------|-------------|------|
| 1 | Setup | go.mod, structure | 30m |
| 2 | Init | `fabbro init` | 45m |
| 3 | Session | `fabbro review --stdin` | 1h |
| 4 | TUI | Basic navigation + annotation | 2h |
| 5 | Parser | FEM → Annotations | 1h |
| 6 | Apply | JSON output | 45m |
| 7 | Test | Integration test | 30m |

**Total estimated time: ~6-7 hours**

After the tracer bullet:
- All packages exist and integrate
- Core workflow is proven
- Can iterate on features (more annotation types, polish, etc.)

---

## Next Steps After Tracer

1. Add remaining annotation types (`{-- delete --}`, `{!! expand !!}`, etc.)
2. Add SPC command palette
3. Add selection mode (multi-line)
4. Add config.yaml parsing
5. Add templates
6. Polish TUI styling with lipgloss
