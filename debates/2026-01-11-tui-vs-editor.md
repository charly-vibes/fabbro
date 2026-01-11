# Debate: TUI-First vs Editor-First UX

**Date**: 2026-01-11  
**Status**: Decided - TUI-first  
**Decision**: Build interactive TUI as primary experience, keep `$EDITOR` as escape hatch

---

## The Problem

The original design assumes users will:
1. Open `$EDITOR` with a `.fem` file
2. Know FEM syntax (`{>> comment <<}`, `{-- DELETE --}`, etc.)
3. Type markup manually while reading

**This creates friction**:
- Users must memorize syntax
- No guidance during annotation
- Doesn't work for read-only contexts (Cursor plans, generated docs)
- Loses the "linear reading" benefit the tool is meant to provide

## Options Considered

### Option A: Raw `$EDITOR` (Original Design)

**Pros**: Simple, uses familiar tools  
**Cons**: Requires FEM knowledge, no guidance, high friction

### Option B: LSP + Editor Plugins

**Pros**: Autocomplete, snippets, diagnostics  
**Cons**: Requires per-editor config, doesn't help read-only contexts, complex setup

### Option C: Web UI

**Pros**: Nice visual UX, mouse-driven  
**Cons**: Requires server, browser, adds complexity, violates CLI-first principle

### Option D: Interactive TUI ← CHOSEN

**Pros**:
- Zero config (ships in binary)
- Works anywhere with a terminal
- Guides user through annotation (no syntax to learn)
- FEM becomes implementation detail
- Works for read-only contexts via line-reference mode
- Consistent with beads, `git add -p`, `lazygit` patterns

**Cons**:
- More implementation work than raw editor
- TUI complexity can creep

---

## Decided Approach

### Primary: TUI Mode (default)

```bash
fabbro review --stdin
# or
fabbro review <file>
```

Launches a pager-like TUI:

```
┌─ Review: session-abc123 ─────────────────────────────────┐
│  42 │ This section explains the authentication flow.    │
│  43 │ First, the user provides credentials...           │
│  44 │ [SELECTED ─────────────────────────────────────]   │
│  45 │ The JWT token is then validated by checking...    │
│  46 │ [SELECTED ─────────────────────────────────────]   │
│  47 │ Finally, the session is established.              │
├──────────────────────────────────────────────────────────┤
│ [c]omment [d]elete [e]xpand [q]uestion │ ? for help     │
└──────────────────────────────────────────────────────────┘
```

**Discoverability: Helix-style SPC menu**

Instead of requiring users to memorize keybindings, use a **Space-triggered command palette** (like Helix, which-key):

```
┌─ Review: session-abc123 ─────────────────────────────────┐
│  44 │ The JWT token is then validated by checking...    │
│  45 │ the signature against the public key stored...    │
│  46 │ Finally, the session is established.              │
├──────────────────────────────────────────────────────────┤
│ ┌─ SPC ──────────────────────────────────┐              │
│ │  c  comment     add comment            │              │
│ │  d  delete      mark for deletion      │              │
│ │  e  expand      request more detail    │              │
│ │  q  question    ask a question         │              │
│ │  k  keep        mark as good           │              │
│ │  u  unclear     flag confusion         │              │
│ │ ───────────────────────────────────────│              │
│ │  v  select      start/end selection    │              │
│ │  a  annotations show all annotations   │              │
│ │  w  write       save and exit          │              │
│ └────────────────────────────────────────┘              │
└──────────────────────────────────────────────────────────┘
```

Press `SPC` → menu appears → press letter → action executes.

**Keybindings** (for power users who skip the menu):
- `j/k` or arrows: navigate
- `v`: start/end selection
- `SPC`: open command palette (Helix-style)
- `c/d/e/q/k/u`: direct annotation (when selection active)
- `/`: search
- `?`: help
- `w` or `Ctrl-S`: save and exit
- `Esc` or `Ctrl-C`: exit with save prompt

**Mouse support** (Phase 1.5):
- Click to position cursor
- Click-drag to select range
- Right-click for context menu (same as SPC menu)

**Vim-surround-style context expansion**:
When selection is active, expand/shrink context:
- `ap`: expand to paragraph
- `ab`: expand to code block (fenced ```)
- `as`: expand to section (next heading)
- `{`/`}`: shrink/expand by one line

This lets users quickly grow a selection to include surrounding context.

**Flow**:
1. User navigates to a section
2. Presses `v` to start selection, moves, presses `v` again
3. Presses `c` for comment
4. Prompt appears: "Comment: _____"
5. User types, presses Enter
6. TUI highlights the annotated section
7. Repeat until done
8. Exit → FEM file saved → `fabbro apply` extracts JSON

### Secondary: Line-Reference Mode (for read-only docs)

For contexts where you can't edit the source (Cursor plans, generated docs):

```bash
# Interactive prompting
fabbro annotate --source plan.md

# Shows content with line numbers, prompts for:
# - Lines to annotate
# - Annotation type
# - Comment text
```

Creates sidecar session in `.fabbro/sessions/` that references the source.

### Escape Hatch: Raw Editor

```bash
fabbro review --stdin --editor
# or
fabbro edit <session-id>
```

Opens raw `.fem` file in `$EDITOR` for power users.

---

## Implementation Notes

### Go TUI Libraries

Consider:
- **bubbletea** (Charm): Modern, composable, well-documented
- **tview**: More batteries-included, grid layouts
- **tcell**: Lower-level, more control

**Recommendation**: Start with bubbletea - good balance of simplicity and power.

### FEM as Implementation Detail

Users never see FEM syntax. The TUI:
1. Displays clean content
2. Captures annotations via prompts
3. Writes FEM internally
4. `fabbro apply` parses FEM → JSON for Claude

### Session File Format

```yaml
---
session_id: abc123
created_at: 2026-01-11T10:00:00Z
source: stdin
content_hash: sha256:abc...
---

Original content here with FEM markup inserted by TUI...

{>> [lines 44-46] This section is unclear. What about edge cases? <<}
```

---

## Risk Mitigations

1. **TUI complexity creep**: Keep v1 minimal (one pane, no mouse, no syntax highlighting)
2. **Terminal compatibility**: Test on common terminals, use standard escape sequences
3. **Line drift**: Store content hash, warn on `apply` if source changed

---

## Timeline Impact

- Phase 1 MVP now includes basic TUI (adds ~2 days)
- Worth it: without good UX, the tool won't be used
- Can ship minimal TUI first, iterate on polish

---

## Summary

The TUI approach solves the core pain point (commenting on long outputs without losing place) while maintaining CLI-first principles. FEM becomes an internal format, not a user-facing syntax. Power users retain `$EDITOR` access.
