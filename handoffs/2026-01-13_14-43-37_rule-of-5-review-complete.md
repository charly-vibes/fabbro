---
date: 2026-01-13T14:43:37-03:00
git_commit: 26a852c
branch: main
beads_issues: fabbro-zwk, fabbro-n1f, fabbro-ywn, fabbro-ydd
status: handoff
---

# Handoff: Rule of 5 Review Complete

## Task(s)

**Completed**: Full Rule of 5 review of all 6 specs and 4 plans

## Critical References

1. `specs/01_initialization.feature` - `specs/06_fem_markup.feature` — all specs reviewed
2. `plans/2026-01-11-tracer-bullet.md` — completed tracer implementation
3. `plans/2026-01-12-post-tracer-features.md` — completed post-tracer features
4. `plans/2026-01-12-tech-debt-cleanup.md` — completed tech debt cleanup

## Review Findings Summary

| Severity | Count |
|----------|-------|
| CRITICAL | 3 |
| HIGH | 4 |
| MEDIUM | 3 |

### Top 3 Critical Issues

1. **JSON Schema Inconsistency** — Spec 04 uses `sessionId`, `startLine`/`endLine`; code uses `session_id`, `line`. Breaks Amp integration.

2. **Specs Present Unimplemented Features as Shipped** — Search, mouse, resume, block deletes, config.yaml all specced but not implemented.

3. **TUI Keybinding Contradictions** — `q` used for both quit and question in different docs.

## Issues Created

| ID | Title | Priority | Dependencies |
|----|-------|----------|--------------|
| fabbro-zwk | Canonicalize JSON schema for fabbro apply | P1 | — |
| fabbro-n1f | Align TUI keybindings - single source of truth | P2 | — |
| fabbro-ywn | Create spec-coverage matrix with implementation status | P2 | — |
| fabbro-ydd | Tag spec scenarios as Implemented/Planned/Not Implemented | P2 | fabbro-ywn |

## Recommended Order

1. **fabbro-zwk** (P1, Critical) — JSON schema is blocking Amp integration
2. **fabbro-n1f** (P2) — Keybindings affect UX, could cause data loss
3. **fabbro-ywn** (P2) — Creates the matrix
4. **fabbro-ydd** (P2) — Uses matrix to tag specs (depends on ywn)

## Verdict

**NEEDS_REVISION** — Specs are well-written as vision docs but mix shipped behavior with aspirational features. After canonicalizing JSON and tagging unimplemented scenarios, specs will be accurate.

## Next Steps

```bash
bd ready  # Shows fabbro-zwk, fabbro-n1f, fabbro-ywn (unblocked)
```

Start with `fabbro-zwk` (JSON schema) — highest impact for Amp dogfooding.

## Notes

- No code changes in this session
- All findings from oracle-assisted Rule of 5 methodology
- Full review output available in thread history
