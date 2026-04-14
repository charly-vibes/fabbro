# Debate: Structured CLI errors and semantic exit codes

**Date:** 2026-04-14
**Status:** Proposal
**Related issue:** `fabbro-uje`

## Goal

Define a contract for machine-readable CLI errors to make fabbro more reliable for pi, Claude Code, and other agent integrations.

## Problem

Current failure behavior is simple:

- Exit code `1`
- Human-readable string on `stderr`

Agents must heuristically guess what went wrong (e.g., "was the ID missing, or was it a network error during fetch?").

## Proposal: Structured JSON errors

When the `--json` flag is active, fabbro should emit errors as structured JSON on `stderr`.

### Format

```json
{
  "error": "Short description",
  "message": "Detailed actionable message",
  "code": "semantic_error_code",
  "details": {
    "key": "value"
  }
}
```

### Proposed error codes

| Code | Meaning |
|------|---------|
| `usage_error` | Invalid flags or arguments |
| `not_found` | Resource (session, file) does not exist |
| `data_error` | Invalid input data (empty stdin, malformed FEM) |
| `io_error` | Filesystem or permission error |
| `drift_error` | Source content has changed since session creation |

## Proposal: Semantic exit codes

Fabbro currently uses `1` for all errors. Adopting standard Unix exit codes (BSD `sysexits.h` style) would help orchestration layers without parsing JSON.

| Code | Name | Meaning |
|------|------|---------|
| 0 | EX_OK | Success |
| 64 | EX_USAGE | The command was used incorrectly |
| 65 | EX_DATAERR | The input data was incorrect in some way |
| 66 | EX_NOINPUT | An input file did not exist or was not readable |
| 69 | EX_UNAVAILABLE | A service or resource was unavailable |
| 74 | EX_IOERR | An error occurred while doing I/O |

## Decisions to make

### 1. Should we always output JSON errors on stderr if --json is passed?

**Option A: Yes, always.**
- Pro: simple, consistent contract
- Con: hides human-readable output even for humans who just wanted JSON for success

**Option B: Only if stderr is not a TTY.**
- Pro: better DX for humans
- Con: inconsistent behavior for agents depending on TTY allocation

**Recommendation:** Option A. Agents always pass `--json` and expect the contract.

### 2. Implementation: global or per-command?

**Option A: Global wrapper in `realMain`.**
- Pro: central, hard to forget
- Con: needs access to the `jsonFlag` which is defined per-command

**Option B: Command-local handling.**
- Pro: allows specific `details` for different error types
- Con: requires repetitive boilerplate in every `RunE`

**Recommendation:** Hybrid. Global wrapper for standard errors, command-local for specific details if needed.

## Action plan

1. Document the *intended* contract in `docs/cli.md` (v1 current vs v2 planned)
2. Add structured error types to `internal/`
3. Refactor `realMain` to handle a new `StructuredError` type
4. Roll out to core agent-facing commands:
   - `fabbro review --stdin --no-interactive`
   - `fabbro apply <id> --json`
   - `fabbro session list --json`
5. Verify with pi extension and update the pi error parsing if needed.

---

## Conclusion

Structured errors are a "Type 1" (high reversibility) decision. We should adopt the BSD-style codes and a simple JSON error shape immediately to unblock more reliable agent integrations.
