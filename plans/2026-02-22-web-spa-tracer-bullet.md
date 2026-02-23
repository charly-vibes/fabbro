# Web SPA Tracer Bullet — Fetch → Annotate → Export

## Goal

Prove the core user journey end-to-end with the thinnest possible slice:
**Paste a GitHub URL → Fetch file content → Select text → Add one comment → Copy summary**

No IndexedDB, no file upload, no sidebar, no autosave. Just the critical path through every layer (app shell, fetch, FEM parser, viewer, editor, toolbar, export) with real data flowing through.

**Parent plan**: `plans/2026-02-18-web-spa-v0.md`

> **Note — parent plan divergence**: The parent plan lists "No GitHub import" as a v0
> non-goal. This tracer bullet intentionally adds GitHub URL fetch as a v0+ experiment
> to validate that the GitHub API works from the browser without a proxy (CORS-free for
> public repos). If successful, the parent plan's non-goals should be updated.

---

## What "Done" Looks Like

1. `just serve-web` opens a page at `localhost:8080`
2. User pastes a URL into the input:
   - **GitHub file URL** (e.g. `https://github.com/owner/repo/blob/main/README.md`) → fetched via GitHub API (public repos only)
   - **Any Cloudflare-enabled page URL** → fetched as markdown via `Accept: text/markdown` header
   - Or pastes raw text directly into a textarea (fallback)
3. App fetches content, creates in-memory session, renders plain-text viewer
4. User drags to select text → floating toolbar appears with "Comment" button
5. User clicks "Comment" → inline textarea → types → Enter → annotation created
6. Annotation highlight (`<mark>`) appears on the selected range
7. User clicks "Finish review" → copyable summary with quoted snippet + line numbers + comment text
8. "Copy" button copies summary to clipboard

**Not in scope**: Suggest/change annotations, notes sidebar, IndexedDB persistence, file upload (drag-drop), `.fem` download, mobile detection, session management, privacy footer.

> **Privacy note**: URL fetching makes outbound network requests to GitHub/third-party
> sites. The landing page should NOT claim "your text never leaves your device" — that
> applies to the paste-text flow only. When fetching URLs, the request goes to the
> remote server directly from the user's browser.

---

## Import Strategies

### Strategy 1: GitHub File URL (primary — no CORS issues)

GitHub API has full CORS support (`Access-Control-Allow-Origin: *`), so browser `fetch()` works directly for public repos.

**URL parsing**: Extract `owner`, `repo`, `branch`, `path` from GitHub blob URLs:
```
https://github.com/{owner}/{repo}/blob/{branch}/{path}
→ regex: /github\.com\/([^/]+)\/([^/]+)\/blob\/([^/]+)\/(.+)/
```

> **Known limitation**: Branch names containing slashes (e.g. `feature/foo`) are
> ambiguous in GitHub blob URLs — the regex cannot distinguish branch from path segments.
> For the tracer bullet, this works reliably for `main`, `master`, and simple branch
> names. Refs with slashes may fail. A future improvement could use the GitHub Refs API
> to resolve ambiguity.

**Fetch**:
```js
// Raw file content — no base64 decoding needed
const res = await fetch(
  `https://api.github.com/repos/${owner}/${repo}/contents/${path}?ref=${branch}`,
  { headers: { 'Accept': 'application/vnd.github.v3.raw' } }
);
const content = await res.text();
```

**Rate limits**: 60 req/hr unauthenticated. Fine for a review tool (one fetch per session). Display remaining limit from `x-ratelimit-remaining` header.

**Source metadata**: Store `sourceFile` as the GitHub path (e.g. `owner/repo/path/to/file.md`) for FEM frontmatter.

### Strategy 2: Cloudflare Markdown for Agents (opportunistic — CORS-dependent)

Cloudflare's `Accept: text/markdown` content negotiation converts HTML→markdown at the edge. Works for any Cloudflare-enabled site (their docs, blog, many others).

**CORS reality**: Most websites don't set `Access-Control-Allow-Origin` for browser requests. This means direct `fetch()` from the SPA will be blocked by the browser for most URLs.

**v0 approach — try it, graceful fallback**:
```js
try {
  const res = await fetch(url, {
    headers: { 'Accept': 'text/markdown, text/html' }
  });
  // If we get here, CORS was allowed
  const contentType = res.headers.get('content-type');
  if (contentType?.includes('text/markdown')) {
    return { content: await res.text(), source: url, format: 'markdown' };
  }
  // Got HTML — could still be useful as plain text
  return { content: await res.text(), source: url, format: 'html' };
} catch (e) {
  // CORS blocked — show error with instructions
  throw new Error('This URL blocked cross-origin requests. Try a GitHub file URL or paste text directly.');
}
```

**Known working sites** (Cloudflare-enabled + CORS-friendly):
- `developers.cloudflare.com/*` — has CORS headers
- Other sites: hit-or-miss

**Future (v1)**: A Cloudflare Worker proxy (~10 lines) would solve CORS for all sites. Not needed for tracer bullet.

### Strategy 3: Paste raw text (always works)

Large textarea: "Or paste text directly". This is the escape hatch — always available, zero network dependencies.

---

## Architecture

### Deployment Layout (GitHub Pages)

```
/ (root)            ← web SPA (static files, served directly)
├── index.html
├── style.css
├── app.js, fetch.js, fem.js, viewer.js, toolbar.js, export.js
└── docs/           ← Hugo docs site (built into subdirectory)
    ├── index.html
    ├── fem/
    ├── cli/
    └── ...
```

**Current state**: Hugo docs deploy to root (`/`) from `site/` directory.
**New state**: Web SPA owns root (`/`), Hugo docs move to `/docs/`.

### Source Layout

```
web/                # SPA source (static, no build step)
├── index.html
├── style.css
├── .nojekyll       # Prevent GitHub Pages Jekyll processing
├── app.js          # App shell: landing → editor → export (in-memory state)
├── fetch.js        # URL detection + GitHub API / Cloudflare fetch
├── fem.js          # FEM parse() + serialize() — port of Go parser
├── fem.test.html   # FEM conformance tests (browser-based, PASS/FAIL)
├── viewer.js       # Plain-text renderer with data-start offsets
├── toolbar.js      # Floating "Comment" button near selection
└── export.js       # Summary renderer + clipboard copy

site/               # Hugo docs source (unchanged)
├── hugo.toml       # baseURL stays at .../fabbro/ (see Phase 5a)
├── content/
└── go.mod
```

### GitHub Pages Workflow (`.github/workflows/docs.yml`)

The workflow merges both outputs into a single Pages artifact:

```
1. Build Hugo docs → site/public/
2. Create combined output directory:
   - Copy web/* → _site/           (SPA at root)
   - Copy site/public/* → _site/docs/  (Hugo docs under /docs/)
3. Upload _site/ as Pages artifact
```

All state lives in a single in-memory `session` object passed between modules. No storage layer.

---

## Phase 1: Scaffold + URL Fetcher + FEM Parser (~2-3h)

### 1a: Project scaffold

**File: `web/index.html`**
- HTML5 boilerplate with `<div id="app"></div>`
- `<script type="module" src="app.js">`
- `<link rel="stylesheet" href="style.css">`

**File: `web/style.css`**
- CSS custom properties: `--font-mono`, `--color-highlight`, `--color-toolbar-bg`
- `.landing` layout: centered container, URL input + textarea
- `.viewer` layout: monospace, line numbers gutter, scrollable
- `.toolbar` floating: absolute positioned, shadow, rounded
- `.highlight` mark styling: yellow background
- `.export` layout: preformatted summary block

**File: `web/app.js`**
- Three views managed by swapping `#app` innerHTML:
  - `landing`: URL input + paste textarea + "Start review" button
  - `editor`: viewer + toolbar (imports viewer.js, toolbar.js)
  - `export`: summary + copy button (imports export.js)
- In-memory session: `{ content, sourceUrl, annotations: [] }`
- On "Start review":
  1. If URL provided → call `fetchContent(url)` from fetch.js
  2. If textarea has text → use directly
  3. Normalize CRLF→LF, store content, render editor
- On "Finish review": render export view
- On "Back to review": render editor view again

### 1b: URL fetcher

**File: `web/fetch.js`**

```js
// Detect URL type and fetch content
export async function fetchContent(url)
// Returns { content: string, source: string, filename: string }

// Internal helpers:
function parseGitHubUrl(url)
// Returns { owner, repo, branch, path } or null

async function fetchGitHub(owner, repo, branch, path)
// GET api.github.com/repos/{owner}/{repo}/contents/{path}?ref={branch}
// Accept: application/vnd.github.v3.raw
// Returns raw text content

async function fetchMarkdown(url)
// GET url with Accept: text/markdown, text/html
// Returns markdown or throws on CORS error
```

**Error states**:
- GitHub 404 → "File not found. Check the URL."
- GitHub rate limit → "GitHub API rate limit reached. Try pasting text directly."
- GitHub file too large → "File too large for the GitHub API. Try pasting text directly."
- CORS error → "This URL blocked cross-origin requests. Try a GitHub URL or paste text."
- Network error → "Could not fetch URL. Check your connection."

**Landing page content**:
```
Review any text. Annotate with comments. Export a structured summary.

[URL input: "Paste a GitHub file URL or any web page URL"]
[Start review button]

— or —

[Large textarea: "Paste text directly"]
[Start review button]
```

### 1c: FEM parser

**File: `web/fem.js`**

Port from `internal/fem/markers.go` + `internal/fem/parser.go`:

```js
export const ANNOTATION_TYPES = [
  { name: 'comment',  open: '{>>', close: '<<}' },
  { name: 'delete',   open: '{--', close: '--}' },
  { name: 'question', open: '{??', close: '??}' },
  { name: 'expand',   open: '{!!', close: '!!}' },
  { name: 'keep',     open: '{==', close: '==}' },
  { name: 'unclear',  open: '{~~', close: '~~}' },
  { name: 'change',   open: '{++', close: '++}' },
];

export function parse(content) → { annotations, cleanContent }
export function serialize(content, annotations) → string (FEM with frontmatter)
```

**Critical: match Go parsing semantics exactly** (from `internal/fem/parser.go`):
- **Parse line-by-line** (`content.split('\n')`) — do NOT use whole-string regex. Go's parser iterates per line; cross-line matches must not occur.
- **Match on original `line`, replace on mutable `cleanLine`**: For each annotation type, find matches in the original line, but apply `replaceAll` to cleanLine. This matters for nested/overlapping cases.
- **ReplaceAll per accepted match**: Once a match of a type is accepted (passes nested-marker check), `replaceAll` removes *all* occurrences of that type's pattern on the line — even ones that were skipped. This is current Go behavior and must be replicated for conformance.
- **Nested marker detection**: `containsNestedMarker(matchText)` checks if the captured text contains any opening delimiter from any type. If so, skip that match.
- **Regex pattern**: `open + \s*(.*?)\s* + close` (non-greedy, whitespace-trimmed). Ensure JS regex matches Go's `\s` semantics (they do for ASCII whitespace).

Key behaviors to match (from `parser_test.go`):
- Single comment extraction: `"Hello {>> comment <<} world"` → text `"comment"`, clean `"Hello  world"`
- All 7 types parse correctly
- Multiple annotations on same line
- Unbalanced markers preserved in clean content
- Nested markers skipped (inner valid type extracted, outer with nested opening marker skipped)
- Empty/whitespace-only annotation text trimmed to `""`
- Frontmatter: `---\nsession_id: ...\ncreated_at: ...\nsource_file: ...\n---\n\n` format matching `session.go`
- YAML single-quote escaping: `quoteYAMLString()` matches Go behavior

**Verification**: `web/fem.test.html` — a simple test page that runs FEM parse/serialize assertions in the browser and logs PASS/FAIL for each test vector from `parser_test.go`. This is the executable spec for FEM conformance (see TDD/SDD note below).

### Success criteria
- [ ] `just serve-web` serves the landing page
- [ ] Paste a GitHub URL → content fetched and displayed (editor view, blank highlights for now)
- [ ] Paste a Cloudflare docs URL → content fetched as markdown (if CORS allows), otherwise graceful error
- [ ] Paste raw text + "Start review" → transitions to editor view
- [ ] `fem.parse()` passes all test cases from Go `parser_test.go` (verified via `fem.test.html`)
- [ ] `fem.serialize()` produces valid frontmatter format
- [ ] `fem.test.html` runs in browser with all PASS results
- [ ] Error states shown clearly (404, rate limit, CORS)

---

## Phase 2: Plain-Text Viewer + Selection Offsets (~1-2h)

**File: `web/viewer.js`**

```js
export function render(container, content)
// Splits content by \n, creates <div class="line" data-start="N"> per line
// Gutter with line numbers, monospace text

export function getOffsets(container, selection)
// Returns { startOffset, endOffset } into canonical content string
// Uses data-start attributes on line divs + selection anchor/focus text offsets
```

Rendering approach:
- Split content on `\n`
- Each line is a `<div class="line" data-start="{cumulative offset}">`
- Line number in `<span class="gutter">`
- Text in `<span class="text">`

**Offset computation** (must remain correct after `<mark>` highlights are added in Phase 3):

Before highlights, displayed text == canonical text and offset computation is trivial.
After highlights, `<mark>` elements split text nodes within `.text` spans, so
`selection.anchorOffset` is relative to a *specific text node*, not the line string.

Robust approach using `Range` (works with or without `<mark>`):
1. Get `range = selection.getRangeAt(0)` (already normalized start < end)
2. For start: find ancestor `.line` div, read `data-start`
3. Walk text nodes in document order within the `.text` span up to `range.startContainer`
4. Sum lengths of preceding text nodes + `range.startOffset` within container
5. Result: `data-start + accumulated text offset` = canonical offset
6. Repeat for end
7. Clamp to content bounds

This tree-walk approach is the **critical correctness requirement** — it ensures
`getOffsets()` works identically before and after highlights are added.

### Success criteria
- [ ] Content renders with line numbers, monospace, scrollable
- [ ] `getOffsets()` returns correct offsets — `content.slice(start, end)` matches selected text
- [ ] Works for single-line and multi-line selections

---

## Phase 3: Selection → Comment Annotation (~2-3h) ⚠️ CRITICAL PATH

**File: `web/toolbar.js`**

```js
export function show(rect, onComment)
// Positions floating toolbar near selection rect
// Single button: "💬 Comment"
// Dismissed on click outside or Esc

export function hide()
```

**Additions to `web/app.js` (editor view)**:
- On `mouseup` in viewer: check `window.getSelection()`
- If non-empty selection: compute offsets via `viewer.getOffsets()`, show toolbar
- On "Comment" click: show inline `<textarea>` below selected text
- On Enter in textarea: create annotation `{ type: 'comment', text, startOffset, endOffset }`, push to `session.annotations`
- Re-render viewer with `<mark>` highlights wrapping annotated ranges
- On Esc: dismiss textarea and toolbar

**Highlight rendering**:
- After annotation is added, re-render lines that contain annotations
- Split text nodes at annotation boundaries, wrap annotated text in `<mark>`
- `getOffsets()` must still work correctly with `<mark>` elements present — this is
  guaranteed by the Range-based text-node-walking approach described in Phase 2
- **Test carefully**: add one annotation, verify offsets still work, then add a second
  annotation on the same line and on a different line — verify offsets for both

### Success criteria
- [ ] Select text → toolbar appears near selection
- [ ] Click "Comment" → textarea appears → type → Enter → annotation stored
- [ ] Yellow highlight visible on annotated text
- [ ] Multiple annotations don't break offset computation
- [ ] Esc dismisses toolbar/textarea

---

## Phase 4: Finish Review + Copy Summary (~1h)

**File: `web/export.js`**

```js
export function render(container, session, { onBack })
// Renders human-readable summary:
//
// ## Review Summary
//
// ### Line 5-7
// > "The original selected text..."
// **Comment:** This paragraph needs more context.
//
// ---
// 1 note total (1 comment)
//
// [Copy to clipboard] [Back to review]
```

- Convert offsets to line numbers: count `\n` in `content.slice(0, offset)`
- Quote snippet: first 80 chars of `content.slice(startOffset, endOffset)`, truncate with `…`
- "Copy" button uses `navigator.clipboard.writeText()` with fallback messaging
  (clipboard API requires HTTPS in production — fine on GitHub Pages; on localhost
  it may require a secure context or fall back to showing a "select all" textarea)
- "Back to review" returns to editor with state preserved

### Success criteria
- [ ] Summary shows correct line numbers, quoted text, comment
- [ ] "Copy" copies formatted text to clipboard
- [ ] "Back" returns to editor with annotations intact
- [ ] Multiple annotations appear in document order

---

## Phase 5: Deployment + justfile + Smoke Test (~1-2h)

### 5a: Hugo baseURL — NO CHANGE needed

**File: `site/hugo.toml`** — keep `baseURL` as-is:
```toml
baseURL = "https://charly-vibes.github.io/fabbro/"
```

The Hugo theme uses `BookSection = "docs"` which already generates content URLs under
`/docs/`. Changing `baseURL` to include `/docs/` would produce double-prefixed URLs
(`/docs/docs/...`) and break theme assets/links. The only change needed is the workflow
copy step (5b) which places Hugo output into `_site/docs/`.

### 5b: GitHub Pages workflow

**File: `.github/workflows/docs.yml`** — updated to deploy SPA + docs:

```yaml
name: Deploy Site
on:
  push:
    branches: [main]
    paths: ['site/**', 'web/**', '.github/workflows/docs.yml']
  workflow_dispatch:

# ... (permissions, concurrency unchanged)

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
          cache: false
      - uses: peaceiris/actions-hugo@v3
        with:
          hugo-version: 'latest'
          extended: true

      # Build Hugo docs
      - working-directory: site
        run: hugo --minify

      # Assemble combined site
      - run: |
          mkdir -p _site/docs
          cp -a web/. _site/              # SPA at root (includes dotfiles like .nojekyll)
          cp -a site/public/. _site/docs/  # Hugo docs under /docs/ (copy contents, not dir)

      - uses: actions/upload-pages-artifact@v3
        with:
          path: _site

  deploy:
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    needs: build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/deploy-pages@v4
        id: deployment
```

The `paths` trigger now includes `web/**` so SPA changes also trigger deployment.
The `deploy` job is preserved from the existing workflow (was omitted in the original plan draft).

> **Note**: Use `cp -a src/. dest/` (not `cp -r src/* dest/`) to correctly copy contents
> including dotfiles and avoid directory nesting surprises.

### 5c: justfile

**File: `justfile` (addition)**:
```
serve-web:
    python3 -m http.server 8080 --directory web
```

### 5d: Manual smoke test

1. `just serve-web`
2. Paste `https://github.com/charly-vibes/fabbro/blob/main/README.md` → content loads
3. Select a heading → Comment "rephrase this"
4. Select a code block → Comment "add error handling"
5. Click "Finish review"
6. Copy summary → paste somewhere → verify it reads correctly
7. Try a Cloudflare docs URL → see if markdown fetch works
8. Try pasting raw text → verify direct input works

### Success criteria
- [ ] Full journey: GitHub URL → 2 comments → finish → copy → readable summary
- [ ] Raw text paste fallback works
- [ ] No JS errors in console
- [ ] Works in both Firefox and Chrome
- [ ] Hugo docs still build with unchanged baseURL
- [ ] Workflow assembles `_site/` with SPA at root + docs at `/docs/`
- [ ] ES module imports use relative paths (work on both localhost and GH Pages)

---

## Summary

| Phase | Deliverable | Est. |
|-------|-------------|------|
| 1 | Scaffold + URL fetcher + FEM parser + `fem.test.html` | 2-3h |
| 2 | Plain-text viewer + robust offset computation | 1-2h |
| 3 | Selection → Comment + highlights (⚠️ critical path) | 2-3h |
| 4 | Export + Copy | 1h |
| 5 | Deployment (GH Pages) + justfile + smoke test | 1-2h |

**Total: ~7-11 hours (1-2 days)**

> Previous estimate was 3.25h — revised upward based on: FEM conformance edge cases,
> robust offset computation with `<mark>` tree-walking, and GitHub Pages deployment
> debugging. Phase 3 is the biggest risk.

## What This Proves

- GitHub API CORS works from browser — no proxy needed for public repos
- Cloudflare markdown fetch works when CORS allows (graceful degradation otherwise)
- ES modules work without a build step
- FEM parser in JS produces correct output matching Go
- Plain-text viewer keeps offset computation tractable (Range-based tree-walk, validates v0 design)
- Selection API → annotation → highlight pipeline works end-to-end
- Export produces a useful, copyable summary

## What Comes Next (after tracer bullet)

1. **Suggest annotations** (add "Suggest" button to toolbar, `{++ ++}` type)
2. **Notes sidebar** (`notes.js` — Phase 6 of parent plan)
3. **IndexedDB persistence** (`storage.js` — Phase 3 of parent plan)
4. **File upload** (drag-and-drop zone on landing page)
5. **Cloudflare Worker proxy** (solves CORS for arbitrary URLs)
6. **`.fem` download** (serialize + Blob download in export view)
7. **Polish** (autosave indicator, session resume, mobile fallback, privacy footer)

## Risks

| Risk | Mitigation |
|------|-----------|
| GitHub API rate limit (60/hr unauthenticated) | Show remaining quota; paste fallback always available |
| GitHub URL parsing fails for slash-branches (`feature/foo`) | Document limitation; works for simple branch names; paste fallback |
| GitHub private repos won't work (no auth) | Public repos only; error message suggests paste fallback |
| GitHub Contents API file size limit | Error message "file too large"; paste fallback |
| Cloudflare markdown CORS blocked on most sites | Graceful error message; GitHub URL and paste always work |
| Offset computation breaks after `<mark>` highlights | Range-based text-node tree-walk (Phase 2); explicit test cases in Phase 3 |
| FEM JS parser diverges from Go | `fem.test.html` conformance suite matching `parser_test.go` vectors |
| FEM per-line vs whole-string regex mismatch | Port Go's exact per-line loop; do not use whole-string regex |
| Clipboard API fails on localhost (insecure context) | Fallback textarea with "select all" prompt |
| Hugo docs URL breakage from baseURL change | DO NOT change baseURL; only change workflow copy destination |

## TDD/SDD Alignment

This tracer bullet uses lighter-weight testing than the main project's Gherkin-driven
TDD, which is appropriate for a throwaway proof-of-concept. However, it still includes:

- **`web/fem.test.html`** — executable conformance spec for the FEM parser, running all
  test vectors from `internal/fem/parser_test.go` in the browser with PASS/FAIL output.
  This is the minimum viable test artifact.
- **Manual smoke test checklist** (Phase 5d) as the integration test.
- If the tracer bullet graduates to production code, proper Gherkin specs should be added.
