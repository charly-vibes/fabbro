---
description: Perform iterative review of implementation plans using the Rule of 5 methodology
---

# Iterative Plan Review (Rule of 5)

Perform thorough implementation plan review using the Rule of 5 - iterative refinement until convergence.

## When to Use

- Reviewing plans before implementation begins
- Validating feasibility and completeness
- Checking alignment with specs and project goals
- Ensuring plans follow TDD methodology

## When NOT to Use

- Plan is a rough draft (finish it first)
- Quick notes or brainstorming
- Already in implementation (too late for major changes)

## Process

Perform 5 passes, each focusing on different aspects. After each pass (starting with pass 2), check for convergence.

### PASS 1 - Feasibility & Risk

**Focus on:**
- Technical feasibility of proposed changes
- Identified risks and mitigations
- Dependencies on external factors
- Assumptions that need validation
- Potential blockers not addressed
- Resource requirements realistic
- Technology choices appropriate

**Output format:**
```
PASS 1: Feasibility & Risk

Issues Found:

[FEAS-001] [CRITICAL|HIGH|MEDIUM|LOW] - Phase/Section
Description: [What's not feasible or risky]
Evidence: [Why this is a concern]
Recommendation: [How to address]

[FEAS-002] ...

Feasibility Assessment: [EXCELLENT|GOOD|FAIR|POOR]
```

### PASS 2 - Completeness & Scope

**Focus on:**
- Missing phases or steps
- Undefined success criteria
- Gaps between current and desired state
- Out of scope clearly defined
- All files/changes identified
- Dependencies between phases
- Rollback strategy present

**Prefix:** COMP-001, etc.

**Convergence Check after Pass 2:**
```
New CRITICAL issues: [count]
Total new issues: [count]
Estimated false positive rate: [percentage]
Status: [CONVERGED | ITERATE | NEEDS_HUMAN]
```

### PASS 3 - TDD & Testing Alignment

**Focus on:**
- Tests planned before implementation
- Success criteria are testable
- Test types appropriate (unit, integration, e2e)
- Verification steps defined
- Red-Green-Refactor cycle clear
- Test coverage goals specified
- Manual verification where automated isn't possible

**Prefix:** TDD-001, etc.

**Convergence Check after Pass 3:**
```
New CRITICAL issues: [count]
Total new issues: [count]
New issues vs Pass 2: [percentage change]
Status: [CONVERGED | ITERATE | NEEDS_HUMAN]
```

### PASS 4 - Ordering & Dependencies

**Focus on:**
- Phases in correct order
- Dependencies between phases clear
- Parallelizable work identified
- Critical path identified
- Blocking issues acknowledged
- Each phase independently verifiable
- Clean handoff points between phases

**Prefix:** ORD-001, etc.

**Convergence Check after Pass 4:**
```
New CRITICAL issues: [count]
Total new issues: [count]
New issues vs Pass 3: [percentage change]
Estimated false positive rate: [percentage]
Status: [CONVERGED | ITERATE | NEEDS_HUMAN]
```

### PASS 5 - Clarity & Executability

**Focus on:**
- Specific enough for implementation
- File paths and changes concrete
- No ambiguous instructions
- Clear handoff points between phases
- Can be executed by different people/agents
- Success criteria unambiguous
- No implicit knowledge required

**Prefix:** EXEC-001, etc.

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
- Uncertain about feasibility OR
- False positive rate > 30%

**If converged before Pass 5:** Stop and report.

## Final Report

```
# Plan Review - Final Report

**Plan:** [path/title]
**Convergence:** Pass [N]

## Summary

Total Issues by Severity:
- CRITICAL: [count] - Must fix before implementing
- HIGH: [count] - Should fix before implementing
- MEDIUM: [count] - Consider addressing
- LOW: [count] - Nice to have

## Top 3 Critical Findings

1. [FEAS-001] [Description] - Phase
   Impact: [Why this matters]
   Fix: [What to do]

2. [TDD-002] [Description] - Phase
   Impact: [Why this matters]
   Fix: [What to do]

3. [EXEC-001] [Description] - Phase
   Impact: [Why this matters]
   Fix: [What to do]

## Quality Assessment

- Feasibility: [EXCELLENT|GOOD|FAIR|POOR]
- Completeness: [EXCELLENT|GOOD|FAIR|POOR]
- TDD Alignment: [EXCELLENT|GOOD|FAIR|POOR]
- Ordering: [EXCELLENT|GOOD|FAIR|POOR]
- Executability: [EXCELLENT|GOOD|FAIR|POOR]

## Recommended Actions

1. [Action 1 - specific and actionable]
2. [Action 2 - specific and actionable]
3. [Action 3 - specific and actionable]

## Verdict

[READY_TO_IMPLEMENT | NEEDS_REVISION | NEEDS_MORE_RESEARCH]

**Rationale:** [1-2 sentences]
```

## Rules

1. **Be specific** - Reference phase/section exactly
2. **Provide alternatives** - Don't just identify problems, suggest solutions
3. **Validate against codebase** - Check that file paths exist, patterns match
4. **Prioritize correctly**:
   - CRITICAL: Blocks implementation, fundamentally flawed
   - HIGH: Significant risk or gaps
   - MEDIUM: Could cause problems, worth fixing
   - LOW: Minor improvements
5. **Stop when converged** - Don't force all 5 passes

## Variations

### For small plans (single phase)

Combine passes:
- PASS 1: Feasibility + Completeness
- PASS 2: TDD Alignment + Executability
- PASS 3: Final Review

### For refactoring plans

Add emphasis on:
- Behavioral preservation guarantees
- Test coverage before refactoring
- Incremental verification at each step
- Rollback points defined

### For high-risk plans

Add verification passes:
- PASS 6: Cross-check with existing code
- PASS 7: Validate assumptions with codebase search
- PASS 8: Confirm no conflicts with in-progress work

## References

- **Steve Yegge's Rule of 5:** https://steve-yegge.medium.com/six-new-tips-for-better-coding-with-agents-d4e9c86e42a9
- **TDD Principles:** Red-Green-Refactor, Test First
- **Plan Quality:** Feasibility, Completeness, TDD, Ordering, Executability
