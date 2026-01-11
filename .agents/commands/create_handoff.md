---
description: Create handoff document for transferring work to another session
---

# Create Handoff

Create a concise handoff document to transfer context to another agent session. The goal is to compact and summarize your context without losing key details.

## Process

### 1. Gather Metadata

```bash
# Get current state
git branch --show-current
git rev-parse --short HEAD
date -Iseconds
```

### 2. Determine Filepath

Create file at `handoffs/YYYY-MM-DD_HH-MM-SS_description.md`:
- If working on a beads issue: `handoffs/YYYY-MM-DD_HH-MM-SS_bd-XXXX_description.md`
- Example: `handoffs/2026-01-11_14-30-00_bd-a1b2_implement-fem-parser.md`
- Without issue: `handoffs/2026-01-11_14-30-00_setup-project-structure.md`

### 3. Write Handoff Document

Use this template:

```markdown
---
date: [ISO timestamp with timezone]
git_commit: [short hash]
branch: [branch name]
beads_issue: [bd-XXXX if applicable]
status: handoff
---

# Handoff: [brief description]

## Task(s)

[Description of tasks worked on with status: completed, in-progress, or planned]

If working from a plan: reference `plans/YYYY-MM-DD-description.md` and note which phase.

## Critical References

[2-3 most important files/docs that must be read to continue]

## Recent Changes

[Files modified in `path/to/file.ext:line` format]

## Learnings

[Important discoveries: patterns, bug root causes, gotchas]

## Artifacts

[Exhaustive list of files created/updated as paths]

## Next Steps

[Prioritized list of what to do next]

## Notes

[Other useful context that doesn't fit above]
```

### 4. Update Beads (if applicable)

If working on a tracked issue:

```bash
# Add comment to issue with handoff reference
bd comments add bd-XXXX "Handoff created: handoffs/YYYY-MM-DD_HH-MM-SS_description.md"

# Sync beads
bd sync
```

### 5. Commit and Respond

```bash
git add handoffs/
git commit -m "Add handoff: [brief description]"
```

Respond to user:

```
Handoff created! Resume in a new session with:

/resume_handoff handoffs/YYYY-MM-DD_HH-MM-SS_description.md
```

## Guidelines

- **More information, not less** - this is the minimum structure
- **Be precise** - include file:line references
- **Avoid large code blocks** - prefer file references
- **Reference beads issues** - link to `bd-XXXX` when applicable
