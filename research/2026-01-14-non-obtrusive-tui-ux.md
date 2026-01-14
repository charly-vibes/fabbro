# Non-Obtrusive TUI UX Patterns for Agent Workflows

**Date:** 2026-01-14  
**Issue:** fabbro-d3d  
**Goal:** Define UX requirements for fabbro TUI that feel native to CLI agent workflows.

---

## Executive Summary

Agent workflows require TUI patterns that minimize context-switching while maximizing visibility into agent activity. This research analyzes existing tools and proposes concrete UX requirements for fabbro.

**Key Findings:**
1. **Floating/overlay patterns** (fzf, tmux popup) work well for quick interactions
2. **Full-screen TUI** (lazygit) excels for complex, focused tasks
3. **Inline prompts** (gum) are ideal for scripted/pipeline workflows
4. **Background + notification** patterns enable async workflows
5. **Multiplexer integration** (tmux/zellij) provides the most seamless experience

---

## Reference Tool Analysis

### 1. fzf: Floating Picker UX

**Approach:** fzf offers three display modes:
- **Fullscreen** (default): Takes over terminal
- **`--height` mode**: Inline rendering below cursor, non-disruptive
- **`--tmux` mode**: Renders in tmux popup window, floating over content

**Key UX Patterns:**
```bash
# Inline mode - renders below prompt, doesn't disrupt scrollback
fzf --height 40% --layout=reverse --border

# tmux popup mode - floats over terminal content
fzf --tmux center,80%,60%

# Style presets for quick visual consistency
fzf --style=minimal
```

**Why It Works:**
- Height mode preserves terminal context (scrollback visible above)
- Tmux popup feels "floating" without requiring terminal graphics
- Pressing Escape cleanly exits without side effects
- Single binary, instant startup (~50ms)

**Relevant for fabbro:**
- Quick file/session pickers should use height mode or tmux popup
- Startup must be near-instant (<100ms)
- ESC should always provide clean exit

---

### 2. lazygit: Fast Full-Screen Experience

**Approach:** Full-screen TUI with consistent panel layout

**Key UX Patterns:**
- **Instant startup:** Written in Go, opens in <200ms
- **Consistent layout:** Always same 6 panels visible
- **Discoverability:** Footer shows keybindings, `?` for full help
- **Trust building:** Interactive confirmations prevent mistakes
- **Context switching:** Can exec to editor and return seamlessly

**Why It Works:**
```
┌─────────────────────────────────────────────────┐
│ Status │ Files   │         Preview              │
│        │         │                              │
│ Local  │ Staging │                              │
│ Branch │         │                              │
│ Remote │ Commits │                              │
│        │         │                              │
│        │ Stash   │                              │
└─────────────────────────────────────────────────┘
Footer: keybindings always visible
```

**Relevant for fabbro:**
- For code review TUI, full-screen with consistent panels makes sense
- Always show "where you are" and "what you can do"
- Startup speed is critical - users notice anything >200ms
- Use alternate screen buffer (preserves terminal history)

---

### 3. gum: Inline Prompt Patterns

**Approach:** Individual composable prompts for shell scripts

**Key UX Patterns:**
```bash
# Choose - inline selection
choice=$(gum choose "feat" "fix" "docs" "refactor")

# Input - inline text prompt
name=$(gum input --placeholder "Enter your name")

# Spin - show progress inline
gum spin --spinner dot --title "Processing..." -- sleep 5

# Confirm - yes/no inline
gum confirm "Deploy to production?"
```

**Why It Works:**
- Each component is atomic and composable
- Output goes to stdout (pipeable)
- Doesn't take over terminal
- Preserves command-line history flow
- Works in scripts without user intervention needed

**Relevant for fabbro:**
- For `fabbro init`, use gum-style inline prompts
- Session selection could be gum choose pattern
- Progress indicators for background operations
- Confirmation before destructive actions

---

### 4. charmbracelet/pop: Notification Patterns

**Approach:** Email TUI with notification-style workflows

**Key Insight:** Pop is about composing with other tools, not standalone operation.

```bash
# Use gum to pick inputs, pop to send
pop --to $(gum input) --subject "Update" --body "..."
```

**Note:** charmbracelet doesn't have a dedicated "toast/notification" library. Notifications in terminal are typically handled via:
- Terminal bell (`\a`)
- Desktop notifications via `notify-send` (Linux) or `osascript` (macOS)
- Status bar updates in tmux/zellij
- Inline progress with spinners

**Relevant for fabbro:**
- Use desktop notifications for background session completion
- tmux/zellij status bar integration for ongoing sessions
- Terminal bell for attention-needed states

---

### 5. Elia: AI Chat TUI

**Approach:** Full-screen keyboard-centric AI chat interface

**Key UX Patterns:**
- **`--inline` mode:** Runs under prompt, doesn't take over screen
- **Full-screen mode:** For extended conversations
- **SQLite persistence:** Conversations saved locally
- **Model switching:** `-m` flag for model selection

```bash
# Inline mode - quick question under prompt
elia --inline "How do I fix this?"

# Full screen - extended session
elia
```

**Why It Works:**
- Choice of inline vs full-screen based on task complexity
- Keyboard-centric (Vim-like navigation)
- Local-first persistence
- Fast launch via Python/Textual

**Relevant for fabbro:**
- Offer both inline and full-screen modes
- `fabbro review --inline` for quick reviews
- `fabbro review` for full code review session
- Local SQLite for session persistence (already planned)

---

### 6. tmux/zellij: Floating Pane Integration

#### tmux display-popup

```bash
# Basic popup
tmux display-popup -w 80% -h 60%

# Run fzf in popup
tmux display-popup -E "fzf"

# Persistent popup (stays open)
tmux display-popup -d "#{pane_current_path}" -w 80% -h 80%
```

**Key Features:**
- Floats over existing panes
- Can be dismissed with Escape
- Inherits environment
- `-E` flag closes on command exit

#### zellij floating panes

```bash
# Toggle floating panes visibility
zellij action toggle-floating-panes

# Run command in floating pane
zellij run --floating -- fabbro review
```

**Key Features:**
- Floating panes stack and can be dragged
- Toggle visibility on/off
- Integrates with zellij plugin system

**Relevant for fabbro:**
- Detect tmux/zellij and use native floating panes
- `fabbro review --popup` to launch in floating pane
- Provide tmux/zellij keybinding examples in docs

---

### 7. Amp: Editor Integration Patterns

**Approach:** Multi-environment agent (CLI + VS Code extension)

**Key UX Patterns:**
- **CLI mode:** Full TUI in terminal
- **VS Code extension:** Native editor integration
- **Thread sharing:** Sessions persist across environments
- **`-x` execute mode:** One-shot queries without TUI

```bash
# Interactive TUI
amp

# Execute mode - answer and exit
amp -x "what is 2+2?"

# Pipe mode
echo "fix this" | amp -x
```

**Why "Open in Editor" Works:**
- Amp CLI can exec to `$EDITOR` and return
- Changes made in editor are visible to agent
- Uses `tea.ExecProcess` pattern (Bubble Tea)
- Preserves terminal state during editor session

**Relevant for fabbro:**
- Support `$EDITOR` integration for annotation editing
- Implement "open file at line" functionality
- Consider VS Code extension for future

---

## UX Pattern Categories

### Pattern 1: Inline/Height Mode
**Best for:** Quick interactions, script integration, pipelines
**Example:** fzf `--height`, gum prompts, elia `--inline`
**When to use in fabbro:** Init wizard, quick session picker, confirmation prompts

### Pattern 2: Floating/Popup Mode  
**Best for:** Focused interaction while maintaining context
**Example:** fzf `--tmux`, tmux display-popup, zellij floating panes
**When to use in fabbro:** Code review, file selection, agent status

### Pattern 3: Full-Screen TUI
**Best for:** Complex, multi-panel workflows requiring focus
**Example:** lazygit, elia full-screen, vim
**When to use in fabbro:** Full code review session, multi-file navigation

### Pattern 4: Background + Notification
**Best for:** Long-running operations
**Example:** Background processes + notify-send
**When to use in fabbro:** Session creation, AI processing, batch reviews

### Pattern 5: Execute Mode (One-Shot)
**Best for:** Scriptable, non-interactive use
**Example:** amp `-x`, gum individual commands
**When to use in fabbro:** `fabbro check file.go`, CI/CD integration

---

## Mapping to Current Specs

Cross-reference with `specs/03_tui_interaction.feature`:

| Requirement | Spec Status | Notes |
|-------------|-------------|-------|
| R1: Display modes | @planned (partial) | Full-screen @implemented; --inline, --popup not in spec |
| R2: Multiplexer integration | Not in spec | Future exploration |
| R3: Startup performance | Implicit | Add explicit timing requirements to spec |
| R4: Background session | Not in spec | Future exploration |
| R5: Clean exit | @implemented | ESC/q/Ctrl+C behaviors defined |
| R6: Editor integration | Not in spec | See 02_review_session @planned --editor |
| R7: Discoverability | @implemented | SPC command palette, ? for help |
| R8: Notification system | Not in spec | Future exploration |

**Already implemented (Phase 1 complete):**
- Full-screen TUI with alternate screen buffer
- Vim-style navigation (j/k/g/G)
- SPC command palette (Helix-style)
- Visual selection (v)

**Next priority (from specs):**
- Search within document (@planned in spec)
- Annotations panel (@planned in spec)
- Help overlay (@planned in spec)

## Concrete UX Requirements for Fabbro

### R1: Multiple Display Modes

```bash
# Default: Smart mode selection based on context
fabbro review

# Explicit modes
fabbro review --inline      # Height mode, below prompt
fabbro review --popup       # tmux/zellij floating pane
fabbro review --fullscreen  # Alternate screen buffer
fabbro review --execute     # Non-interactive, stdout output
```

### R2: Multiplexer Integration

```bash
# Auto-detect tmux/zellij
if in_tmux:
    use display-popup for floating UIs
elif in_zellij:
    use floating panes
else:
    use --height mode
```

**Implementation:**
- Check `$TMUX` and `$ZELLIJ` environment variables
- Provide `--no-popup` flag to override

### R3: Startup Performance

| Operation | Target | Notes |
|-----------|--------|-------|
| Cold start | <200ms | First launch |
| Warm start | <100ms | Subsequent launches |
| Session load | <500ms | Loading existing session |
| File picker | <50ms | fzf-like instant |

**Implementation:**
- Single Go binary (no runtime dependencies)
- Lazy load session data
- Defer AI model connections

### R4: Background Session Creation

```bash
# Start review in background
fabbro review --background file.go

# Check status
fabbro status

# Notification when ready
# Desktop: notify-send "fabbro: Review ready for file.go"
# tmux: set status-right message
# zellij: plugin notification
```

### R5: Clean Exit Semantics

| Key | Action |
|-----|--------|
| ESC | Cancel/back (never exit from root) |
| q | Quit from root view |
| Ctrl+C | Force quit (with confirmation if unsaved) |

**Implementation:**
- Use alternate screen buffer
- Restore terminal state on any exit
- Never leave garbage in scrollback

### R6: Editor Integration

```bash
# Open annotation in $EDITOR
# On save, return to fabbro with changes applied

# Open file at specific line
# Uses fabbro's keybinding (e.g., 'e' on a hunk)
```

**Implementation:**
- Use `tea.ExecProcess` from Bubble Tea
- Pass `+line` argument to editors that support it
- Handle editor exit codes

### R7: Discoverability

- Footer bar with context-sensitive keybindings (like lazygit)
- `?` key opens help overlay
- Command palette (Ctrl+P or similar)
- Onboarding hints for first-time users

### R8: Notification System

| Context | Notification Method |
|---------|---------------------|
| tmux | Status bar message + bell |
| zellij | Plugin notification |
| Plain terminal | Desktop notification (if available) + bell |
| Headless/SSH | Exit code + stdout message |

---

## Recommended Approach for Fabbro

### Phase 1: Core TUI (Current Focus)
1. **Full-screen mode first** - Simpler to implement correctly
2. **Alternate screen buffer** - Clean terminal restoration
3. **Fast startup** - <200ms target
4. **Basic keybindings** - hjkl navigation, q to quit, ? for help

### Phase 2: Display Mode Flexibility
1. **Inline mode** - `--inline` flag for quick interactions
2. **Execute mode** - `--execute` for scriptable output
3. **Height mode fallback** - When not in multiplexer

### Phase 3: Multiplexer Integration
1. **tmux popup detection** - Auto-use display-popup
2. **zellij floating panes** - Auto-use floating panes
3. **Status bar integration** - Background session status

### Phase 4: Background Operations (Future Exploration)
1. **Background flag** - `--background` for async start
2. **Notification system** - Desktop + terminal notifications
3. **Status command** - `fabbro status` to check running sessions

> **Note**: Phases 2-4 are future exploration. Current focus is Phase 1 (core TUI) which is largely complete per `03_tui_interaction.feature`.

---

## Tradeoffs

### Full-Screen vs Floating
| Aspect | Full-Screen | Floating |
|--------|-------------|----------|
| Focus | High | Medium |
| Context | Low (hidden) | High (visible behind) |
| Complexity | Lower | Higher (multiplexer deps) |
| Universality | Works everywhere | Requires tmux/zellij |

**Recommendation:** Start with full-screen, add floating as enhancement.

### Inline vs Separate Process
| Aspect | Inline (height mode) | Separate (popup/window) |
|--------|---------------------|------------------------|
| Integration | Feels native | Feels like tool |
| Scrollback | Visible above | Hidden |
| Terminal state | Shared | Isolated |

**Recommendation:** Use inline for quick operations, full-screen for sessions.

### TUI vs GUI
| Aspect | TUI | GUI (Electron/Tauri) |
|--------|-----|---------------------|
| Resource usage | Low | High |
| Startup time | Fast | Slow |
| Agent workflow fit | Native | Foreign |
| Cross-platform | Easy | Complex |

**Recommendation:** TUI-first. GUI is out of scope for fabbro's mission.

---

## References

- [fzf README - Display Modes](https://github.com/junegunn/fzf#display-modes)
- [lazygit - TUI Design](https://github.com/jesseduffield/lazygit)
- [charmbracelet/gum](https://github.com/charmbracelet/gum)
- [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea)
- [Elia TUI](https://github.com/darrenburns/elia)
- [Zellij Floating Panes](https://zellij.dev/news/floating-panes-tmux-mode/)
- [tmux display-popup](https://man7.org/linux/man-pages/man1/tmux.1.html)
- [Amp Code Manual](https://ampcode.com/manual)
- [Terminal UI Development in Go](https://dev.to/bmf_san/understanding-terminal-specifications-to-help-with-tui-development-749)
