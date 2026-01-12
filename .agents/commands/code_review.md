---
description: Perform iterative code review using the Rule of 5 methodology
---

# Iterative Code Review

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

Perform 5 passes, each focusing on different aspects. After each pass, check for convergence.

### PASS 1 - Security & Safety

Focus on:
- Input validation and sanitization
- Authentication and authorization
- SQL injection, XSS, CSRF vulnerabilities
- Secret management and data exposure
- Error handling that doesn't leak information

Output format:
- Issue ID (SEC-001, etc.)
- Severity: CRITICAL | HIGH | MEDIUM | LOW
- Location (file:line)
- Description
- Recommendation with code example

### PASS 2 - Performance & Scalability

Focus on:
- Time complexity (O(n²), O(n³) patterns)
- Database queries (N+1, missing indexes)
- Memory allocation and potential leaks
- Unnecessary loops or iterations
- Caching opportunities

Prefix: PERF-001, etc.

### PASS 3 - Maintainability & Readability

Focus on:
- Code clarity and naming
- Documentation (comments, docstrings)
- Pattern consistency
- Technical debt indicators
- DRY violations
- Magic numbers and hard-coded values

Prefix: MAINT-001, etc.

### PASS 4 - Correctness & Requirements

Focus on:
- Does it do what it's supposed to do?
- Edge case handling
- Test coverage gaps
- Behavioral correctness
- Requirements satisfaction

Prefix: REQ-001, etc.

### PASS 5 - Operations & Reliability

Focus on:
- Failure modes and error handling
- Timeout and retry logic
- Observability (logging, metrics)
- Resource management (connections, files)
- Deployment considerations

Prefix: OPS-001, etc.

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
- Recommended next actions
- Convergence assessment

## Rules

1. Be specific with file:line references
2. Provide actionable code examples for fixes
3. Don't just list potential issues - confirm they exist
4. Prioritize: CRITICAL security > CRITICAL other > HIGH > MEDIUM > LOW
5. If converged before pass 5, stop and report

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

## References

- Steve Yegge's "Six New Tips for Better Coding with Agents"
- Original discovery by Jeffrey Emanuel
