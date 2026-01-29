# AIX and Desire Path Research

**Date:** 2026-01-29
**Question:** What are the best practices for AIX (AI Experience) for the `fabbro` tool, focusing on self-healing patterns, clear error messages, and AI-focused `--help` commands, viewed through the lens of the 'Desire Path' philosophy?

## Summary

The `fabbro` CLI is built with `cobra`, a conventional choice providing a standard command structure. Its current help text is concise and human-readable, and its error handling is functional, typically returning simple error strings. From an AIX (AI Experience) and "Desire Path" perspective, the current implementation presents an opportunity to provide more structured, machine-readable, and context-aware information. An AI agent's "desire path" is to receive unambiguous signals about command success or failure and clear guidance on how to proceed, which could be better supported by more detailed help text and structured error responses.

## Detailed Findings

### CLI Structure

**Location:** `cmd/fabbro/main.go`

**Purpose:** This file is the entry point for the `fabbro` executable. It defines the root command and all subcommands using the `github.com/spf13/cobra` library.

**Key files:**
- `main.go`: Contains the `main` function and builder functions (`build<CommandName>Cmd`) for all CLI commands.

**How it works:**
1. `main()` calls `realMain()`, which orchestrates command execution.
2. `buildRootCmd()` creates the main `fabbro` command and attaches all subcommands (`init`, `review`, `apply`, etc.).
3. Each `build<CommandName>Cmd()` function constructs a `cobra.Command` struct, defining its `Use`, `Short` description, arguments (`Args`), and execution logic (`RunE`).

#### Desire Path Analysis
- A human user can easily infer the purpose of `review [file]`. An AI agent, however, lacks the context. It might "desire" to know: Is the file argument optional? What happens if it's omitted?
- The clear `[command] [subcommand]` structure (e.g., `session list`) provides a predictable "paved road" for an agent to discover functionality.
- However, for arguments and flags, the current help text does not explicitly detail constraints and options, requiring the agent to learn through trial and error, a less efficient "desire path."

#### Desire Path Analysis
- The `fabbro not initialized` error successfully paves a "desire path" by suggesting the exact fix. This is a key principle of self-healing.
- The `must provide either session-id or --file` error is less helpful for an agent. The agent's "desire" is to know *how* to provide input. An ideal error would suggest valid flags or arguments.
- While error recovery is a key part of the desire path, the "happy path" (the most efficient workflow) is equally important. The current error messages do not explicitly guide an agent toward a more efficient sequence of commands.
- All errors are unstructured strings. An agent would benefit from structured errors (e.g., JSON with `{"code": "E_NOT_INITIALIZED", "message": "...", "suggestion": "..."}`) which can be programmatically parsed and acted upon.

#### Desire Path Analysis
- For an AI agent, the "happy path" workflow heavily relies on predictable, parsable output. The `--json` flag is the most important "paved road" in the entire CLI for this purpose. It is a deliberate and highly effective feature for AIX.
- The agent's "desire" is to always receive structured data. The presence of the `--json` flag on some but not all commands creates an inconsistent experience. The ideal desire path would be for all commands that return data to support a `--json` flag.

## Code References
| File | Lines | Description |
|------|-------|-------------|
| `cmd/fabbro/main.go` | 42-63 | `buildRootCmd` defines the top-level command and adds subcommands. |
| `cmd/fabbro/main.go` | 119-174| `buildReviewCmd` shows typical command construction, argument handling, and error wrapping. |
| `cmd/fabbro/main.go` | 200-201| Example of a `--json` flag being defined for the `apply` command. |
| `cmd/fabbro/main.go` | 254-260| Example of JSON output generation in the `apply` command. |
| `cmd/fabbro/main.go` | 280-281| Example of a `--json` flag being defined for the `session list` command. |

## Open Questions

- What is the intended level of AI agent autonomy? The answer would influence how much effort is put into creating structured, machine-readable output vs. human-readable text.
- Should all data-returning commands be required to implement a `--json` output option for a consistent AI experience?

## Conclusions from Open Questions

**Date:** 2026-01-29

Based on discussion, the following decisions have been made to guide the future development of `fabbro`'s AI Experience:

1.  **Level of Autonomy: Human-in-the-Loop.**
    - `fabbro` should be designed to support an AI agent that assists a human user.
    - This means that while machine-readable formats are important for the agent, clear, human-readable output must be maintained as a priority. The CLI should be effective for both direct human use and agent-assisted use.

2.  **`--json` Output Requirement: Enforced.**
    - For a consistent and predictable AI experience, all `fabbro` commands that return data must be required to implement a `--json` output option.
    - This provides a reliable "paved road" for the agent, fulfilling its "desire path" for structured data and simplifying its logic for interacting with the tool.
