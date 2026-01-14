# TUI Integration Test Workflow

Run integration tests on fabbro's TUI using the mcp-tui-test MCP server.

## Prerequisites

Start Amp with the TUI testing MCP:
```bash
just amp-tui
```

## Test Workflow

Use the MCP tools to test fabbro's TUI behavior. The MCP provides these tools:

| Tool | Description |
|------|-------------|
| `launch_tui` | Start a TUI application |
| `send_keys` | Send keyboard input (use `\n` for Enter, `\x1b` for Escape) |
| `send_ctrl` | Send Ctrl+key combinations |
| `capture_screen` | Capture current screen output |
| `expect_text` | Wait for specific text to appear |
| `assert_contains` | Verify expected text is present |
| `close_session` | Clean up the TUI session |

## Critical Setup Notes

### 1. Use bash -c wrapper for commands

Bubbletea apps require a proper PTY. Launch with `bash -c '...'`:

```
# WRONG - will fail with TTY errors
launch_tui(command="/path/to/fabbro review file.md")

# CORRECT - bash provides proper PTY handling
launch_tui(command="bash -c 'cd /test/dir && /path/to/fabbro review file.md 2>&1'")
```

### 2. Initialize fabbro in test directory first

Before launching the TUI, create a temp directory and initialize fabbro:

```bash
TESTDIR=$(mktemp -d)
cp test/fixtures/sample-review.md $TESTDIR/
cd $TESTDIR && fabbro init
```

### 3. Use buffer mode for full TUIs

```
launch_tui(command="...", mode="buffer", dimensions="80x24")
```

Stream mode captures output but doesn't maintain screen state.

### 4. Send keys one at a time with delays

Multi-character sequences may be dropped. Send separately:

```
# WRONG - may only register first character
send_keys("jjj")

# CORRECT - send one at a time
send_keys("j", delay=0.2)
send_keys("j", delay=0.2)
send_keys("j", delay=0.2)
```

### 5. Enter key: use Ctrl+M

The `\n` escape doesn't work reliably. Use `send_ctrl("m")` instead:

```
# WRONG - may not register
send_keys("\n")

# CORRECT - Ctrl+M is Enter
send_ctrl("m")
```

### 6. Double-key commands (like gg)

Send each key separately with delay:

```
send_keys("g", delay=0.2)
send_keys("g", delay=0.2)
```

### 7. Escape key limitation

The `\x1b` escape sequence doesn't work reliably through the MCP string interface. 
This is a known limitation - the escape character gets string-escaped before reaching pexpect.

### 8. Use assert_contains for verification

After sending keys, verify screen state:

```
send_keys("v")
capture_screen()  # View what's on screen
assert_contains("●")  # Verify selection indicator appeared
```

## Test Scenarios

### 1. Basic Navigation Test

```
1. Initialize fabbro in a temp directory
2. Launch: fabbro review test/fixtures/sample-review.md
3. Verify: Screen shows "Review:" header and line numbers
4. Send: j (move down)
5. Verify: Cursor indicator ">" moves to line 2
6. Send: G (jump to end)
7. Verify: Cursor is on last line
8. Send: gg (jump to start)
9. Verify: Cursor is on line 1
10. Send: Q (quit)
```

### 2. Selection Test

```
1. Launch fabbro with sample content
2. Send: v (start selection)
3. Verify: Selection indicator "●" appears
4. Send: jj (extend selection down 2 lines)
5. Verify: 3 lines show selection indicator
6. Send: v (toggle selection off)
7. Verify: Selection indicators removed
8. Cleanup
```

### 3. Annotation Test

```
1. Launch fabbro with sample content (dimensions: 80x24, mode: buffer)
2. Navigate to line 5: jjjjj
3. Start selection: v
4. Add comment: c
5. Verify: "Comment:" prompt appears
6. Type: "test comment\n"
7. Verify: Selection cleared, annotation stored
8. Save: w
9. Quit: Q
10. Verify: Session file created with annotation
11. Cleanup: close_session (always, even on failure)
```

### 4. Command Palette Test

```
1. Launch fabbro with sample content
2. Select a line: v
3. Open palette: <space>
4. Verify: Palette shows "[c]omment [d]elete [q]uestion"
5. Select comment: c
6. Verify: Comment prompt appears
7. Cancel: <escape>
8. Verify: Back to normal mode
```

## Running a Test

Example test execution:

```
# In an amp-tui session, ask:

"Run the Basic Navigation Test:
1. Create a temp directory and run 'fabbro init'
2. Use launch_tui to start 'fabbro review test/fixtures/sample-review.md' 
   with mode='buffer' and dimensions='80x24'
3. Use get_screen to verify the TUI shows the Review header
4. Use send_keys to send 'j' and verify cursor moves
5. Use send_keys to send 'G' and verify jump to end
6. Use send_keys to send 'gg' and verify jump to start  
7. Use send_keys to send 'Q' to quit
8. Use close_session to cleanup
9. Report pass/fail for each step"
```

**Important:** Always call `close_session` at the end, even if a test fails. This prevents orphaned TUI processes.

## Expected Screen Format

The fabbro TUI displays content like:

```
─── Review: 20260114-abcd ─────────────────────────
>    1 │ # Sample Document for TUI Testing
     2 │ 
     3 │ This is a sample document...
     4 │ ...
──────────────────────────────────────────────────
[v]select [SPC]palette [c]omment [d]elete [Q]uit
```

- `>` indicates cursor position
- `●` indicates selected lines
- Line numbers are right-aligned in column 1-4
- `│` separates line numbers from content

## Assertions

When verifying screen output:

- **Cursor position**: Look for `>` at start of line
- **Selection**: Look for `●` indicator
- **Palette open**: Look for "Annotations" header
- **Input mode**: Look for prompt like "Comment:" or "Question:"
- **Session ID**: Appears in title bar after "Review:"
