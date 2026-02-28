# Spec Coverage Matrix

This document tracks implementation status for all scenarios in the fabbro specifications.

**Legend:**
- ✅ Implemented — Working in current build
- 🔶 Partial — Core functionality works, some aspects missing
- ❌ Not Implemented — Planned but not yet built
- 🚫 Deprecated — Removed from roadmap

**Last updated:** 2026-02-27

---

## 01_initialization.feature

| # | Scenario | Status | Notes |
|---|----------|--------|-------|
| 1 | Initializing a new project | 🔶 | Creates `.fabbro/sessions/`; no `templates/`, `config.yaml`, or `.gitignore` |
| 2 | Initializing an already initialized project | ✅ | Works correctly |
| 3 | Quiet initialization | ❌ | `--quiet` flag not implemented |
| 4 | Initializing in a subdirectory of an initialized project | ❌ | No parent detection/warning |
| 5 | Initializing with agent integration scaffolding | ❌ | `--agents` flag not implemented |
| 6 | Initializing with agents updates AGENTS.md | ❌ | `--agents` flag not implemented |
| 7 | Agent scaffolding detects available agents | ❌ | `--agents` flag not implemented |

**Summary:** 1 ✅, 1 🔶, 5 ❌

---

## 02_review_session.feature

| # | Scenario | Status | Notes |
|---|----------|--------|-------|
| 1 | Creating a review session from stdin | ✅ | Works with `fabbro review --stdin` |
| 2 | Creating a review session from a file | ✅ | Works with `fabbro review <file>` |
| 3 | Creating a review session with a custom session ID | ❌ | `--id` flag not implemented |
| 4 | Session file contains metadata header | ✅ | YAML frontmatter with `session_id`, `created_at` |
| 5 | Session file preserves original content | ✅ | Content preserved correctly |
| 6 | Attempting to review without initialization | ✅ | Returns error with exit code 1 |
| 7 | Attempting to review a non-existent file | ✅ | Returns "file not found" error |
| 8 | Attempting to review with no input | ✅ | Returns error suggesting `--stdin` or file path |
| 9 | Opening session in external editor instead of TUI | ❌ | `--editor` flag not implemented |
| 10 | Non-interactive mode creates session without opening anything | ❌ | `--no-interactive` not implemented |

**Summary:** 7 ✅, 0 🔶, 3 ❌

---

## 03_tui_interaction.feature

| # | Scenario | Status | Notes |
|---|----------|--------|-------|
| 1 | Navigating with keyboard | ✅ | `j`/`k`/arrows work |
| 2 | Page navigation | ✅ | `Ctrl+d`/`Ctrl+u` work |
| 3 | Jump to beginning and end | ✅ | `gg`/`G` work |
| 4 | Center cursor in viewport | ✅ | `zz`/`zt`/`zb` work |
| 5 | Search within document | ✅ | `/` search with `n`/`N` navigation |
| 6 | Opening the command palette | ✅ | `Space` opens palette |
| 7 | Selecting action from command palette | ✅ | Works correctly |
| 8 | Dismissing the command palette | ✅ | `Esc` closes palette |
| 9 | Selecting a single line | ✅ | `v` toggle works |
| 10 | Selecting a range of lines | ✅ | Multi-line selection works |
| 11 | Canceling selection | ✅ | `Esc` clears selection |
| 12 | Expand selection to paragraph | ✅ | `ap` text object works |
| 13 | Expand selection to code block | ✅ | `ab` text object works |
| 14 | Expand selection to section | ✅ | `as` text object works |
| 15 | Shrink and grow selection by line | ✅ | `{`/`}` work |
| 16 | Adding a comment annotation | ✅ | `c` prompts and adds |
| 17 | Adding a delete annotation | ✅ | `d` prompts and adds |
| 18 | Adding a question annotation | ✅ | `q` prompts and adds |
| 19 | Adding an expand annotation | ✅ | `e` prompts and adds |
| 20 | Adding a keep annotation | ✅ | `k` adds without prompt |
| 21 | Adding a change annotation | ✅ | `r` prompts for replacement text |
| 22 | Canceling annotation input | ✅ | `Esc` cancels input |
| 23 | Text input wraps when content is long | ✅ | Text wraps in input area |
| 24 | Adding newlines in annotation input | ✅ | `Shift+Enter` inserts newline |
| 25 | Editing annotation text on current line | ✅ | `e` opens editor with pre-filled text |
| 26 | Picking annotation when multiple exist on same line | ✅ | Annotation picker appears |
| 27 | Editing annotation range | ❌ | `R` range editing not implemented |
| 28 | Canceling annotation edit | ✅ | `Esc` cancels edit |
| 29 | No annotation on current line | ✅ | Shows error message |
| 30 | Visual indicator for annotated lines | ✅ | `●` indicator shown |
| 31 | Show annotation preview when cursor on annotated line | ✅ | Preview panel appears |
| 32 | Multiple annotations show count | ✅ | Shows "(1 of N annotations)" |
| 33 | Annotation preview disappears when cursor leaves | ✅ | Returns to normal help text |
| 34 | Annotation range highlighting in preview | ✅ | `▐` range highlight indicator |
| 35 | Tab cycling updates annotation range highlighting | ✅ | Tab cycles through annotations |
| 36 | Viewing all annotations in session | ✅ | `a` opens annotations panel |
| 37 | Jumping to annotation from list | ✅ | Enter jumps to annotation |
| 38 | Clicking to position cursor | ❌ | Mouse not implemented |
| 39 | Click-drag to select range | ❌ | Mouse not implemented |
| 40 | Right-click context menu | ❌ | Mouse not implemented |
| 41 | Opening inline editor for direct content changes | ✅ | `i` opens inline editor |
| 42 | Saving inline edit | ✅ | `Ctrl+S`/`Ctrl+Enter` saves |
| 43 | Canceling inline edit | ✅ | `Esc` twice or `Ctrl+C` cancels |
| 44 | Saving and exiting the review | ✅ | `w` saves and exits |
| 45 | Quitting without saving | ✅ | `Q`/`Ctrl+C` quits immediately |
| 46 | Exiting with confirmation prompt | ❌ | Confirmation not implemented |
| 47 | Viewing help | ✅ | `?` help panel works |

**Summary:** 43 ✅, 0 🔶, 4 ❌

---

## 04_apply_feedback.feature

| # | Scenario | Status | Notes |
|---|----------|--------|-------|
| 1 | Applying feedback outputs human-readable summary | ✅ | Works without `--json` |
| 2 | Applying feedback as JSON | ✅ | `--json` outputs valid JSON |
| 3 | JSON contains all annotation fields | 🔶 | Has `sessionId`, `sourceFile`, `startLine`, `endLine`; missing `createdAt` |
| 4 | JSON includes all annotation types | ✅ | All types supported including `change` |
| 5 | Parsing inline comment annotation | ✅ | `{>> ... <<}` works |
| 6 | Parsing block delete annotation | ❌ | Block markers `{--/--}` not implemented |
| 7 | Parsing question annotation | ✅ | `{?? ... ??}` works |
| 8 | Parsing expand annotation | ✅ | `{!! ... !!}` works |
| 9 | Parsing keep annotation | ✅ | `{== ... ==}` works |
| 10 | Parsing unclear annotation | ✅ | `{~~ ... ~~}` works |
| 11 | Annotations reference original line numbers | ✅ | Frontmatter offset handled |
| 12 | Multi-line annotations span correct range | ✅ | StartLine/EndLine correct for multi-line |
| 13 | Applying non-existent session | ✅ | Returns error |
| 14 | Applying session with malformed FEM | ❌ | No FEM syntax validation |
| 15 | Warning when source content has changed | ❌ | Content hash not implemented |
| 16 | Compact JSON output for piping | ❌ | `--compact` not implemented |
| 17 | Pretty-printed JSON output | ✅ | Default is pretty-printed |
| 18 | Apply by source file path | ✅ | `--file` flag finds session by source |
| 19 | Apply by file returns latest session | ✅ | Returns most recent session for file |
| 20 | Apply by file not found | ✅ | Error when no session for file |
| 21 | Cannot use both session ID and --file | ✅ | Mutual exclusivity enforced |
| 22 | JSON output includes sourceFile | ✅ | `sourceFile` field in JSON output |
| 23 | stdin session has empty sourceFile | ✅ | `sourceFile: ""` for stdin sessions |

**Summary:** 18 ✅, 1 🔶, 4 ❌

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
| 9 | Change annotation syntax | ✅ | `{++ ... ++}` works |
| 10 | Multi-line change annotation | ✅ | `[lines N-M] ->` format works |
| 11 | Emphasize syntax | ❌ | `{** ... **}` not implemented |
| 12 | Section annotation | ❌ | `{## ... ##}` not implemented |
| 13 | Multiple annotations on single line | ✅ | Works |
| 14 | Overlapping annotations from different selections | ✅ | Works correctly |
| 15 | Escaped markup is not parsed | ❌ | Escaping not implemented |
| 16 | Session file with YAML frontmatter | ✅ | Works |
| 17 | Annotations preserve surrounding whitespace | ✅ | Works |
| 18 | Newlines in annotation text | ❌ | Not supported |
| 19 | Empty annotation text | ✅ | Works |
| 20 | Nested braces in annotation text | ❌ | Not handled |
| 21 | Unclosed annotation marker | ❌ | No syntax error reporting |

**Summary:** 13 ✅, 0 🔶, 8 ❌

---

## 07_web_notes_sidebar.feature

| # | Scenario | Status | Notes |
|---|----------|--------|-------|
| 1 | Notes panel appears in editor view | ✅ | Right-side panel with annotation count |
| 2 | Empty state when no annotations exist | ✅ | Shows placeholder message |
| 3 | Note card displays annotation details | ✅ | Badge, snippet, text, line number |
| 4 | Snippet preview is truncated at 60 characters | ✅ | Truncation with "…" |
| 5 | Notes are sorted by position in document | ✅ | Sorted by offset |
| 6 | Counter updates when annotations change | ✅ | Reactive count |
| 7 | Comment annotation shows Comment badge | ✅ | Blue styling |
| 8 | Suggest annotation shows Suggest badge | ✅ | Green styling |
| 9 | Clicking a note scrolls to its highlight | ✅ | Scroll + flash |
| 10 | Clicking a highlight scrolls to its note | ✅ | Bidirectional navigation |
| 11 | Deleting an annotation via the sidebar | ✅ | Removes annotation and highlight |
| 12 | Delete button does not trigger note click | ✅ | Proper event isolation |

**Summary:** 12 ✅, 0 🔶, 0 ❌

---

## 08_web_search.feature

| # | Scenario | Status | Notes |
|---|----------|--------|-------|
| 1 | Open search bar with / key | ✅ | Opens search bar in viewer |
| 2 | Dismiss search with Escape | ✅ | Clears query and highlights |
| 3 | Confirm search with Enter | ✅ | Closes bar, keeps highlights |
| 4 | / key is ignored when typing in textarea or input | ✅ | Early return for input elements |
| 5 | Matches highlight as the user types | ✅ | Incremental via `findMatches` |
| 6 | Highlights update incrementally | ✅ | Re-renders on each keystroke |
| 7 | No matches found | ✅ | Counter shows "0/0" |
| 8 | Search is case-insensitive | ✅ | Uses `toLowerCase()` |
| 9 | Match counter shows current position | ✅ | Shows "N/M" format |
| 10 | Counter updates on navigation | ✅ | Updates on navigate() |
| 11 | Navigate to next match with n | ✅ | `n` key in viewer |
| 12 | Navigate to previous match with N | ✅ | `N` key in viewer |
| 13 | Navigation wraps around at end | ✅ | Modulo arithmetic |
| 14 | Navigation wraps around at beginning | ✅ | Modulo arithmetic |
| 15 | First match is scrolled to when search begins | ✅ | `scrollToCurrentMatch` on update |

**Summary:** 15 ✅, 0 🔶, 0 ❌

---

## Overall Summary

| Spec | ✅ Implemented | 🔶 Partial | ❌ Not Implemented | Total |
|------|---------------|-----------|-------------------|-------|
| 01_initialization | 1 | 1 | 5 | 7 |
| 02_review_session | 7 | 0 | 3 | 10 |
| 03_tui_interaction | 43 | 0 | 4 | 47 |
| 04_apply_feedback | 18 | 1 | 4 | 23 |
| 05_session_management | 0 | 0 | 18 | 18 |
| 06_fem_markup | 13 | 0 | 8 | 21 |
| 07_web_notes_sidebar | 12 | 0 | 0 | 12 |
| 08_web_search | 15 | 0 | 0 | 15 |
| **TOTAL** | **109** | **2** | **42** | **153** |

**Coverage: 111/153 scenarios (73%)**

---

## Priority Implementation Recommendations

### High Priority (Core Workflow)
1. Session management commands (`sessions`, `show`, `resume`, `delete`)
2. Block delete markers for multi-line annotations
3. `createdAt` field in JSON output

### Medium Priority (UX Improvements)
1. Custom session ID (`--id` flag)
2. Confirmation prompt before discarding unsaved changes
3. Annotation range editing (`R` key)
4. Agent integration scaffolding (`--agents` flag)

### Low Priority (Nice to Have)
1. Mouse support
2. `--editor` and `--no-interactive` modes
3. `emphasize` and `section` annotation types
4. FEM syntax escaping, nested braces, and error reporting
5. Compact JSON output (`--compact` flag)
6. Session cleaning and export commands
