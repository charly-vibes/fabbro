---
description: Create detailed implementation plans for fabbro features
---

# Create Implementation Plan

Create detailed implementation plans following fabbro's Spec-Driven Development methodology.

## When Invoked

1. **If a spec file is provided**: Read it fully and begin planning
2. **If no parameters**: Ask for the feature/task description

## Process

### Step 1: Understand the Requirement

1. Read any mentioned spec files in `specs/` completely
2. Check existing research in `research/` for related work
3. Review debates in `debates/` for past discussions on the topic
4. Understand the scope and constraints

### Step 2: Research the Codebase

1. Find relevant existing patterns and code
2. Identify integration points
3. Note conventions to follow
4. Track research in a todo list

### Step 3: Design Options (if applicable)

Present design options with pros/cons. Get alignment before detailed planning.

### Step 4: Write the Plan

Save to `plans/YYYY-MM-DD-description.md`:

```markdown
# [Feature Name] Implementation Plan

## Overview
[Brief description of what we're implementing]

## Related
- Spec: `specs/XX_feature.feature`
- Research: `research/YYYY-MM-DD-topic.md` (if applicable)

## Current State
[What exists now, what's missing]

## Desired End State
[What will exist after implementation, how to verify]

## Out of Scope
[What we're NOT doing to prevent scope creep]

## Phase 1: [Name]

### Changes Required
- File: `path/to/file.ext`
- Changes: [Description]

### Success Criteria

#### Automated:
- [ ] Tests pass: `just test`
- [ ] Type checking passes (if applicable)

#### Manual:
- [ ] [Verification step]

---

## Phase 2: [Name]
[Continue phases as needed]

---

## Testing Strategy
[Following TDD - tests before implementation]

## References
- Related spec: `specs/XX_feature.feature`
```

### Step 5: Review and Iterate

Present the plan for feedback. Iterate until approved.

## Guidelines

1. **Align with Spec-Driven Development**: Plans should reference or result in `.feature` specs
2. **Follow TDD**: Plan tests before implementation
3. **Be specific**: Include file paths and concrete changes
4. **Track progress**: Use todo list throughout planning
5. **No open questions**: Resolve all questions before finalizing
