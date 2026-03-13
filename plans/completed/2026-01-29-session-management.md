# Session Management Implementation Plan

## Overview

Implement the `fabbro sessions`, `show`, `resume`, `delete`, `clean`, and `export` commands for managing review sessions.

**Spec**: `specs/05_session_management.feature`

## Current State

- Sessions are stored in `.fabbro/sessions/` as `.fem` files
- No commands exist to list, show, resume, delete, or export sessions
- `session.Load()` and `session.List()` exist in `internal/session/`

## Desired End State

- `fabbro sessions` lists all sessions with metadata
- `fabbro show <id>` displays session details and annotation breakdown
- `fabbro resume <id>` reopens session in TUI
- `fabbro delete <id>` removes session (with confirmation)
- `fabbro clean --older-than <duration>` removes old sessions
- `fabbro export <id>` outputs session content

---

## Phase 1: List Sessions Command (~1 hour)

**Spec scenarios**: "Listing all sessions", "Listing sessions in JSON format", "No sessions exist"

### Changes Required

**File: cmd/fabbro/main.go**
- Add `sessions` subcommand with `--json` flag

**File: internal/session/session.go**
- Add `ListWithMetadata()` returning slice of session summaries:
  ```go
  type SessionSummary struct {
      ID             string
      CreatedAt      time.Time
      Source         string
      AnnotationCount int
  }
  ```

### Output Format

Human-readable (default):
```
SESSION ID    CREATED              SOURCE        ANNOTATIONS
abc123        2026-01-29 10:00     stdin         5
def456        2026-01-28 14:30     document.md   12
```

JSON (`--json`):
```json
[
  {"id": "abc123", "createdAt": "2026-01-29T10:00:00Z", "source": "stdin", "annotations": 5}
]
```

### Success Criteria

- [ ] `fabbro sessions` lists all sessions sorted newest-first
- [ ] `fabbro sessions --json` outputs valid JSON array
- [ ] Empty sessions directory shows helpful message

---

## Phase 2: Show Session Details (~45 min)

**Spec scenarios**: "Showing session details", "Showing session with annotation breakdown"

### Changes Required

**File: cmd/fabbro/main.go**
- Add `show <session-id>` subcommand

**File: internal/session/session.go**
- Add `Session.AnnotationBreakdown() map[string]int`

### Output Format

```
Session ID:     abc123
Created:        2026-01-29 10:00:00
Source:         document.md
Content lines:  100

Annotations (6 total):
  comment:  3
  question: 2
  delete:   1
```

### Success Criteria

- [ ] `fabbro show <id>` displays session metadata
- [ ] Annotation breakdown shows count by type
- [ ] Non-existent session returns error with exit code 1

---

## Phase 3: Resume Session (~30 min)

**Spec scenarios**: "Resuming an interrupted review", "Resuming in editor mode"

### Changes Required

**File: cmd/fabbro/main.go**
- Add `resume <session-id>` subcommand with `--editor` flag

### Behavior

- Load existing session file
- Open TUI (or `$EDITOR` with `--editor` flag)
- Existing annotations should be visible and editable

### Success Criteria

- [ ] `fabbro resume <id>` opens session in TUI
- [ ] `fabbro resume <id> --editor` opens in `$EDITOR`
- [ ] Non-existent session returns error

---

## Phase 4: Delete Session (~30 min)

**Spec scenarios**: "Deleting a session", "Deleting a session with --force"

### Changes Required

**File: cmd/fabbro/main.go**
- Add `delete <session-id>` subcommand with `--force` flag

### Behavior

- Without `--force`: prompt for confirmation
- With `--force`: delete immediately
- Print success message after deletion

### Success Criteria

- [ ] `fabbro delete <id>` prompts for confirmation
- [ ] `fabbro delete <id> --force` deletes without prompt
- [ ] Non-existent session returns error

---

## Phase 5: Clean Old Sessions (~45 min)

**Spec scenarios**: "Cleaning sessions older than threshold", "Dry-run cleaning"

### Changes Required

**File: cmd/fabbro/main.go**
- Add `clean` subcommand with `--older-than`, `--dry-run`, and `--force` flags

### Behavior

- Parse duration: `7d`, `14d`, `30d`, etc.
- **Safety limit**: Minimum threshold is `1d` (reject `0d`, `0h`, etc.)
- List sessions that match criteria
- Prompt for confirmation (unless `--dry-run`)
- Delete matching sessions

### Validation

```go
duration, err := parseDuration(olderThan)
if err != nil {
    return fmt.Errorf("invalid duration: %s", olderThan)
}
if duration < 24*time.Hour {
    return fmt.Errorf("minimum --older-than is 1d (safety limit). Use --force to override")
}
```

### Success Criteria

- [ ] `fabbro clean --older-than 7d` lists and deletes old sessions
- [ ] `fabbro clean --older-than 7d --dry-run` lists but doesn't delete
- [ ] Confirmation prompt before deletion
- [ ] `fabbro clean --older-than 0d` rejected with safety error
- [ ] `fabbro clean --older-than 0d --force` works (bypass safety)

---

## Phase 6: Export Session (~30 min)

**Spec scenarios**: "Exporting session as standalone file", "Exporting session to stdout"

### Changes Required

**File: cmd/fabbro/main.go**
- Add `export <session-id>` subcommand with `--output` flag

### Behavior

- Without `--output`: print session content to stdout
- With `--output <path>`: write to file

### Success Criteria

- [ ] `fabbro export <id>` outputs to stdout
- [ ] `fabbro export <id> --output review.fem` writes to file
- [ ] Non-existent session returns error

---

## Phase 7: Partial Session ID Matching (~30 min)

**Spec scenarios**: "Partial session ID matching", "Ambiguous partial session ID"

### Changes Required

**File: internal/session/session.go**
- Modify `Load()` to support partial matching
- Return error for ambiguous matches listing candidates

### Matching Rules

1. **Exact match first**: If input exactly matches a session ID, use it
2. **Prefix match**: If input is prefix of exactly one session ID, use it
3. **Ambiguous**: If input matches multiple session IDs, error with candidates
4. **No match**: If input matches nothing, error with suggestions

```go
func (s *Store) LoadPartial(partial string) (*Session, error) {
    sessions, _ := s.List()
    var matches []string
    
    for _, sess := range sessions {
        if sess.ID == partial {
            return s.Load(sess.ID)  // Exact match
        }
        if strings.HasPrefix(sess.ID, partial) {
            matches = append(matches, sess.ID)
        }
    }
    
    switch len(matches) {
    case 0:
        return nil, fmt.Errorf("no session matching '%s'", partial)
    case 1:
        return s.Load(matches[0])
    default:
        return nil, fmt.Errorf("ambiguous session ID '%s' matches: %s", 
            partial, strings.Join(matches, ", "))
    }
}
```

### Success Criteria

- [ ] `fabbro show abc1` matches `review-abc123`
- [ ] Ambiguous matches return error listing candidates
- [ ] Exact match takes priority over prefix match

---

## Summary

| Phase | Deliverable | Time |
|-------|-------------|------|
| 1 | `sessions` list command | 1h |
| 2 | `show` details command | 45m |
| 3 | `resume` command | 30m |
| 4 | `delete` command | 30m |
| 5 | `clean` command (with safety validation) | 1h |
| 6 | `export` command | 30m |
| 7 | Partial ID matching | 45m |

**Total: ~5 hours**

## Dependencies

All phases can run in sequence. Phase 7 (partial matching) is independent and can run anytime after Phase 1.
