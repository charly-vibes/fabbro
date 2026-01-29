# AIX Best Practices for `fabbro`

**Date**: 2026-01-29

**Author**: Gemini

## 1. Guiding Principles for AI Experience (AIX)

The primary user of `fabbro`'s CLI is envisioned to be an AI agent. This requires a shift in CLI design from human-centric to agent-centric interaction. The core principles are:

- **Clarity and Precision**: Commands, flags, and outputs should be unambiguous.
- **Discoverability**: An agent should be able to determine command functionality and options programmatically.
- **Predictability**: The tool's behavior should be consistent.
- **Self-Healing**: Error messages should not just report failure, but guide the agent toward a solution.

## 2. Self-Healing Patterns

When a command fails, the error message is the primary mechanism for recovery. Errors should be structured and informative.

### Don't:

```
Error: invalid argument
```

### Do:

```json
{
  "error": "Invalid argument for --session-id",
  "code": "E_INVALID_ARG",
  "message": "The session ID '123' does not exist.",
  "suggestion": "Use 'fabbro session list' to see available session IDs. Or, create a new session with 'fabbro review'."
}
```

This structured JSON output allows an agent to parse the error, understand the context, and programmatically execute the suggested command to recover.

**Implementation Strategy:**
- Define a set of structured error codes (e.g., `E_NOT_INITIALIZED`, `E_SESSION_NOT_FOUND`, `E_FILE_NOT_FOUND`).
- Create a custom error handling function that intercepts errors, enriches them with a context-aware suggestion, and prints them as JSON to stderr.

## 3. AI-Focused Help Commands

The `--help` flag is the primary discovery mechanism for an agent. Help text should be verbose and structured, acting as a "man page" for the agent.

### Current `fabbro review --help`:

```
Start a review session

Usage:
  fabbro review [flags]

Flags:
  -h, --help   help for review
```

### Proposed `fabbro review --help`:

```
Command: review

Description:
  Starts a new interactive code review session in the terminal UI.
  This command is the primary entry point for annotating code.
  It requires the repository to be initialized with 'fabbro init'.

Usage:
  fabbro review [flags]

Flags:
  --file <path>         (Optional) The relative path to a specific file to open upon starting the session. If not provided, the TUI will start with no file open.
  --session-id <id>     (Optional) The ID of an existing session to resume. If not provided, a new session is created. Use 'fabbro session list' to find existing IDs.
  -h, --help              Show this help message.

Pre-conditions:
  - The current directory must be a git repository.
  - 'fabbro init' must have been run.

Post-conditions (on success):
  - A new terminal UI session is started.
  - A new session entry is created if --session-id was not provided.

Examples:
  # Start a new review session
  fabbro review

  # Start a new session and open a specific file
  fabbro review --file internal/tui/model.go

  # Resume an existing session
  fabbro review --session-id 20260129-100000
```

**Key Improvements:**
- **Verbose Descriptions**: Explicitly states what the command does, its purpose, and its context.
- **Explicit Optionality**: Clearly marks which flags are optional or required.
- **Pre- and Post-conditions**: Defines the expected state of the system before and after the command runs. This is critical for an agent to build reliable workflows.
- **Concrete Examples**: Provides copy-pasteable examples for common use cases.

**Implementation Strategy:**
- Use the `Long` and `Example` fields in the `cobra.Command` struct to provide this detailed information.
- Standardize the format across all commands for consistency.
