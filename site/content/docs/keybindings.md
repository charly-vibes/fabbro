---
title: "Keybindings"
weight: 30
---

# TUI Keybindings Reference

This is the **authoritative source** for fabbro TUI keybindings.

## Navigation

| Key | Action |
|-----|--------|
| `j` / `↓` | Move cursor down |
| `k` / `↑` | Move cursor up |
| `Ctrl+d` | Scroll down half page |
| `Ctrl+u` | Scroll up half page |
| `gg` | Jump to first line |
| `G` | Jump to last line |
| `zz` | Center cursor line in viewport |
| `zt` | Move cursor line to top of viewport |
| `zb` | Move cursor line to bottom of viewport |

## Search

| Key | Action |
|-----|--------|
| `/` | Open search prompt |
| `Enter` | Perform search and jump to first match |
| `Esc` | Cancel search (in search mode) / Clear search results (in normal mode) |
| `n` | Jump to next match |
| `N` / `p` | Jump to previous match |

Search uses fuzzy matching (characters must appear in order, but not necessarily adjacent).
Matches are highlighted with a `◎` indicator. The current match is shown in orange, other matches in yellow.
A match counter (e.g., `2/13`) is displayed in the status bar.

## Selection

| Key | Action |
|-----|--------|
| `v` | Toggle line selection (start/clear) |
| `Esc` | Clear selection |
| `ap` | Expand selection to paragraph (blank-line delimited) |
| `ab` | Expand selection to code block (fenced ``` markers) |
| `as` | Expand selection to section (heading to next heading) |
| `{` | Shrink selection by one line |
| `}` | Grow selection by one line |

## Annotations (require selection)

Direct keys (normal mode with selection):

| Key | Annotation Type | Prompt |
|-----|-----------------|--------|
| `c` | comment | "Comment:" |
| `d` | delete | "Reason for deletion:" |
| `q` | question | "Question:" |
| `e` | expand | "What to expand:" |
| `u` | unclear | "What's unclear:" |
| `r` | change | "Replacement text:" |

Palette-only keys (press `Space` first):

| Key | Annotation Type | Prompt |
|-----|-----------------|--------|
| `k` | keep | "Reason to keep:" |

## Command Palette

| Key | Action |
|-----|--------|
| `Space` | Open annotation palette |
| `Esc` | Close palette |

## Save & Exit

| Key | Action |
|-----|--------|
| `w` | Save session |
| `Ctrl+C Ctrl+C` | Quit (with confirmation prompt) |
| `Space` → `Q` | Force quit immediately (no confirmation) |

Quit confirmation shows "Quit? [y/n]" or "⚠ Unsaved changes! Quit anyway? [y/n]" if dirty.

## Input Mode

| Key | Action |
|-----|--------|
| `Enter` | Submit annotation |
| `Esc` | Cancel input |
| `Backspace` | Delete character |

---

## Design Notes

- **Case-sensitive**: `q` = question, `Q` = quit
- **Vim-inspired**: Navigation uses `hjkl` style, `gg`/`G` for jumps
- **Helix-inspired**: `Space` opens discoverable command palette

## Annotation Preview

When the cursor is on an annotated line (shown with `●` indicator), the annotation content is automatically displayed in a preview panel below the code area. This replaces the help text.

- Shows annotation type, line range, and full text
- For multiple annotations on the same line, shows "(X of N annotations)"
- Preview disappears when cursor moves to a non-annotated line
- **Range highlighting**: Lines covered by the current annotation show a `▐` indicator

| Key | Action |
|-----|--------|
| `Tab` | Cycle to next annotation (when multiple exist) |
| `Shift+Tab` | Cycle to previous annotation |

Moving the cursor resets the preview to the first annotation.

### Visual Indicators

| Indicator | Meaning |
|-----------|---------|
| `●` | Line has at least one annotation |
| `▐` | Line is within the range of the currently previewed annotation |
| `◎` | Line contains a search match |

## Editing Existing Annotations

| Key | Action |
|-----|--------|
| `e` (no selection) | Edit annotation at cursor line |
| In picker: `j`/`k` | Navigate annotation list |
| In picker: `Enter` | Select annotation to edit |
| In picker: `Esc` | Cancel picker |

When multiple annotations exist on the same line, a picker opens to select which one to edit.

## Help Panel

| Key | Action |
|-----|--------|
| `?` | Open help panel (press any key to close) |

The help panel displays all available keybindings organized by category (Navigation, Selection, Annotations, General).

## Not Yet Implemented

The following are planned but not yet implemented:

- Mouse support
- `a` annotations list panel
- `R` edit annotation range (adjust start/end lines)
