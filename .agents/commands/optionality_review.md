---
description: Review strategic flexibility and reversibility of decisions using optionality thinking
---

# Optionality Review

Review this [CODE/ARCHITECTURE/PLAN/DECISION] through the lens of strategic optionality.

(See Quick Reference section below for optionality principles and red/green flags)

## Work to Review

[PASTE YOUR WORK OR SPECIFY FILE PATH]

## Context (optional)

[Any relevant context: timeline, constraints, uncertainty level]

---

## Phase 1: Decision Classification

First, classify the key decisions using Bezos's framework:

**Type 1 (One-Way Doors):** Consequential and irreversible
- Require careful, methodical consideration
- Need broad consultation
- Examples: Core architecture, public APIs, major vendor commitments

**Type 2 (Two-Way Doors):** Changeable and reversible
- Should be made quickly with ~70% information
- Can be changed if wrong
- Examples: Internal tooling, feature flags, config choices

For each major decision in this work:

| Decision | Type | Reversibility | Decision Deadline | Justification |
|----------|------|---------------|-------------------|---------------|
| [What's being decided] | [1 or 2] | [EASY/MODERATE/HARD/IRREVERSIBLE] | [When must we decide?] | [Why this classification] |

**Early Exit:** If NO Type 1 decisions identified, output abbreviated report:
- List Type 2 decisions with brief notes
- Recommend: "Proceed quickly - all decisions are reversible"
- Skip Phases 2-6

---

## Phase 2: Alternative Paths

For Type 1 decisions, evaluate the option space (2-4 alternatives):

### Current Approach
- Description: [Brief summary]
- Locks in: [What future choices this constrains]
- Enables: [What this makes possible]

### Alternative A
- Description: [Different approach]
- Would preserve: [What options stay open]
- Trade-off: [What we'd give up]

### Alternative B (if applicable)
- Description: [Another approach]
- Would preserve: [What options stay open]
- Trade-off: [What we'd give up]

### Alternative C, D... (add more if needed for complex decisions)

**Assessment:** Are alternatives adequately explored? [YES/NO]
If NO, what should be investigated before committing?

**For Type 2 decisions:** Don't deep-dive. Briefly note 1-2 alternatives, then move on - speed matters more than exhaustive analysis for reversible decisions.

---

## Phase 3: Exit Costs & Escape Hatches

For Type 1 decisions from Phase 1, map the exit strategy:

| Decision | Reversal Cost | Reversal Time | Escape Hatch |
|----------|---------------|---------------|--------------|
| [Choice made] | [LOW/MED/HIGH] | [hours/days/weeks/months] | [Concrete path to undo if needed] |
| [Another choice] | [LOW/MED/HIGH] | [hours/days/weeks/months] | [How to exit] |

**Red Flags:**
- [ ] Irreversible decisions without clear justification
- [ ] No escape hatches for major commitments
- [ ] Assumptions treated as facts
- [ ] "We can refactor later" without a concrete path

**Reversibility Score:** [1-10] (10 = easily reversible)

---

## Phase 4: Failure Modes & Fallbacks

What if things go wrong?

| Failure Scenario | Probability | Impact | Fallback Plan | Recovery Cost |
|------------------|-------------|--------|---------------|---------------|
| [What could fail] | [L/M/H] | [L/M/H/CRITICAL] | [Plan B] | [time/effort] |

**External Dependencies:**
- [ ] Vendor lock-in: [None / Acceptable / Concerning]
- [ ] Technology bets: [Reversible / Locked]
- [ ] Regulatory assumptions: [Validated / Assumed]
- [ ] Single points of failure identified: [YES/NO]

**Resilience Score:** [1-10] (10 = multiple fallbacks, graceful degradation)

**Convergence Check:** If Reversibility ≥7 AND Resilience ≥7, consider skipping to Final Report.

---

## Phase 5: Future Value Assessment

Does this create or destroy options?

### Options CREATED
1. [New capability enabled]
   - Enables: [What becomes possible]
   - Value: [Why this matters]

2. [Another capability]
   - Enables: [What becomes possible]
   - Value: [Why this matters]

### Options DESTROYED (with justification)
1. [Closed off possibility]
   - Prevents: [What becomes harder/impossible]
   - Justified because: [Why trade-off is acceptable]

### Growth Potential
- Can scale to 10x: [YES/PARTIALLY/NO]
- Can add features without rewrite: [YES/PARTIALLY/NO]
- Can be extracted/reused: [YES/PARTIALLY/NO]
- Learning/telemetry built in: [YES/NO]

**Future Value Score:** [1-10] (10 = maximum option creation)

---

## Phase 6: Decision Points & Triggers

Where should we reassess?

| Milestone | What to Assess | Reassess Trigger | Go/No-Go Criteria |
|-----------|----------------|------------------|-------------------|
| [When] | [What we're evaluating] | [What would cause review] | [How to decide] |
| [Next milestone] | [Another assessment] | [Trigger condition] | [Decision criteria] |
| [Future checkpoint] | [Long-term assessment] | [Warning sign] | [How to decide] |

**Assumption Validation:**
- [ ] Key assumptions explicitly stated
- [ ] Validation checkpoints defined
- [ ] Early warning metrics identified

---

## Final Report

### Scores Summary
- Reversibility: [X/10]
- Resilience: [X/10]
- Future Value: [X/10]
- **Overall Optionality: [X/10]** (average, or lowest score if any ≤3)

### Verdict

**Optionality Assessment:** [FLEXIBLE | BALANCED | LOCKED_IN | CONCERNING]
- FLEXIBLE (8-10): Excellent strategic flexibility
- BALANCED (5-7): Acceptable trade-offs, some lock-in justified
- LOCKED_IN (3-4): Significant constraints, ensure they're intentional
- CONCERNING (1-2): High risk, reconsider approach

**Key Findings:**
1. [Most significant optionality issue or strength]
2. [Second finding]
3. [Third finding]

### Recommendations

**Before proceeding:**
- [Critical actions to preserve flexibility]

**Build in over time:**
- [Strategic improvements for future flexibility]

### The Bottom Line

[2-3 sentences: Is the level of lock-in appropriate for our uncertainty level? Are we preserving the right options? What's the key trade-off?]

**Human review recommended:** [YES/NO]
**Reason:** [Why human judgment needed, or why assessment is sufficient]

---

## Quick Reference: Optionality Principles

### The Meta-Principle

The value of flexibility is highest when:
1. **Uncertainty is high** - Can't predict what we'll need
2. **Stakes are high** - Mistakes are costly
3. **Learning is possible** - We can gain information
4. **Time is available** - We can defer without critical cost

When these hold, *invest* in optionality rather than *spending* it.

### Key Frameworks

**Bezos's 70% Rule:** "Most decisions should probably be made with somewhere around 70 percent of the information you wish you had."

**Taleb's Barbell:** 85-95% conservative/reversible + 5-15% high-risk/high-upside bets. Avoid the middle.

**Real Options Types:**
- Option to defer (wait for information)
- Option to expand (scale up if successful)
- Option to contract (scale down if needed)
- Option to abandon (exit and salvage)
- Option to switch (change direction)
- Compound options (options that create options)

### Red Flags

- "We'll never need to change this"
- "This is the only way"
- "We can refactor later" (with no path shown)
- Single points of failure without fallbacks
- Irreversible choices without stated justification
- Hard-coded assumptions

### Green Flags

- Multiple approaches evaluated
- Reversibility designed in
- Decision points clearly marked
- Fallback options exist
- Assumptions explicit
- Built to learn, not just to work

---

## When Optionality Doesn't Matter

Don't over-engineer flexibility when:

1. **Constraints are genuine** - Regulatory requirements, physical laws, third-party APIs
2. **Flexibility has negative value** - Security (limiting is the point), standards (consistency matters)
3. **Decision is reversible AND cheap** - Config changes, styling, naming

**The test:** "What's the cost if we're wrong?"
- LOW cost → Don't over-engineer flexibility
- HIGH cost → Optionality thinking essential
