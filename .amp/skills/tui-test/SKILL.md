---
name: tui-test
description: Run TUI integration tests for fabbro's terminal UI via mcp-tui-test MCP
---

# TUI Integration Testing Skill

You are running integration tests on fabbro's Terminal User Interface using the `mcp-tui-test` MCP server.

## Usage

Start Amp with the MCP configured, then load this skill:
```bash
just amp-tui
# Then in Amp: /skill tui-test
# Then: "run all TUI integration tests"
```

## Available MCP Tools

Tools from the `tui-test` MCP server (verify exact names on first use):

- `launch_tui` - Launch a TUI application
- `send_keys` - Send keyboard input
- `send_ctrl` - Send Ctrl+key combinations  
- `get_screen` - Capture screen output
- `wait_for_text` - Wait for text to appear
- `assert_text` - Verify text is present
- `close_session` - Close session

> **Note:** Tool names may be prefixed with `mcp__tui-test__` depending on MCP integration. Check available tools if calls fail.

## Test Setup

Before running TUI tests:

1. Create a temporary test directory
2. Initialize fabbro: `fabbro init`
3. Use the test fixture: `test/fixtures/sample-review.md`

## Test Execution Pattern

For each test:

```
1. Setup: Create temp dir, fabbro init
2. Launch: launch_tui with fabbro review command, mode="buffer", dimensions="80x24"
3. Wait: wait_for_text for "Review:" to confirm TUI loaded
4. Execute: send_keys for test actions
5. Verify: get_screen and check expected output
6. Cleanup: close_session (ALWAYS - even on failure)
7. Report: Pass/Fail with details
```

**Important:** Always call `close_session` even if a test fails mid-execution. Orphaned TUI processes can block subsequent tests.

**Buffer vs Stream Mode:**
- Use `mode="buffer"` for full TUI apps (fabbro) - provides screen coordinates
- Use `mode="stream"` for simple CLI tools - faster, less overhead

## Key Mappings for send_keys

| Key | send_keys value |
|-----|-----------------|
| Enter | `\n` |
| Escape | `\x1b` |
| Tab | `\t` |
| Down arrow | `\x1b[B` |
| Up arrow | `\x1b[A` |
| Regular keys | just the character, e.g., `j`, `k`, `v` |

For Ctrl combinations, use `send_ctrl` with the key letter.

## Screen Parsing

The fabbro TUI format:
```
─── Review: {session-id} ───────────────
{cursor}{sel} {linenum} │ {content}
...
──────────────────────────────────────
{hotkey bar or prompt}
```

Where:
- `{cursor}` = `>` if current line, ` ` otherwise
- `{sel}` = `●` if selected, ` ` otherwise
- `{linenum}` = right-aligned 3-digit line number

## Test Cases

### TC1: Basic Launch and Quit
- Launch fabbro review with sample file
- Verify "Review:" appears in output
- Send `Q` to quit
- Verify session closes cleanly

### TC2: Navigation (j/k/G/gg)
- Launch TUI
- Send `j` 3 times, verify cursor on line 4
- Send `G`, verify cursor on last line
- Send `gg`, verify cursor on line 1
- Quit

### TC3: Selection Toggle
- Launch TUI
- Send `v` to start selection
- Verify `●` appears on current line
- Send `v` again to cancel
- Verify `●` removed
- Quit

### TC4: Multi-line Selection
- Launch TUI
- Send `v` then `jjj` (select 4 lines)
- Verify 4 lines show `●`
- Send `\x1b` (Escape) to cancel
- Verify selection cleared
- Quit

### TC5: Add Comment Annotation
- Launch TUI
- Navigate: `jjj`
- Select: `v`
- Comment: `c`
- Verify "Comment:" prompt appears
- Type: `test annotation\n`
- Verify selection cleared
- Save: `w`
- Quit: `Q`
- Verify session file contains annotation

## Reporting

After each test, report:
```
✓ TC1: Basic Launch and Quit - PASS
✗ TC2: Navigation - FAIL: cursor not on expected line
  Expected: line 4, Got: line 3
```
