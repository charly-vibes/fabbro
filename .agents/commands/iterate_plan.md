---
description: Iterate on existing implementation plans with updates and research
---

# Iterate Implementation Plan

Update existing implementation plans based on user feedback, grounded in codebase reality.

## When Invoked

1. **If NO plan file provided**: Ask for the plan path (`ls plans/` to list)
2. **If plan file provided but NO feedback**: Ask what changes to make
3. **If BOTH provided**: Proceed directly

## Process

### Step 1: Understand Current Plan

1. Read the existing plan file completely
2. Understand the structure, phases, and scope
3. Note the success criteria and implementation approach

### Step 2: Research If Needed

**Only if changes require new technical understanding:**

1. Create a todo list for research tasks
2. Search for relevant patterns in the codebase
3. Check `research/` for related documentation
4. Read relevant spec files in `specs/`

### Step 3: Confirm Understanding

Before making changes, confirm:

```
Based on your feedback, I understand you want to:
- [Change 1 with specific detail]
- [Change 2 with specific detail]

My research found:
- [Relevant code pattern or constraint]

I plan to update the plan by:
1. [Specific modification]
2. [Another modification]

Does this align with your intent?
```

Get user confirmation before proceeding.

### Step 4: Update the Plan

1. Make focused, precise edits to the existing plan
2. Maintain existing structure unless explicitly changing it
3. Update success criteria if needed
4. Ensure consistency:
   - If adding a phase, follow existing pattern
   - If modifying scope, update "Out of Scope" section
   - Maintain automated vs manual success criteria distinction

### Step 5: Present Changes

```
I've updated the plan at `plans/[filename].md`

Changes made:
- [Specific change 1]
- [Specific change 2]

Would you like any further adjustments?
```

## Guidelines

1. **Be Skeptical**: Question vague feedback, verify technical feasibility
2. **Be Surgical**: Precise edits, preserve good content
3. **Be Thorough**: Read entire plan, research only what's necessary
4. **Be Interactive**: Confirm understanding before making changes
5. **No Open Questions**: Ask immediately if changes raise questions

## Example Flows

**Scenario 1: Everything upfront**
```
User: /iterate_plan plans/2025-01-10-feature.md - add phase for error handling
```

**Scenario 2: Plan file only**
```
User: /iterate_plan plans/2025-01-10-feature.md
Agent: I've found the plan. What changes would you like to make?
```

**Scenario 3: No arguments**
```
User: /iterate_plan
Agent: Which plan would you like to update? (ls plans/ to list)
```
