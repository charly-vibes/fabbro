---
description: Generate comprehensive PR descriptions from git diff
---

# Generate PR Description

Generate a pull request description based on the changes in the current branch.

## Process

### Step 1: Identify the PR

1. Check if current branch has an associated PR:
   ```bash
   gh pr view --json url,number,title,state 2>/dev/null
   ```

2. If no PR exists, list open PRs:
   ```bash
   gh pr list --limit 10 --json number,title,headRefName,author
   ```

3. Ask user which PR to describe if not clear

### Step 2: Gather Information

1. Get the full PR diff:
   ```bash
   gh pr diff {number}
   ```

2. Get commit history:
   ```bash
   gh pr view {number} --json commits
   ```

3. Get PR metadata:
   ```bash
   gh pr view {number} --json url,title,number,state,baseRefName
   ```

### Step 3: Analyze Changes

Think deeply about:
- Purpose and impact of each change
- User-facing changes vs internal implementation
- Breaking changes or migration requirements
- How changes relate to specs in `specs/`
- Related plans in `plans/` if any

### Step 4: Generate Description

Use this template:

```markdown
## Summary

[Brief description of what this PR does]

## Changes

- [Specific change 1]
- [Specific change 2]

## Related

- Plan: `plans/YYYY-MM-DD-description.md` (if applicable)
- Spec: `specs/XX_feature.feature` (if applicable)

## Testing

- [ ] All tests pass
- [ ] [Manual verification steps]

## Notes

[Any additional context, breaking changes, or migration notes]
```

### Step 5: Update the PR

1. Show the user the generated description
2. If approved, update the PR:
   ```bash
   gh pr edit {number} --body "[description]"
   ```

## Guidelines

- Focus on the "why" as much as the "what"
- Be thorough but concise - descriptions should be scannable
- Include breaking changes prominently
- Reference related specs and plans
- Run verification commands when possible to check off items
