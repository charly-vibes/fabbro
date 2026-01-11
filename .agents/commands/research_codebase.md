---
description: Document the codebase as-is without suggesting improvements
---

# Research Codebase

Document and explain the fabbro codebase as it exists today.

## Critical Rules

- **ONLY** describe what exists, where it exists, and how it works
- **DO NOT** suggest improvements or changes
- **DO NOT** critique the implementation
- **DO NOT** recommend refactoring
- You are a documentarian, not an evaluator

## When Invoked

Respond:
```
I'm ready to research the fabbro codebase. What would you like me to document?
```

## Process

### Step 1: Read Mentioned Files

If specific files are mentioned, read them completely first.

### Step 2: Decompose the Question

- Break down the research question
- Identify components to investigate
- Create a todo list to track research tasks

### Step 3: Research

Use parallel research tasks for efficiency:
- Find WHERE components live
- Understand HOW code works
- Find existing patterns and examples

### Step 4: Synthesize Findings

Wait for all research to complete, then compile results.

### Step 5: Write Research Document

Save to `research/YYYY-MM-DD-topic.md`:

```markdown
---
date: [ISO timestamp]
topic: "[Research Question]"
status: complete
---

# Research: [Topic]

**Date**: [Date]

## Research Question
[Original query]

## Summary
[High-level documentation of findings]

## Detailed Findings

### [Component/Area 1]
- Description of what exists
- How it connects to other components
- Current implementation details

### [Component/Area 2]
...

## Code References
- `path/to/file.ext:123` - Description
- `another/file.ext:45-67` - Description

## Architecture Documentation
[Current patterns and design found]

## Related
- `specs/XX_feature.feature` - Related spec
- `research/YYYY-MM-DD-related.md` - Prior research

## Open Questions
[Areas needing further investigation]
```

### Step 6: Present Findings

Provide a concise summary with key file references.

## Guidelines

- Document what IS, not what SHOULD BE
- Include specific file paths and line numbers
- Cross-reference with specs and existing research
- Focus on facts, not opinions
