# Spec Coverage Matrix

This document tracks implementation status for all scenarios in the fabbro specifications.

**Legend:**
- ✅ Implemented — Working in current build
- 🔶 Partial — Core functionality works, some aspects missing
- ❌ Not Implemented — Planned but not yet built
- 🚫 Deprecated — Removed from roadmap

**Last updated:** 2026-04-14

---

## 01_initialization.feature

| # | Scenario | Status | Notes |
|---|----------|--------|-------|
| 1 | Initializing a new project | 🔶 | Creates `.fabbro/sessions/`; no `templates/`, `config.yaml`, or `.gitignore` |
| 2 | Initializing an already initialized project | ✅ | Works correctly |
| 3 | Quiet initialization | ✅ | `--quiet` implemented |
| 4 | Initializing in a subdirectory of an initialized project | ✅ | Detects parent root and warns |
| 5 | Initializing with agent integration scaffolding | ✅ | `--agents` implemented |
| 6 | Initializing with agents updates AGENTS.md | ✅ | Appends `## fabbro workflow` section when missing |
| 7 | Agent scaffolding detects available agents | ✅ | Creates `.claude/` and `.cursor/` command files only when those dirs exist |

**Summary:** 6 ✅, 1 🔶, 0 ❌

---

## 02_review_session.feature

| # | Scenario | Status | Notes |
|---|----------|--------|-------|
| 1 | Creating a review session from stdin | ✅ | Works with `fabbro review --stdin` |
| 2 | Creating a review session from a file | ✅ | Works with `fabbro review <file>` |
| 3 | Creating a review session with a custom session ID | ✅ | `--id` implemented |
| 4 | Session file contains metadata header | ✅ | YAML frontmatter with `session_id`, `created_at` |
| 5 | Session file preserves original content | ✅ | Content preserved correctly |
| 6 | Attempting to review without initialization | ✅ | Returns error with exit code 1 |
| 7 | Attempting to review a non-existent file | ✅ | Returns "file not found" error |
| 8 | Attempting to review with no input | ✅ | Returns error suggesting `--stdin` or file path |
| 9 | Opening session in external editor instead of TUI | ✅ | `--editor` implemented |
| 10 | Non-interactive mode creates session without opening anything | ✅ | `--no-interactive` implemented |

**Summary:** 10 ✅, 0 🔶, 0 ❌

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
| 3 | JSON contains all annotation fields | ✅ | Includes `sessionId`, `sourceFile`, `createdAt`, `annotations` |
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
| 15 | Warning when source content has changed | ✅ | Content hash verification implemented |
| 16 | Compact JSON output for piping | ✅ | `--compact` implemented |
| 17 | Pretty-printed JSON output | ✅ | Default is pretty-printed |
| 18 | Apply by source file path | ✅ | `--file` flag finds session by source |
| 19 | Apply by file returns latest session | ✅ | Returns most recent session for file |
| 20 | Apply by file not found | ✅ | Error when no session for file |
| 21 | Cannot use both session ID and --file | ✅ | Mutual exclusivity enforced |
| 22 | JSON output includes sourceFile | ✅ | `sourceFile` field in JSON output |
| 23 | stdin session has empty sourceFile | ✅ | `sourceFile: ""` for stdin sessions |

**Summary:** 21 ✅, 0 🔶, 2 ❌

---

## 05_session_management.feature

| # | Scenario | Status | Notes |
|---|----------|--------|-------|
| 1 | Listing all sessions | ✅ | `fabbro session list` implemented |
| 2 | Listing sessions in JSON format | ✅ | `fabbro session list --json` implemented |
| 3 | No sessions exist | ✅ | Helpful empty-state output |
| 4 | Showing session details | ✅ | `fabbro session show` implemented |
| 5 | Showing session with annotation breakdown | ✅ | Annotation counts by type shown |
| 6 | Showing non-existent session | ✅ | Returns error |
| 7 | Resuming an interrupted review | ✅ | `fabbro session resume` implemented |
| 8 | Resuming in editor mode | ✅ | `--editor` implemented |
| 9 | Resuming non-existent session | ✅ | Returns error |
| 10 | Deleting a session | ✅ | Confirmation prompt implemented |
| 11 | Deleting a session with --force | ✅ | Implemented |
| 12 | Deleting non-existent session | ✅ | Returns error |
| 13 | Cleaning sessions older than threshold | ✅ | Implemented |
| 14 | Dry-run cleaning | ✅ | Implemented |
| 15 | Exporting session as standalone file | ✅ | `fabbro session export --output` implemented |
| 16 | Exporting session to stdout | ✅ | Implemented |
| 17 | Partial session ID matching | ✅ | Prefix matching implemented |
| 18 | Ambiguous partial session ID | ✅ | Ambiguity error lists matches |

**Summary:** 18 ✅, 0 🔶, 0 ❌

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

## 09_web_docx_upload.feature

| # | Scenario | Status | Notes |
|---|----------|--------|-------|
| 1 | Dropping a .docx file starts a review session | ❌ | Not implemented |
| 2 | Extracted text preserves paragraph structure | ❌ | Not implemented |
| 3 | Drop zone label includes .docx | ❌ | Not implemented |
| 4 | Handling a corrupt .docx file | ❌ | Not implemented |
| 5 | Handling an empty .docx file | ❌ | Not implemented |
| 6 | Rejecting .doc (legacy Word) files | ❌ | Not implemented |

**Summary:** 0 ✅, 0 🔶, 6 ❌

---

## 10_web_html_to_text.feature

| # | Scenario | Status | Notes |
|---|----------|--------|-------|
| 1 | Markdown response is used as-is | ❌ | Not implemented |
| 2 | HTML response is converted to plain text | ❌ | Not implemented |
| 3 | Plain text response is used as-is | ❌ | Not implemented |
| 4 | Script and style elements are removed | ❌ | Not implemented |
| 5 | Navigation and footer elements are removed | ❌ | Not implemented |
| 6 | Article element content is preferred | ❌ | Not implemented |
| 7 | Main element content is preferred when no article exists | ❌ | Not implemented |
| 8 | Body content is used as fallback | ❌ | Not implemented |
| 9 | Markdown token count is surfaced when available | ❌ | Not implemented |
| 10 | Missing token count header returns null | ❌ | Not implemented |
| 11 | GitHub URLs still use the GitHub API | ❌ | Not implemented |

**Summary:** 0 ✅, 0 🔶, 11 ❌

---

## Overall Summary

| Spec | ✅ Implemented | 🔶 Partial | ❌ Not Implemented | Total |
|------|---------------|-----------|-------------------|-------|
| 01_initialization | 6 | 1 | 0 | 7 |
| 02_review_session | 10 | 0 | 0 | 10 |
| 03_tui_interaction | 43 | 0 | 4 | 47 |
| 04_apply_feedback | 21 | 0 | 2 | 23 |
| 05_session_management | 18 | 0 | 0 | 18 |
| 06_fem_markup | 13 | 0 | 8 | 21 |
| 07_web_notes_sidebar | 12 | 0 | 0 | 12 |
| 08_web_search | 15 | 0 | 0 | 15 |
| 09_web_docx_upload | 0 | 0 | 6 | 6 |
| 10_web_html_to_text | 0 | 0 | 11 | 11 |
| **TOTAL** | **138** | **1** | **31** | **170** |

**Coverage: 139/170 scenarios (82%)**

---

## Priority Implementation Recommendations

### High Priority (Core Workflow)
1. Block delete markers for multi-line annotations
2. FEM syntax validation/reporting for malformed markup
3. Remaining TUI interaction gaps (`R` range editing, quit confirmation)

### Medium Priority (UX Improvements)
1. Mouse support
2. `emphasize` and `section` annotation types
3. FEM escaping and nested-brace handling
4. Formal JSON schema contract + CI validation

### Low Priority (Nice to Have)
1. Additional agent-focused machine-readable error contracts
2. Semantic exit codes
3. Advanced diff-aware workflows
4. MCP integration if CLI-native workflows prove insufficient
