# Contributing to fabbro

## Development Philosophy

fabbro follows these core principles:

1. **Local First** — No server components, works offline
2. **Tidy First** — Small refactorings before new features (see Kent Beck's "Tidy First?")
3. **Spec-Driven Development** — Features start with Gherkin specs in `specs/`
4. **Test-Driven Development** — Red → Green → Refactor

## Getting Started

```bash
# Clone the repo
git clone https://github.com/charly-vibes/fabbro.git
cd fabbro

# Install dependencies
go mod download

# Run tests
just test

# Run full CI
just ci
```

## Project Structure

```
fabbro/
├── cmd/fabbro/      # CLI entry point
├── internal/        # Internal packages
│   ├── config/      # Initialization and config
│   ├── fem/         # FEM parser
│   ├── session/     # Session management
│   └── tui/         # Terminal UI
├── specs/           # Gherkin feature specs
├── plans/           # Implementation plans
├── research/        # Research documents
├── handoffs/        # Session handoff docs
└── docs/            # User documentation
```

## Development Workflow

### 1. Create a Spec

Before implementing, write a Gherkin spec in `specs/`:

```gherkin
Feature: New feature name

  Scenario: Basic usage
    Given fabbro is initialized
    When I run the new command
    Then it should produce expected output
```

### 2. Write Failing Tests

Implement tests that exercise the spec scenarios. Run tests and confirm they fail.

### 3. Implement

Write the minimal code to make tests pass.

### 4. Refactor

Clean up while keeping tests green.

### 5. Commit

Follow conventional commits:

```bash
# Tidy First refactorings (separate commits)
git commit -m "refactor: extract parser into dedicated package"

# Features
git commit -m "feat(tui): add Helix-style command palette"

# Bug fixes
git commit -m "fix(fem): handle unclosed annotation markers"

# Tests
git commit -m "test(session): add error path coverage"
```

## Just Commands

```bash
just test          # Run unit tests
just test-all      # Run all test tiers
just ci            # Full CI pipeline
just lint          # Run linters
just fmt           # Format code
just build         # Build binary
just help          # Show all commands
```

## Code Style

- Follow standard Go conventions
- Run `just fmt` before committing
- No comments unless code is genuinely complex
- Error messages should be actionable

## Testing

- **Unit tests** — Fast, isolated, in `*_test.go` files
- **Integration tests** — Tagged with `//go:build integration`
- **Coverage threshold** — 65% minimum

## Pull Requests

1. Create a branch from `main`
2. Make focused, atomic commits
3. Ensure `just ci` passes
4. Open PR with clear description
5. Address review feedback

## Specifications

All features are defined as [Gherkin](https://cucumber.io/docs/gherkin/) `.feature` files in `specs/`. These specs are the **source of truth** for how fabbro behaves — they serve as living documentation, drive TDD, and define acceptance criteria.

### Spec inventory

| File | Feature | Description |
|------|---------|-------------|
| `01_initialization.feature` | Project Initialization | `fabbro init`, directory scaffolding, agent integration |
| `02_review_session.feature` | Review Session Creation | Creating sessions from stdin/files, metadata, error handling |
| `03_tui_interaction.feature` | TUI Interaction | Navigation, selection, annotations, command palette, inline editing |
| `04_apply_feedback.feature` | Apply Feedback | Extracting annotations as human-readable or JSON output |
| `05_session_management.feature` | Session Management | Listing, resuming, deleting, and cleaning sessions |
| `06_fem_markup.feature` | FEM Markup Language | FEM syntax reference, parsing rules, edge cases |
| `07_web_notes_sidebar.feature` | Web Notes Sidebar | Web UI annotation sidebar, navigation, deletion |
| `08_web_search.feature` | Web Incremental Search | `/`-triggered search with incremental highlighting |

### Implementation status tags

Every scenario is tagged with its current status:

| Tag | Meaning |
|-----|---------|
| `@implemented` | Working in the current build |
| `@partial` | Core behavior works, some aspects still missing |
| `@planned` | Designed but not yet implemented |

When you implement a scenario, update its tag from `@planned` to `@implemented` in the same commit.

### Writing specs

- Use user-centric language: _"As a user, I want to..."_
- One scenario per behavior — keep them atomic
- Include error and edge-case scenarios alongside happy paths
- Add comments for implementation notes (e.g., `# Block markers {--/--} not yet implemented`)

## Task Tracking

We use [beads](https://github.com/steveyegge/beads) for task tracking. Beads is an AI-native, git-native issue tracker that lives inside the repository (`.beads/`), syncs via git, and works entirely from the CLI — no web UI required.

### Why beads?

- **Git-native** — Issues are stored in `.beads/issues.jsonl` and travel with the repo
- **AI-friendly** — CLI-first design integrates with AI coding agents (Amp, Claude Code)
- **Offline-first** — Works without network; syncs when you push
- **Dependency-aware** — Issues can declare dependencies so `bd ready` shows only unblocked work

### Essential commands

```bash
bd ready                              # Show unblocked issues (start here)
bd list                               # All issues
bd show <id>                          # Issue details with dependencies
bd create "Title of the issue"        # Create a new issue
bd update <id> --status=in_progress   # Claim work
bd close <id>                         # Mark complete
bd sync                               # Sync issues with git remote
```

### Typical workflow

1. Run `bd ready` to see what's available to work on
2. Pick an issue and `bd update <id> --status=in_progress`
3. Implement following TDD (see [Development Workflow](#development-workflow) above)
4. When done, `bd close <id>` and commit your code
5. Run `bd sync` before pushing to keep issues in sync across clones
