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

| Flag | Required | Description |
|------|----------|-------------|
| `--stdin` | Yes | Read content from standard input |

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
| `session-id` | Yes | The session ID (shown when session was created) |

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
---

func main() {
    fmt.Println("Hello") {>> Consider using log <<}
}
```
