# Interactive Shell Detection Research

**Date:** 2026-01-14  
**Issue:** fabbro-9y0  
**Topic:** How gemini-cli detects interactive shell commands for merge commits

## Summary

Gemini-cli uses **pseudo-terminal (PTY) support** with the `node-pty` library to handle interactive commands. Rather than "detecting" which commands are interactive upfront, they provide a full PTY environment that works for both interactive and non-interactive commands. This approach is more robust than command pattern matching.

> **Scope Note**: This research is for *future* fabbro features. Currently, fabbro doesn't execute shell commands - it's a code review annotation tool. This research applies if/when fabbro adds:
> - Shell command execution in TUI (e.g., running tests)
> - Git operations (merge, rebase) from within fabbro
> - Editor integration via `$EDITOR`

## Key Findings

### 1. Gemini-CLI's Approach: PTY-Based Execution

Gemini-cli doesn't try to predict which commands will be interactive. Instead, they:

1. **Always use PTY when enabled** via `tools.shell.enableInteractiveShell` setting
2. **Spawn commands in a pseudo-terminal** using `node-pty` (or `@lydell/node-pty`)
3. **Render terminal state** using `@xterm/headless` for serialization
4. **Allow user interaction** via `ctrl+f` to focus the interactive shell

From their [shell tool documentation](https://geminicli.com/docs/tools/shell/):
> "The `run_shell_command` tool now supports interactive commands by integrating a pseudo-terminal (pty). This allows you to run commands that require real-time user input, such as text editors (`vim`, `nano`), terminal-based UIs (`htop`), and interactive version control operations (`git rebase -i`)."

### 2. Shell Execution Service Architecture

The [`shellExecutionService.ts`](https://github.com/google-gemini/gemini-cli/blob/main/packages/core/src/services/shellExecutionService.ts) has two execution paths:

```typescript
// Two execution modes:
// 1. PTY mode (interactive) - uses node-pty
// 2. Child process fallback (non-interactive) - uses child_process.spawn
```

**PTY Execution Flow:**
1. Spawn shell via `node-pty` with `xterm-256color` terminal type
2. Create headless terminal (`@xterm/headless`) for state tracking
3. Stream output via terminal serializer
4. Handle resize events (`resizePty`)
5. Accept user input (`writeToPty`)
6. Track scrollback (300,000 lines for large context)

**Key Configuration:**
```typescript
const ptyProcess = ptyInfo.module.spawn(executable, args, {
  cwd,
  name: 'xterm-256color',
  cols,
  rows,
  env: {
    ...sanitizeEnvironment(process.env, sanitizationConfig),
    GEMINI_CLI: '1',        // Marker for scripts to detect
    TERM: 'xterm-256color',
    PAGER: 'cat',           // Disable pager for non-interactive
    GIT_PAGER: 'cat',
  },
  handleFlowControl: true,
});
```

### 3. Environment Variable Detection

Gemini-cli sets `GEMINI_CLI=1` in subprocess environments. This allows:
- Scripts to detect they're running under gemini-cli
- Commands to adapt behavior accordingly

### 4. Inactivity Timeout Handling

Instead of detecting interactive commands, gemini-cli uses:
- **Inactivity timeout**: Cancels commands that produce no output for a configurable period
- User can still interact during this window if the command is interactive

## Go Implementation Patterns

### TTY Detection in Go

The standard approach uses `golang.org/x/term`:

```go
import "golang.org/x/term"

func IsTerminal(fd int) bool {
    return term.IsTerminal(fd)
}

// Usage
if term.IsTerminal(int(os.Stdin.Fd())) {
    fmt.Println("Running in terminal")
}
```

For cross-platform compatibility (including Git Bash on Windows), use [`mattn/go-isatty`](https://github.com/mattn/go-isatty):

```go
import "github.com/mattn/go-isatty"

if isatty.IsTerminal(os.Stdout.Fd()) {
    fmt.Println("Is Terminal")
} else if isatty.IsCygwinTerminal(os.Stdout.Fd()) {
    fmt.Println("Is Cygwin/MSYS terminal")
}
```

### PTY Library for Go: `creack/pty`

The [`creack/pty`](https://github.com/creack/pty) library (1,074+ importers) provides PTY support:

```go
import (
    "os/exec"
    "github.com/creack/pty"
)

// Start command with PTY
cmd := exec.Command("bash")
ptmx, err := pty.Start(cmd)
if err != nil {
    return err
}
defer ptmx.Close()

// Handle terminal size
pty.InheritSize(os.Stdin, ptmx)

// Copy I/O
go io.Copy(ptmx, os.Stdin)
io.Copy(os.Stdout, ptmx)
```

For interactive shell with size handling:

```go
import (
    "os/signal"
    "syscall"
    "golang.org/x/term"
    "github.com/creack/pty"
)

func runInteractive(cmd *exec.Cmd) error {
    ptmx, err := pty.Start(cmd)
    if err != nil {
        return err
    }
    defer ptmx.Close()

    // Handle resize signals
    ch := make(chan os.Signal, 1)
    signal.Notify(ch, syscall.SIGWINCH)
    go func() {
        for range ch {
            pty.InheritSize(os.Stdin, ptmx)
        }
    }()
    ch <- syscall.SIGWINCH // Initial resize
    defer func() { signal.Stop(ch); close(ch) }()

    // Put terminal in raw mode
    oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
    if err != nil {
        return err
    }
    defer term.Restore(int(os.Stdin.Fd()), oldState)

    // I/O copy
    go io.Copy(ptmx, os.Stdin)
    io.Copy(os.Stdout, ptmx)

    return nil
}
```

## Interactive Command Patterns

### Commands That Typically Need PTY

| Category | Examples |
|----------|----------|
| **Editors** | `vim`, `nano`, `emacs -nw` |
| **Git interactive** | `git rebase -i`, `git add -p`, `git commit` (with editor) |
| **REPLs** | `python`, `node`, `irb`, `ghci` |
| **TUI apps** | `htop`, `top`, `mc`, `less` |
| **Prompts** | `npm init`, `ng new`, password prompts |

### Detection Strategies

1. **Command prefix matching** (brittle, not recommended):
   ```go
   // Don't do this - too fragile
   interactiveCommands := []string{"vim", "nano", "git rebase"}
   ```

2. **Environment-based** (gemini-cli approach):
   ```go
   // Set environment so child knows context
   cmd.Env = append(os.Environ(), "FABBRO=1")
   ```

3. **Always use PTY** (most robust):
   ```go
   // Use PTY for all shell commands when in TUI mode
   if runningInTUI {
       runWithPTY(cmd)
   } else {
       runWithChildProcess(cmd)
   }
   ```

4. **Fallback with timeout**:
   ```go
   // Run with child_process, but with inactivity timeout
   // User can cancel and retry with PTY if command hangs
   ```

## Recommended Approach for Fabbro

### Option A: Always PTY (Gemini-CLI approach)

**Pros:**
- Works for all interactive commands automatically
- No command detection logic needed
- Full terminal capabilities

**Cons:**
- More complex implementation
- PTY overhead for simple commands
- Need to handle terminal serialization for review display

### Option B: Opt-in PTY Mode

**Pros:**
- Simpler default path
- PTY only when needed

**Cons:**
- User must know when to enable
- Can still hang on unexpected interactive commands

### Option C: Hybrid with Detection (Recommended for Fabbro)

1. **Detect if stdin is a terminal**: Use `term.IsTerminal()`
2. **Check for known interactive patterns**: Git merge/rebase, editor commands
3. **Use PTY for detected interactive commands**
4. **Implement inactivity timeout** for unexpected hangs
5. **Allow user override** via flag or config

```go
type ShellConfig struct {
    ForceInteractive bool          // Always use PTY
    Timeout          time.Duration // Inactivity timeout
}

func ExecuteShell(cmd string, config ShellConfig) error {
    needsPTY := config.ForceInteractive || 
                isKnownInteractive(cmd) ||
                term.IsTerminal(int(os.Stdin.Fd()))
    
    if needsPTY {
        return executeWithPTY(cmd, config.Timeout)
    }
    return executeWithChildProcess(cmd, config.Timeout)
}

func isKnownInteractive(cmd string) bool {
    patterns := []string{
        "git rebase -i",
        "git merge",      // May open editor
        "git commit",     // Opens editor without -m
        "vim", "nvim", "nano", "emacs",
    }
    for _, p := range patterns {
        if strings.Contains(cmd, p) {
            return true
        }
    }
    return false
}
```

## Implementation Considerations

### 1. TUI Integration

When fabbro runs in TUI mode (Bubble Tea), PTY interaction requires:
- Suspending TUI temporarily for interactive commands
- Restoring terminal state after command completion
- Handling SIGWINCH for resize

### 2. Review Mode vs Shell Mode

- **Review mode**: Display command output in TUI (may need serialization)
- **Shell mode**: Full terminal passthrough for interactive commands

### 3. Git-Specific Handling

For `git merge`, `git rebase -i`, `git commit`:
- Set `GIT_EDITOR` to control editor behavior
- Consider using `--no-edit` flags when appropriate
- Detect `GIT_SEQUENCE_EDITOR` for rebase

### 4. Error Recovery

If PTY spawning fails (sandbox restrictions, etc.):
- Fall back to child_process
- Log warning for user
- Similar to gemini-cli's fallback pattern

## Dependencies for Go Implementation

```go
// go.mod additions
require (
    github.com/creack/pty v1.1.24      // PTY support
    golang.org/x/term v0.39.0          // Terminal utilities
    github.com/mattn/go-isatty v0.0.20 // Cross-platform TTY detection
)
```

## Next Steps

When fabbro needs shell/editor integration:

1. **Create beads issue**: `bd create --title="Add $EDITOR integration for annotation editing" --type=task`
2. **Add `creack/pty` dependency**: For TUI suspension pattern
3. **Implement `tea.ExecProcess`**: Bubble Tea's built-in editor handoff

## Integration with Bubble Tea

For TUI → editor → TUI flow, use Bubble Tea's `tea.ExecProcess`:

```go
// Suspend TUI, open editor, resume TUI
cmd := tea.ExecProcess(exec.Command("vim", "+42", "file.go"), func(err error) tea.Msg {
    return editorClosedMsg{err: err}
})
return m, cmd
```

This is the pattern lazygit and other Bubble Tea apps use. No PTY management needed.

## References

1. [Gemini-CLI Shell Tool Documentation](https://geminicli.com/docs/tools/shell/)
2. [Gemini-CLI Interactive Shell Announcement](https://developers.googleblog.com/en/say-hello-to-a-new-level-of-interactivity-in-gemini-cli/)
3. [creack/pty - PTY Interface for Go](https://github.com/creack/pty)
4. [golang.org/x/term - Terminal Support](https://pkg.go.dev/golang.org/x/term)
5. [mattn/go-isatty - Cross-platform TTY Detection](https://github.com/mattn/go-isatty)
6. [node-pty - Used by Gemini-CLI](https://github.com/lydell/node-pty)
