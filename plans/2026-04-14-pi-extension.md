# Plan: pi.dev extension for fabbro

**Date:** 2026-04-14
**Status:** In Progress

## Goal

Create a pi package/extension that integrates fabbro's CLI-native review workflow into pi without embedding fabbro's Bubble Tea TUI.

## Progress

- [x] Phase 1: Scaffold package
- [x] Phase 2: Session creation and priming
- [x] Phase 3: Feedback retrieval
- [x] Phase 4: Session discovery and helper UX
- [x] Phase 5: Reuse assessment and workflow validation
- [x] Phase 6: Contract hardening (follow-up)

## Approach

Use fabbro's existing CLI as the primary integration surface:

- `fabbro prime --json`
- `fabbro review --stdin --no-interactive`
- `fabbro apply <id> --json`
- `fabbro session list --json`

The human review UI remains external:

- pi creates or inspects sessions
- the human resumes the session with `fabbro session resume <id>`
- pi later reapplies the structured feedback

## Non-goals (v1)

- embedding fabbro's Bubble Tea TUI inside pi
- reusing browser UI from `web/` directly inside pi
- MCP integration
- redesigning fabbro's core review workflow

## Phases

### Phase 1: Scaffold package

Create the package skeleton and prove pi can load it.

Deliverables:
- package directory and `package.json`
- extension entrypoint
- runtime check for `fabbro` availability
- README or local usage notes

Verification:
- load the extension in pi successfully
- confirm missing `fabbro` yields a clear user-facing error

### Phase 2: Session creation and priming

Implement the extension surfaces needed to:
- fetch fabbro primer data
- create review sessions from generated text via stdin/non-interactive mode

Verification:
- create a session from pi using generated text
- capture the returned session ID reliably

### Phase 3: Feedback retrieval

Implement the extension surfaces needed to:
- fetch feedback from `fabbro apply <id> --json`
- handle parseable success/failure cases cleanly

Verification:
- apply feedback from an existing session and surface it in pi as structured data

### Phase 4: Session discovery and helper UX

Implement:
- `fabbro session list --json`
- helper commands or UI to guide the user to `fabbro session resume <id>`

Verification:
- list sessions from pi
- select or inspect a session without manual shell glue

### Phase 5: Reuse assessment and workflow validation

Assess whether anything from the web app should be shared with the pi integration.

Likely reusable candidates:
- `web/fem.js`

Likely non-reusable areas:
- DOM editor and notes UI
- IndexedDB persistence

Verification:
- produce a short research note or update the existing research with explicit reuse boundaries
- validate the full flow end-to-end in pi

### Phase 6: Contract hardening (follow-up)

Improve fabbro's machine-readable contract if needed:
- structured JSON errors
- semantic exit-code policy
- documented stable output expectations

This phase is valuable but not required to prove the pi integration concept.

## Success Criteria

- pi extension can create fabbro review sessions from generated text
- pi extension can retrieve feedback from fabbro as structured data
- pi extension can list sessions
- the v1 integration does not attempt to embed the fabbro TUI
- any reuse from the web implementation is explicit and intentional
