# Testing Strategy: 98% Coverage in Subsecond

## Overview

This plan defines how fabbro achieves **98% code coverage** with a **subsecond unit test suite**, **tiered integration/fuzz testing**, and **unified local/CI workflow**.

**Constraints:**

| Constraint | Target |
|------------|--------|
| Unit test coverage | ≥98% |
| Unit test duration | <1 second |
| Full test suite (all tiers) | <15 seconds |
| Workflow parity | CI runs same commands as local |
| Pre-push gate | Tests must pass to push |

## Related

- Implementation: `plans/2026-01-11-tracer-bullet.md`
- Design: `research/2026-01-09-fabbro-design-document.md`

---

## Architecture for Testability

### Package Layout

```
cmd/fabbro/           # Thin shell: Cobra wiring only, no logic
internal/
├── core/             # Orchestration (pure functions over interfaces)
├── parser/           # FEM parser (pure functions, no I/O)
├── session/          # Session management (I/O behind interfaces)
├── tui/              # Bubbletea model/update/view (pure)
└── cliio/            # I/O abstractions (filesystem, env, time)
```

### Core Principles

1. **Logic in internal packages** — cmd/ contains no business logic
2. **I/O at the edges** — All side effects behind interfaces
3. **Pure Update/View** — Bubbletea model returns `tea.Cmd`, never calls I/O directly
4. **Dependency injection** — All packages accept interfaces, tests inject fakes

---

## Test Tiers

All tiers are **required**, running at different frequencies:

| Tier | Tag | When | Duration | Purpose |
|------|-----|------|----------|---------|
| 1: Unit | (default) | Pre-push, CI | <1s | Core logic, 98% coverage |
| 2: Integration | `integration` | CI, `just test-all` | <10s | Full TUI, real files, CLI flows |
| 3: Property | `fuzz` | CI, `just test-all` | <5s | Fuzzing, edge cases |

**Total budget: <15 seconds** for all tiers combined.

### Tier 1: Unit Tests (Default)

Fast, pure-function tests. No build tags required.

```go
func TestParse_Comment(t *testing.T) {
    tests := []struct {
        name  string
        input string
        want  []Annotation
    }{
        {"simple", "{>> hello <<}", []Annotation{{Type: "comment", Text: "hello"}}},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, _, _ := Parse(tt.input)
            // assert
        })
    }
}
```

### Tier 2: Integration Tests

Real filesystem, full TUI program, end-to-end CLI flows.

```go
//go:build integration

package integration_test

func TestFullReviewFlow(t *testing.T) {
    dir := t.TempDir()
    
    // Init
    runCLI(t, dir, "init")
    
    // Create session from stdin
    runCLIWithStdin(t, dir, "review --stdin", "test content")
    
    // Verify session file exists
    sessions := glob(t, filepath.Join(dir, ".fabbro/sessions/*.fem"))
    if len(sessions) != 1 {
        t.Fatalf("expected 1 session, got %d", len(sessions))
    }
}

func TestTUIProgram(t *testing.T) {
    m := tui.NewModel(realDeps(), initialState{})
    p := tea.NewProgram(m, tea.WithoutRenderer())
    
    // Send quit message, verify clean exit
    go func() {
        p.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
    }()
    
    if _, err := p.Run(); err != nil {
        t.Fatalf("program error: %v", err)
    }
}
```

### Tier 3: Property-Based Tests (Fuzzing)

Go native fuzzing for parser edge cases.

```go
//go:build fuzz

package parser_test

import "testing"

func FuzzParse(f *testing.F) {
    // Seed corpus
    f.Add("{>> comment <<}")
    f.Add("{-- delete --}...{--/--}")
    f.Add("plain text with no markup")
    f.Add("{>> nested {>> bad <<} <<}")
    
    f.Fuzz(func(t *testing.T, input string) {
        // Should never panic
        annotations, clean, err := Parse(input)
        
        // Invariants that must hold
        if err == nil {
            // Clean content should not contain markers
            if strings.Contains(clean, "{>>") {
                t.Errorf("clean output contains marker: %q", clean)
            }
        }
    })
}
```

---

## Unified Workflow: Justfile

**Critical requirement**: CI and local use the exact same commands.

### justfile

```justfile
# fabbro justfile - unified local/CI workflow

set shell := ["bash", "-uc"]

# Default: run fast unit tests
default: test

# === Test Commands ===

# Run unit tests (Tier 1) - pre-push gate
test:
    go test ./... -race -count=1

# Run unit tests with coverage
test-cover:
    go test ./... -race -coverprofile=coverage.out -coverpkg=./...
    go tool cover -func=coverage.out

# Run integration tests (Tier 2)
test-integration:
    go test ./... -race -tags=integration -count=1

# Run fuzz tests (Tier 3) - short mode for CI
test-fuzz:
    go test ./... -tags=fuzz -fuzz=. -fuzztime=3s

# Run ALL test tiers
test-all: test test-integration test-fuzz
    @echo "All test tiers passed"

# === Coverage Commands ===

# Check coverage meets 98% threshold
check-coverage: test-cover
    #!/usr/bin/env bash
    set -euo pipefail
    pct=$(go tool cover -func=coverage.out | grep '^total:' | awk '{print substr($3, 1, length($3)-1)}')
    echo "Total coverage: ${pct}%"
    if (( $(echo "$pct < 98.0" | bc -l) )); then
        echo "❌ Coverage ${pct}% is below required 98%"
        exit 1
    fi
    echo "✅ Coverage ${pct}% meets threshold"

# Generate HTML coverage report
cover-html: test-cover
    go tool cover -html=coverage.out -o coverage.html
    @echo "Open coverage.html in browser"

# === Lint Commands ===

# Run all linters
lint:
    go vet ./...
    @command -v staticcheck >/dev/null && staticcheck ./... || echo "staticcheck not installed, skipping"

# === Build Commands ===

# Build the binary
build:
    go build -o bin/fabbro ./cmd/fabbro

# === CI Commands ===

# Full CI pipeline (what GitHub Actions runs)
ci: lint test-all check-coverage build
    @echo "✅ CI pipeline passed"

# Pre-push checks (fast gate)
pre-push: lint test
    @echo "✅ Pre-push checks passed"

# === Setup Commands ===

# Setup development environment
setup:
    go mod download
    go install honnef.co/go/tools/cmd/staticcheck@latest
    @echo "Installing lefthook..."
    go install github.com/evilmartians/lefthook@latest
    lefthook install
    @echo "✅ Development environment ready"

# Symlink .agents/commands to .claude/commands for Claude compatibility
setup-claude:
    mkdir -p .claude
    ln -sfn ../.agents/commands .claude/commands
    @echo "Symlinked .agents/commands → .claude/commands"
```

---

## Pre-Push Hook: Lefthook

Lefthook enforces the test gate before pushing. Bypass with `--no-verify` only for WIP commits.

### lefthook.yml

```yaml
# lefthook.yml - pre-commit/pre-push hooks

pre-commit:
  parallel: true
  commands:
    lint:
      glob: "*.go"
      run: go vet ./...
    format-check:
      glob: "*.go"
      run: test -z "$(gofmt -l .)"

pre-push:
  commands:
    test:
      run: just pre-push
      fail_text: |
        ❌ Tests failed. Push blocked.
        
        To bypass (WIP only): git push --no-verify
        
        ⚠️  --no-verify requires justification:
        - Commit message MUST contain "WIP" or "wip"
        - You MUST fix tests before merging
```

### .lefthook-local.yml (optional, gitignored)

For personal overrides:

```yaml
# Override for slower machines
pre-push:
  commands:
    test:
      run: just test  # Skip integration tests locally
```

---

## CI Workflow: GitHub Actions

Uses the **same justfile commands** as local development.

### .github/workflows/ci.yml

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

env:
  GO_VERSION: '1.22'

jobs:
  ci:
    runs-on: ubuntu-latest
    timeout-minutes: 5

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Install just
        uses: extractions/setup-just@v2

      - name: Install tools
        run: |
          go install honnef.co/go/tools/cmd/staticcheck@latest

      - name: Run CI pipeline
        run: just ci

      - name: Upload coverage
        uses: actions/upload-artifact@v4
        with:
          name: coverage
          path: |
            coverage.out
            coverage.html
```

---

## Workflow Parity Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                      Developer Workflow                      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Local Development              CI (GitHub Actions)          │
│  ──────────────────            ─────────────────────        │
│                                                              │
│  $ just test                   - name: Run CI pipeline       │
│  $ just lint                     run: just ci                │
│  $ just check-coverage                                       │
│  $ just ci        ─────────────────────────────────────►     │
│                          (same commands)                     │
│                                                              │
│  Pre-push hook (lefthook)                                    │
│  ─────────────────────────                                   │
│  $ git push                                                  │
│    └── lefthook runs: just pre-push                         │
│        └── Blocks if tests fail                             │
│                                                              │
│  Bypass (WIP only):                                          │
│  $ git push --no-verify                                     │
│    └── Commit message must contain "WIP"                    │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Test Categories (Detailed)

### 1. Parser Tests (internal/parser)

**Tier 1 - Unit:**

```go
func TestParse_AllAnnotationTypes(t *testing.T) {
    tests := []struct {
        name  string
        input string
        want  AnnotationType
    }{
        {"comment", "{>> note <<}", Comment},
        {"delete", "{-- remove --}text{--/--}", Delete},
        {"expand", "{!! more !!}", Expand},
        {"unclear", "{~~ confusing ~~}", Unclear},
        {"keep", "{== good ==}", Keep},
        {"question", "{?? why ??}", Question},
    }
    // ...
}
```

**Tier 3 - Fuzz:**

```go
//go:build fuzz

func FuzzParse(f *testing.F) {
    f.Fuzz(func(t *testing.T, input string) {
        _, _, _ = Parse(input)  // Must not panic
    })
}
```

### 2. Session Tests (internal/session)

**Tier 1 - Unit (in-memory store):**

```go
func TestSession_CreateLoad(t *testing.T) {
    store := NewMemoryStore()
    // ...
}
```

**Tier 2 - Integration (real filesystem):**

```go
//go:build integration

func TestSession_FileStore(t *testing.T) {
    dir := t.TempDir()
    store := NewFileStore(dir)
    // ...
}
```

### 3. TUI Tests (internal/tui)

**Tier 1 - Unit (model only):**

```go
func TestModel_Navigation(t *testing.T) {
    m := NewModel(fakeDeps{}, state{lines: 10})
    
    m, _ = m.Update(key('j'))
    if m.cursor != 1 { t.Fatal("expected cursor at 1") }
    
    m, _ = m.Update(key('k'))
    if m.cursor != 0 { t.Fatal("expected cursor at 0") }
}
```

**Tier 2 - Integration (program):**

```go
//go:build integration

func TestTUI_FullProgram(t *testing.T) {
    p := tea.NewProgram(NewModel(realDeps{}, state{}), tea.WithoutRenderer())
    // ...
}
```

### 4. CLI Tests (cmd/fabbro)

**Tier 1 - Unit (command parsing):**

```go
func TestInitCommand_Flags(t *testing.T) {
    cmd := NewRootCmd()
    cmd.SetArgs([]string{"init", "--quiet"})
    // ...
}
```

**Tier 2 - Integration (full flow):**

```go
//go:build integration

func TestCLI_InitReviewApply(t *testing.T) {
    dir := t.TempDir()
    
    // Full user journey
    run(t, dir, "init")
    run(t, dir, "review", "--stdin", "content")
    run(t, dir, "apply", "session-id", "--json")
}
```

---

## Coverage by Design

```go
// main.go pattern for coverage
func main() {
    os.Exit(realMain())
}

func realMain() int {
    if err := fabbroCLI(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
        fmt.Fprintln(os.Stderr, err)
        return 1
    }
    return 0
}
```

Tests call `fabbroCLI()` directly — the 2-line `main()` wrapper is the only uncovered code.

---

## Guardrails

### What to Avoid

| Anti-pattern | Why it hurts | Alternative |
|--------------|--------------|-------------|
| Snapshot tests of full TUI | Brittle, slow | Assert key substrings |
| Real file I/O in unit tests | Slow | In-memory fakes |
| `time.Sleep` in tests | Flaky, slow | Fake clock interface |
| Testing `main()` directly | Can't capture coverage | Test `realMain()` |
| Golden file comparisons | Slow disk I/O | String assertions |
| Long fuzz durations | Slow CI | 3s fuzztime in CI |

### Coverage Exceptions (2% Buffer)

- `main()` function (2 lines)
- Panic/fatal paths
- Platform-specific code blocks

---

## Metrics

Track in CI:

| Metric | Target | Alert |
|--------|--------|-------|
| Unit test duration | <1s | Fail if >2s |
| Full suite duration | <15s | Fail if >30s |
| Coverage | ≥98% | Fail if <98% |

---

## Branch Protection

Configure in GitHub Settings → Branches → main:

- ✓ Require status checks to pass before merging
- ✓ Required checks: `ci`
- ✓ Require branches to be up to date before merging
- ✓ Dismiss stale approvals on new commits

---

## Summary

| Goal | Implementation |
|------|----------------|
| 98% coverage | Pure-function design, `realMain()` pattern |
| <1s unit tests | No I/O, small fixtures, table-driven |
| <15s full suite | Tiered tests with build tags |
| Workflow parity | Justfile commands used by both local and CI |
| Pre-push gate | Lefthook runs `just pre-push` |
| WIP bypass | `--no-verify` with "WIP" in commit message |
