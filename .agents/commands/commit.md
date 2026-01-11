---
description: Create git commits with user approval and no Claude attribution
---

# Commit Changes

Create git commits for changes made during this session.

## Process

1. **Review what changed:**
   - Review the conversation history
   - Run `git status` to see current changes
   - Run `git diff` to understand modifications
   - Consider whether changes should be one or multiple commits

2. **Plan your commit(s):**
   - Identify which files belong together
   - Draft clear, descriptive commit messages
   - Use imperative mood ("Add feature" not "Added feature")
   - Focus on why, not just what

3. **Present your plan:**
   ```
   I plan to create [N] commit(s):

   Commit 1: [message]
   - file1.ext
   - file2.ext

   Commit 2: [message]
   - file3.ext

   Shall I proceed?
   ```

4. **Execute upon confirmation:**
   - Use `git add` with specific files (avoid `-A` or `.`)
   - Create commits with planned messages
   - Show result: `git log --oneline -n [number]`

5. **Sync with beads (if applicable):**
   - If working on a tracked issue: `bd sync`
   - Push if instructed: `git push`

## Important

- **NEVER** add co-author information or Claude attribution
- Commits should be authored solely by the user
- Do not include "Generated with Claude" messages
- Do not add "Co-Authored-By" lines
- Write commit messages as if the user wrote them

## Remember

- You have full context of what was done
- Group related changes together
- Keep commits focused and atomic
- The user trusts your judgment
