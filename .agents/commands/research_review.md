---
description: Perform iterative review of research documents using the Rule of 5 methodology
---

# Iterative Research Review (Rule of 5)

Perform thorough research document review using the Rule of 5 - iterative refinement until convergence.

## When to Use

- Reviewing research documents before finalizing
- Validating findings and conclusions
- Checking for gaps in analysis
- Ensuring research is actionable

## When NOT to Use

- Research is clearly incomplete (finish it first)
- Quick notes or brainstorming (not ready for review)
- External documents you can't modify

## Process

Perform 5 passes, each focusing on different aspects. After each pass (starting with pass 2), check for convergence.

### PASS 1 - Accuracy & Sources

**Focus on:**
- Claims backed by evidence
- Source credibility and recency
- Correct interpretation of sources
- Factual accuracy of technical details
- Version/date relevance (outdated information)
- Broken links or missing references
- Misquotations or paraphrasing errors

**Output format:**
```
PASS 1: Accuracy & Sources

Issues Found:

[ACC-001] [CRITICAL|HIGH|MEDIUM|LOW] - Section/Paragraph
Description: [What's inaccurate]
Evidence: [Why this is wrong, correct information]
Recommendation: [How to fix]

[ACC-002] ...

Accuracy Assessment: [EXCELLENT|GOOD|FAIR|POOR]
```

### PASS 2 - Completeness & Scope

**Focus on:**
- Missing important topics or considerations
- Unanswered questions that should be addressed
- Gaps in the analysis
- Scope creep (irrelevant tangents)
- Depth appropriate for the topic
- Key stakeholders or perspectives missing
- Relevant alternatives not considered

**Prefix:** COMP-001, etc.

**Convergence Check after Pass 2:**
```
New CRITICAL issues: [count]
Total new issues: [count]
Estimated false positive rate: [percentage]
Status: [CONVERGED | ITERATE | NEEDS_HUMAN]
```

### PASS 3 - Clarity & Structure

**Focus on:**
- Logical flow and organization
- Clear definitions of terms
- Appropriate headings and sections
- Readability for target audience
- Jargon explained or avoided
- Transitions between sections
- Visual aids where helpful

**Prefix:** CLAR-001, etc.

**Convergence Check after Pass 3:**
```
New CRITICAL issues: [count]
Total new issues: [count]
New issues vs Pass 2: [percentage change]
Status: [CONVERGED | ITERATE | NEEDS_HUMAN]
```

### PASS 4 - Actionability & Conclusions

**Focus on:**
- Clear takeaways and recommendations
- Conclusions supported by the research
- Practical applicability
- Trade-offs clearly articulated
- Next steps identified
- Decision criteria provided
- Risks and limitations acknowledged

**Prefix:** ACT-001, etc.

**Convergence Check after Pass 4:**
```
New CRITICAL issues: [count]
Total new issues: [count]
New issues vs Pass 3: [percentage change]
Estimated false positive rate: [percentage]
Status: [CONVERGED | ITERATE | NEEDS_HUMAN]
```

### PASS 5 - Integration & Context

**Focus on:**
- Alignment with existing research
- Connections to other documents
- Relevance to current project goals
- Contradictions with established decisions
- Impact on existing plans
- Cross-references properly linked
- Consistency with team knowledge

**Prefix:** INT-001, etc.

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
- Uncertain about accuracy or conclusions OR
- False positive rate > 30%

**If converged before Pass 5:** Stop and report.

## Final Report

```
# Research Review - Final Report

**Document:** [path/title]
**Convergence:** Pass [N]

## Summary

Total Issues by Severity:
- CRITICAL: [count] - Must fix before using
- HIGH: [count] - Should fix before sharing
- MEDIUM: [count] - Consider addressing
- LOW: [count] - Nice to have

## Top 3 Critical Findings

1. [ACC-001] [Description] - Section
   Impact: [Why this matters]
   Fix: [What to do]

2. [COMP-002] [Description] - Section
   Impact: [Why this matters]
   Fix: [What to do]

3. [ACT-001] [Description] - Section
   Impact: [Why this matters]
   Fix: [What to do]

## Quality Assessment

- Accuracy: [EXCELLENT|GOOD|FAIR|POOR]
- Completeness: [EXCELLENT|GOOD|FAIR|POOR]
- Clarity: [EXCELLENT|GOOD|FAIR|POOR]
- Actionability: [EXCELLENT|GOOD|FAIR|POOR]
- Integration: [EXCELLENT|GOOD|FAIR|POOR]

## Recommended Actions

1. [Action 1 - specific and actionable]
2. [Action 2 - specific and actionable]
3. [Action 3 - specific and actionable]

## Verdict

[READY | NEEDS_REVISION | NEEDS_MORE_RESEARCH]

**Rationale:** [1-2 sentences]
```

## Rules

1. **Be specific** - Reference section/paragraph exactly
2. **Provide corrections** - Don't just identify errors, give correct information
3. **Verify claims** - Check cited sources when possible
4. **Prioritize correctly**:
   - CRITICAL: Factual errors that could mislead decisions
   - HIGH: Significant gaps or unclear conclusions
   - MEDIUM: Clarity issues, minor gaps
   - LOW: Style, formatting, minor improvements
5. **Stop when converged** - Don't force all 5 passes

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
- Less emphasis on polish and structure

### For decision documents

Add emphasis on:
- Decision criteria clarity
- All options fairly presented
- Risks and trade-offs explicit
- Recommendation clearly justified
- Dissenting views acknowledged

## References

- **Steve Yegge's Rule of 5:** https://steve-yegge.medium.com/six-new-tips-for-better-coding-with-agents-d4e9c86e42a9
- **Research Quality Principles:** Accuracy, Completeness, Clarity, Actionability, Integration
