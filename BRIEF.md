# Project status brief

## Purpose / current state

none

## Focus

none

## In flight

- fabbro-85b2 — release: publish prebuilt binaries for the five targets on GitHub Releases (status: open; priority: 1; age: 2d)
- fabbro-fyy6 — Correct fabbro prime q-key guidance (status: open; priority: 1; age: 106d)
- fabbro-q803 — Prevent edited frontmatter session IDs from orphaning sessions (status: open; priority: 1; age: 106d)
- fabbro-32p — Implement structured JSON error output for agent-facing CLI commands (status: open; priority: 2; age: 216d)
- fabbro-3e7 — Explore session-to-file sync using git (status: open; priority: 2; age: 297d)
- fabbro-4yff — docs: apply README front-door standard v0.4 (§5.1) (status: open; priority: 2; age: 2d)
- fabbro-c66 — Review: editor mode (--editor flag) (status: in_progress; priority: 2; age: 258d)
- fabbro-dqtc — Fix license metadata mismatch: README says MIT, LICENSE is GPL-3.0 (status: open; priority: 2; age: 50d)
- fabbro-kwt — Refactor TUI launch to use dependency injection (status: open; priority: 2; age: 292d)
- fabbro-uk4 — CLI Ergonomics and DX Review (status: open; priority: 2; age: 293d)
- fabbro-14s — FEM Emphasize and Section Markers (status: open; priority: 3; age: 293d)
- fabbro-29t — Web: Inline edit mode for multi-line replacements (status: in_progress; priority: 3; age: 267d)
- fabbro-381 — TUI Edit Annotation Range (status: open; priority: 3; age: 293d)
- fabbro-67b — TUI Mouse Support (status: open; priority: 3; age: 293d)
- fabbro-9g7 — Complete .fabbro directory structure (status: open; priority: 3; age: 293d)
- fabbro-a0u — FEM Parse Error Reporting (status: open; priority: 3; age: 293d)
- fabbro-ayq — Init Command Enhancements (status: open; priority: 3; age: 293d; blocked by: fabbro-9g7, fabbro-bhz, fabbro-e4s, fabbro-ylf)
- fabbro-bhz — Add --quiet flag to init command (status: open; priority: 3; age: 293d)
- fabbro-e24 — FEM Nested Braces and Multi-line Text (status: open; priority: 3; age: 293d)
- fabbro-e4s — Add --agents flag for agent scaffolding (status: open; priority: 3; age: 293d)
- fabbro-ejx — Epic: Pluggable annotation parser architecture (status: open; priority: 3; age: 308d; blocked by: fabbro-gt1, fabbro-kvs, fabbro-nts)
- fabbro-gt1 — Build markdown-comment annotation test corpus (20-50 cases) (status: open; priority: 3; age: 308d)
- fabbro-kvs — Create annotation format evaluation harness (status: open; priority: 3; age: 308d; blocked by: fabbro-gt1)
- fabbro-n41 — Normalize line endings in FEM parser (status: open; priority: 3; age: 295d)
- fabbro-nts — Add token count analysis for FEM vs markdown-comment formats (status: open; priority: 3; age: 308d; blocked by: fabbro-kvs)
- fabbro-rdz — Optimize highlight computation for wrapped lines (status: open; priority: 3; age: 303d)
- fabbro-vnl — FEM Block Delete Syntax (status: open; priority: 3; age: 293d)
- fabbro-ylf — Add subdirectory detection to init (status: open; priority: 3; age: 293d)
- fabbro-zr0 — FEM Escape Sequences (status: open; priority: 3; age: 293d)

## Blocked / waiting

- fabbro-ayq — Init Command Enhancements (status: open; priority: 3; age: 293d; blocked by: fabbro-9g7, fabbro-bhz, fabbro-e4s, fabbro-ylf)
- fabbro-ejx — Epic: Pluggable annotation parser architecture (status: open; priority: 3; age: 308d; blocked by: fabbro-gt1, fabbro-kvs, fabbro-nts)
- fabbro-kvs — Create annotation format evaluation harness (status: open; priority: 3; age: 308d; blocked by: fabbro-gt1)
- fabbro-nts — Add token count analysis for FEM vs markdown-comment formats (status: open; priority: 3; age: 308d; blocked by: fabbro-kvs)

## Next / ready

- fabbro-85b2 — release: publish prebuilt binaries for the five targets on GitHub Releases (status: open; priority: 1; age: 2d)
- fabbro-fyy6 — Correct fabbro prime q-key guidance (status: open; priority: 1; age: 106d)
- fabbro-q803 — Prevent edited frontmatter session IDs from orphaning sessions (status: open; priority: 1; age: 106d)
- fabbro-32p — Implement structured JSON error output for agent-facing CLI commands (status: open; priority: 2; age: 216d)
- fabbro-3e7 — Explore session-to-file sync using git (status: open; priority: 2; age: 297d)
(18 more omitted)

## Parked

none

### Needs re-triage

none

## Recent

- cc194ba2e1dca87e4afe2a07fae8e78221ee60ce docs: apply issue-review fixes to README/release tickets (2026-11-16T11:38:52Z)
- 1ec5005f56f9ed8433475f9918e6c65d780a0dea docs: file README front-door and release tickets (2026-11-16T00:28:42Z)
- 874fed7a91516e7bc5ea24b6e2ebf75b0d98070b beads: issue-review remediation for fabbro-dqtc (files, base_commit, Must gate) (2026-09-29T16:21:01Z)
- 30e129d81bbe146f1b27108209a05153054c76e5 beads: file license-mismatch bug (fabbro-dqtc: README MIT vs LICENSE GPL-3.0) (2026-09-29T16:17:24Z)
- 7cfcc5cc41bf42bc109c0ae54ae0146cb14440ac docs(bd): integration blocks forbid dolt push/pull — embedded + no-db:true, issues.jsonl in git is the sync mechanism (2026-09-29T14:10:55Z)
(195 more omitted)

## Health + confidence

Health: **amber** — basis: signals — owner: none — reason: orphan_commits: 2 commits in the 14-day window link no tracker id or open work-item path: 1ec5005f56f9ed8433475f9918e6c65d780a0dea (docs: file README front-door and release tickets); cc194ba2e1dca87e4afe2a07fae8e78221ee60ce (docs: apply issue-review fixes to README/release tickets)
- signal: orphan commits — 2 commits in the 14-day window link no tracker id or open work-item path: 1ec5005f56f9ed8433475f9918e6c65d780a0dea (docs: file README front-door and release tickets); cc194ba2e1dca87e4afe2a07fae8e78221ee60ce (docs: apply issue-review fixes to README/release tickets)

## Sources & freshness

| source | status | observed_at | staleness | detail |
| --- | --- | --- | --- | --- |
| beads | present | 2026-11-16T11:33:59Z | fresh | 1 malformed line(s) skipped in .beads/issues.jsonl |
| git | present | 2026-11-16T11:38:52Z | fresh | none |
| openspec | missing | none | unknown | openspec/ directory is missing or unreadable |
| whisper | missing | none | unknown | none |
| curation | present | 2026-11-18T21:12:09Z | fresh | none |

Curation: updated_at 2026-11-18T21:12:09Z; confirmed_at none.

Freshness: generated_at 2026-11-18T21:12:09Z; per-source observed_at and staleness are listed above.
