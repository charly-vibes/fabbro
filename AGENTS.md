# AGENTS.md - Fabbro Development Workflow

This document outlines the development philosophy and process for building `fabbro`. It is designed to ensure high-quality, maintainable code through a structured, test-driven methodology.

## Project Structure

```
fabbro/
├── specs/           # Gherkin .feature files (living documentation)
├── research/        # Research documents (YYYY-MM-DD-topic.md)
├── plans/           # Implementation plans (YYYY-MM-DD-description.md)
├── debates/         # Design debates and decision records
├── .agents/commands/  # Agent workflow commands
└── .claude/commands/  # Claude Code slash commands (symlinked)
```

## Available Commands

Commands are available as slash commands in Claude Code:

- `/create_plan` - Create implementation plans → outputs to `plans/`
- `/implement_plan` - Execute approved plans following TDD
- `/research_codebase` - Document codebase as-is → outputs to `research/`

## Core Philosophy

1.  **Local First**: `fabbro` is a tool for individual developers on their local machine. The architecture must prioritize simplicity, reliability, and offline-first functionality. There are no server components.
2.  **Tidy First**: We adhere to the "Tidy First" principle. Before adding new functionality, we take the time for small refactorings to make the new code easier to write. A clean workspace is a productive workspace.
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


# Task tracking

Use 'bd' for task tracking
