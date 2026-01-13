# Spec Coverage Matrix

This document tracks implementation status for all scenarios in the fabbro specifications.

**Legend:**
- ✅ Implemented — Working in current build
- 🔶 Partial — Core functionality works, some aspects missing
- ❌ Not Implemented — Planned but not yet built
- 🚫 Deprecated — Removed from roadmap

**Last updated:** 2026-01-13

---

## 01_initialization.feature

| # | Scenario | Status | Notes |
|---|----------|--------|-------|
| 1 | Initializing a new project | 🔶 | Creates `.fabbro/sessions/`; no `templates/`, `config.yaml`, or `.gitignore` |
| 2 | Initializing an already initialized project | ✅ | Works correctly |
| 3 | Quiet initialization | ❌ | `--quiet` flag not implemented |
| 4 | Initializing in a subdirectory of an initialized project | ❌ | No parent detection/warning |

**Summary:** 1 ✅, 1 🔶, 2 ❌

---

## 02_review_session.feature

| # | Scenario | Status | Notes |
|---|----------|--------|-------|
| 1 | Creating a review session from stdin | ✅ | Works with `fabbro review --stdin` |
| 2 | Creating a review session from a file | ❌ | Only `--stdin` supported currently |
| 3 | Creating a review session with a custom session ID | ❌ | `--id` flag not implemented |
| 4 | Session file contains metadata header | ✅ | YAML frontmatter with `session_id`, `created_at` |
| 5 | Session file preserves original content | ✅ | Content preserved correctly |
| 6 | Attempting to review without initialization | ✅ | Returns error with exit code 1 |
| 7 | Attempting to review a non-existent file | ❌ | File review not implemented |
| 8 | Attempting to review with no input | ✅ | Returns error suggesting `--stdin` |
| 9 | Opening session in external editor instead of TUI | ❌ | `--editor` flag not implemented |
| 10 | Non-interactive mode creates session without opening anything | ❌ | `--no-interactive` not implemented |

**Summary:** 5 ✅, 0 🔶, 5 ❌

---

## 03_tui_interaction.feature

| # | Scenario | Status | Notes |
|---|----------|--------|-------|
| 1 | Navigating with keyboard | ✅ | `j`/`k`/arrows work |
| 2 | Page navigation | ✅ | `Ctrl+d`/`Ctrl+u` work |
| 3 | Jump to beginning and end | ✅ | `gg`/`G` work |
| 4 | Search within document | ❌ | `/` search not implemented |
| 5 | Opening the command palette | ✅ | `Space` opens palette |
| 6 | Selecting action from command palette | ✅ | Works correctly |
| 7 | Dismissing the command palette | ✅ | `Esc` closes palette |
| 8 | Selecting a single line | ✅ | `v` toggle works |
| 9 | Selecting a range of lines | ✅ | Multi-line selection works |
| 10 | Canceling selection | ✅ | `Esc` clears selection |
| 11 | Expand selection to paragraph | ❌ | Text objects not implemented |
| 12 | Expand selection to code block | ❌ | Text objects not implemented |
| 13 | Expand selection to section | ❌ | Text objects not implemented |
| 14 | Shrink and grow selection by line | ❌ | `{`/`}` not implemented |
| 15 | Adding a comment annotation | ✅ | `c` prompts and adds |
| 16 | Adding a delete annotation | ✅ | `d` prompts and adds |
| 17 | Adding a question annotation | ✅ | `q` prompts and adds |
| 18 | Adding an expand annotation | ✅ | `e` prompts and adds |
| 19 | Adding a keep annotation | ✅ | `k` adds without prompt |
| 20 | Canceling annotation input | ✅ | `Esc` cancels input |
| 21 | Viewing all annotations in session | ❌ | Annotations panel not implemented |
| 22 | Jumping to annotation from list | ❌ | Annotations panel not implemented |
| 23 | Clicking to position cursor | ❌ | Mouse not implemented |
| 24 | Click-drag to select range | ❌ | Mouse not implemented |
| 25 | Right-click context menu | ❌ | Mouse not implemented |
| 26 | Saving and exiting the review | ✅ | `w` saves and exits |
| 27 | Quitting without saving | ✅ | `Q`/`Ctrl+C` quits immediately |
| 28 | Exiting with confirmation prompt | ❌ | Confirmation not implemented |
| 29 | Viewing help | ❌ | `?` help panel not implemented |

**Summary:** 17 ✅, 0 🔶, 12 ❌

---

## 04_apply_feedback.feature

| # | Scenario | Status | Notes |
|---|----------|--------|-------|
| 1 | Applying feedback outputs human-readable summary | ✅ | Works without `--json` |
| 2 | Applying feedback as JSON | ✅ | `--json` outputs valid JSON |
| 3 | JSON contains all annotation fields | 🔶 | Has `sessionId`, `startLine`, `endLine`; missing `sourceFile`, `createdAt` |
| 4 | JSON includes all annotation types | ✅ | All 6 types supported |
| 5 | Parsing inline comment annotation | ✅ | `{>> ... <<}` works |
| 6 | Parsing block delete annotation | ❌ | Block markers `{--/--}` not implemented |
| 7 | Parsing question annotation | ✅ | `{?? ... ??}` works |
| 8 | Parsing expand annotation | ✅ | `{!! ... !!}` works |
| 9 | Parsing keep annotation | ✅ | `{== ... ==}` works |
| 10 | Parsing unclear annotation | ✅ | `{~~ ... ~~}` works |
| 11 | Annotations reference original line numbers | ✅ | Frontmatter offset handled |
| 12 | Multi-line annotations span correct range | ❌ | Only single-line currently |
| 13 | Applying non-existent session | ✅ | Returns error |
| 14 | Applying session with malformed FEM | ❌ | No FEM syntax validation |
| 15 | Warning when source content has changed | ❌ | Content hash not implemented |
| 16 | Compact JSON output for piping | ❌ | `--compact` not implemented |
| 17 | Pretty-printed JSON output | ✅ | Default is pretty-printed |

**Summary:** 11 ✅, 1 🔶, 5 ❌

---

## 05_session_management.feature

| # | Scenario | Status | Notes |
|---|----------|--------|-------|
| 1 | Listing all sessions | ❌ | `fabbro sessions` not implemented |
| 2 | Listing sessions in JSON format | ❌ | Not implemented |
| 3 | No sessions exist | ❌ | Not implemented |
| 4 | Showing session details | ❌ | `fabbro show` not implemented |
| 5 | Showing session with annotation breakdown | ❌ | Not implemented |
| 6 | Showing non-existent session | ❌ | Not implemented |
| 7 | Resuming an interrupted review | ❌ | `fabbro resume` not implemented |
| 8 | Resuming in editor mode | ❌ | Not implemented |
| 9 | Resuming non-existent session | ❌ | Not implemented |
| 10 | Deleting a session | ❌ | `fabbro delete` not implemented |
| 11 | Deleting a session with --force | ❌ | Not implemented |
| 12 | Deleting non-existent session | ❌ | Not implemented |
| 13 | Cleaning sessions older than threshold | ❌ | `fabbro clean` not implemented |
| 14 | Dry-run cleaning | ❌ | Not implemented |
| 15 | Exporting session as standalone file | ❌ | `fabbro export` not implemented |
| 16 | Exporting session to stdout | ❌ | Not implemented |
| 17 | Partial session ID matching | ❌ | Not implemented |
| 18 | Ambiguous partial session ID | ❌ | Not implemented |

**Summary:** 0 ✅, 0 🔶, 18 ❌

---

## 06_fem_markup.feature

| # | Scenario | Status | Notes |
|---|----------|--------|-------|
| 1 | Inline comment syntax | ✅ | Works |
| 2 | Comment with line reference (sidecar style) | ❌ | Sidecar not implemented |
| 3 | Block delete with reason | ❌ | Block markers not implemented |
| 4 | Inline delete (single line) | ✅ | Works |
| 5 | Question syntax | ✅ | Works |
| 6 | Expand syntax | ✅ | Works |
| 7 | Keep syntax | ✅ | Works |
| 8 | Unclear syntax | ✅ | Works |
| 9 | Emphasize syntax | ❌ | `{** ... **}` not implemented |
| 10 | Section annotation | ❌ | `{## ... ##}` not implemented |
| 11 | Multiple annotations on single line | ✅ | Works |
| 12 | Escaped markup is not parsed | ❌ | Escaping not implemented |
| 13 | Session file with YAML frontmatter | ✅ | Works |
| 14 | Annotations preserve surrounding whitespace | ✅ | Works |
| 15 | Newlines in annotation text | ❌ | Not supported |
| 16 | Empty annotation text | ✅ | Works |
| 17 | Nested braces in annotation text | ❌ | Not handled |
| 18 | Unclosed annotation marker | ❌ | No syntax error reporting |

**Summary:** 10 ✅, 0 🔶, 8 ❌

---

## Overall Summary

| Spec | ✅ Implemented | 🔶 Partial | ❌ Not Implemented | Total |
|------|---------------|-----------|-------------------|-------|
| 01_initialization | 1 | 1 | 2 | 4 |
| 02_review_session | 5 | 0 | 5 | 10 |
| 03_tui_interaction | 17 | 0 | 12 | 29 |
| 04_apply_feedback | 11 | 1 | 5 | 17 |
| 05_session_management | 0 | 0 | 18 | 18 |
| 06_fem_markup | 10 | 0 | 8 | 18 |
| **TOTAL** | **44** | **2** | **50** | **96** |

**Coverage: 46/96 scenarios (48%)**

---

## Priority Implementation Recommendations

### High Priority (Core Workflow)
1. Session management commands (`sessions`, `show`, `resume`, `delete`)
2. File input for `fabbro review` (not just stdin)
3. Block delete markers for multi-line annotations

### Medium Priority (UX Improvements)
1. Search (`/`) in TUI
2. Help panel (`?`) in TUI  
3. Custom session ID (`--id` flag)
4. Confirmation prompt before discarding unsaved changes

### Low Priority (Nice to Have)
1. Mouse support
2. Text objects (`ap`, `ab`, `as`)
3. `--editor` and `--no-interactive` modes
4. `emphasize` and `section` annotation types
5. FEM syntax escaping and error reporting
