# Research: pi extension validation and web reuse boundaries

**Date:** 2026-04-14
**Related plan:** `plans/2026-04-14-pi-extension.md`
**Related issues:** `fabbro-ulj`, `fabbro-0ji`

## Summary

The current `pi-fabbro` extension is sufficient for a v1 human-in-the-loop review workflow without embedding fabbro's TUI.

The strongest architectural conclusion is:

- **reuse the fabbro CLI contract directly in pi**
- **do not reuse the web UI modules in pi**
- **do not extract `web/fem.js` right now** because it is browser-oriented and materially less complete than the current Go FEM parser

## Validation performed

Validated from the `fabbro/` repo root with the extension loaded via `pi -e ./pi-fabbro`.

### Commands exercised

- `/fabbro-help`
- `/fabbro-status`
- `/fabbro-prime`
- `/fabbro-review ...`
- `/fabbro-apply <session-id>`
- `/fabbro-sessions`
- `cd pi-fabbro && npm run smoke`
- `just test`

### Observed workflow

1. pi creates a session from generated text with `/fabbro-review ...`
2. pi returns a session ID and tells the human exactly what to run next:
   - `fabbro session resume <id>`
3. the human performs review outside pi in fabbro's native TUI
4. pi can later retrieve structured feedback with `/fabbro-apply <id>`
5. pi can discover sessions with `/fabbro-sessions`

### Outcome

This is enough for v1.

The extension does **not** need to embed Bubble Tea. The external-resume step is explicit, understandable, and already aligned with fabbro's native workflow.

## Reuse audit of web modules

Audited files:

- `web/app.js`
- `web/editor.js`
- `web/notes.js`
- `web/storage.js`
- `web/fem.js`

### `web/app.js`

**Role:** browser application shell.

**Why not reusable in pi:**

- built around DOM mounting and event wiring
- manages drag-and-drop, file upload, browser fetch flows, tutorial startup, and landing-page rendering
- coordinates IndexedDB-backed sessions and browser-only navigation/state

**Conclusion:** not reusable for pi.

### `web/editor.js`

**Role:** browser editor/viewer interaction layer.

**Why not reusable in pi:**

- built around DOM selection APIs, mouse events, scrolling, key handlers, and HTML rendering
- depends on browser-only modules like viewer, toolbar, notes sidebar, help overlay, search, and palette UI
- represents a direct browser interaction model rather than a CLI/orchestration layer

**Conclusion:** not reusable for pi.

### `web/notes.js`

**Role:** browser notes sidebar renderer.

**Why not reusable in pi:**

- entirely DOM-driven card rendering and click/edit/delete interactions
- assumes browser-side text snippets and note focus behaviors
- no clean non-UI core separated from rendering

**Conclusion:** not reusable for pi.

### `web/storage.js`

**Role:** browser session persistence.

**Why not reusable in pi:**

- hard-coded to IndexedDB
- session shape and persistence semantics are web-app-local, not CLI-native
- pi integration already relies on fabbro's real session storage and CLI commands

**Conclusion:** not reusable for pi.

### `web/fem.js`

**Role:** JavaScript FEM parser/serializer for the web app.

**Potentially reusable in theory:** yes, because it is the least DOM-dependent module.

**Why it should not be reused yet:**

1. the current pi extension does not need a JS FEM parser at all because it uses:
   - `fabbro review --stdin --no-interactive`
   - `fabbro apply <id> --json`
   - `fabbro session list --json`
2. the Go parser in `internal/fem/parser.go` is more capable than `web/fem.js`
3. `web/fem.js` currently looks like a simplified port, not an authoritative shared implementation

### Important capability gap

Compared with `internal/fem/parser.go`, `web/fem.js` does **not** clearly implement several current parser behaviors, including:

- block delete region handling
- multiline annotation parsing
- unclosed-marker error reporting
- escaped brace handling
- sidecar line reference rewriting like `[line N]` / `[lines N-M]`

That means extracting `web/fem.js` into a shared module right now would risk spreading a **less correct** implementation into a new integration.

**Conclusion:** `web/fem.js` is the only plausible future extraction candidate, but it is **not worth extracting for the pi extension v1**.

## Recommendation

### For pi v1

Keep the architecture exactly where it is headed:

- pi package as orchestration layer
- fabbro CLI as the stable integration surface
- human review stays in fabbro's external TUI
- no direct reuse of `web/` modules

### For future reuse work

If reuse becomes desirable later, the right direction is **not** to import `web/fem.js` into pi immediately.

Instead:

1. decide which FEM implementation is authoritative
2. align behavior across Go and JS implementations with tests
3. extract only a deliberately shared, tested parser module
4. adopt that shared module only if a real consumer needs it

At today's scope, there is no evidence that this extraction would pay for itself.

## Friction notes

### 1. PATH fallback behavior is valuable

The extension successfully falls back to `go run ./cmd/fabbro` when the PATH binary lacks `--no-interactive` support. This is especially useful during local development.

### 2. Resume step must stay explicit

The main user handoff is:

- pi creates the session
- human runs `fabbro session resume <id>` outside pi
- pi later applies structured feedback

This should be documented plainly rather than hidden. It is the key mental model for the integration.

### 3. Machine-readable contracts still matter

The current CLI is sufficient for v1, but reliability would improve with:

- documented JSON success/failure shapes
- structured JSON errors
- clearer exit-code semantics

That remains follow-up work under `fabbro-uje`.

## Final call

**Reuse boundary:**

- **Do not reuse** `web/app.js`, `web/editor.js`, `web/notes.js`, or `web/storage.js` in pi
- **Do not extract** `web/fem.js` yet
- **Do reuse** fabbro's CLI-native contract as the integration boundary

This is the lowest-risk and highest-leverage path for the pi integration.