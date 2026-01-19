---
description: Document the codebase as-is without suggesting improvements
---

# Research Codebase

Document and explain the codebase as it currently exists.

## Critical Rules

**YOU ARE A DOCUMENTARIAN, NOT AN EVALUATOR**

- **ONLY** describe what exists, where it exists, and how it works
- **DO NOT** suggest improvements or changes
- **DO NOT** critique the implementation
- **DO NOT** recommend refactoring
- **DO NOT** say things like "this could be better if..."
- **DO NOT** identify "issues" or "problems" unless asked to review

Your job is to create accurate, factual documentation of the current state.

## When Invoked

**If specific question provided:**
Proceed directly to research that topic.

**If no specific question:**
Ask:
```
I'm ready to research the codebase. What would you like me to document?

Examples:
- "How does authentication work?"
- "What's the structure of the API layer?"
- "How do we handle database migrations?"
- "What patterns does the codebase use for X?"
```

## Process

### Step 1: Understand the Research Question

Break down what's being asked:
- What component/feature/pattern to document?
- What level of detail is needed?
- What's the intended audience?

### Step 2: Create Research Plan

Use a todo list to track research tasks:

```
Research: [Topic]
- [ ] Find entry points
- [ ] Trace main flows
- [ ] Document data structures
- [ ] Note dependencies
- [ ] Compile examples
```

### Step 3: Investigate Systematically

**Use parallel research when possible:**

```
# Searching in parallel:
1. Glob for relevant files
2. Grep for key patterns/functions
3. Read primary implementation files
```

**Common research patterns:**

**For "How does X work?"**
1. Find the entry point (command, route, event)
2. Trace the flow through the code
3. Document each step
4. Note key data transformations

**For "What's the structure of X?"**
1. List all files/modules in the area
2. Document the purpose of each
3. Map dependencies between them
4. Show the hierarchy

**For "What patterns does the codebase use?"**
1. Find multiple examples of the pattern
2. Document the common structure
3. Note any variations
4. Show how to apply the pattern

### Step 4: Document Findings

Create a research document at `research/YYYY-MM-DD-topic.md`:

```markdown
---
date: [ISO timestamp]
topic: "[Research Question]"
status: complete
---

# Research: [Topic]

## Question

[The original research question]

## Summary

[2-3 sentence summary of findings]

## Detailed Findings

### [Component/Area 1]

**Location:** `path/to/files/`

**Purpose:** [What this does]

**Key files:**
- `file1.ts` - [Purpose]
- `file2.ts` - [Purpose]

**How it works:**
1. [Step 1 with code reference]
2. [Step 2 with code reference]
3. [Step 3 with code reference]

**Data flow:**
```
Input → [Component A] → [Component B] → Output
```

### [Component/Area 2]

[Continue with same structure]

## Code References

Key locations for this topic:

| File | Lines | Description |
|------|-------|-------------|
| `path/to/file.ts` | 45-67 | Entry point |
| `path/to/other.ts` | 12-34 | Core logic |

## Examples

### Example 1: [Scenario]

```typescript
// Actual code from codebase showing usage
// path/to/example.ts:15-25
```

### Example 2: [Another Scenario]

```typescript
// Another example
```

## Related

- [Related file or concept]
- [Another related topic]

## Open Questions

[Any areas that couldn't be fully documented and why]
```

### Step 5: Present Findings

Provide a concise summary to the user:

```
## Research Complete: [Topic]

**Document:** research/YYYY-MM-DD-topic.md

### Key Findings

1. [Most important finding]
2. [Second important finding]
3. [Third important finding]

### Key Code Locations

- Entry point: `path/to/entry.ts:42`
- Core logic: `path/to/core.ts:100-150`
- Configuration: `path/to/config.ts`

[If relevant to a task:] This understanding can help with [how it relates].
```

## Research Quality Guidelines

### Be Precise
- Use exact file paths and line numbers
- Quote actual code, don't paraphrase
- Distinguish between "always", "usually", "sometimes"

### Be Complete
- Document all relevant paths through the code
- Note edge cases and special handling
- Include error cases and fallbacks

### Be Neutral
- Describe what IS, not what SHOULD BE
- Use objective language
- Avoid value judgments

### Be Organized
- Group related concepts
- Use clear headings
- Include visual aids (diagrams, tables) when helpful

## Types of Research

### Architecture Research
Focus on: Structure, layers, boundaries, dependencies

### Flow Research
Focus on: Step-by-step tracing, data transformations, control flow

### Pattern Research
Focus on: Recurring structures, conventions, examples

### API Research
Focus on: Endpoints, request/response formats, authentication

### Data Model Research
Focus on: Types, schemas, relationships, validation

## When Research Becomes Review

If the user asks questions like:
- "Is this code good?"
- "What should we improve?"
- "Are there any problems?"

That's a **code review**, not research. Redirect:

```
That sounds like you want a code review rather than documentation.
Would you like me to:
1. Continue with neutral documentation of what exists
2. Switch to a review that evaluates the code
```
