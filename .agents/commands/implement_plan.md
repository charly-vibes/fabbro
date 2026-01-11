---
description: Implement approved plans from plans/ directory following TDD
---

# Implement Plan

Implement an approved plan from `plans/` following fabbro's TDD methodology.

## Getting Started

When given a plan path:
1. Read the plan completely
2. Check for existing checkmarks (completed work)
3. Read related specs in `specs/`
4. Check `bd ready` for the next unblocked issue
5. Create a todo list to track progress
6. Start implementing from the first unchecked phase

If no plan path provided, list available plans: `ls plans/`

## Beads Integration

Plans should have corresponding beads issues for tracking. Before starting:

```bash
bd ready                    # Show unblocked work
bd show <issue-id>          # Review issue details
bd update <id> --status=in_progress  # Claim the phase
```

After completing a phase:
```bash
bd close <id>               # Mark phase complete
bd ready                    # See what's unblocked next
```

## Implementation Philosophy

Follow the **Red, Green, Refactor** cycle from AGENTS.md:

1. **Red**: Write a failing test for the scenario
2. **Green**: Write minimal code to pass the test
3. **Refactor**: Clean up while keeping tests green

## Workflow

### For Each Phase:

1. **Read the phase requirements**
2. **Write/update the test first** (TDD)
3. **Run tests** - confirm they fail for the right reason
4. **Implement the code** - minimal to pass the test
5. **Run tests** - confirm they pass
6. **Refactor** if needed
7. **Check off completed items** in the plan file

### After Completing a Phase:

Run success criteria checks and inform the user:

```
Phase [N] Complete - Ready for Verification

Automated verification:
- [x] Tests pass
- [x] Type checking passes

Manual verification needed:
- [ ] [Items from plan requiring human check]

Let me know when verified so I can proceed to Phase [N+1].
```

## When Things Don't Match

If the plan doesn't match reality:

```
Issue in Phase [N]:
Expected: [what the plan says]
Found: [actual situation]
Why this matters: [explanation]

How should I proceed?
```

## Resuming Work

If the plan has checkmarks:
- Trust completed work is done
- Start from first unchecked item
- Verify previous work only if something seems off

## Key Reminders

- Tests first, then implementation
- Check off items as you complete them
- Minimal code to pass tests
- Refactor only with green tests
- Update beads: `bd update <id> --status=in_progress` when starting, `bd close <id>` when done
