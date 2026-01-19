---
description: Create handoff document for transferring work to another session
---

# Create Handoff Document

Create a concise handoff document to transfer context to another agent session. The goal is to compact and summarize your context without losing key details.

## Process

### Step 1: Gather Session Context

Collect from your current session:

1. **Tasks worked on**: What was attempted, completed, or blocked
2. **Decisions made**: Key choices and their rationale
3. **Problems encountered**: Bugs, blockers, unexpected situations
4. **Learnings**: Insights about the codebase, patterns, gotchas
5. **Artifacts created**: Files, commits, PRs, issues

### Step 2: Get Current Git State

```bash
# Get current state
git branch --show-current
git rev-parse --short HEAD
git status --short
git log --oneline -5

# Get date for filename
date +%Y-%m-%d_%H-%M-%S
```

### Step 3: Create Handoff File

Save to `handoffs/YYYY-MM-DD_HH-MM-SS_description.md`:

```markdown
---
date: [ISO timestamp]
git_commit: [short hash from step 2]
branch: [branch name]
status: handoff
---

# Handoff: [Brief Description of Work]

## Summary

[2-3 sentences: What was the goal? How far did we get? What's the state?]

## Tasks

### Completed
- [Task 1 that was finished]
- [Task 2 that was finished]

### In Progress
- [Task that was started but not finished]
  - **Status**: [Where it stands]
  - **Blocked by**: [If applicable]

### Not Started
- [Task that was planned but not begun]

## Critical Context

### Key Decisions Made
1. [Decision 1]: [Rationale]
2. [Decision 2]: [Rationale]

### Learnings / Gotchas
- [Learning 1 with file:line reference if applicable]
- [Gotcha about the codebase that future sessions should know]

### Problems Encountered
- [Problem 1]: [How it was resolved or current state]

## Files Changed

List all files modified during this session:

```
path/to/file1.ext:line-range  - [what changed]
path/to/file2.ext:line-range  - [what changed]
```

## Artifacts Created

- [x] Commit: `abc1234` - "commit message"
- [x] File: `path/to/new/file.ext`
- [ ] PR: #123 (draft/ready)
- [ ] Issue: #456

## Next Steps

**Priority order for resuming work:**

1. [Highest priority next action]
   - Files: [relevant files]
   - Approach: [brief description]

2. [Second priority action]
   - Files: [relevant files]
   - Approach: [brief description]

3. [Third priority action]

## Open Questions

- [Question that couldn't be resolved this session]
- [Question for the user to consider]

## References

- Plan: `plans/YYYY-MM-DD-name.md` (if applicable)
- Research: `research/YYYY-MM-DD-topic.md` (if applicable)
- Related handoffs: `handoffs/YYYY-MM-DD_earlier.md` (if continuing previous work)
```

### Step 4: Commit the Handoff

```bash
git add handoffs/
git commit -m "docs: add handoff for [brief description]"
```

### Step 5: Report to User

```
Handoff created at: handoffs/YYYY-MM-DD_HH-MM-SS_description.md

To resume in a new session:
/resume_handoff handoffs/YYYY-MM-DD_HH-MM-SS_description.md

Key items for next session:
1. [Most important next step]
2. [Second priority]
```

## Guidelines

1. **Be thorough but concise**: Include everything needed, nothing more
2. **Use file:line references**: Make it easy to find relevant code
3. **Capture the "why"**: Decisions without rationale lose value
4. **Note blockers explicitly**: Don't bury blockers in prose
5. **Order next steps by priority**: Most important first
6. **Include open questions**: Unresolved issues are valuable context
7. **Reference related docs**: Link to plans, research, specs

## When to Create Handoffs

**Always create a handoff when:**
- Ending a session with incomplete work
- Work will be continued by a different agent/session
- Complex context that would be lost
- Multiple sessions expected for a task

**Skip handoffs when:**
- Task is fully complete
- Context is trivial (simple, isolated change)
- User explicitly says no handoff needed

## Handoff Quality Checklist

Before finishing:
- [ ] Git state captured (branch, commit)
- [ ] All modified files listed
- [ ] Key decisions documented with rationale
- [ ] Blockers explicitly called out
- [ ] Next steps prioritized and actionable
- [ ] References to related docs included
- [ ] Open questions listed
