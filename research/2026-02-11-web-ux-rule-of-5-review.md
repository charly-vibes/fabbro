# Research: Rule of 5 UX/Product Design Review — fabbro Web for General Audience

**Date:** 2026-02-11
**Status:** Complete
**Methodology:** Rule of 5 iterative review adapted for product/UX design (5 passes: User Viability, Technical Feasibility, Product Coherence, Competitive & Market, Scope & Prioritization)
**Convergence:** Partial convergence after pass 4; pass 5 still surfaced scope CRITICALs

---

## Executive Summary

The debate doc correctly identifies the core tension (vim-speed vs web-approachability) and makes sound calls (Docs-style default, autosave, rendered markdown). However, it **overestimates what non-technical users will tolerate** (GitHub branches/commits, "snapshot" semantics, source toggles) and **underestimates the complexity of highlight-based annotations with reliable anchoring** in a zero-backend SPA.

**Net recommendation:** You need a *simpler*, more opinionated MVP — "Paste or upload → highlight → comment → copy summary" — and push GitHub/Docs import behind "Advanced" or to a later version.

### Top 3 Findings

1. **UV-001 (CRITICAL):** GitHub repo browsing + "snapshot review" metadata is not usable for non-technical users without heavy framing and defaults
2. **TF-001 (CRITICAL):** "Rendered markdown + highlight annotations mapped to line-based FEM" is the hardest technical problem in the spec and is currently hand-waved as "solvable"
3. **SC-001 (CRITICAL):** MVP is not actually scoped as an MVP — it includes too many hard problems simultaneously

---

## Issue Summary

| Severity | Count | Categories |
|----------|-------|------------|
| CRITICAL | 6 | User Viability (2), Technical Feasibility (2), Scope (2) |
| HIGH | 10 | User Viability (3), Product Coherence (3), Competitive (2), Scope (2) |
| MEDIUM | 11 | All passes |
| LOW | 8 | All passes |

---

## PASS 1: User Viability

*Will non-technical users actually understand and use this?*

### UV-001: "Import from GitHub" assumes developer mental models (CRITICAL)

**Section:** Decision 3a (GitHub Repository Browsing)

**Problem:** Non-technical users don't understand `owner/repo`, branches, file trees, commit SHAs, or why a "snapshot" matters. This creates immediate drop-off.

**Recommendation:**
- Move GitHub import to **Advanced** or v2
- If kept: accept **full URL only** (not `owner/repo`), auto-detect default branch, hide commit SHA
- Provide tooltip: "This imports a copy for review; changes in GitHub won't update this review"

---

### UV-002: Google Docs import sets up expectations you can't meet (CRITICAL)

**Section:** Decision 3c

**Problem:** If users can already comment natively in Docs, why import into fabbro? The moment you offer "Import from Google Docs," users expect round-trip (export annotations back as Docs comments/suggestions). The "import-as-copy" disclaimer will feel like a broken feature.

**Recommendation:**
- Defer Docs import entirely until you can articulate a materially better outcome than native Docs commenting
- Or: frame it purely as "paste text" with a convenience Google Docs fetcher, not as an integration

---

### UV-003: "Rendered markdown with optional source view" presumes markdown literacy (HIGH)

**Section:** Decision 2

**Problem:** "View source / split view" invites confusion — users may not know why the same text appears twice.

**Recommendation:**
- Default to **Reading view only**
- Put "View source" behind Advanced settings, label it "Show markup" (not "source")
- Avoid split view in v1

---

### UV-004: Annotation types still feel "tool-internal" (HIGH)

**Section:** Decision 6

**Problem:** "Keep / Unclear / Expand" read like rubric labels. "Suggest replacement" can be misconstrued as collaborative editing (Docs suggestion mode), which you don't support.

**Recommendation:**
- MVP: reduce to **Comment** and **Suggestion** only
- Add microcopy under toolbar: "Comments don't change the text; suggestions propose replacement text"
- Rename "Expand" → "Needs more detail", "Unclear" → "Confusing" if you add them later

---

### UV-005: "No quit concept" misses that users want a "done" moment (HIGH)

**Section:** Decision 1 implications

**Problem:** Users want a clear "Done / Share / Export / Close review" moment. Without it, reviews feel perpetually unfinished.

**Recommendation:**
- Add a **"Finish review"** action that takes them to the summary/export screen
- Autosave continues in background, but "Finish" is the intentional completion step

---

### UV-006: Autosave-only can reduce trust without visible feedback (MEDIUM)

**Problem:** Non-technical users expect visible confirmation. Autosave without clear "Saved" state creates anxiety.

**Recommendation:**
- Status indicator: "Saving…" → "Saved locally"
- Prominent "Download" action early so users don't fear lock-in to one browser

---

## PASS 2: Technical Feasibility

*Are the proposed implementations achievable in a zero-backend SPA?*

### TF-001: Highlight anchoring over rendered markdown is the core technical risk (CRITICAL)

**Section:** Technical Considerations (Line-based FEM vs Text-range UI)

**Problem:** The spec says "DOM selection → compute line numbers under the hood" + snippet hash. This is underspecified and likely to fail because:
- Markdown rendering changes whitespace/line wrapping vs original text
- Normalization differences (CRLF/LF, HTML entities, list formatting)
- Code blocks with syntax highlighting insert spans, complicating offset calculations
- If anchoring fails, exported FEM references wrong lines → catastrophic trust loss

**Recommendation (simplest viable):**
- Anchor to **plain-text offsets in the original raw text**, not rendered DOM structure
- Keep a hidden "canonical text" buffer (the imported raw content)
- Render from that buffer, but compute selection based on canonical text positions
- Export to FEM by converting offset-ranges → line-ranges at export time, with drift warnings
- **Consider:** MVP could display plain text with minimal formatting (not full rendered markdown) to avoid the entire mapping problem

---

### TF-002: OAuth PKCE "no backend" is operationally brittle (CRITICAL)

**Section:** Auth for SPA, Decision 3a/3c

**Problem:** OAuth redirect flows break "unsaved state" unless carefully handled. Token lifetimes, refresh, revoked access, Google Picker origins, CORS edge cases all create support burden.

**Recommendation:**
- Do not include OAuth import in MVP
- Start with **file upload + paste** only
- If needed later: implement OAuth as a *separate import page* that produces a local copy, then returns to the editor

---

### TF-003: Browser selection APIs over rendered markdown yield messy ranges (HIGH)

**Problem:** Partial node selection, cross-element selection, whitespace normalization — mapping browser `Selection` ranges back to canonical text is non-trivial, especially with rich markdown rendering.

**Recommendation:**
- Constrain selection to a controlled text layer with minimal DOM segmentation
- Or: MVP starts as **plain text/markdown display without rich rendering**, adds rendering later
- This is contrarian but dramatically reduces anchoring failures

---

### TF-004: WASM size/startup may harm first-run conversion (MEDIUM)

**Problem:** 4MB WASM load + runtime cost feels slow for general audience. Syntax highlighting is not core value for reviewing AI prose output.

**Recommendation:**
- Lazy-load WASM highlighting only for code blocks
- Or: use a JS highlighter (Shiki/Prism) for web, skip WASM highlighting entirely
- FEM parsing in JS is ~50 lines of regex — consider whether WASM is worth it for MVP at all

---

### TF-005: IndexedDB-only breaks cross-device and sharing expectations (MEDIUM)

**Problem:** Non-technical users expect "send link to someone." No backend = no shareable URLs for session state.

**Recommendation:**
- Make "Share" explicitly mean **export** (download file or copy summary)
- If mentioning "shareable URLs," clarify "This link works only on this device/browser"

---

## PASS 3: Product Coherence

*Does the feature set tell one clear story?*

### PC-001: Two interaction modes complicate the "simple tool" narrative (HIGH)

**Section:** Decision 1 (Docs + vim)

**Problem:** Supporting two modes isn't just engineering cost — it confuses messaging and onboarding ("Is this a keyboard tool or a Docs-like tool?").

**Recommendation:**
- MVP: ship **one** default interaction model (Docs-like)
- Add keyboard shortcuts gradually, but don't brand it "vim mode" — call it "Keyboard shortcuts"

---

### PC-002: "Snapshot review, not PR diff review" contradicts GitHub import expectations (HIGH)

**Problem:** Users hear "GitHub import" and expect PR review or comments back to GitHub. A "snapshot" is a convenience file loader, so it shouldn't carry GitHub mental-model baggage.

**Recommendation:**
- Rename to "Import a file from GitHub (copy)"
- De-emphasize branches/commit SHA in UI

---

### PC-003: Sidebar "threads" imply collaboration, but tool is single-user (HIGH)

**Section:** Decision 7

**Problem:** Threads + resolved/open + type icons looks like multi-person review (Docs/Figma). If single-user only, creates jarring mismatch.

**Recommendation:**
- MVP: frame as **"Notes list"** not "Threads"
- Avoid multi-author affordances (avatars, timestamps implying others)
- Add explicit "Export to share" action

---

### PC-004: Missing "what is this tool?" framing for the target audience (MEDIUM)

**Problem:** The spec is feature-focused, but general audience needs a clear promise on first visit.

**Recommendation:**
- Add a first-run onboarding screen: 2–3 sentences + single primary action (Paste/upload)
- "Review AI outputs. Annotate with comments and suggestions. Export a structured summary."

---

### PC-005: Export tier hierarchy doesn't match user priority (MEDIUM)

**Problem:** `.fem` download is Tier 2, but for a no-backend tool, "Download your work" may need to be *more* prominent than "Copy summary" (trust/safety).

**Recommendation:**
- "Finish review" view shows two primary CTAs: **Copy summary** + **Download file**
- JSON under Advanced; integrations labeled "Coming later"

---

## PASS 4: Competitive & Market

*How does this compare to existing tools?*

### MK-001: Market expectation is "share link + collaborate" — you can't do that (CRITICAL)

**Problem:** Docs/Figma/Notion/GitHub reviews all support sharing + multi-user. Your UI borrows their patterns (threads, resolved), so users assume the same capabilities.

**Recommendation:**
- Position explicitly as **personal/offline review + export** in v1
- Avoid "collaboration-shaped" UI language until supported
- Turn the limitation into a feature: **"Your reviews stay private. Nothing leaves your browser."**

---

### MK-002: Google Docs import invites direct comparison to native Docs commenting (HIGH)

**Problem:** Users can already comment in Docs. Why import into fabbro? The differentiator must be stronger than "we have annotation types."

**Recommendation:**
- Make the differentiator explicit: "Fabbro turns reviews into structured, exportable feedback summaries — something Docs comments can't do"
- If you can't deliver a materially better outcome, defer Docs import

---

### MK-003: Privacy posture is a competitive advantage, but messaging is absent (MEDIUM)

**Problem:** A no-backend SPA is a *huge* selling point for reviewing sensitive AI outputs, but the spec never mentions it.

**Recommendation:**
- Add persistent footer: "Runs locally in your browser. Your text stays on your device."
- This is potentially the strongest differentiator vs Docs/Notion

---

### MK-004: Missing "AI review" specific workflows competitors emphasize (MEDIUM)

**Problem:** AI output evaluation tools lean on rubrics, scoring, batch review. The spec focuses on document annotation (good) but doesn't clarify the intended single-output workflow.

**Recommendation:**
- Clarify: "Review one AI output at a time; export a summary to send back to the team or feed back to the AI"

---

## PASS 5: Scope & Prioritization

*Is the MVP scoped correctly?*

### SC-001: MVP includes too many hard problems simultaneously (CRITICAL)

**Section:** Decisions 1–7 collectively

**Problem:** The spec's "MVP" includes: complex anchoring (rendered markdown + highlights → line FEM), sidebar threads with resolved state, GitHub import (API browsing), Google Docs import (OAuth + conversion), dual interaction mode (Docs + vim), tiered export formats. This is multiple MVPs.

**Recommended MVP cutline:**

| In MVP | Out of MVP |
|--------|-----------|
| Paste + file upload | GitHub import |
| Plain text / light markdown display | Full rendered markdown |
| Highlight → Comment / Suggest | Question, Delete, Keep, Unclear, Expand |
| Notes list (simple sidebar) | Threaded sidebar with resolved/open |
| Autosave + "Finish review" | Vim mode toggle |
| Copy summary + Download .fem | JSON export, integrations |

---

### SC-002: Imports should be sequenced by user value vs fragility (HIGH)

**Recommended roadmap:**
- **v0:** paste/upload → annotate → export summary
- **v1:** `.fem` round-trip + summary export polish + onboarding
- **v2:** GitHub public-only "fetch a file by URL" (no repo browser)
- **v3:** OAuth imports (GitHub private / Google Docs)

---

### SC-003: Annotation types should be more aggressive reduction (HIGH)

**Recommendation:** MVP types: **Comment** and **Suggestion** only. Add others when you have evidence they're needed. Internally map to FEM types, but don't expose the full taxonomy early.

---

### SC-004: Sidebar threads not necessary to prove value (MEDIUM)

**Recommendation:** MVP has a simple "Notes list" with jump-to-highlight. Skip resolved/open status, skip thread framing.

---

### SC-005: Mobile should be explicitly "no" for MVP (MEDIUM)

**Recommendation:** Message "Desktop-only for now." Ensure it doesn't break catastrophically on mobile (read-only fallback, not a blank page).

---

## Convergence Assessment

### Repeated Themes (strong convergence)

| Theme | Passes | Issues |
|-------|--------|--------|
| GitHub concepts alienate non-technical users | 1, 3, 5 | UV-001, PC-002, SC-002 |
| Collaboration expectations vs no-backend reality | 2, 3, 4 | TF-005, PC-003, MK-001 |
| Highlight anchoring is the make-or-break technical risk | 2, 5 | TF-001, TF-003, SC-001 |
| MVP scope too broad | All | SC-001 reinforced everywhere |
| Privacy as differentiator (under-leveraged) | 4 | MK-003 |

### Diminishing Returns?

Not fully converged: pass 5 produced multiple CRITICAL/HIGH scope issues, indicating the spec's "MVP" definition needs revision before execution begins.

---

## Recommended Next Actions (spec changes, not code)

### Immediate

1. **Rewrite MVP section** with a single core user journey: paste/upload → annotate → finish → export
2. **Explicitly define non-goals for v1:** no collaboration, no share-by-link across devices, no round-trip to GitHub/Docs comments, no vim mode
3. **Specify anchoring strategy** — canonical text offsets, export-time line mapping, drift warnings
4. **Move GitHub/Docs import to "Later"** and adjust language

### Short-term

5. **Tighten annotation types** to Comment + Suggestion for MVP
6. **Add privacy messaging** as a first-class differentiator
7. **Design "Finish review" flow** as the completion moment
8. **Add onboarding screen** with clear value proposition

### Deferred

9. **GitHub import (v2)** — simplified "fetch file by URL" first, repo browser later
10. **Google Docs import (v3)** — only once differentiator is proven
11. **Vim mode** — as opt-in "Keyboard shortcuts" after core UX is validated

---

## References

- Debate doc: `debates/2026-02-11-web-ux-for-general-audience.md`
- WASM research: `research/2026-02-10-wasm-web-spa-integration.md`
- Rule of 5 methodology: Steve Yegge's "Six New Tips for Better Coding with Agents"
- Prior code review: `research/2026-01-14-rule-of-5-code-review.md`
