# AGENTS.md - Fabbro Development Workflow

This document outlines the development philosophy and process for building `fabbro`. It is designed to ensure high-quality, maintainable code through a structured, test-driven methodology.

## Project Structure

```
fabbro/
├── specs/           # Gherkin .feature files (living documentation)
├── research/        # Research documents (YYYY-MM-DD-topic.md)
├── plans/           # Implementation plans (YYYY-MM-DD-description.md)
├── handoffs/        # Session handoff documents for context transfer
├── debates/         # Design debates and decision records
├── .agents/commands/  # Agent workflow commands
└── .claude/commands/  # Claude Code slash commands (symlinked)
```

## Running Locally

Use `just` commands to build and test fabbro during development:

```bash
# Build and run with arguments (fastest iteration)
just run init
just run review --stdin < file.go

# Install to ~/go/bin for testing from any directory
just install

# Run tests
just test
```

## Available Commands

Commands are available as slash commands in Claude Code:

- `/create_plan` - Create implementation plans → outputs to `plans/`
- `/implement_plan` - Execute approved plans following TDD
- `/research_codebase` - Document codebase as-is → outputs to `research/`
- `/commit` - Create git commits with user approval
- `/create_handoff` - Create handoff document → outputs to `handoffs/`
- `/resume_handoff` - Resume work from a handoff document

## Core Philosophy

1.  **Local First**: `fabbro` is a tool for individual developers on their local machine. The architecture must prioritize simplicity, reliability, and offline-first functionality. There are no server components.

2.  **Tidy First**: We follow Kent Beck's "Tidy First?" approach rigorously:
    - **Before adding new code**, look for small structural improvements (tidyings)
    - Tidyings are tiny, safe refactorings: rename, extract, inline, reorder
    - Each tidying is a separate commit (prefixed with `refactor:`)
    - Tidyings make the subsequent behavior change easier to write and review
    - If you can't tidy first, note it and proceed—but prefer tidying
    - A clean workspace is a productive workspace

3.  **Spec-Driven Development (SDD)**: All new functionality begins with a specification. We use Gherkin (`.feature` files) to describe how a feature should behave from the user's perspective. These specs are human-readable, serve as living documentation, and form the foundation of our test suite.

4.  **Test-Driven Development (TDD)**: The specs are implemented as automated tests *before* the feature code is written. The development cycle is "Red, Green, Refactor":
    *   **Red**: Write a failing test that implements a single scenario from the spec.
    *   **Green**: Write the simplest possible production code to make the test pass.
    *   **Refactor**: Clean up the production and test code while keeping the test green.

## Agent/Developer Workflow

All contributions to `fabbro` must follow this process:

1.  **Create/Update a Spec File**: In the `specs/` directory, create or modify a `.feature` file that describes the desired functionality. Use clear, user-centric Gherkin syntax.
    *   *Example*: `specs/01_initialization.feature`

2.  **Implement the Failing Test**: Write the test code that executes the scenario defined in the spec. Run the test suite and confirm that it fails for the expected reason.

3.  **Write Production Code**: Implement the feature, focusing only on what is necessary to make the failing test pass.

4.  **Run Tests**: Run the test suite again and confirm that all tests now pass.

5.  **Refactor**: With passing tests as a safety net, refactor the code for clarity, efficiency, and adherence to style guidelines. Re-run tests to ensure nothing was broken.

6.  **Repeat**: Continue this cycle for all scenarios in the spec file. Once all scenarios for a feature are implemented, the feature is considered complete.

This structured approach ensures that `fabbro` is built on a solid foundation of clear specifications and comprehensive tests, making it robust and easy to maintain.

## Commit Workflow

**NEVER commit without explicit user approval.** Before running any `git add` or `git commit`:

1. Run `git status` and `git diff` to understand all changes
2. **Check if documentation needs updating:**
   - New CLI commands → update README.md usage section
   - New keybindings → update docs/keybindings.md
   - Changed behavior → update relevant spec files (@planned → @implemented)
3. Present a clear commit plan to the user:
   ```
   I plan to create [N] commit(s):

   Commit 1: [type]: [message]
   - file1.ext
   - file2.ext

   Shall I proceed?
   ```
4. Wait for user confirmation before executing
5. Use `git add` with specific files (avoid `-A` or `.`)

See `.agents/commands/commit.md` for the full commit process.

## Conventional Commits

All commits MUST follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Commit Types

| Type | Description |
|------|-------------|
| `feat:` | New feature for the user |
| `fix:` | Bug fix |
| `refactor:` | Code restructuring (Tidy First commits) |
| `test:` | Adding or updating tests |
| `docs:` | Documentation changes |
| `chore:` | Build, CI, tooling changes |
| `style:` | Formatting, whitespace (no code change) |

### Examples

```bash
# Tidy First refactoring (always separate commits)
refactor: extract FEM parser into dedicated package
refactor: rename Session to ReviewSession for clarity

# Feature implementation
feat(tui): add Helix-style SPC command palette
feat(cli): implement fabbro init command

# Bug fixes
fix(fem): handle unclosed annotation markers gracefully

# Tests
test(init): add scenarios for idempotent initialization

# Documentation
docs: add FEM syntax reference to README
```

### Rules

1. **Tidyings get `refactor:` prefix** - Always separate from feature commits
2. **One logical change per commit** - Atomic, focused commits
3. **Imperative mood** - "Add feature" not "Added feature"
4. **No period at end** of subject line
5. **72 character limit** on subject line

## Landing the Plane (Session Completion)

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   bd sync
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds


# Task Tracking with Beads

Use `bd` for task tracking. Issues persist across sessions and sync with git.

## Essential Commands

```bash
bd ready                    # Show unblocked work
bd list                     # All issues
bd show <id>                # Issue details with dependencies
bd update <id> --status=in_progress  # Claim work
bd close <id>               # Mark complete
bd sync                     # Sync with git remote
```

## Subagent Coordination

When implementing plans with multiple phases, use the Amp `Task` tool to spawn subagents:

### Sequential Phases (with dependencies)
For phases that depend on each other (Phase 2 depends on Phase 1):
- Work sequentially in the main agent
- Complete and close each beads issue before starting the next
- Use `bd ready` to see what's unblocked

### Parallel Work (independent tasks)
For independent work within a phase:
- Spawn parallel subagents using Amp's `Task` tool
- Each subagent should work on a distinct file/component
- Never have multiple subagents edit the same file

### Subagent Best Practices

1. **Include full context** in the Task prompt:
   - Reference the plan file path
   - Specify which phase/section to implement
   - Include the beads issue ID to update
   - Specify success criteria from the plan

2. **Example subagent invocation:**
   ```
   Task: Implement Phase 3 session creation
   
   Plan: plans/2026-01-11-tracer-bullet.md (Phase 3)
   Issue: fabbro-wwh
   
   Requirements:
   - Create internal/session/session.go
   - Implement Session struct and Create/Load functions
   - Write tests first (TDD)
   - Run `go test ./...` to verify
   
   When complete:
   - Run `bd close fabbro-wwh`
   - Report what was created
   ```

3. **Handoff between subagents:**
   - Subagent closes its beads issue when done
   - Parent agent runs `bd ready` to see newly unblocked work
   - Next subagent picks up the next phase

## Plan-to-Issues Workflow

When a new plan is approved:

1. **Create issues for each phase:**
   ```bash
   bd create --title="Phase 1: Setup" --type=task
   bd create --title="Phase 2: Init command" --type=task
   # ... etc
   ```

2. **Add dependencies:**
   ```bash
   bd dep add <phase-2-id> <phase-1-id>  # Phase 2 depends on Phase 1
   ```

3. **Start implementation:**
   ```bash
   bd ready  # Shows Phase 1 (only unblocked item)
   ```
