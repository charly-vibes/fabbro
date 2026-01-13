# Claude Code Integration

fabbro is designed to work with Claude Code (and similar AI coding assistants) to enable structured human-AI code review workflows.

## Concept

The typical workflow:

1. **AI generates code** — Claude Code writes or modifies code
2. **Human reviews with fabbro** — You annotate the code with feedback
3. **AI processes annotations** — Claude Code reads the JSON output and addresses feedback
4. **Iterate** — Repeat until the code meets your standards

## Setup

Add fabbro to your project:

```bash
fabbro init
```

## Workflow Example

### 1. Review AI-generated code

After Claude Code generates a file:

```bash
cat src/new_feature.go | fabbro review --stdin
```

Add annotations in the TUI:
- "This function should handle errors"
- "Consider extracting this to a separate module"
- "Add tests for edge cases"

### 2. Extract annotations

```bash
fabbro apply <session-id> --json > feedback.json
```

### 3. Feed back to Claude Code

Share the JSON with Claude Code:

```
Here's my feedback on the code you generated:

{paste feedback.json contents}

Please address each annotation.
```

### 4. Review again

After Claude Code makes changes, review again if needed.

## JSON Output Format

The `--json` output is designed for easy parsing by AI assistants:

```json
{
  "sessionId": "abc12345",
  "annotations": [
    {
      "type": "comment",
      "text": "Add error handling for network failures",
      "startLine": 15,
      "endLine": 15
    },
    {
      "type": "comment",
      "text": "This logic should be extracted to a helper function",
      "startLine": 42,
      "endLine": 42
    }
  ]
}
```

## Future Integration

Planned features for tighter Claude Code integration:

- **AGENTS.md commands** — Custom slash commands for fabbro workflows
- **Automatic session creation** — Claude Code triggers review sessions
- **Inline feedback** — Annotations flow directly back to Claude Code context
- **Diff-aware reviews** — Review only changed lines in a commit

## Tips

- Keep annotations actionable and specific
- Reference line numbers when discussing changes
- Use the JSON output for precise, parseable feedback
- Consider creating review templates for common patterns
