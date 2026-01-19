---
description: Resume work from a handoff document with context analysis
---

# Resume from Handoff

Resume work from a handoff document through an interactive process.

## Getting Started

**Three scenarios:**

1. **Handoff path provided**: Read and analyze immediately
2. **No path provided**: List available handoffs and ask which to resume
3. **Multiple recent handoffs**: Show options sorted by date

```bash
# List available handoffs
ls -la handoffs/

# Show most recent
ls -t handoffs/ | head -5
```

## Process

### Step 1: Read the Handoff Completely

Read the entire handoff file without limit/offset to capture full context:
- Summary and current state
- Completed vs in-progress vs not-started tasks
- Key decisions and their rationale
- Learnings and gotchas
- Files changed and artifacts created
- Next steps (prioritized)
- Open questions

### Step 2: Verify Current State

Check if codebase matches handoff expectations:

```bash
# Verify we're on the right branch
git branch --show-current

# Check current commit vs handoff
git log --oneline -5

# See any changes since handoff
git status

# If handoff has specific commit, compare
git log [handoff_commit]..HEAD --oneline
```

**Identify any drift:**
- New commits since handoff?
- Uncommitted changes present?
- Different branch than expected?

### Step 3: Read Referenced Documents

If the handoff references other files, read them:
- Plans in `plans/`
- Research in `research/`
- Specs in `specs/`
- Related handoffs

### Step 4: Create Analysis Summary

Present to user:

```
## Handoff Analysis

**Original work**: [Brief description from handoff]
**Handoff date**: [Date]
**Git state at handoff**: [branch] @ [commit]
**Current git state**: [branch] @ [commit]

### State Comparison

[One of:]
- ✓ Codebase matches handoff exactly
- ⚠ [N] commits added since handoff: [brief description]
- ⚠ Uncommitted changes present
- ✗ Different branch - expected [X], on [Y]

### Tasks Status

**Completed** (from handoff):
- [Task 1]
- [Task 2]

**In Progress** (needs continuation):
- [Task with status from handoff]

**Not Started** (remaining):
- [Task 1]
- [Task 2]

### Key Context to Carry Forward

- [Decision 1]: [Rationale - still valid?]
- [Gotcha]: [Still relevant?]

### Recommended Next Actions

Based on handoff priorities and current state:

1. **[Action]**: [Why this first]
2. **[Action]**: [Why this second]

### Open Questions

- [Question from handoff - do we have an answer now?]

---

Shall I proceed with [recommended first action]?
```

### Step 5: Get User Confirmation

Wait for user to:
- Confirm the analysis is accurate
- Approve the recommended first action
- Or redirect to a different priority

### Step 6: Begin Work

Once confirmed:

1. **Create a todo list** from the prioritized next steps
2. **Mark first task as in_progress**
3. **Apply learnings/gotchas** from handoff
4. **Reference the handoff** if questions arise during work

## Handling Edge Cases

### Codebase Has Diverged

If commits were made after the handoff:

```
The codebase has changed since this handoff:

Commits since handoff:
- abc1234: "message 1"
- def5678: "message 2"

These changes might affect:
- [Assessment of impact on planned work]

Options:
1. Proceed with original plan (changes are unrelated)
2. Review new commits first to understand context
3. Ask user for guidance

Recommended: [Your recommendation]
```

### Handoff References Missing Files

If referenced files don't exist:

```
Handoff references files that don't exist:
- plans/2026-01-10-feature.md (not found)
- research/2026-01-09-topic.md (not found)

This might mean:
- Files were renamed or moved
- Work was restructured
- Files are on a different branch

I can:
1. Search for similar files
2. Proceed without those references
3. Ask user for updated locations
```

### Multiple Related Handoffs

If there's a chain of handoffs:

```
Found related handoffs in chronological order:
1. handoffs/2026-01-09_session1.md
2. handoffs/2026-01-10_session2.md
3. handoffs/2026-01-11_session3.md (requested)

I'll read all three to understand the full history, focusing on the most recent for current state.
```

### Stale Handoff (Old Date)

If handoff is more than a few days old:

```
This handoff is from [N] days ago. The codebase may have changed significantly.

Recommended approach:
1. Verify all "completed" items are still in place
2. Check if "in progress" items were finished elsewhere
3. Validate "next steps" are still relevant
4. Look for newer handoffs or related work

Shall I do a full validation before proceeding?
```

## Guidelines

1. **Read completely first**: Don't start work before understanding full context
2. **Verify state**: Always check git state matches expectations
3. **Respect decisions**: Previous decisions had reasons; don't undo without cause
4. **Carry forward learnings**: Apply gotchas and insights from handoff
5. **Track continuity**: Reference handoff when making decisions
6. **Update if needed**: Create new handoff if this session doesn't finish

## Integration with Workflows

**After resuming successfully:**
- Use todo list to track remaining work
- Follow original plan if one exists
- Apply learnings from handoff throughout

**Before ending this session:**
- Create new handoff if work incomplete
- Reference the original handoff for continuity
- Note what was accomplished since resuming
