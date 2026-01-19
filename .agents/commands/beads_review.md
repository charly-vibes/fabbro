---
description: Perform iterative review of beads issues using the Rule of 5 methodology
---

# Iterative Beads Review (Rule of 5)

Perform thorough beads issue review using the Rule of 5 - iterative refinement until convergence.

## When to Use

- After creating issues from a plan
- Before starting implementation work
- When issues seem stale or misaligned
- Validating issue quality and dependencies

## When NOT to Use

- Single issue review (just read and fix)
- Issues are placeholders (not ready for review)
- Informal tracking (notes, todos)

## Setup

First, gather the issues to review:

```bash
bd list                    # All issues
bd ready                   # Unblocked issues
bd graph                   # Dependency visualization
bd show <id>               # Individual issue details
bd dep tree                # Dependency tree
bd dep cycles              # Check for circular dependencies
```

## Process

Perform 5 passes, each focusing on different aspects. After each pass (starting with pass 2), check for convergence.

### PASS 1 - Completeness & Clarity

**Focus on:**
- Title clearly describes the work
- Description has enough context to implement
- File paths and changes are concrete (not vague)
- Success criteria or tests are defined
- No ambiguous or vague language
- Acceptance criteria clear

**Output format:**
```
PASS 1: Completeness & Clarity

Issues Found:

[CLRT-001] [CRITICAL|HIGH|MEDIUM|LOW] - Issue ID
Title: [Issue title]
Description: [What's unclear or incomplete]
Evidence: [Why this is a problem]
Recommendation: [How to fix - specific command or action]

[CLRT-002] ...
```

**What to look for:**
- Vague titles: "Fix auth" (fix what?)
- No description or minimal description
- "Implement X" without saying how or where
- Missing file paths
- No verification steps
- Unclear done criteria

**Fix commands:**
```bash
bd edit <id> description     # Update description
bd update <id> --title="New clear title"
```

### PASS 2 - Scope & Atomicity

**Focus on:**
- Each issue represents one logical unit of work
- Issues not too large (should complete in one session)
- Issues not too small (trivial changes bundled appropriately)
- Clear boundaries between issues
- No overlapping scope between issues
- Each issue independently valuable

**Prefix:** SCOPE-001, etc.

**What to look for:**
- "Implement entire authentication system" (too large)
- "Fix typo in README line 42" (maybe too small, could bundle)
- Two issues both say "update user model"
- Issue requires changes across 10+ files
- Issue mixes refactoring with feature work

**Fix commands:**
```bash
# Split large issues
bd create --title="Phase 1: ..." --description="..."
bd create --title="Phase 2: ..." --description="..."
bd dep add phase2-id phase1-id  # phase2 depends on phase1

# Merge small issues
bd close small-issue-1 --reason="merged into main-issue"
bd update main-issue --description="Now includes work from small-issue-1"
```

**Convergence Check after Pass 2:**
```
New CRITICAL issues: [count]
Total new issues: [count]
Estimated false positive rate: [percentage]
Status: [CONVERGED | ITERATE | NEEDS_HUMAN]
```

### PASS 3 - Dependencies & Ordering

**Focus on:**
- Dependencies correctly defined
- No missing dependencies (B needs A but not linked)
- No circular dependencies (A→B→C→A)
- Critical path is sensible
- Parallelizable work not falsely serialized
- Dependency rationale is clear

**Prefix:** DEP-001, etc.

**What to look for:**
- Issue requires another to be done but not linked
- Circular dependency chains
- Everything depends on one issue (bottleneck)
- No dependencies when clear order exists
- Dependencies prevent parallel work unnecessarily

**Fix commands:**
```bash
bd dep cycles                           # Find circular dependencies
bd dep tree                            # Visualize dependencies
bd dep add <blocked-id> <blocker-id>   # Add missing dependency
bd dep remove <blocked-id> <blocker-id> # Remove incorrect dependency
```

**Convergence Check after Pass 3:**
```
New CRITICAL issues: [count]
Total new issues: [count]
New issues vs Pass 2: [percentage change]
Status: [CONVERGED | ITERATE | NEEDS_HUMAN]
```

### PASS 4 - Plan & Spec Alignment

**Focus on:**
- Issues trace back to plan phases
- Plan references in descriptions
- Related specs linked where applicable
- TDD approach clear (tests defined before impl)
- All plan phases have corresponding issues
- Issue breakdown matches plan structure

**Prefix:** ALIGN-001, etc.

**What to look for:**
- Plan has 5 phases but only 3 issues
- Issue doesn't reference source plan
- Plan says "test first" but issue doesn't mention tests
- Spec requirements not covered by any issue
- Issue contradicts plan approach

**Fix commands:**
```bash
bd update <id> --description="...

Ref: plans/2026-01-12-feature.md#phase-2"
```

**Convergence Check after Pass 4:**
```
New CRITICAL issues: [count]
Total new issues: [count]
New issues vs Pass 3: [percentage change]
Estimated false positive rate: [percentage]
Status: [CONVERGED | ITERATE | NEEDS_HUMAN]
```

### PASS 5 - Executability & Handoff

**Focus on:**
- Can be picked up by any developer/agent
- No implicit knowledge required
- Verification steps clear and specific
- Handoff points defined for multi-issue work
- Priority and labels appropriate
- Estimation realistic (if used)

**Prefix:** EXEC-001, etc.

**What to look for:**
- "You know what to do" (no, they don't)
- Assumes knowledge of previous conversations
- "Test it" without saying how
- No verification steps
- Priority/labels missing or incorrect

**Final Convergence Check:**
```
New CRITICAL issues: [count]
Total new issues: [count]
New issues vs Pass 4: [percentage change]
Estimated false positive rate: [percentage]
Status: [CONVERGED | NEEDS_ITERATION | ESCALATE_TO_HUMAN]
```

## Convergence Criteria

**CONVERGED** if:
- No new CRITICAL issues AND
- New issue rate < 10% vs previous pass AND
- False positive rate < 20%

**ITERATE** if:
- New issues found that need addressing

**ESCALATE_TO_HUMAN** if:
- After 5 passes, still finding CRITICAL issues OR
- Uncertain about scope or dependencies OR
- False positive rate > 30%

**If converged before Pass 5:** Stop and report.

## Final Report

```
# Beads Review - Final Report

**Scope:** [All issues / Feature X / Milestone Y]
**Convergence:** Pass [N]

## Summary

Total Issues Reviewed: [count]

Issues Found by Severity:
- CRITICAL: [count] - Must fix before work starts
- HIGH: [count] - Should fix before work starts
- MEDIUM: [count] - Consider addressing
- LOW: [count] - Nice to have

## Top 3 Critical Findings

1. [DEP-001] Circular dependency detected
   Issues: #42 → #43 → #44 → #42
   Impact: Cannot start any of these issues
   Fix: `bd dep remove 42 44` to break cycle

2. [SCOPE-002] Issue too large to complete
   Issue: #38 "Implement authentication system"
   Impact: Unmanageable scope, blocks other work
   Fix: Split into 5 issues for each plan phase

3. [CLRT-003] Missing implementation details
   Issue: #29 "Update API"
   Impact: Cannot implement without more info
   Fix: Add file paths, endpoints, and success criteria

## Recommended bd Commands

```bash
# Fix circular dependency
bd dep remove 42 44

# Split large issue
bd create --title="Phase 1: JWT tokens" --description="..."
bd close 38 --reason="split into phase issues"

# Update missing details
bd edit 29 description
```

## Issue Quality Assessment

- Clarity: [EXCELLENT|GOOD|FAIR|POOR]
- Scope: [EXCELLENT|GOOD|FAIR|POOR]
- Dependencies: [EXCELLENT|GOOD|FAIR|POOR]
- Completeness: [EXCELLENT|GOOD|FAIR|POOR]

## Verdict

[READY_TO_WORK | NEEDS_UPDATES | NEEDS_REPLANNING]

**Rationale:** [1-2 sentences]
```

## Rules

1. **Reference beads issue IDs** - Use exact IDs
2. **Provide actionable bd commands** - Show how to fix
3. **Check actual content** - Don't assume, verify with `bd show`
4. **Prioritize correctly**:
   - CRITICAL: Blocks all work (circular deps, missing info)
   - HIGH: Blocks specific work or causes confusion
   - MEDIUM: Could be clearer but workable
   - LOW: Minor improvements
5. **Stop when converged** - Don't force all 5 passes

## Variations

### For small issue sets (<5 issues)

Combine passes:
- PASS 1: Completeness + Scope
- PASS 2: Dependencies + Alignment
- PASS 3: Final Review

### For epic-level review

Add emphasis on:
- Epic structure and milestones clear
- Cross-cutting concerns identified
- Risk distribution across phases
- Checkpoint issues for validation
- Integration points between epics

### For inherited/stale issues

Add verification passes:
- PASS 6: Validate against current codebase state
- PASS 7: Check for already-completed work
- PASS 8: Confirm assumptions still valid

## References

- **Steve Yegge's Rule of 5:** https://steve-yegge.medium.com/six-new-tips-for-better-coding-with-agents-d4e9c86e42a9
- **Beads Documentation:** Use `bd help` for command reference
