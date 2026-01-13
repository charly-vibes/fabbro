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

## Selection

| Key | Action |
|-----|--------|
| `v` | Toggle line selection (start/clear) |
| `Esc` | Clear selection |

## Annotations (require selection)

| Key | Annotation Type | Prompt |
|-----|-----------------|--------|
| `c` | comment | "Comment:" |
| `d` | delete | "Reason:" |
| `q` | question | "Question:" |
| `e` | expand | "Expand:" |
| `k` | keep | (no prompt) |
| `u` | unclear | "Unclear:" |

## Command Palette

| Key | Action |
|-----|--------|
| `Space` | Open annotation palette |
| `Esc` | Close palette |

## Save & Exit

| Key | Action |
|-----|--------|
| `w` | Save and exit |
| `Q` | Quit immediately (no save) |
| `Ctrl+C` | Quit immediately (no save) |

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

## Not Yet Implemented

The following are planned but not yet implemented:

- `/` search
- `?` help panel
- Mouse support
- Confirmation prompt on unsaved exit
- `a` annotations list panel
- Text objects (`ap`, `ab`, `as`)
