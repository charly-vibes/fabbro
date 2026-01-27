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
- "How does the FEM parser work?"
- "What's the structure of the TUI layer?"
- "How do sessions get created and loaded?"
- "What patterns does the codebase use for X?"
```

## Process

### Step 1: Understand the Research Question

Break down what's being asked:
- What component/feature/pattern to document?
- What level of detail is needed?
- What's the intended audience?

### Step 2: Create Research Plan

Track research with beads if there's an associated issue:

```bash
bd update <issue-id> --status=in_progress
```

Mental checklist:
- [ ] Find entry points
- [ ] Trace main flows
- [ ] Document data structures
- [ ] Note dependencies
- [ ] Compile examples

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
1. Find the entry point (command, package, function)
2. Trace the flow through the code
3. Document each step
4. Note key data transformations

**For "What's the structure of X?"**
1. List all files/packages in the area
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
# [Topic] Research

**Date:** YYYY-MM-DD  
**Issue:** fabbro-xxx (if applicable)  
**Question:** [The research question being answered]

## Summary

[2-3 sentence summary of findings]

## Detailed Findings

### [Component/Area 1]

**Location:** `internal/component/`

**Purpose:** [What this does]

**Key files:**
- `file.go` - [Purpose]
- `file_test.go` - [Purpose]

**How it works:**
1. [Step 1 with code reference]
2. [Step 2 with code reference]

### [Component/Area 2]

[Continue with same structure]

## Code References

| File | Lines | Description |
|------|-------|-------------|
| `internal/pkg/file.go` | 45-67 | Entry point |
| `cmd/fabbro/main.go` | 12-34 | CLI wiring |

## Examples

```go
// From internal/session/session.go:15-25
func Create(path string) (*Session, error) {
    // ...
}
```

## Related

- `research/YYYY-MM-DD-related-topic.md`
- `specs/XX_feature.feature`

## Open Questions

[Any areas that couldn't be fully documented and why]
```

### Step 5: Present Findings

Provide a concise summary:

```
## Research Complete: [Topic]

**Document:** research/YYYY-MM-DD-topic.md

### Key Findings

1. [Most important finding]
2. [Second important finding]
3. [Third important finding]

### Key Code Locations

- Entry point: `cmd/fabbro/cmd_review.go:42`
- Core logic: `internal/session/session.go:100-150`
- Config: `internal/config/config.go`
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
- Include visual aids (mermaid diagrams, tables) when helpful

## Types of Research

### Architecture Research
Focus on: Structure, packages, boundaries, dependencies

### Flow Research
Focus on: Step-by-step tracing, data transformations, control flow

### Pattern Research
Focus on: Recurring structures, conventions, examples

### CLI Research
Focus on: Commands, flags, argument handling, output formats

### Data Model Research
Focus on: Types, structs, relationships, validation

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
