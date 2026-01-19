---
description: Perform iterative code review using the Rule of 5 methodology
---

# Iterative Code Review (Rule of 5)

Perform thorough code review using the Rule of 5 - iterative refinement until convergence.

## When to Use

- Reviewing your own code before committing
- Doing a quick quality check on a feature
- Working solo without a code review partner
- You want high-quality review without setting up multi-agent systems

## When NOT to Use

- Security-critical code (use full multi-agent review instead)
- Production systems with high reliability requirements (add human review)
- Very large changes (>1000 LOC) - consider chunking first

## Process

Perform 5 passes, each focusing on different aspects. After each pass (starting with pass 2), check for convergence.

### PASS 1 - Security & Safety

**Focus on:**
- Input validation and sanitization
- Authentication and authorization checks
- SQL injection, XSS, CSRF vulnerabilities
- Secret management and data exposure
- Error handling that doesn't leak information
- File path traversal, command injection
- Cryptographic issues

**Output format:**
```
PASS 1: Security & Safety

Issues Found:

[SEC-001] [CRITICAL|HIGH|MEDIUM|LOW] - file.ts:line
Description: [What the security issue is]
Evidence: [Code snippet or explanation]
Recommendation: [How to fix with code example]

[SEC-002] ...

Security Assessment: [EXCELLENT|GOOD|FAIR|POOR]
```

### PASS 2 - Performance & Scalability

**Focus on:**
- Time complexity (O(n²), O(n³) patterns)
- Database queries (N+1, missing indexes)
- Memory allocation and potential leaks
- Unnecessary loops or iterations
- Caching opportunities
- Network round trips
- Resource cleanup

**Prefix:** PERF-001, etc.

**Convergence Check after Pass 2:**
```
New CRITICAL issues: [count]
Total new issues: [count]
Estimated false positive rate: [percentage]
Status: [CONVERGED | ITERATE | NEEDS_HUMAN]
```

### PASS 3 - Maintainability & Readability

**Focus on:**
- Code clarity and naming
- Documentation (comments, docstrings)
- Pattern consistency
- Technical debt indicators
- DRY violations
- Magic numbers and hard-coded values
- Function/method length
- Cyclomatic complexity

**Prefix:** MAINT-001, etc.

**Convergence Check after Pass 3:**
```
New CRITICAL issues: [count]
Total new issues: [count]
New issues vs Pass 2: [percentage change]
Status: [CONVERGED | ITERATE | NEEDS_HUMAN]
```

### PASS 4 - Correctness & Requirements

**Focus on:**
- Does it do what it's supposed to do?
- Edge case handling
- Test coverage gaps
- Behavioral correctness
- Requirements satisfaction
- Off-by-one errors
- Null/undefined handling
- Type safety

**Prefix:** REQ-001, etc.

**Convergence Check after Pass 4:**
```
New CRITICAL issues: [count]
Total new issues: [count]
New issues vs Pass 3: [percentage change]
Estimated false positive rate: [percentage]
Status: [CONVERGED | ITERATE | NEEDS_HUMAN]
```

### PASS 5 - Operations & Reliability

**Focus on:**
- Failure modes and error handling
- Timeout and retry logic
- Observability (logging, metrics)
- Resource management (connections, files)
- Deployment considerations
- Configuration management
- Graceful degradation
- Recovery mechanisms

**Prefix:** OPS-001, etc.

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
- Uncertain about severity or correctness OR
- False positive rate > 30%

**If converged before Pass 5:** Stop and report. Don't continue unnecessarily.

## Final Report

```
# Code Review - Final Report

**Files Reviewed:** [list]
**Convergence:** Pass [N]

## Summary

Total Issues by Severity:
- CRITICAL: [count] - Must fix before merge
- HIGH: [count] - Should fix before merge
- MEDIUM: [count] - Consider addressing
- LOW: [count] - Nice to have

## Top 3 Critical Findings

1. [SEC-001] [Description] - file.ts:line
   Impact: [Why this matters]
   Fix: [What to do]

2. [PERF-002] [Description] - file.ts:line
   Impact: [Why this matters]
   Fix: [What to do]

3. [REQ-001] [Description] - file.ts:line
   Impact: [Why this matters]
   Fix: [What to do]

## Quality Assessment

- Security: [EXCELLENT|GOOD|FAIR|POOR]
- Performance: [EXCELLENT|GOOD|FAIR|POOR]
- Maintainability: [EXCELLENT|GOOD|FAIR|POOR]
- Correctness: [EXCELLENT|GOOD|FAIR|POOR]
- Operations: [EXCELLENT|GOOD|FAIR|POOR]

## Recommended Actions

1. [Action 1 - specific and actionable]
2. [Action 2 - specific and actionable]
3. [Action 3 - specific and actionable]

## Verdict

[APPROVE | APPROVE_WITH_NOTES | NEEDS_CHANGES | NEEDS_REWORK]

**Rationale:** [1-2 sentences]
```

## Rules

1. **Be specific** - Use file:line references
2. **Provide code examples** - Show how to fix, not just what's wrong
3. **Validate claims** - Don't flag potential issues, confirm they exist
4. **Prioritize correctly**:
   - CRITICAL: Security vulnerabilities, data loss, crashes
   - HIGH: Significant bugs, performance issues
   - MEDIUM: Code quality issues, minor bugs
   - LOW: Style, documentation, minor improvements
5. **Stop when converged** - Don't force all 5 passes

## Variations

### For small changes (< 100 LOC)

Combine passes:
- PASS 1: Security + Performance
- PASS 2: Maintainability + Correctness
- PASS 3: Operations + Final Review

Check convergence after pass 2.

### For refactoring (not new features)

Add a PASS 0 before everything:

**PASS 0: Behavioral Preservation**
- Does refactored code have identical behavior?
- Are all tests still passing?
- Are edge cases still handled the same way?

Then proceed with standard passes focusing on "did refactoring introduce new issues?"

### For security-sensitive code

Emphasize PASS 1 with additional checks:
- Authentication/authorization flows
- Cryptographic implementations
- Data encryption at rest and in transit
- Audit logging
- Input validation boundaries

## References

- **Steve Yegge's Article:** https://steve-yegge.medium.com/six-new-tips-for-better-coding-with-agents-d4e9c86e42a9
- **Original Discovery:** Jeffrey Emanuel
- **Gastown Implementation:** https://github.com/steveyegge/gastown
