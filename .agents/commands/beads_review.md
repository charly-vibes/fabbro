---
description: Perform iterative review of beads issues using the Rule of 5 methodology
---

# Iterative Beads Review

Perform thorough beads issue review using the Rule of 5 - iterative refinement until convergence.

## When to Use

- After creating issues from a plan
- Before starting implementation work
- When issues seem stale or misaligned
- Validating issue quality and dependencies

## Setup

First, gather the issues to review:

```bash
bd list                    # All issues
bd ready                   # Unblocked issues
bd graph                   # Dependency visualization
bd show <id>               # Individual issue details
```

## Process

Perform 5 passes, each focusing on different aspects. After each pass, check for convergence.

### PASS 1 - Completeness & Clarity

Focus on:
- Title clearly describes the work
- Description has enough context to implement
- File paths and changes are concrete
- Success criteria / tests are defined
- No ambiguous or vague language

Output format:
- Issue ID (CLRT-001, etc.)
- Severity: CRITICAL | HIGH | MEDIUM | LOW
- Beads Issue: fabbro-xxx
- Description
- Recommendation

### PASS 2 - Scope & Atomicity

Focus on:
- Each issue represents one logical unit of work
- Issues not too large (should complete in one session)
- Issues not too small (trivial changes bundled)
- Clear boundaries between issues
- No overlapping scope between issues

Prefix: SCOPE-001, etc.

### PASS 3 - Dependencies & Ordering

Focus on:
- Dependencies correctly defined (use `bd dep tree`)
- No missing dependencies
- No circular dependencies (use `bd dep cycles`)
- Critical path is sensible
- Parallelizable work not falsely serialized

Prefix: DEP-001, etc.

### PASS 4 - Plan & Spec Alignment

Focus on:
- Issues trace back to plan phases
- Plan references in descriptions (`Ref: plans/...`)
- Related specs linked where applicable
- TDD approach clear (tests defined before impl)
- All plan phases have corresponding issues

Prefix: ALIGN-001, etc.

### PASS 5 - Executability & Handoff

Focus on:
- Can be picked up by any developer/agent
- No implicit knowledge required
- Verification steps clear
- Handoff points defined for multi-issue work
- Priority and labels appropriate

Prefix: EXEC-001, etc.

## Convergence Check

After each pass (starting with pass 2), report:
1. Number of new CRITICAL issues found
2. Number of new issues vs previous pass
3. Estimated false positive rate
4. Convergence status:
   - **CONVERGED**: No new CRITICAL, <10% new issues, <20% false positives
   - **ITERATE**: Continue to next pass
   - **NEEDS_HUMAN**: Found blocking issues that need human judgment

## Final Report

Provide:
- Total issues by severity
- Top 3 most critical findings
- Recommended issue updates (specific `bd update` commands)
- Convergence assessment
- Verdict: READY_TO_WORK | NEEDS_UPDATES | NEEDS_REPLANNING

## Rules

1. Reference beads issue IDs specifically (fabbro-xxx)
2. Provide actionable `bd` commands for fixes
3. Check actual issue content with `bd show`
4. Prioritize: Dependencies > Scope > Clarity > Alignment
5. If converged before pass 5, stop and report

## Fixing Issues

Common fixes with beads commands:

```bash
# Update description
bd edit <id> description

# Add dependency
bd dep add <blocked-id> <blocker-id>

# Remove incorrect dependency
bd dep remove <blocked-id> <blocker-id>

# Update title
bd update <id> --title="New title"

# Add reference to plan
bd update <id> --description="...\n\nRef: plans/YYYY-MM-DD-name.md#phase-N"

# Check for cycles
bd dep cycles
```

## Variations

### For small issue sets (<5 issues)

Combine passes:
- PASS 1: Completeness + Scope
- PASS 2: Dependencies + Alignment
- PASS 3: Final Review

### For epic-level review

Add emphasis on:
- Epic structure and milestones
- Cross-cutting concerns identified
- Risk distribution across phases
- Checkpoint issues for validation

### For inherited/stale issues

Add verification passes:
- PASS 6: Validate against current codebase state
- PASS 7: Check for already-completed work
- PASS 8: Confirm assumptions still valid
