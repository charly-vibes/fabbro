# CLI Commands

fabbro provides a command-line interface for managing code review sessions.

## Global Options

```
fabbro [command] [flags]
```

## Commands

### `fabbro init`

Initialize fabbro in the current directory.

```bash
fabbro init
```

Creates a `.fabbro/` directory with:
- `sessions/` - Directory for review session files

If already initialized, prints a message and exits successfully.

### `fabbro review`

Start a new review session.

```bash
fabbro review --stdin
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--stdin` | Read content from standard input |

You must provide either a file path or `--stdin`, but not both.

**Example:**

```bash
# Review a single file
cat src/main.go | fabbro review --stdin

# Review a git diff
git diff HEAD~1 | fabbro review --stdin

# Review multiple files concatenated
cat src/*.go | fabbro review --stdin
```

After reading input, launches the TUI for annotation. On save, creates a session file in `.fabbro/sessions/<id>.fem`.

### `fabbro apply`

Extract annotations from a saved session.

```bash
fabbro apply <session-id> [flags]
```

**Arguments:**

| Argument | Required | Description |
|----------|----------|-------------|
| `session-id` | Yes (or use `--file`) | The session ID (shown when session was created) |

**Flags:**

| Flag | Description |
|------|-------------|
| `--json` | Output annotations as JSON |

**Example:**

```bash
# Human-readable output
fabbro apply abc12345
# Output:
# Session: abc12345
# Annotations: 2
#   Line 5: [comment] Consider error handling
#   Line 12: [comment] Extract to function

# JSON output for programmatic use
fabbro apply abc12345 --json
```

**JSON Output Format:**

```json
{
  "sessionId": "abc12345",
  "sourceFile": "src/main.go",
  "annotations": [
    {
      "type": "comment",
      "text": "Consider error handling",
      "startLine": 5,
      "endLine": 5
    },
    {
      "type": "comment",
      "text": "Extract to function",
      "startLine": 12,
      "endLine": 12
    }
  ]
}
```

**Note:** `sourceFile` is empty for stdin sessions.

### `fabbro session`

Manage editing sessions.

#### `fabbro session list`

List all editing sessions.

```bash
fabbro session list
```

Shows all sessions with their ID, creation date, and source file (if any).

#### `fabbro session resume`

Resume a previous editing session.

```bash
fabbro session resume <session-id>
```

Opens the TUI with the session content and any existing annotations, allowing you to continue reviewing.

### `fabbro tutor`

Start the interactive tutorial.

```bash
fabbro tutor
```

Launches a guided lesson in the TUI that teaches fabbro basics:
- Navigation (j/k, gg, G, Ctrl+d/u)
- Selection (v, Esc)
- Annotations (c, d, q, e, u)
- Command palette (Space)
- Saving and quitting

The tutorial is hands-on—you practice each feature on real content. Tutorial sessions are not saved; they're just for practice.

**Note:** Works without `fabbro init`. You can run the tutor from any directory.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error (not initialized, file not found, etc.) |

## Session Files

Session files are stored in `.fabbro/sessions/` with the format `<id>.fem`. They contain YAML frontmatter followed by the annotated content:

```yaml
---
session_id: abc12345
created_at: 2026-01-11T12:00:00Z
source_file: 'src/main.go'
---

func main() {
    fmt.Println("Hello") {>> Consider using log <<}
}
```

**Note:** `source_file` is omitted for stdin sessions.
