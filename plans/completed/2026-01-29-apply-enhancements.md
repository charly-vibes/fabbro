# Apply Command Enhancements Plan

## Overview

Extend `fabbro apply` with compact JSON output, content hash verification, and custom session ID support for `fabbro review`.

**Spec**: `specs/02_review_session.feature`, `specs/04_apply_feedback.feature`

## Current State

- `fabbro apply <id> --json` outputs pretty-printed JSON
- No compact output option
- Content hash stored but not verified
- `fabbro review` uses UUID for session ID
- `fabbro review --editor` and `--no-interactive` not implemented

## Desired End State

- `--compact` outputs minified JSON for piping
- Content hash mismatch produces warning
- `fabbro review --id <name>` uses custom session ID
- `fabbro review --editor` opens `$EDITOR`
- `fabbro review --no-interactive` creates session without opening anything

---

## Phase 1: Compact JSON Output (~20 min)

**Spec scenario**: "Compact JSON output for piping"

### Changes Required

**File: cmd/fabbro/main.go**

Add `--compact` flag to apply command:
```go
compactFlag := applyCmd.Bool("compact", false, "Output minified JSON")
```

Use `json.Marshal` instead of `json.MarshalIndent` when compact:
```go
var output []byte
if *compactFlag {
    output, _ = json.Marshal(result)
} else {
    output, _ = json.MarshalIndent(result, "", "  ")
}
```

### Success Criteria

- [ ] `fabbro apply <id> --json --compact` outputs single line
- [ ] Suitable for piping to jq or other tools

---

## Phase 2: Content Hash Verification (~45 min)

**Spec scenario**: "Warning when source content has changed"

### Changes Required

**File: internal/session/session.go**

Add hash verification:
```go
func (s *Session) VerifySourceHash() (bool, error) {
    if s.Source == "" || s.Source == "stdin" {
        return true, nil  // Can't verify stdin
    }
    
    content, err := os.ReadFile(s.Source)
    if err != nil {
        return false, fmt.Errorf("source file not found: %s", s.Source)
    }
    
    currentHash := computeHash(content)
    return currentHash == s.ContentHash, nil
}
```

**File: cmd/fabbro/main.go**

In apply command:
```go
if valid, err := session.VerifySourceHash(); err != nil {
    fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
} else if !valid {
    fmt.Fprintf(os.Stderr, "Warning: source file has changed since session was created. Line numbers may have drifted.\n")
}
```

### Success Criteria

- [ ] Warning printed when source file hash differs
- [ ] Warning printed when source file not found
- [ ] No warning for stdin sessions
- [ ] Annotations still output after warning

---

## Phase 3: Custom Session ID (~30 min)

**Spec scenario**: "Creating a review session with a custom session ID"

### Changes Required

**File: cmd/fabbro/main.go**

Add `--id` flag to review command:
```go
idFlag := reviewCmd.String("id", "", "Custom session ID")
```

In session creation:
```go
sessionID := *idFlag
if sessionID == "" {
    sessionID = generateUUID()
}
```

### Validation

Session ID must be a valid filename:
- **Allowed characters**: `a-z`, `A-Z`, `0-9`, `-`, `_`
- **Max length**: 64 characters
- **Reserved IDs**: `tutor`, `_tutor_` (used by tutor command)
- **No path separators**: `/`, `\` rejected

```go
var validSessionID = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)
var reservedIDs = map[string]bool{"tutor": true, "_tutor_": true}

func validateSessionID(id string) error {
    if !validSessionID.MatchString(id) {
        return fmt.Errorf("invalid session ID: must be 1-64 alphanumeric characters, dash, or underscore")
    }
    if reservedIDs[id] {
        return fmt.Errorf("session ID '%s' is reserved", id)
    }
    return nil
}
```

- Check for existing session with same ID

### Success Criteria

- [ ] `fabbro review --stdin --id my-review` creates `my-review.fem`
- [ ] Invalid IDs rejected with error (spaces, unicode, slashes)
- [ ] IDs over 64 chars rejected
- [ ] Reserved IDs (`tutor`) rejected with clear error
- [ ] Duplicate IDs rejected with error

---

## Phase 4: Editor Mode (~30 min)

**Spec scenario**: "Opening session in external editor instead of TUI"

### Changes Required

**File: cmd/fabbro/main.go**

Add `--editor` flag:
```go
editorFlag := reviewCmd.Bool("editor", false, "Open in $EDITOR instead of TUI")
```

Implementation:
```go
if *editorFlag {
    editor := os.Getenv("EDITOR")
    if editor == "" {
        editor = "vi"
    }
    cmd := exec.Command(editor, sessionPath)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}
```

### Success Criteria

- [ ] `fabbro review --stdin --editor` opens `$EDITOR`
- [ ] Falls back to `vi` if `$EDITOR` unset
- [ ] TUI not launched when `--editor` used

---

## Phase 5: Non-Interactive Mode (~20 min)

**Spec scenario**: "Non-interactive mode creates session without opening anything"

### Changes Required

**File: cmd/fabbro/main.go**

Add `--no-interactive` flag:
```go
noInteractiveFlag := reviewCmd.Bool("no-interactive", false, "Create session without opening TUI or editor")
```

Implementation:
```go
if *noInteractiveFlag {
    // Just create session and print ID
    fmt.Println(session.ID)
    return nil
}
```

### Success Criteria

- [ ] `fabbro review --stdin --no-interactive` creates session
- [ ] Session ID printed to stdout
- [ ] No TUI or editor launched
- [ ] Exit code 0 on success

---

## Phase 6: JSON Output Completeness (~30 min)

**Spec scenario**: "JSON contains all annotation fields" (currently @partial)

### Current Issue

JSON output missing `sourceFile` and `createdAt`.

### Changes Required

**File: internal/session/session.go or cmd/fabbro/main.go**

Ensure JSON output includes:
```json
{
  "sessionId": "abc123",
  "sourceFile": "document.md",
  "createdAt": "2026-01-29T10:00:00Z",
  "annotations": [...]
}
```

### Success Criteria

- [ ] `sourceFile` included in JSON output
- [ ] `createdAt` included in JSON output
- [ ] Empty string for sourceFile when stdin

---

## Summary

| Phase | Deliverable | Time |
|-------|-------------|------|
| 1 | Compact JSON (`--compact`) | 20m |
| 2 | Content hash verification | 45m |
| 3 | Custom session ID (`--id`) | 30m |
| 4 | Editor mode (`--editor`) | 30m |
| 5 | Non-interactive mode | 20m |
| 6 | JSON output completeness | 30m |

**Total: ~2.75 hours**

## Dependencies

All phases are independent.
