# TUI (Terminal User Interface)

fabbro includes a terminal-based UI for reviewing and annotating code.

## Overview

The TUI launches when you run `fabbro review --stdin`. It displays the content with line numbers and allows you to navigate, select lines, and add comments.

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
| `j` | Move cursor down one line |
| `k` | Move cursor up one line |
| `↓` | Move cursor down one line |
| `↑` | Move cursor up one line |

### Selection

| Key | Action |
|-----|--------|
| `v` | Toggle selection on current line |

Selecting a line marks it for annotation. Press `v` again to deselect.

### Commenting

| Key | Action |
|-----|--------|
| `c` | Enter comment mode (requires selection) |

When in comment mode:

| Key | Action |
|-----|--------|
| Type | Add characters to comment |
| `Backspace` | Delete last character |
| `Enter` | Save comment and return to normal mode |
| `Esc` | Cancel comment and return to normal mode |

### Session Control

| Key | Action |
|-----|--------|
| `w` | Save session and quit |
| `q` | Quit without saving |
| `Ctrl+C` | Quit without saving |

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
