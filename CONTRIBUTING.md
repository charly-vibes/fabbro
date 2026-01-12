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

## Task Tracking

We use [beads](https://github.com/charly-vibes/beads) for task tracking:

```bash
bd ready           # Show unblocked work
bd list            # All issues
bd show <id>       # Issue details
```
