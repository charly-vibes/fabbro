# Debate: Web UX for a General (Non-Technical) Audience

**Date:** 2026-02-11
**Status:** Revised (post Rule-of-5 review)
**Context:** The WASM/vanilla JS SPA research doc (2026-02-10) translates every TUI feature 1:1. This debate reframes the web version for a broader, non-technical audience.
**Review:** See `research/2026-02-11-web-ux-rule-of-5-review.md` for the full Rule of 5 analysis that drove these revisions.

## Core Positioning

**"Review AI outputs privately. Annotate with comments and suggestions. Export a structured summary."**

### Key Differentiator: Privacy

> Runs locally in your browser. Your text never leaves your device.

This is the strongest competitive advantage vs Google Docs, Notion, and other cloud tools for reviewing sensitive AI output. It must be front-and-center in the UI (persistent footer) and marketing.

### What fabbro Web Is NOT (v1 non-goals)

- ❌ Not a collaborative/multi-user review tool (no share-by-link across devices)
- ❌ Not a GitHub PR review tool (no round-trip to PR comments)
- ❌ Not a Google Docs replacement (no round-trip to Docs suggestions)
- ❌ Not a vim-in-the-browser experiment
- ❌ Not mobile-friendly (desktop-only, with graceful read-only fallback on mobile)

## Decision 1: Interaction Model — Docs-Style Only

~~Option B: Google Docs-style review as default, vim mode as opt-in~~

**Revised:** Ship **one** interaction model. No dual modes in v1.

- **Mouse selection** → floating context toolbar ("Comment", "Suggest")
- **Keyboard shortcuts** added gradually (not branded as "vim mode") — accessible via a "Keyboard shortcuts" help page, not a mode toggle
- **Scroll + click** for navigation, not `j/k`

### Core User Flow

```
Select text (drag) → Toolbar appears → Click "Comment" or "Suggest" → Type → Submit
```

### Implications
- Selection is **text highlight** (drag to select), not line-based cursor
- `ap/ab/as` equivalents become toolbar buttons: "Select paragraph", "Select section", "Select code block"
- **Autosave** (IndexedDB) with visible "Saved locally" indicator
- **"Finish review"** button as the intentional completion step → leads to export screen

### Resolved: UV-005, PC-001
*Users need a "done" moment. One interaction mode simplifies messaging and onboarding.*

## Decision 2: Document Display — Reading View Only

~~Option B: Rendered markdown with optional source view~~

**Revised:** Reading view only. No source view toggle in v1.

- **Default:** Rendered markdown (reading-first experience)
- Code blocks get syntax highlighting (JS-based, not WASM — see Technical Decisions)
- Line numbers hidden from the user entirely — annotations anchor to **text highlights**, not line numbers
- "Show markup" available under Settings/Advanced in a later version, labeled clearly (not "View source")

### Resolved: UV-003, TF-003
*Avoids confusing "why do I see this twice?" and reduces anchoring complexity by not needing to support selection in two rendering modes.*

## Decision 3: Import Flows — Paste + Upload Only (v1)

### v1: Paste + File Upload

The only import flows in v1:

**Paste:**
- Large textarea on landing page: "Paste text to review"
- Immediately creates session on submit

**File upload:**
- Drop-zone below paste area: "Or drop a file here"
- Accept `.md`, `.txt`, `.fem` (round-trip), `.go`, `.py`, etc.
- Single file → immediately create session + autosave

**First-run onboarding** (above paste area):
> Review AI outputs. Annotate with comments and suggestions. Export a structured summary.
> Your text stays on your device — nothing is uploaded.

### v2: GitHub "Fetch File by URL" (Later)

Simplified import — NOT a repo browser:
- User pastes a **full GitHub URL** to a file (e.g., `https://github.com/owner/repo/blob/main/README.md`)
- App fetches the raw file content via GH API (public repos only)
- Creates a session with the fetched content
- No branch selector, no file tree, no commit SHA visible
- Labeled: "Import a copy from GitHub" with tooltip: "This imports a snapshot for review"

### v3: GitHub Private Repos + Google Docs (Later)

Only after core value is proven:
- GitHub OAuth PKCE for private repos (isolated import page, not inline)
- Google Docs import via OAuth PKCE + Google Picker (import-as-copy only)
- Both require a clear differentiator vs native commenting in those tools

### Resolved: UV-001, UV-002, TF-002, SC-002, MK-002
*GitHub repo browsing and Docs import are high-effort, high-support, and not essential to proving core value. Paste + upload is the simplest path to "aha moment."*

## Decision 4: Export/Output Flows

### "Finish Review" Screen

When user clicks "Finish review", they see two primary actions:

**Primary CTA 1: Copy Summary**
- Human-readable summary grouped by document sections
- Includes quoted text snippets + annotation content
- Designed to paste into Slack, email, or a document
- One-click copy to clipboard

**Primary CTA 2: Download Your Work**
- `.fem` download (lossless, CLI-compatible round-trip)
- "Download as annotated markdown" option (readable, slightly lossy)

**Advanced (collapsed/hidden):**
- "Export JSON" — same schema as `fabbro apply --json`

**Future (labeled "Coming later"):**
- GitHub PR comments (requires PR diff review mode)
- Google Docs comments (requires range index mapping)

### Resolved: PC-005, SC-001
*Download is equally prominent as copy — critical for trust in a no-backend tool. Users must feel they can always get their work out.*

## Decision 5: What's In and Out of Web v1

| In v1 | Out of v1 |
|-------|----------|
| Paste + file upload | GitHub import (v2) |
| Plain reading view (rendered markdown) | Source/split view |
| Mouse highlight → Comment / Suggest | Question, Delete, Keep, Unclear, Expand |
| Notes list (simple sidebar) | Threaded sidebar with resolved/open |
| Autosave + "Finish review" flow | Vim mode / keyboard shortcut mode |
| Copy summary + Download .fem | JSON export (hidden in Advanced) |
| Privacy messaging | Integrations (GitHub, Docs export) |
| Desktop-only | Mobile support |

## Decision 6: Annotation Types — Comment + Suggest Only (v1)

**Visible in floating toolbar:**
- **Comment** — general feedback on selected text
- **Suggest** — propose replacement text (merges "change" and "inline edit")

**Microcopy under toolbar:**
> Comments add feedback. Suggestions propose replacement text.

**Internal FEM mapping:**
- Comment → `{>> text <<}`
- Suggest → `{=> replacement text <=}` (change type)

**Added later (when usage data justifies):**
- Question ("Ask about this")
- Delete ("Remove this")
- Keep ("This is good")
- Unclear → labeled "Confusing"
- Expand → labeled "Needs more detail"

### Resolved: UV-004, SC-003
*Two actions is maximally simple. Users can always express "delete this" or "this is unclear" as a comment. Structured types are a power-user optimization.*

## Decision 7: Notes List (Not Threads)

~~Sidebar threads with type icon, resolved/open status~~

**Revised:** Simple **Notes list** in a right-side panel:

- Each annotation shows: type badge (Comment/Suggest), text snippet preview, annotation text
- Click note → scroll to highlight in document; click highlight → focus note in list
- No "resolved/open" status in v1 (implies multi-user workflow)
- No avatars, no timestamps (single-user, no collaboration framing)
- Counter: "3 notes" in the panel header

**Later:** Add resolved/open status, filtering, and sorting when collaboration or workflow tracking is added.

### Resolved: PC-003, SC-004, MK-001
*Avoids creating collaboration expectations that can't be met. "Notes" is honest framing for single-user review.*

## Technical Decisions

### Anchoring Strategy (addresses TF-001 — the hardest problem)

The spec previously hand-waved "DOM selection → compute line numbers." This is now fully specified:

**Architecture: Canonical Text Buffer**

```
┌──────────────────────────────────────┐
│           User sees:                 │
│   Rendered markdown (reading view)   │
│   ← mouse selection happens here     │
└──────────┬───────────────────────────┘
           │ Selection API
           ▼
┌──────────────────────────────────────┐
│        Mapping Layer (JS)            │
│   DOM range → canonical text offset  │
│   (each rendered element carries a   │
│    data-offset attribute pointing    │
│    to its position in raw text)      │
└──────────┬───────────────────────────┘
           │
           ▼
┌──────────────────────────────────────┐
│      Canonical Text Buffer           │
│   (original raw text, immutable)     │
│   Annotations stored as:             │
│   { startOffset, endOffset,          │
│     snippetHash, type, text }        │
└──────────┬───────────────────────────┘
           │ At export time
           ▼
┌──────────────────────────────────────┐
│      FEM Line Mapping                │
│   offset → line number conversion    │
│   Drift warning if snippet hash      │
│   doesn't match content at offset    │
└──────────────────────────────────────┘
```

**Key principles:**
1. **Canonical text buffer** — the original raw text is stored immutably. All annotations anchor to byte offsets in this buffer.
2. **`data-offset` attributes** — the markdown renderer tags each text node with its start/end position in the canonical buffer. This is how DOM selections map back to text positions.
3. **Snippet hash** — each annotation stores a hash of the selected text for drift detection.
4. **Line numbers computed at export time** — when generating `.fem` or JSON, convert offsets to 1-based line numbers by counting newlines in the canonical buffer.
5. **No split view in v1** — this eliminates the entire problem of anchoring across two different DOM representations.

**Fallback (if rendered markdown anchoring proves too fragile in practice):**
- Display as lightly-styled plain text (monospace, preserve line breaks, highlight code blocks with background color) instead of full markdown rendering. This makes offset mapping trivial at the cost of visual polish.

### Syntax Highlighting — JS Only (No WASM for v1)

**Decision:** Use a JavaScript syntax highlighter (Shiki or Prism) instead of WASM chroma.

**Rationale:**
- FEM parsing is ~50 lines of regex — reimplement in JS rather than shipping 4MB WASM for MVP
- Syntax highlighting is not core value for reviewing prose (most AI output is text, not code)
- Lazy-load highlighting only when code blocks are present
- Keeps bundle < 500KB total
- WASM can be added later if single-source-of-truth parsing becomes important

**What this means:** The web version has its own FEM parser (JS), not the Go WASM module. The FEM spec is simple enough that divergence risk is low, and we can add conformance tests to catch it.

### Auth — None in v1

No OAuth, no tokens, no API calls. Content comes from paste or file upload only.

### Storage — IndexedDB with Export Safety Net

- Sessions stored in IndexedDB `fabbro-sessions` store
- Autosave on every annotation change with "Saved locally" indicator
- "Download your work" always available — users are never locked in
- Clear messaging: "Your reviews are saved in this browser only"

### Mobile — Desktop-Only with Graceful Degradation

- Explicit "Best on desktop" message on small screens
- Read-only fallback on mobile: can view a previously exported review, but cannot create annotations
- No touch selection support in v1

## Resolved Open Questions

| Question | Decision | Rationale |
|----------|----------|-----------|
| Vim mode as URL param? | No. No vim mode in v1. | Ship one interaction model. |
| Mobile support? | Desktop-only. Read-only fallback. | Touch selection is too fragile. |
| Collaborative review? | Explicitly not in v1. Punt. | Requires backend. Frame as "personal review." |
| Syntax highlighting? | JS (Shiki/Prism), not WASM chroma. | Smaller bundle, simpler build. |
| Web version location? | `web/` directory in this repo. | Shares specs, FEM conformance tests. |

## Roadmap

### v0 — MVP ("Can I review and export?")

Paste/upload → highlight → Comment/Suggest → Notes list → Finish review → Copy summary / Download .fem

- Single interaction model (mouse + floating toolbar)
- Rendered markdown reading view
- Two annotation types (Comment, Suggest)
- Simple notes list sidebar
- Autosave + "Finish review" export screen
- Privacy footer
- Onboarding copy
- Desktop-only

### v1 — Polish ("Is this pleasant to use?")

- Keyboard shortcuts (not modal, not branded as vim)
- Search within document
- More annotation types (Question, Delete, Keep, Confusing, Needs more detail)
- Notes list filtering and sorting
- `.fem` round-trip import (upload .fem from CLI, continue in web)
- Summary export formatting improvements

### v2 — GitHub Import ("Can I review files from repos?")

- "Paste a GitHub URL" → fetch raw file (public repos only)
- No repo browser, no branch selector, no OAuth
- Source file metadata stored for reference

### v3 — Integrations ("Can I connect to my tools?")

- GitHub OAuth PKCE for private repos
- Google Docs import (copy-only, OAuth PKCE + Picker)
- Resolved/open annotation status
- Export to GitHub/Docs comments (only if differentiator is proven)
