# Research: Inline Edit Keybindings

**Date:** 2026-01-28  
**Issue:** fabbro-sp0  
**Goal:** Change inline edit to use Enter=accept, Shift+Enter=newline

## Current Implementation

The inline editor is configured in [internal/tui/handlers.go](file:///var/home/sasha/para/areas/dev/gh/charly/fabbro/internal/tui/handlers.go):

### Text Area Setup (line 56)

```go
ta.KeyMap.InsertNewline.SetKeys("shift+enter", "ctrl+m")
```

This already configures `Shift+Enter` for newlines.

### Input Mode Handler (lines 322-371)

```go
func (m Model) handleInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.Type {
    case tea.KeyEnter:
        // Submits the annotation
        ...
    case tea.KeyEsc:
        // Cancels
        ...
    }
    // Falls through to textarea.Update() for other keys
    var cmd tea.Cmd
    *m.inputTA, cmd = m.inputTA.Update(msg)
    return m, cmd
}
```

## Two Different Editing Modes

There are **two** editing modes with different keybindings:

### 1. `modeInput` (quick annotations) ✓ Already correct
- **Enter** → Submit
- **Shift+Enter** → Newline
- Shows: `"(Shift+Enter for newline, Enter to submit)"`

### 2. `modeEditor` (change annotations) ✗ Needs fix
- **Ctrl+S** → Save
- **Esc Esc** → Cancel  
- **Enter** → Newline (default textarea behavior)
- Shows: `"(Ctrl+S save, Esc Esc cancel)"`

## Required Changes (Oracle-reviewed)

### 1. handlers.go — `handleEditorMode()` 

Add `case "enter":` (using `msg.String()` not `tea.KeyEnter` to avoid modifier ambiguity):

```go
case "enter":
    m.saveEditorContent()
    return m, nil
```

Keep `Ctrl+S` as alternative save for power users.

### 2. handlers.go — editor textarea setup

Configure keymap in **both** entry points (critical!):
- `openEditorForAnnotation()`
- `openEditor()`

Use `ctrl+j` instead of `ctrl+m` (Enter = `^M` in many terminals, which could cause conflicts):

```go
ta.KeyMap.InsertNewline.SetKeys("shift+enter", "ctrl+j")
```

### 3. view.go — update help text (line 155)

```go
b.WriteString("┌─ Edit (Enter save, Shift+Enter newline, Esc Esc cancel) ┐\n")
```

### 4. Do NOT change cancel behavior

Keep `Esc Esc` as-is — changing it is a separate concern with its own edge cases.

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| `ctrl+m` = Enter in some terminals | Use `ctrl+j` instead |
| Accidental Enter submits | Standard UX; saves content (no data loss) |
| Inconsistent behavior | Apply keymap in both editor entry points |
