# WASM Integration for fabbro in a Zero-Dependency Web SPA

**Date:** 2026-02-10
**Status:** Research

## Executive Summary

fabbro can be partially compiled to WASM today. The **FEM parser** and **highlighter** are pure-computation packages with no OS dependencies and can be exposed to JS immediately. The **TUI** and **session** layers are tightly coupled to the terminal and filesystem respectively, and must be replaced with web-native equivalents. The recommended approach is to compile a thin Go→WASM "core" library and build a vanilla JS/TS SPA around it.

## Architecture Analysis

### Package Compatibility Matrix

| Package | WASM-safe? | Blockers | Notes |
|---------|-----------|----------|-------|
| `internal/fem` | **Yes** | None | Pure `strings` + `regexp`. No I/O. |
| `internal/highlight` | **Yes** | None | Uses `chroma` (pure Go lexer/tokenizer). Output is ANSI escape codes → need to swap to CSS/HTML rendering. |
| `internal/session` | **No** | `os`, `filepath`, `crypto/rand`, `config.GetSessionsDir()` | Filesystem-centric. Needs an in-browser storage adapter (localStorage / IndexedDB). |
| `internal/config` | **No** | `os.Getwd()`, `os.Stat()`, `os.MkdirAll()` | Entirely filesystem. Not needed in web context. |
| `internal/tui` | **No** | `bubbletea`, `bubbles/textarea`, `lipgloss`, ANSI terminal | Terminal UI framework. Must be reimplemented as DOM-based UI. |
| `cmd/fabbro` | **No** | `cobra`, `os.Stdin/Stdout`, `tea.NewProgram()` | CLI entry point. Not applicable to web. |

### What Goes Into the WASM Module

The WASM module should expose two capabilities:

1. **FEM Parse/Render** – Parse content → extract annotations + clean content. Serialize annotations back into FEM syntax.
2. **Syntax Highlighting** – Tokenize source code into `[]{text, color}` spans for the web UI to render with CSS classes instead of ANSI codes.

Estimated WASM binary size: ~2-4 MB (chroma lexer tables are the largest contributor). Can be reduced with `tinygo` or by limiting included lexers.

## Integration Strategy

### Option A: Go `syscall/js` (standard `GOOS=js GOARCH=wasm`)

```
┌─────────────────────────────────────┐
│           Vanilla JS SPA            │
│  ┌──────────┐  ┌─────────────────┐  │
│  │ Editor   │  │ Annotation      │  │
│  │ (DOM)    │  │ Panel (DOM)     │  │
│  └────┬─────┘  └────────┬────────┘  │
│       │                 │           │
│       ▼                 ▼           │
│  ┌──────────────────────────────┐   │
│  │     JS Bridge (wasm_exec.js) │   │
│  └──────────────┬───────────────┘   │
│                 │                   │
│  ┌──────────────▼───────────────┐   │
│  │      fabbro-core.wasm        │   │
│  │  • fem.Parse()               │   │
│  │  • fem.Serialize()           │   │
│  │  • highlight.Tokenize()      │   │
│  └──────────────────────────────┘   │
└─────────────────────────────────────┘
         │
         ▼ (persistence)
   localStorage / IndexedDB
```

**Pros:** Standard Go toolchain, full stdlib regex support, zero external build tools.
**Cons:** Larger binary (~4 MB+), requires `wasm_exec.js` runtime shim, GC overhead.

### Option B: TinyGo WASM

Same architecture but compiled with `tinygo`. Produces ~500 KB-1 MB binaries.

**Pros:** Much smaller binary, faster startup.
**Cons:** TinyGo may have issues with `chroma` (heavy use of reflection and init functions). Needs validation. `regexp` support can be limited in older TinyGo versions.

### Option C: Rewrite core in JS/TS, skip WASM

The FEM parser is ~50 lines of regex. The highlighter can use an existing JS library (e.g., Shiki, Prism). No WASM needed.

**Pros:** Smallest bundle, simplest build, native DOM integration.
**Cons:** Two implementations to maintain, divergence risk.

### Recommendation: Option A first, with Option C as fallback

Option A gives you a single source of truth for FEM parsing (Go), which matters for correctness. If binary size becomes a blocker, Option C is trivial to implement for the parser alone.

## WASM Module API Design

```go
// wasm/main.go — WASM entry point (build tag: js,wasm)
package main

import "syscall/js"

func main() {
    js.Global().Set("fabbro", js.ValueOf(map[string]interface{}{
        "parse":     js.FuncOf(parse),
        "serialize": js.FuncOf(serialize),
        "highlight": js.FuncOf(highlightTokens),
    }))
    <-make(chan struct{}) // keep alive
}

// parse(content: string) → { annotations: [...], cleanContent: string }
func parse(this js.Value, args []js.Value) interface{} { ... }

// serialize(cleanContent: string, annotations: [...]) → string
func serialize(this js.Value, args []js.Value) interface{} { ... }

// highlight(code: string, filename: string) → [{text, color}, ...]
func highlightTokens(this js.Value, args []js.Value) interface{} { ... }
```

### JS Usage (zero-dep SPA)

```html
<script src="wasm_exec.js"></script>
<script>
  const go = new Go();
  WebAssembly.instantiateStreaming(fetch("fabbro-core.wasm"), go.importObject)
    .then(result => {
      go.run(result.instance);
      // fabbro.parse(), fabbro.serialize(), fabbro.highlight() now available
    });
</script>
```

## Web SPA Architecture (No Framework Dependencies)

```
web/
├── index.html              # Single page, loads WASM + CSS + app.js
├── style.css               # All styling (vim-like theme)
├── wasm_exec.js            # Go WASM runtime (copy from Go SDK)
├── fabbro-core.wasm        # Built from wasm/main.go
├── app.js                  # Application entry, router
├── editor.js               # Code viewer with line numbers, selection, annotations
├── annotations.js          # Annotation panel, FEM type picker
├── storage.js              # Session persistence (localStorage/IndexedDB)
├── keybindings.js          # Vim-like key handler (j/k/v/c/d/q/e/u/r/w)
└── highlight.js            # Maps WASM token output → DOM spans with CSS classes
```

### Key Web UI Components

1. **Code Viewer** (`<pre>` + line-numbered `<div>`s) — replaces BubbleTea viewport
2. **Selection system** — CSS classes on line divs, toggled by `v` key
3. **Annotation input** — `<textarea>` overlay, triggered by `c`/`d`/`q` etc.
4. **Annotation markers** — gutter dots + preview panel, like current TUI
5. **Persistence** — `storage.js` wraps IndexedDB, stores sessions as JSON blobs with same frontmatter schema

### Keybinding Parity

The current TUI keybindings (`j/k/gg/G/Ctrl+d/v/c/d/q/e/u/r/w/Space/?/`) can be mapped 1:1 via `document.addEventListener('keydown', ...)`. No framework needed.

## Build Pipeline

```makefile
# In justfile or Makefile
build-wasm:
    GOOS=js GOARCH=wasm go build -o web/fabbro-core.wasm ./wasm/
    cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/

serve:
    python3 -m http.server 8080 --directory web/
```

## Persistence Without Filesystem

| TUI/CLI concept | Web equivalent |
|----------------|----------------|
| `.fabbro/sessions/*.fem` | IndexedDB `fabbro-sessions` store |
| `session.Create()` | `storage.createSession()` → IndexedDB put |
| `session.Load(id)` | `storage.loadSession(id)` → IndexedDB get |
| `session.List()` | `storage.listSessions()` → IndexedDB getAll |
| `config.FindProjectRoot()` | Not needed (single implicit "project") |
| File content input | `<textarea>` paste, drag-and-drop, or File API |

## Effort Estimate

| Task | Effort |
|------|--------|
| Create `wasm/` package with `syscall/js` bridge | 1-2 days |
| Build pipeline (justfile recipes) | 0.5 day |
| Web editor component (code viewer + selection) | 2-3 days |
| Annotation input/display UI | 1-2 days |
| Keybinding system | 1 day |
| IndexedDB persistence layer | 1 day |
| Syntax highlighting CSS mapping | 1 day |
| Testing + polish | 2 days |
| **Total** | **~9-12 days** |

## Spec-by-Spec Feature Coverage Analysis

Cross-referencing all 6 spec files against the web SPA translation. Only `@implemented` features are in scope (we're translating what exists, not building planned features). `@planned` features are noted for awareness.

### Spec 01: Initialization (`01_initialization.feature`)

| Scenario | Status | Web Translation | Notes |
|----------|--------|-----------------|-------|
| Init new project | @partial | **Replace**: No `.fabbro/` dir. On first visit, create IndexedDB database. Auto-init, no explicit command needed. | Trivial — implicit init. |
| Init already initialized | @implemented | **N/A**: IndexedDB is always available. Idempotent by design. | |
| Quiet init | @planned | Skip | |
| Init in subdirectory | @planned | Skip | |
| Agent integration scaffolding | @planned | Skip | |

**Risk: None.** Init is simpler in web context.

### Spec 02: Review Session Creation (`02_review_session.feature`)

| Scenario | Status | Web Translation | Notes |
|----------|--------|-----------------|-------|
| Create from stdin | @implemented | **Replace**: Paste into textarea, or drag-and-drop file. No stdin concept. | Core UX change. |
| Create from file | @implemented | **Replace**: File picker (`<input type="file">`) or drag-and-drop. | Need File API. |
| Session file contains metadata | @implemented | **Keep**: Store same frontmatter schema in IndexedDB. `session_id`, `created_at`, `source_file`. | Direct translation. |
| Session preserves content | @implemented | **Keep**: Content stored verbatim in IndexedDB. | Direct translation. |
| Error: no init | @implemented | **N/A**: No init required in web. | |
| Error: non-existent file | @implemented | **N/A**: Files come from browser, always exist at input time. | |
| Error: no input | @implemented | **Keep**: Validate non-empty input before creating session. | Direct translation. |
| Custom session ID | @planned | Skip | |
| Editor fallback | @planned | Skip (no `$EDITOR` in browser) | |
| Non-interactive mode | @planned | Skip | |

**Risk: Low.** Input source changes (stdin/file → paste/drag-and-drop) but session model is identical.

### Spec 03: TUI Interaction (`03_tui_interaction.feature`)

This is the largest spec and the core of the translation effort.

#### Navigation

| Scenario | Status | Web Translation | Notes |
|----------|--------|-----------------|-------|
| j/k, ↑/↓ navigation | @implemented | **Keep**: `keydown` listener, move `.cursor-active` class between line `<div>`s. | 1:1 mapping. |
| Ctrl+d/Ctrl+u half-page | @implemented | **Keep**: Calculate visible lines from container height, scroll programmatically. | Need `scrollTop` math. |
| gg/G jump first/last | @implemented | **Keep**: Set cursor to 0 or `lines.length - 1`. | 1:1 mapping. |
| zz/zt/zb viewport control | @implemented | **Keep**: `element.scrollIntoView()` with `block: 'center'/'start'/'end'`. | Browser API makes this easier. |
| `/` search + `n`/`N` navigation | @implemented | **Keep**: Fuzzy match logic runs in WASM or JS. Highlight matches with CSS classes. | Fuzzy match in WASM or reimplement in JS (trivial). |

#### Command Palette (SPC Menu)

| Scenario | Status | Web Translation | Notes |
|----------|--------|-----------------|-------|
| Open palette with Space | @implemented | **Keep**: Show overlay `<div>` with annotation options. | CSS overlay. |
| Select action from palette | @implemented | **Keep**: Key handler routes to annotation input. | 1:1 mapping. |
| Dismiss with Esc | @implemented | **Keep**: Hide overlay. | Trivial. |

#### Selection

| Scenario | Status | Web Translation | Notes |
|----------|--------|-----------------|-------|
| Single line `v` | @implemented | **Keep**: Toggle `.selected` CSS class. Track `anchor`/`cursor` in JS state. | 1:1 mapping. |
| Range selection (v + j/k) | @implemented | **Keep**: Extend `.selected` class to range. | 1:1 mapping. |
| Cancel selection (Esc) | @implemented | **Keep**: Clear `.selected` classes. | Trivial. |
| `ap` expand to paragraph | @implemented | **Keep**: Port `FindParagraph()` logic (blank-line boundaries). | Pure logic, WASM or JS. |
| `ab` expand to code block | @implemented | **Keep**: Port `FindCodeBlock()` logic (fence detection). | Pure logic, WASM or JS. |
| `as` expand to section | @implemented | **Keep**: Port `FindSection()` logic (heading detection). | Pure logic, WASM or JS. |
| `{`/`}` shrink/grow | @implemented | **Keep**: Adjust selection boundaries by ±1. | Trivial. |

#### Annotations

| Scenario | Status | Web Translation | Notes |
|----------|--------|-----------------|-------|
| Add comment (c) | @implemented | **Keep**: Show `<textarea>` overlay, create annotation on submit. | UI component change, logic identical. |
| Add delete (d) | @implemented | **Keep**: Same as comment with different type. | |
| Add question (q) | @implemented | **Keep**: Same pattern. | |
| Add expand (e with selection) | @implemented | **Keep**: Same pattern. | |
| Add keep (k via palette) | @implemented | **Keep**: No text prompt needed. | |
| Add unclear (u) | @implemented | **Keep**: Same pattern. | |
| Add change/replace (r) | @implemented | **Keep**: Same pattern, includes `[lines X-Y] ->` prefix. | |
| Cancel input (Esc) | @implemented | **Keep**: Hide overlay, no annotation created. | |
| Text input wraps | @implemented | **Keep**: Native `<textarea>` wraps automatically. | Easier in web. |
| Multiline input (Ctrl+J) | @implemented | **Keep**: `<textarea>` supports multiline natively. Map Enter=submit, Shift+Enter or Ctrl+J=newline, or reverse. | **⚠ UX decision**: Ctrl+J for newline is unusual in web. Consider Shift+Enter=newline, Enter=submit (standard web pattern). |

#### Editing Existing Annotations

| Scenario | Status | Web Translation | Notes |
|----------|--------|-----------------|-------|
| Edit annotation on line (e) | @implemented | **Keep**: Open editor overlay with pre-filled text. | |
| Picker for multiple annotations | @implemented | **Keep**: Show picker overlay listing annotations. | |
| Cancel edit (Esc Esc) | @implemented | **Keep**: Double-Esc or single-Esc to cancel. | |
| No annotation error | @implemented | **Keep**: Show flash message. | |
| Editing annotation range (R) | @planned | Skip | |

#### Viewing Annotations

| Scenario | Status | Web Translation | Notes |
|----------|--------|-----------------|-------|
| ● indicator on annotated lines | @implemented | **Keep**: Render `●` in gutter `<span>`. | 1:1 mapping. |
| Preview panel on hover/cursor | @implemented | **Keep**: Show annotation content in panel below editor. | |
| Multiple annotations count | @implemented | **Keep**: "(X of N annotations)" in preview. | |
| Preview disappears on leave | @implemented | **Keep**: Conditional rendering. | |
| ▐ range highlighting | @implemented | **Keep**: CSS class on lines within annotation range. | |
| Tab cycling through annotations | @implemented | **Keep**: Tab key cycles `previewIndex`. | |

#### Direct Content Editing

| Scenario | Status | Web Translation | Notes |
|----------|--------|-----------------|-------|
| Inline editor (i) | @implemented | **Keep**: `<textarea>` overlay with selected content. | |
| Save inline edit | @implemented | **Keep**: Creates "change" annotation with edited text. | |
| Cancel inline edit | @implemented | **Keep**: Esc Esc or Ctrl+C. | |

#### Save & Exit

| Scenario | Status | Web Translation | Notes |
|----------|--------|-----------------|-------|
| Save (w) | @implemented | **Replace**: Serialize annotations into session, write to IndexedDB. No TUI exit — stay in app. | **⚠ Behavior change**: `w` saves but doesn't navigate away. |
| Quit (Ctrl+C Ctrl+C) | @implemented | **Replace**: "Close session" → return to session list or landing page. No process to exit. | **⚠ Behavior change**: No process exit in browser. |
| Dirty indicator / confirmation | @implemented | **Keep**: Track `dirty` flag, show "unsaved changes" warning. | |

#### Other

| Scenario | Status | Web Translation | Notes |
|----------|--------|-----------------|-------|
| Help panel (?) | @implemented | **Keep**: Show help overlay with keybindings. | |
| Mouse interaction | @planned | Skip (but web makes this trivial to add later) | |
| Annotations list panel (a) | @planned | Skip | |

### Spec 04: Apply Feedback (`04_apply_feedback.feature`)

| Scenario | Status | Web Translation | Notes |
|----------|--------|-----------------|-------|
| Human-readable summary | @implemented | **Keep**: Render annotation list in a "results" view. | DOM-based rendering. |
| JSON output | @implemented | **Keep**: "Export as JSON" button, copies to clipboard or downloads. | |
| All annotation types in JSON | @implemented | **Keep**: Same JSON schema. | |
| Parse all FEM types (comment, delete, question, expand, keep, unclear, change) | @implemented | **Keep**: WASM `fabbro.parse()` — identical behavior. | Single source of truth. |
| Line numbers match original | @implemented | **Keep**: Same 1-indexed, frontmatter-excluded numbering. | |
| Multi-line annotation range | @implemented | **Keep**: `startLine`/`endLine` preserved. | |
| Error: non-existent session | @implemented | **Keep**: Check IndexedDB, show error. | |
| Apply by source file | @implemented | **Keep**: Query IndexedDB by `sourceFile` field. | |
| Apply returns latest session | @implemented | **Keep**: Sort by `createdAt`, return newest. | |
| JSON includes sourceFile | @implemented | **Keep**: Included in export. | |
| Block delete markers | @planned | Skip | |
| Content hash verification | @planned | Skip | |

**Risk: None.** FEM parsing runs in WASM unchanged. JSON export is a UI button instead of CLI flag.

### Spec 05: Session Management (`05_session_management.feature`)

All scenarios are `@planned`. The only implemented session management is `session list` and `session resume`.

| Feature | Status | Web Translation | Notes |
|---------|--------|-----------------|-------|
| Session list (list command) | @implemented (in code, @planned in spec) | **Keep**: Query IndexedDB, render as list view. | |
| Session resume | @implemented (in code, @planned in spec) | **Keep**: Load from IndexedDB, open editor with parsed annotations. | |
| Session delete | @planned | Skip | |
| Session clean | @planned | Skip | |
| Session export | @planned | Skip (but "download .fem" is trivial to add) | |

**Risk: None.** Basic list/resume translates directly.

### Spec 06: FEM Markup Language (`06_fem_markup.feature`)

| Scenario | Status | Web Translation | Notes |
|----------|--------|-----------------|-------|
| All 7 implemented annotation types | @implemented | **Keep**: WASM `fabbro.parse()` handles all. | Zero risk — same Go code. |
| Multiple annotations per line | @implemented | **Keep**: Same parsing. | |
| Empty annotation text | @implemented | **Keep**: Same parsing. | |
| Whitespace handling | @implemented | **Keep**: Same parsing. | |
| YAML frontmatter parsing | @implemented | **Keep**: Same parsing in WASM. | |
| Overlapping annotations | @implemented | **Keep**: Same model. | |
| Emphasize `{** **}` | @planned | Skip | |
| Section `{## ##}` | @planned | Skip | |
| Escape syntax | @planned | Skip | |
| Multi-line annotations | @planned | Skip | |
| Nested braces | @planned | Skip | |
| Unclosed markers | @planned | Skip | |

**Risk: None.** FEM parsing is the strongest case for WASM — zero translation needed.

---

## Summary: Feature Translation Coverage

### Fully Translatable (zero risk)
- All 7 FEM annotation types (parsing + serialization)
- All navigation keybindings (j/k/gg/G/Ctrl+d/Ctrl+u/zz/zt/zb)
- Selection system (v, ap, ab, as, {, })
- All annotation operations (c/d/q/e/u/r/k + palette)
- Annotation preview + range highlighting + Tab cycling
- Inline editing (i) with change annotation creation
- Edit existing annotations (e without selection)
- Multi-annotation picker
- Search (/n/N)
- Help panel (?)
- Session creation + persistence
- Session list + resume
- Apply/export as JSON

### Behavior Changes (low risk, intentional)
| CLI/TUI Behavior | Web Behavior | Reason |
|------------------|-------------|--------|
| `w` saves and exits TUI | `w` saves, stays in editor | No process to exit |
| `Ctrl+C Ctrl+C` quits process | `Ctrl+C Ctrl+C` closes session / returns to list | No process |
| stdin / file arg input | Paste / drag-drop / file picker | No stdin in browser |
| `.fabbro/sessions/` on disk | IndexedDB `fabbro-sessions` store | No filesystem |
| `fabbro init` explicit | Auto-init on first use | No CLI |
| Ctrl+J for newline in input | Shift+Enter (web convention) or keep Ctrl+J | UX preference |

### Not Translatable (intentionally excluded)
- Shell completion (`fabbro completion`)
- `$EDITOR` fallback
- Agent scaffolding (`--agents`)
- `fabbro prime` (AI context output — CLI-only concern)
- `fabbro tutor` (could be translated but low priority)

### Features Gained by Web
- Shareable URLs (session ID in URL hash)
- Copy annotation JSON to clipboard (one click)
- Drag-and-drop file input
- Potential for mouse selection (planned in CLI, trivial in web)
- Right-click context menu (planned in CLI, native in web)
- No install required

## Open Questions

1. **Binary size tolerance?** — Standard Go WASM is 4 MB+. Acceptable for a dev tool? Could lazy-load.
2. **Offline-first?** — Service worker for full offline? Or just in-memory + localStorage?
3. **Export/import?** — Should the web version be able to export `.fem` files for CLI consumption (round-trip)?
4. **Shared codebase vs fork?** — Should the web version live in this repo (`web/` dir) or a separate repo?
5. **Mobile support?** — Touch events for selection? Or desktop-only for now?
6. **Tutor translation?** — Port `fabbro tutor` as an interactive web tutorial? Good for onboarding but not critical.
7. **`w` behavior?** — Save-and-stay vs save-and-close? Or make it configurable?
8. **Syntax highlighting strategy?** — Use WASM chroma (heavier binary) or switch to a JS highlighter like Shiki/Prism (lighter, separate dep)?
