---
description: Perform iterative review of implementation plans using the Rule of 5 methodology
---

# Iterative Plan Review

Perform thorough implementation plan review using the Rule of 5 - iterative refinement until convergence.

## When to Use

- Reviewing plans before implementation begins
- Validating feasibility and completeness
- Checking alignment with specs and project goals
- Ensuring plans follow TDD and SDD methodology

## Process

Perform 5 passes, each focusing on different aspects. After each pass, check for convergence.

### PASS 1 - Feasibility & Risk

Focus on:
- Technical feasibility of proposed changes
- Identified risks and mitigations
- Dependencies on external factors
- Assumptions that need validation
- Potential blockers not addressed

Output format:
- Issue ID (FEAS-001, etc.)
- Severity: CRITICAL | HIGH | MEDIUM | LOW
- Location (phase/section)
- Description
- Recommendation

### PASS 2 - Completeness & Scope

Focus on:
- Missing phases or steps
- Undefined success criteria
- Gaps between current and desired state
- Out of scope clearly defined
- All files/changes identified

Prefix: COMP-001, etc.

### PASS 3 - Spec & TDD Alignment

Focus on:
- Links to spec files in `specs/`
- Tests planned before implementation
- Success criteria are testable
- Scenarios from specs covered
- Verification steps defined

Prefix: TDD-001, etc.

### PASS 4 - Ordering & Dependencies

Focus on:
- Phases in correct order
- Dependencies between phases clear
- Parallelizable work identified
- Critical path identified
- Rollback strategy if needed

Prefix: ORD-001, etc.

### PASS 5 - Clarity & Executability

Focus on:
- Specific enough for implementation
- File paths and changes concrete
- No ambiguous instructions
- Clear handoff points between phases
- Beads issues can be created from phases

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
- Recommended revisions
- Convergence assessment
- Verdict: READY_TO_IMPLEMENT | NEEDS_REVISION | NEEDS_MORE_RESEARCH

## Rules

1. Be specific with phase/section references
2. Provide actionable suggestions for improvements
3. Validate against fabbro's SDD methodology
4. Prioritize: Feasibility > TDD Alignment > Completeness > Clarity
5. If converged before pass 5, stop and report

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
