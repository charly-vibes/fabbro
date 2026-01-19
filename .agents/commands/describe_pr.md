---
description: Generate comprehensive PR descriptions from git diff
---

# Describe Pull Request

Generate a comprehensive pull request description based on git changes.

## Process

### Step 1: Identify What to Describe

**Option A: Current branch (most common)**
```bash
# Get current branch
git branch --show-current

# Get base branch (usually main or master)
git rev-parse --abbrev-ref HEAD@{upstream} 2>/dev/null || echo "main"
```

**Option B: Existing PR**
```bash
# If PR number provided
gh pr view [number] --json title,body,baseRefName,headRefName
```

### Step 2: Gather Change Information

```bash
# Get all commits on this branch (not on main)
git log main..HEAD --oneline

# Get the full diff
git diff main...HEAD

# Get list of changed files
git diff main...HEAD --stat

# Get detailed file changes
git diff main...HEAD --name-status
```

### Step 3: Analyze the Changes

For each changed file, understand:
- **What changed**: New code, modifications, deletions
- **Why it changed**: Feature, bug fix, refactor, etc.
- **Impact**: What this affects in the system

**Group changes by purpose:**
- Core feature changes
- Supporting changes (types, utils, configs)
- Test changes
- Documentation changes

### Step 4: Read Related Context

Check for related documents:
- `plans/` - Implementation plans this PR follows
- `specs/` - Specifications being implemented
- `research/` - Research informing the approach

### Step 5: Generate PR Description

Use this template:

```markdown
## Summary

[1-2 sentences describing what this PR does and why]

## Changes

### [Category 1: e.g., "Core Feature"]
- [Specific change with file reference]
- [Another change]

### [Category 2: e.g., "Tests"]
- [Test additions/changes]

### [Category 3: e.g., "Configuration"]
- [Config changes if any]

## Technical Details

[If the implementation is non-obvious, explain the approach]

**Key decisions:**
- [Decision 1]: [Why]
- [Decision 2]: [Why]

## Testing

**Automated:**
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Type checking passes
- [ ] Linting passes

**Manual verification:**
- [ ] [Specific manual test 1]
- [ ] [Specific manual test 2]

## Related

- Plan: `plans/YYYY-MM-DD-name.md`
- Spec: `specs/XX_feature.feature`
- Issue: #123

## Checklist

- [ ] Code follows project conventions
- [ ] Tests added for new functionality
- [ ] Documentation updated (if applicable)
- [ ] No security vulnerabilities introduced
- [ ] No breaking changes (or documented if necessary)

## Screenshots

[If UI changes, include before/after screenshots]
```

### Step 6: Present to User

Show the generated description and ask for approval:

```
Here's the PR description I've generated:

---
[Full description]
---

Would you like me to:
1. Use this description as-is
2. Modify specific sections
3. Generate a different version
```

### Step 7: Apply Description

**For new PR:**
```bash
gh pr create --title "[title]" --body "[description]"
```

**For existing PR:**
```bash
gh pr edit [number] --body "[description]"
```

## Description Quality Guidelines

### Good Summaries
- Start with action verb: "Add", "Fix", "Refactor", "Update"
- Explain the "what" and "why" concisely
- Avoid vague terms like "improve" or "update" without specifics

**Bad:** "Update authentication"
**Good:** "Add JWT token refresh to prevent session timeout during long operations"

### Good Change Lists
- Group by logical purpose, not by file
- Include file references for easy navigation
- Explain non-obvious changes

**Bad:**
- Changed auth.ts
- Changed user.ts
- Changed types.ts

**Good:**
- Add token refresh logic (`src/auth/refresh.ts:15-45`)
- Update User type to include `refreshToken` field (`src/types/user.ts`)
- Add refresh endpoint handler (`src/api/auth.ts:67-89`)

### Good Technical Details
- Explain the "why" behind non-obvious choices
- Note any trade-offs made
- Reference prior art or patterns followed

### Good Testing Sections
- Be specific about what to verify
- Include both automated and manual checks
- Note any areas with limited test coverage

## Variations

### For Bug Fixes

Emphasize:
- What was broken (observed behavior)
- What was expected
- Root cause
- How it's fixed
- How to verify the fix

### For Refactoring

Emphasize:
- What motivated the refactor
- What changed structurally
- What stayed the same (behavior preservation)
- Before/after comparison if helpful

### For New Features

Emphasize:
- What the feature does
- How to use it
- Configuration options
- Edge cases handled
- Future work (if any)

### For Large PRs

If the PR is large:
1. Consider if it should be split
2. Add a "How to Review" section
3. Suggest review order for related changes
4. Note which changes are low-risk vs need careful review
