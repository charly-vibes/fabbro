# fabbro Implementation Evaluation

**Date**: 2026-01-11  
**Status**: Research complete  
**Inputs**: Design document v1.0, existing specs, AGENTS.md workflow

---

## Current State

The repository is in early setup phase:
- **specs/**: One Gherkin spec (`01_initialization.feature`) defining `fabbro init`
- **research/**: Design document v1.0 (comprehensive, 2000+ lines)
- **plans/**: Empty (ready for implementation plans)
- **debates/**: Empty
- **No source code yet**

---

## Key Recommendations

### 1. Language Choice: Go CLI + npm Wrapper

**Recommendation**: Implement core in Go, with a thin npm wrapper for Claude Code Web.

| Approach | Pros | Cons |
|----------|------|------|
| **Go CLI (recommended)** | Single binary, fast startup, easy cross-compile, mirrors beads | Requires Go toolchain |
| TypeScript core | Native npm, familiar to JS devs | Node runtime dependency, slower startup, complex packaging for single binary |

The design doc mentions both Go and npm installation - this is solved by:
- Go binary as source of truth
- npm package that downloads/launches the binary (like esbuild, ripgrep wrappers)

### 2. Phase 1 MVP Scope

Focus on **one complete workflow** (Workflow 2: Long Async Edits):

```
fabbro init → fabbro review --stdin → $EDITOR → fabbro apply <session-id> --json
```

**Phase 1 deliverables**:
1. `fabbro init` - Creates `.fabbro/{sessions,templates,config.yaml,.gitignore}`
2. `fabbro review --stdin` - Creates session, opens editor
3. `fabbro apply <session-id> --json` - Parses FEM, outputs JSON
4. Minimal FEM subset: `{>> <<}`, `{-- --}`, `{!! !!}`, `{?? ??}`

**Explicitly defer**:
- `fabbro resume`
- Advanced templates
- SQLite caching
- Rich FEM features (nested blocks, attributes)

### 3. Testing Approach (TDD/SDD Alignment)

Per AGENTS.md, use **specs as living docs + standard Go tests**:

```go
// Implements specs/01_initialization.feature - "Initializing a new project"
func TestInit_NewProject(t *testing.T) {
    dir := t.TempDir()
    // ... run fabbro init, assert filesystem state
}
```

**Testing strategy**:
- Integration-style tests against CLI entry point
- Inject editor interface for testing (`NoopEditor`, `TestEditor`)
- Parse apply JSON via structs, not string comparison
- No Gherkin runner (godog) initially - manual mapping is fine for <10 specs

### 4. Storage Format

```
.fabbro/
├── config.yaml          # User config (editor override, etc.)
├── .gitignore           # Contains "sessions/"
├── sessions/            # One .fem file per review session
│   └── 20260111-123456-xy7z.fem
└── templates/           # Future: review templates
```

**Key decisions**:
- Sessions stored as plain text `.fem` files (git-friendly, transparent)
- No SQLite/JSONL index in Phase 1 (YAGNI - just scan `sessions/*.fem`)
- Config in YAML for human-friendliness

### 5. Editor Integration

Priority order:
1. `FABBRO_EDITOR`
2. `$VISUAL`
3. `$EDITOR`
4. OS default (nano/vi/notepad)

Implementation: `exec.Command` with blocking call, attach stdin/stdout/stderr.

---

## Recommended Project Structure

```
fabbro/
├── cmd/fabbro/          # CLI entry point
│   ├── main.go
│   └── main_test.go     # Integration tests
├── internal/
│   ├── config/          # .fabbro/config.yaml handling
│   ├── tui/             # Bubbletea TUI components
│   ├── fem/             # FEM parser
│   └── session/         # Session management
├── specs/               # Gherkin specs (living docs)
├── research/            # Research documents
├── plans/               # Implementation plans
└── go.mod
```

---

## Go TUI Library Stack

### Primary: Charm Ecosystem

The [Charm](https://charm.sh/) libraries provide a cohesive, well-maintained TUI stack:

| Library | Purpose | Stars | Use in fabbro |
|---------|---------|-------|---------------|
| **[bubbletea](https://github.com/charmbracelet/bubbletea)** | TUI framework (Elm architecture) | 38k | Core framework |
| **[bubbles](https://github.com/charmbracelet/bubbles)** | Pre-built components | 5k+ | Viewport, Help, Key bindings |
| **[lipgloss](https://github.com/charmbracelet/lipgloss)** | Styling/layout | 8k+ | Borders, colors, layout |
| **[huh](https://github.com/charmbracelet/huh)** | Forms/prompts | 6k+ | Comment input prompts |

### Supporting Libraries

| Library | Purpose | Use in fabbro |
|---------|---------|---------------|
| **[bubblezone](https://github.com/lrstanley/bubblezone)** | Mouse event tracking | Click-to-select, zone detection |

### Component Mapping

| fabbro Feature | Library/Component |
|----------------|-------------------|
| Document viewport | `bubbles/viewport` |
| Helix-style SPC menu | Custom bubble + `bubbles/list` |
| Help bar | `bubbles/help` |
| Key bindings | `bubbles/key` |
| Comment input | `huh` form fields or `bubbles/textinput` |
| Mouse click zones | `bubblezone` |
| Styling | `lipgloss` |

### UX Implementation Details

**Helix-style SPC Menu**:
- Build as custom bubble component
- On `SPC` keypress, render overlay with `lipgloss`
- Use `bubbles/list` for menu items
- Each item maps to annotation type

**Mouse Support (Phase 1.5)**:
- Use `bubblezone` to mark selectable regions
- `tea.WithMouseCellMotion()` for mouse events
- Click-drag selection via tracking mouse down/up positions

**Vim-surround Context Expansion**:
- Implement as custom key handlers
- Parse content for paragraph/block/section boundaries
- `ap` = expand to blank-line-delimited paragraph
- `ab` = expand to fenced code block (``` boundaries)
- `as` = expand to markdown heading section

---

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| FEM grammar creep | Version FEM explicitly, start with 4-5 tags only |
| Editor incompatibilities | Fail fast with clear errors, later add `--no-editor` flag |
| npm wrapper drift | Keep npm as binary-fetcher only, no logic |
| Overcomplicated testing | Stick to Go tests, revisit BDD runner if >20 specs |

---

## Next Steps

1. **Create implementation plan** for Phase 1 via `/create_plan`
2. **Spec out remaining features** - Add Gherkin specs for `review` and `apply`
3. **Initialize Go module** - `go mod init github.com/charly-vibes/fabbro`
4. **TDD cycle**: Red → Green → Refactor per spec scenario

---

## Summary

The design document is comprehensive and well-thought-out. The beads-inspired architecture (CLI-first, git-backed, session hooks) is solid. The recommended path forward:

- **Go CLI** as core implementation
- **Phase 1 MVP** targeting the complete "long async edit" workflow
- **Standard Go tests** mapping to Gherkin specs
- **Plain text `.fem` sessions** with minimal indexing

Estimated effort for Phase 1: **L (2-4 days focused work)** to reach a working end-to-end flow with TUI.

---

## Dependencies Summary

```go
// go.mod
module github.com/charly-vibes/fabbro

go 1.21

require (
    github.com/charmbracelet/bubbletea v1.3+
    github.com/charmbracelet/bubbles v0.20+
    github.com/charmbracelet/lipgloss v1.0+
    github.com/charmbracelet/huh v0.6+
    github.com/lrstanley/bubblezone v1.0+
    github.com/spf13/cobra v1.8+          // CLI framework
    gopkg.in/yaml.v3 v3.0+                // Config parsing
)
```
