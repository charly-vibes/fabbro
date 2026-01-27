# Tutor Command Implementation Plan

## Overview

Add a `fabbro tutor` subcommand that launches an interactive tutorial, similar to `vimtutor`. The tutor guides new users through fabbro's core workflow: navigation, selection, annotation, and saving.

**Goal**: Make fabbro self-documenting with a hands-on learning experience.

## Related

- CLI docs: `docs/cli.md`
- Keybindings: `docs/keybindings.md`
- TUI implementation: `internal/tui/`
- Main CLI: `cmd/fabbro/main.go`

## Inspiration

| Tool | Command | Behavior |
|------|---------|----------|
| Vim | `vimtutor` | Opens a copy of tutorial file for editing |
| Emacs | `C-h t` | Interactive tutorial buffer |
| Helix | `:tutor` | Similar to vimtutor |

**Approach**: Follow vimtutor's pattern—open a pre-written tutorial file in the TUI. User practices on real content with guided exercises.

---

## Phase 1: Minimal Tutor (~1 hour)

### 1.1 Create Tutorial Content

**File: `internal/tutor/content.go`**

Embed the tutorial text as a Go string (using `//go:embed` or raw string):

```go
package tutor

import _ "embed"

//go:embed tutorial.txt
var Content string
```

**File: `internal/tutor/tutorial.txt`**

```
===============================================================================
=    W e l c o m e   t o   t h e   F a b b r o   T u t o r i a l              =
===============================================================================

Fabbro is a code review annotation tool. This tutorial teaches you the
basics through hands-on practice.

TIP: This is a real fabbro session. Your annotations won't be saved unless
     you press 'w'. Feel free to experiment!

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
                          Lesson 1: NAVIGATION
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Use these keys to move around:

    j - move DOWN one line
    k - move UP one line
    gg - jump to the FIRST line
    G - jump to the LAST line
    Ctrl+d - scroll DOWN half a page
    Ctrl+u - scroll UP half a page

>>> Practice: Use j and k to move to line 30, then use G to jump to the end.

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
                          Lesson 2: SELECTION
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Before annotating, you must SELECT text:

    v - toggle selection on current line
    v + j/k - extend selection to multiple lines
    Esc - clear selection

The selected line(s) will be highlighted.

>>> Practice: Press v to select this line, then press Esc to deselect.

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
                          Lesson 3: ADDING COMMENTS
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

With a line selected, press 'c' to add a comment:

    1. Select a line with 'v'
    2. Press 'c'
    3. Type your comment
    4. Press Enter to save

>>> Practice: Select the line below and add the comment "This is a test"

    function example() {
        return 42;
    }

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
                          Lesson 4: ANNOTATION TYPES
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Fabbro supports several annotation types. With a line selected:

    c - comment    "Add a note"
    d - delete     "Mark for deletion"
    q - question   "Ask a question"
    e - expand     "Request more detail"
    u - unclear    "Mark as confusing"

For 'k' (keep), use the command palette (next lesson).

>>> Practice: Select a line above and try pressing 'd' to mark for deletion.

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
                          Lesson 5: COMMAND PALETTE
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Press SPACE to open the command palette. It shows all available actions:

    Space → c - comment
    Space → d - delete
    Space → k - keep (only available here!)
    Space → Q - force quit

Press Esc to close the palette without acting.

>>> Practice: Press Space to see the palette, then press Esc to close it.

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
                          Lesson 6: SAVING & QUITTING
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

When you're done annotating:

    w - save session (creates .fabbro/sessions/<id>.fem)
    Ctrl+C Ctrl+C - quit with confirmation

After saving, extract annotations with:

    fabbro apply <session-id> --json

>>> Practice: When ready, press Ctrl+C twice to exit (don't save this session).

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
                          TUTORIAL COMPLETE
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

You now know fabbro basics! Next steps:

    fabbro review myfile.go     Review a file
    fabbro review --stdin       Review piped input
    fabbro session list         See your sessions
    fabbro apply <id> --json    Extract annotations

For keybinding reference: see docs/keybindings.md

Happy reviewing!
===============================================================================
```

### 1.2 Add Tutor Command

**File: `cmd/fabbro/main.go`**

Add `buildTutorCmd()`:

```go
func buildTutorCmd(stdout io.Writer) *cobra.Command {
    return &cobra.Command{
        Use:   "tutor",
        Short: "Start the interactive tutorial",
        Long: `Launch an interactive tutorial that teaches fabbro basics.

The tutor opens a guided lesson file in the TUI where you can
practice navigation, selection, and annotation. Like vimtutor,
this is hands-on learning.

Your practice session is temporary and won't affect any real files.`,
        RunE: func(cmd *cobra.Command, args []string) error {
            // Create a temporary session with tutorial content
            sess := &session.Session{
                ID:         "tutor",
                Content:    tutor.Content,
                SourceFile: "",
                CreatedAt:  time.Now(),
            }

            fmt.Fprintln(stdout, "Welcome to the fabbro tutor!")
            fmt.Fprintln(stdout, "")

            model := tui.NewWithFile(sess, "(tutorial)")
            p := tea.NewProgram(model)
            if _, err := p.Run(); err != nil {
                return fmt.Errorf("TUI error: %w", err)
            }

            return nil
        },
    }
}
```

Register in `buildRootCmd()`:

```go
rootCmd.AddCommand(buildTutorCmd(stdout))
```

### 1.3 Handle Tutor Session Save

The tutor uses a special session ID "tutor". Options:
1. **Don't save** - Modify TUI to skip save for tutor sessions
2. **Save normally** - Let users save if they want (simpler)
3. **Save to temp** - Save to `/tmp/` instead of `.fabbro/`

**Recommendation**: Option 2 (save normally). If user saves, they get a real session file they can inspect. Simple and educational.

### Success Criteria

#### Automated:
- [ ] `go test ./...` passes
- [ ] `fabbro tutor --help` shows usage

#### Manual:
- [ ] `fabbro tutor` opens TUI with tutorial content
- [ ] All lessons are readable and accurate
- [ ] User can practice each described action

---

## Phase 2: Tutor Without Init (~30 min)

### Problem

Currently all commands (except `init`) require `fabbro init` first. The tutor should work anywhere, even without initialization.

### Solution

Skip the `config.IsInitialized()` check for tutor. Since tutor sessions don't save to `.fabbro/` by default, no init is needed.

**Option A**: Don't call `session.Create()` for tutor—pass content directly to TUI.

**Option B**: Create an in-memory session that doesn't persist.

**Recommendation**: Option A—simplest. The tutor command creates a `Session` struct directly without calling `session.Create()`.

### Changes

**File: `cmd/fabbro/main.go`** (already shown in Phase 1)

The `buildTutorCmd` creates the session struct directly, bypassing `session.Create()` which requires init.

**File: `internal/tui/model.go`**

Ensure TUI handles sessions without a real ID gracefully. On save, if session ID is "tutor", show a message like "Tutorial session saved as tutor.fem" or skip persistence.

### Success Criteria

- [ ] `fabbro tutor` works in any directory (no init required)
- [ ] Tutorial session save behavior is clear to user

---

## Phase 3: Tutor Save Behavior (~30 min)

### Decision: What happens on 'w' (save)?

**Option A: Skip save entirely**
- Show message: "Tutorial sessions are not saved. Exit with Ctrl+C."
- Pro: Clean, no clutter
- Con: User doesn't practice the save command

**Option B: Save to temp directory**
- Save to `/tmp/fabbro-tutor-<timestamp>.fem`
- Pro: User practices save, file is discoverable
- Con: Leaves temp files

**Option C: Require init to save** (current behavior)
- If init'd: save normally to `.fabbro/sessions/tutor.fem`
- If not init'd: show error "Initialize fabbro to save sessions"
- Pro: Teaches real workflow
- Con: Confusing UX for tutorial

**Recommendation**: Option A for simplicity. The tutorial teaches the concept; users will practice saving on real sessions.

### Implementation

**File: `internal/tui/handlers.go`**

In save handler, check if session ID is "tutor":

```go
func (m *Model) handleSave() {
    if m.session.ID == "tutor" {
        m.setStatusMessage("Tutorial sessions are not saved")
        return
    }
    // existing save logic
}
```

### Success Criteria

- [ ] Pressing 'w' in tutor shows "Tutorial sessions are not saved"
- [ ] No file is created

---

## File Structure

```
internal/
└── tutor/
    ├── content.go      # //go:embed directive
    └── tutorial.txt    # The actual tutorial text
```

---

## Summary

| Phase | Deliverable | Time |
|-------|-------------|------|
| 1 | `fabbro tutor` command + tutorial content | 1h |
| 2 | Works without `fabbro init` | 30m |
| 3 | Handle save gracefully | 30m |

**Total: ~2 hours**

---

## Future Enhancements (Out of Scope)

- **Localization**: Tutorial text in other languages
- **Progress tracking**: Remember which lessons completed
- **Advanced tutor**: Separate `fabbro tutor advanced` for power features
- **Interactive validation**: Detect if user completed exercise correctly
- **`:tutor` command**: Open tutor from within existing TUI session

---

## Open Questions

1. Should tutor work from within an existing TUI session (`:tutor` command)?
2. Should we track tutorial completion for gamification?
3. Include a "quick reference" card at the end users can copy?
