# AIX Improvements Implementation Plan

**Date**: 2026-01-29

## Overview

This plan details the implementation of key improvements to `fabbro`'s command-line interface to enhance its AI Experience (AIX). The goal is to make the CLI more consistent, predictable, and self-healing for AI agents, while maintaining excellent usability for human users, based on our established design principles.

## Related

- Research: `research/2026-01-29-aix-desire-path-research.md`

## Current State

The `fabbro` CLI is functional but has several inconsistencies from an AIX perspective. Help text is minimal, not all data-returning commands have a `--json` output option, and some error messages could be more suggestive.

## Desired End State

- All commands will have verbose, structured help text with examples.
- All commands that return data will consistently provide a `--json` output option.
- Error messages will be more helpful and suggest corrective actions where possible.
- The CLI will be significantly easier for an AI agent to interact with reliably.

**How to verify:**
- Run `fabbro [command] --help` for every command and verify the output is detailed and structured.
- Run all data-returning commands with the `--json` flag and verify they produce valid JSON.
- Intentionally trigger common errors (e.g., not initialized, missing arguments) and verify the error messages are more suggestive.

## Out of Scope

- A full transition to structured, JSON-only error messages. This plan will improve error strings, but a full structured error system is a larger task for future consideration.
- Changing the core logic of any command beyond what is necessary to implement these AIX improvements.

## Risks & Mitigations

- **Risk:** Manually updating all commands is tedious and prone to inconsistency.
- **Mitigation:** Update all command help text in a single commit or development session. The first command updated (`buildInitCmd`) will serve as the explicit template for all subsequent command help text in the same session.

---

## Phase 1: Standardize Command Help Text

### Changes Required

**File: `cmd/fabbro/main.go`**

- **Changes:**
    - For each `build<CommandName>Cmd` function (`buildInitCmd`, `buildReviewCmd`, `buildApplyCmd`, `buildSessionListCmd`, `buildSessionResumeCmd`, `buildTutorCmd`):
        - Add a detailed `Long` description explaining what the command does, its pre-conditions, and post-conditions.
        - Add an `Example` field showing 1-2 common use cases for the command.
        - The `review` command from the original research document can be used as a template.
- **Tests:**
    - This is a non-functional change to help text. Manual verification is sufficient. No new automated tests are required for this phase.

### Implementation Approach

1.  Start with `buildInitCmd`.
2.  Update its `Long` and `Example` fields to be highly descriptive.
3.  Proceed to the next command, using the previous one as a template for consistency.
4.  Repeat for all commands.

### Success Criteria

#### Manual:
- [ ] Run `fabbro init --help` and verify the new help text is displayed.
- [ ] Run `fabbro review --help` and verify the new help text is displayed.
- [ ] Run `fabbro apply --help` and verify the new help text is displayed.
- [ ] Run `fabbro session list --help` and verify the new help text is displayed.
- [ ] Run `fabbro session resume --help` and verify the new help text is displayed.

---

## Phase 2: Enforce `--json` Flag for Data-Returning Commands

### Changes Required

**File: `cmd/fabbro/main.go`**

- **Audit all commands:** Identify which commands return data to `stdout`.
    - `apply`: Already has `--json`.
    - `session list`: Already has `--json`.
    - `init`: Returns a string, but no structured data. No change needed.
    - `review`: Returns a session ID string. Could benefit from a JSON output for consistency.
    - `session resume`: Doesn't return data on `stdout`. No change needed.
- **Changes:**
    - In `buildReviewCmd`, add a `--json` flag.
    - If the `--json` flag is present, the "Created session: [ID]" output should be changed to a JSON object like `{"sessionId": "[ID]"}`.
- **Tests:**
    - In `cmd/fabbro/main_test.go`, add a new test case to verify that `fabbro review --json` produces the correct JSON output.

### Implementation Approach

1.  Add the `jsonFlag` variable and flag definition to `buildReviewCmd`.
2.  Wrap the final `fmt.Fprintf` in an `if/else` block. If `jsonFlag` is true, print the JSON object; otherwise, print the human-readable string.
3.  Write a new test function in `main_test.go` that executes the `review` command with the `--json` flag and asserts that the output is the expected JSON.

### Success Criteria

#### Automated:
- [ ] All existing tests pass: `go test ./...`
- [ ] New test for `review --json` passes.

#### Manual:
- [ ] Run `fabbro review [file] --json` and verify it prints a JSON object with the session ID.
- [ ] Run `fabbro review [file]` (without the flag) and verify it still prints the human-readable string.

---

## Testing Strategy

**Following TDD (for Phase 2):**
1.  Write a test for `review --json` that asserts the expected JSON output.
2.  Watch the test fail because the flag and logic do not exist yet.
3.  Implement the `--json` flag and the conditional output logic.
4.  Run the test again and watch it pass.

**Test types needed:**
- Unit tests: A new unit test for the `review --json` functionality will be added to `cmd/fabbro/main_test.go`.

## Rollback Strategy

- Changes can be reverted using `git restore` or `git checkout` on `cmd/fabbro/main.go`. No complex rollback is needed.

---

## Phase 3: Improve Error Message Suggestions

### Changes Required

**File: `cmd/fabbro/main.go`**

- **Changes:**
    - Audit all `fmt.Errorf` calls within the command definitions.
    - Identify error messages that state a problem without suggesting a solution.
    - Update these error messages to be more suggestive.
        - **Example 1:** In `buildReviewCmd`, change `return fmt.Errorf("no input provided. Use --stdin or provide a file path")` to something more directive like `return fmt.Errorf("Error: No input file specified. Please provide a file path as an argument or pipe content via --stdin.")`
        - **Example 2:** In `buildApplyCmd`, change `return fmt.Errorf("must provide either session-id or --file")` to `return fmt.Errorf("Error: You must specify a session to apply. Use a session-id argument or find a session by source with the --file flag. Use 'fabbro session list' to see available sessions.")`

- **Tests:**
    - In `cmd/fabbro/main_test.go`, find the existing tests that verify the error conditions being updated.
    - Modify these tests to assert that the new, more descriptive error messages are returned.

### Implementation Approach

1.  Search for all instances of `fmt.Errorf` in `cmd/fabbro/main.go`.
2.  For each instance, evaluate if the message can be improved to guide the user toward a solution.
3.  Update the string with a more helpful message.
4.  Find the corresponding unit test in `cmd/fabbro/main_test.go` and update its error assertion to match the new string.

### Success Criteria

#### Automated:
- [ ] All existing tests pass: `go test ./...`
- [ ] Updated tests for error messages pass, verifying the new strings.

#### Manual:
- [ ] Trigger the "no input provided" error for `fabbro review` and verify the new, more helpful message is displayed.
- [ ] Trigger the "must provide either session-id or --file" error for `fabbro apply` and verify the new, more helpful message is displayed.
