# TUI (Terminal User Interface)

fabbro includes a terminal-based UI for reviewing and annotating code.

## Overview

The TUI launches when you run `fabbro review <file>` or `fabbro review --stdin`. It displays the content with line numbers and allows you to navigate, select lines, and add comments.

## Layout

```
─── Review: abc12345 ─────────────────────────────
   1 │ package main
   2 │ 
>  3 │ func main() {
●  4 │     fmt.Println("Hello")
   5 │ }
──────────────────────────────────────────────────
[v]select [c]omment [w]rite [q]uit
```

- `>` indicates the cursor position
- `●` indicates a selected line
- The status bar shows available commands

## Keybindings

### Navigation

| Key | Action |
|-----|--------|
| `j` / `↓` | Move cursor down one line |
| `k` / `↑` | Move cursor up one line |
| `Ctrl+d` | Scroll down half page |
| `Ctrl+u` | Scroll up half page |
| `gg` | Jump to first line |
| `G` | Jump to last line |

### Selection

| Key | Action |
|-----|--------|
| `v` | Toggle selection on current line |
| `Esc` | Clear selection |

Selecting a line marks it for annotation. You can navigate while selected to extend the selection range.

### Annotations (require selection)

| Key | Annotation Type | Prompt |
|-----|-----------------|--------|
| `c` | comment | "Comment:" |
| `d` | delete | "Reason for deletion:" |
| `q` | question | "Question:" |
| `e` | expand | "What to expand:" |
| `u` | unclear | "What's unclear:" |
| `r` | change | "Replacement text:" |

### Command Palette

| Key | Action |
|-----|--------|
| `Space` | Open annotation palette (when selected) |
| `Esc` | Close palette |

The palette provides all annotation types including `k` (keep) which is only available via palette.

### Input Mode

When typing an annotation:

| Key | Action |
|-----|--------|
| Type | Add characters |
| `Backspace` | Delete last character |
| `Enter` | Submit annotation and return to normal mode |
| `Esc` | Cancel and return to normal mode |

### Session Control

| Key | Action |
|-----|--------|
| `w` | Save session |
| `Ctrl+C Ctrl+C` | Quit (with confirmation prompt) |
| `Space` → `Q` | Force quit immediately (no confirmation) |

**Note:** `q` (lowercase) is the question annotation. Quit requires double `Ctrl+C` or palette.

## Workflow

1. **Navigate** to a line of interest using `j`/`k`
2. **Select** the line with `v`
3. **Comment** by pressing `c`, then type your annotation
4. **Submit** with `Enter` (or cancel with `Esc`)
5. **Repeat** for additional annotations
6. **Save** with `w` when done

## Tips

- You can navigate while a line is selected
- Comments are embedded as FEM markup: `{>> your comment <<}`
- Session files are saved to `.fabbro/sessions/<id>.fem`
- Use `fabbro apply <id> --json` to extract annotations programmatically
