---
description: Resume work from a handoff document with context analysis
---

# Resume Handoff

Resume work from a handoff document through an interactive process.

## Initial Response

1. **If a handoff path was provided:**
   - Read the handoff document FULLY
   - Read any plans/research it references
   - Begin analysis and propose next steps

2. **If a beads issue ID was provided (e.g., bd-a1b2):**
   - Run `bd show bd-XXXX` to get issue context
   - Look for handoffs: `ls handoffs/*bd-XXXX*`
   - Use the most recent handoff (by timestamp in filename)
   - If none found, inform the user

3. **If no parameters:**
   ```
   I'll help you resume from a handoff. Available options:

   - Provide a handoff path: `/resume_handoff handoffs/YYYY-MM-DD_description.md`
   - Provide a beads issue: `/resume_handoff bd-a1b2`

   Recent handoffs:
   [list contents of handoffs/ directory]
   ```

## Process

### Step 1: Read and Analyze

1. **Read handoff completely** (no limit/offset)
2. **Extract sections:**
   - Tasks and statuses
   - Recent changes
   - Learnings
   - Artifacts
   - Next steps

3. **Read referenced files:**
   - Plans from `plans/`
   - Research from `research/`
   - Specs from `specs/`
   - Files mentioned in "Recent Changes"

4. **Check beads context (if applicable):**
   ```bash
   bd show bd-XXXX
   bd comments bd-XXXX
   ```

### Step 2: Verify Current State

1. **Check for changes since handoff:**
   ```bash
   git log --oneline [handoff_commit]..HEAD
   git diff [handoff_commit]
   ```

2. **Verify mentioned files still exist**

3. **Identify any conflicts or drift**

### Step 3: Present Analysis

```
I've analyzed the handoff from [date]. Current situation:

**Original Tasks:**
- [Task 1]: [status] → [current state]
- [Task 2]: [status] → [current state]

**Key Learnings:**
- [Learning with file:line] - [still valid?]

**Changes Since Handoff:**
- [commits or modifications since]

**Recommended Next Actions:**
1. [Priority action from handoff]
2. [Second priority]

**Issues Identified:**
- [Any conflicts or problems]

Shall I proceed with [action 1]?
```

### Step 4: Create Action Plan

1. **Create todo list** from handoff's next steps
2. **Add any new tasks** discovered during analysis
3. **Get confirmation** before starting

### Step 5: Begin Work

- Start with first approved task
- Apply learnings from handoff
- Update progress as you go

## Scenarios

### Clean Continuation
- All changes present, no conflicts
- Proceed with recommended actions

### Diverged Codebase
- Changes since handoff
- Reconcile differences, adapt plan

### Incomplete Work
- Tasks marked "in-progress"
- Complete unfinished work first

### Stale Handoff
- Significant time passed
- Re-evaluate strategy before proceeding

## Guidelines

- **Verify before acting** - don't assume handoff matches current state
- **Leverage learnings** - apply documented patterns and avoid noted pitfalls
- **Track continuity** - use todo list to maintain progress
- **Consider new handoff** - create one when session ends
