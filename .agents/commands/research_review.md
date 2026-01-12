---
description: Perform iterative review of research documents using the Rule of 5 methodology
---

# Iterative Research Review

Perform thorough research document review using the Rule of 5 - iterative refinement until convergence.

## When to Use

- Reviewing research documents before finalizing
- Validating findings and conclusions
- Checking for gaps in analysis
- Ensuring research is actionable

## Process

Perform 5 passes, each focusing on different aspects. After each pass, check for convergence.

### PASS 1 - Accuracy & Sources

Focus on:
- Claims backed by evidence
- Source credibility and recency
- Correct interpretation of sources
- Factual accuracy of technical details
- Version/date relevance (outdated information)

Output format:
- Issue ID (ACC-001, etc.)
- Severity: CRITICAL | HIGH | MEDIUM | LOW
- Location (section/paragraph)
- Description
- Recommendation

### PASS 2 - Completeness & Scope

Focus on:
- Missing important topics or considerations
- Unanswered questions that should be addressed
- Gaps in the analysis
- Scope creep (irrelevant tangents)
- Depth appropriate for the topic

Prefix: COMP-001, etc.

### PASS 3 - Clarity & Structure

Focus on:
- Logical flow and organization
- Clear definitions of terms
- Appropriate headings and sections
- Readability for target audience
- Jargon explained or avoided

Prefix: CLAR-001, etc.

### PASS 4 - Actionability & Conclusions

Focus on:
- Clear takeaways and recommendations
- Conclusions supported by the research
- Practical applicability to fabbro
- Trade-offs clearly articulated
- Next steps identified

Prefix: ACT-001, etc.

### PASS 5 - Integration & Context

Focus on:
- Alignment with existing research in `research/`
- Connections to specs in `specs/`
- Relevance to current project goals
- Contradictions with established decisions
- Impact on existing plans

Prefix: INT-001, etc.

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
- Verdict: READY | NEEDS_REVISION | NEEDS_MORE_RESEARCH

## Rules

1. Be specific with section/paragraph references
2. Provide actionable suggestions for improvements
3. Don't nitpick style - focus on substance
4. Prioritize: Accuracy > Completeness > Actionability > Clarity
5. If converged before pass 5, stop and report

## Variations

### For quick research notes

Combine passes:
- PASS 1: Accuracy + Completeness
- PASS 2: Clarity + Actionability
- PASS 3: Integration + Final Review

### For exploratory research (early stage)

Relax standards:
- Accept more uncertainty in conclusions
- Focus on identifying promising directions
- Flag areas needing deeper investigation
