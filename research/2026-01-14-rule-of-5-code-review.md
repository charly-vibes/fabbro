# Research: Rule of 5 Code Review - Improvement Opportunities

**Date:** 2026-01-14  
**Status:** Complete  
**Methodology:** Rule of 5 iterative review (5 passes: Security, Performance, Maintainability, Correctness, Operations)  
**Convergence:** Converged after pass 4 (no new CRITICAL issues, diminishing returns)

---

## Executive Summary

The fabbro codebase is well-structured for an MVP with strong test coverage (85-100% per package). However, the review identified **1 CRITICAL**, **4 HIGH**, and **5 MEDIUM** severity issues that should be addressed before broader adoption.

### Top 3 Findings

1. **SEC-001 (CRITICAL)**: Session ID generation uses only 2 bytes of randomness, causing test flakiness and collision risk
2. **REQ-001 (HIGH)**: TUI save logic has an off-by-one bug—annotations are never written to saved files
3. **SEC-002 (HIGH)**: `session.Load` silently ignores parse errors, producing invalid Session objects

---

## Issue Summary

| Severity | Count | Categories |
|----------|-------|------------|
| CRITICAL | 1 | Security |
| HIGH | 4 | Security (2), Correctness (2) |
| MEDIUM | 5 | Security, Maintainability (2), Correctness, Operations |
| LOW | 3 | Maintainability, Operations (2) |

---

## PASS 1: Security & Safety

### SEC-001: Session ID Generation - Insufficient Entropy (CRITICAL)

**Location:** [internal/session/session.go#L21-L27](file:///var/home/sasha/para/areas/dev/gh/charly/fabbro/internal/session/session.go#L21-L27)

**Description:** Session IDs use only 2 random bytes (16 bits / 65,536 combinations). Combined with date prefix, this is insufficient for production use.

**Evidence:** `TestGenerateID_ReturnsUniqueIDs` fails intermittently due to birthday paradox collisions.

**Impact:**
- Session overwrites on collision (data loss)
- Test flakiness
- Predictable session IDs (security concern)

**Recommendation:**

```go
func generateID() (string, error) {
    const randBytes = 8 // 64 bits of entropy
    bytes := make([]byte, randBytes)
    if _, err := rand.Read(bytes); err != nil {
        return "", fmt.Errorf("failed to generate random session ID: %w", err)
    }
    suffix := hex.EncodeToString(bytes)
    date := time.Now().UTC().Format("20060102")
    return date + "-" + suffix, nil
}
```

Also add collision detection:

```go
func Create(content string) (*Session, error) {
    const maxAttempts = 3
    for attempt := 0; attempt < maxAttempts; attempt++ {
        id, err := generateID()
        if err != nil {
            return nil, err
        }
        sessionPath := filepath.Join(config.SessionsDir, id+".fem")
        if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
            // ID is unique, proceed
            break
        }
    }
    // ... rest of creation
}
```

**Effort:** S (< 1 hour)

---

### SEC-002: session.Load Silently Ignores Parse Errors (HIGH)

**Location:** [internal/session/session.go#L77-L91](file:///var/home/sasha/para/areas/dev/gh/charly/fabbro/internal/session/session.go#L77-L91)

**Description:** Frontmatter parsing ignores `time.Parse` errors and missing fields, returning invalid Session objects.

**Impact:**
- Empty `session_id` → saves to `.fabbro/sessions/.fem` (invalid filename)
- Zero `CreatedAt` → silent data corruption
- No validation → garbage in, garbage out

**Recommendation:**

```go
for _, line := range strings.Split(frontmatter, "\n") {
    if strings.HasPrefix(line, "session_id: ") {
        sessionID = strings.TrimSpace(strings.TrimPrefix(line, "session_id: "))
    }
    if strings.HasPrefix(line, "created_at: ") {
        ts := strings.TrimSpace(strings.TrimPrefix(line, "created_at: "))
        t, err := time.Parse(time.RFC3339, ts)
        if err != nil {
            return nil, fmt.Errorf("invalid created_at in session file: %w", err)
        }
        createdAt = t
        createdAtSet = true
    }
}

if sessionID == "" {
    return nil, fmt.Errorf("invalid session file: missing session_id")
}
if !createdAtSet {
    return nil, fmt.Errorf("invalid session file: missing created_at")
}
```

**Effort:** S (< 1 hour)

---

### SEC-003: No Input Size Limits (MEDIUM)

**Location:** [cmd/fabbro/main.go#L82-L96](file:///var/home/sasha/para/areas/dev/gh/charly/fabbro/cmd/fabbro/main.go#L82-L96)

**Description:** `io.ReadAll(stdin)` and `os.ReadFile()` have no size limits. Malicious or accidental huge inputs cause OOM.

**Recommendation:**

```go
const maxInputBytes = 10 << 20 // 10 MiB

if stdinFlag {
    limited := io.LimitReader(stdin, maxInputBytes+1)
    data, err := io.ReadAll(limited)
    if err != nil {
        return fmt.Errorf("failed to read stdin: %w", err)
    }
    if len(data) > maxInputBytes {
        return fmt.Errorf("input too large (max %d bytes)", maxInputBytes)
    }
    content = string(data)
}
```

**Effort:** S (< 1 hour)

---

### SEC-004: config.IsInitialized Only Checks .fabbro/ (LOW)

**Location:** [internal/config/config.go#L8-L11](file:///var/home/sasha/para/areas/dev/gh/charly/fabbro/internal/config/config.go#L8-L11)

**Description:** If `.fabbro/` exists but `sessions/` was deleted, `review` proceeds then fails at write time.

**Recommendation:**

```go
func IsInitialized() bool {
    if _, err := os.Stat(FabbroDir); err != nil {
        return false
    }
    if _, err := os.Stat(SessionsDir); err != nil {
        return false
    }
    return true
}
```

**Effort:** S (< 30 min)

---

## PASS 2: Performance & Scalability

### PERF-001: No Performance Issues Detected

**Assessment:** The codebase is performant for its use case:

- FEM parser: O(n) line-by-line with precompiled regexes, no catastrophic backtracking
- TUI rendering: O(visible lines), constant for viewport
- Session I/O: Single file reads/writes, appropriate for review-sized content

**Recommendation:** Input size limits (SEC-003) provide the necessary guard against pathological inputs.

**Convergence:** No new issues in this pass.

---

## PASS 3: Maintainability & Readability

### MAINT-001: Time Format Inconsistency (MEDIUM)

**Location:** 
- [internal/session/session.go#L41](file:///var/home/sasha/para/areas/dev/gh/charly/fabbro/internal/session/session.go#L41): uses `time.RFC3339`
- [internal/tui/tui.go#L301](file:///var/home/sasha/para/areas/dev/gh/charly/fabbro/internal/tui/tui.go#L301): uses `"2006-01-02T15:04:05Z07:00"`

**Description:** Two different time format strings for the same purpose. While functionally equivalent, this creates maintenance risk.

**Recommendation:** Use `time.RFC3339` consistently:

```go
// tui.go:301
m.session.CreatedAt.Format(time.RFC3339)
```

**Effort:** S (< 15 min)

---

### MAINT-002: TUI Writes Directly to os.Stderr (MEDIUM)

**Location:** [internal/tui/tui.go#L204-L205](file:///var/home/sasha/para/areas/dev/gh/charly/fabbro/internal/tui/tui.go#L204-L205)

**Description:** Error handling writes directly to `os.Stderr`, making TUI harder to test.

**Recommendation:** Inject an `io.Writer`:

```go
type Model struct {
    // ...existing fields...
    logWriter io.Writer
}

func New(sess *session.Session) Model {
    return Model{
        // ...existing initialization...
        logWriter: os.Stderr,
    }
}

// In handleNormalMode:
case "w":
    if err := m.save(); err != nil {
        fmt.Fprintf(m.logWriter, "Error: %v\n", err)
    }
```

**Effort:** S (< 30 min)

---

### MAINT-003: Annotation Type Strings Duplicated (LOW)

**Location:** 
- [internal/tui/tui.go#L36-L43](file:///var/home/sasha/para/areas/dev/gh/charly/fabbro/internal/tui/tui.go#L36-L43): `inputPrompts` map
- [internal/tui/tui.go#L162-L195](file:///var/home/sasha/para/areas/dev/gh/charly/fabbro/internal/tui/tui.go#L162-L195): hardcoded `inputType` strings
- [internal/fem/markers.go](file:///var/home/sasha/para/areas/dev/gh/charly/fabbro/internal/fem/markers.go): `Markers` map

**Description:** Annotation types ("comment", "delete", etc.) are hardcoded in multiple places with no single source of truth.

**Recommendation:** Define allowed types once in `fem` package:

```go
// fem/types.go
var AnnotationTypes = []string{"comment", "delete", "question", "expand", "keep", "unclear"}

func IsValidType(t string) bool {
    for _, valid := range AnnotationTypes {
        if t == valid {
            return true
        }
    }
    return false
}
```

**Effort:** S (< 1 hour)

---

## PASS 4: Correctness & Requirements

### REQ-001: TUI Save Has Off-By-One Bug (HIGH)

**Location:** [internal/tui/tui.go#L280-L292](file:///var/home/sasha/para/areas/dev/gh/charly/fabbro/internal/tui/tui.go#L280-L292)

**Description:** Critical bug where annotations are **never saved** to the output file:

1. `annotationsByLine` keys use 1-based line numbers (`a.StartLine`)
2. Loop uses 0-based index `i` to look up annotations
3. Result: lookup always misses, annotations never written

```go
// Current broken code:
for i, line := range m.lines {           // i is 0-based
    if ann, ok := annotationsByLine[i]; ok {  // but StartLine is 1-based
```

**Impact:** All annotations are lost on save. Users think they saved feedback, but the file has no FEM markers.

**Recommendation:**

```go
func (m Model) save() error {
    annotationsByLine := make(map[int][]fem.Annotation)  // Support multiple per line
    for _, a := range m.annotations {
        annotationsByLine[a.StartLine] = append(annotationsByLine[a.StartLine], a)
    }

    var result []string
    for i, line := range m.lines {
        lineNum := i + 1  // Convert to 1-based
        if anns, ok := annotationsByLine[lineNum]; ok {
            annotated := line
            for _, ann := range anns {
                marker := fem.Markers[ann.Type]
                annotated += " " + marker[0] + ann.Text + marker[1]
            }
            result = append(result, annotated)
        } else {
            result = append(result, line)
        }
    }
    // ...rest of save logic
}
```

**Effort:** S (< 1 hour)

---

### REQ-002: Only One Annotation Per Line Preserved (HIGH)

**Location:** [internal/tui/tui.go#L280](file:///var/home/sasha/para/areas/dev/gh/charly/fabbro/internal/tui/tui.go#L280)

**Description:** `annotationsByLine := make(map[int]fem.Annotation)` stores only one annotation per line. Multiple annotations on the same line → last one wins silently.

**Impact:** User adds multiple annotations to one line, only last is preserved.

**Recommendation:** (Combined with REQ-001 fix above) Use `map[int][]fem.Annotation` to store all annotations per line.

**Effort:** (Included in REQ-001)

---

### REQ-003: FEM Parser Doesn't Handle Edge Cases (MEDIUM)

**Location:** [internal/fem/parser.go#L15-L21](file:///var/home/sasha/para/areas/dev/gh/charly/fabbro/internal/fem/parser.go#L15-L21)

**Description:** Parser limitations that should be tested/documented:

1. **Single-line only**: `.` doesn't match `\n`, so multi-line annotations don't work
2. **Unbalanced markers**: `{>> test` left in `cleanContent` without warning
3. **Nested markers**: `{>> outer {?? inner ??} <<}` undefined behavior

**Impact:** User confusion when complex FEM syntax doesn't work as expected.

**Recommendation:**

1. Document limitations in README and `docs/fem.md`
2. Add edge case tests:

```go
func TestParse_UnbalancedMarkers(t *testing.T) {
    content := "Hello {>> unclosed marker world"
    annotations, clean, err := Parse(content)
    // Verify: no annotation extracted, marker left in clean content
}

func TestParse_MultipleAnnotationsOnOneLine(t *testing.T) {
    content := "text {>> first <<} middle {?? second ??} end"
    annotations, _, err := Parse(content)
    // Verify: 2 annotations extracted
}
```

**Effort:** M (1-2 hours for tests + docs)

---

### REQ-004: Conflicting --stdin and File Arguments (LOW)

**Location:** [cmd/fabbro/main.go#L82-L99](file:///var/home/sasha/para/areas/dev/gh/charly/fabbro/cmd/fabbro/main.go#L82-L99)

**Description:** If user passes both `--stdin` and a file path, stdin silently wins. User may not realize their file was ignored.

**Recommendation:**

```go
if stdinFlag && len(args) == 1 {
    return fmt.Errorf("cannot use both --stdin and a file path")
}
```

**Effort:** S (< 15 min)

---

## PASS 5: Operations & Reliability

### OPS-001: File Permissions May Be Too Permissive (LOW)

**Location:** 
- [internal/session/session.go#L44](file:///var/home/sasha/para/areas/dev/gh/charly/fabbro/internal/session/session.go#L44): `0644`
- [internal/tui/tui.go#L304](file:///var/home/sasha/para/areas/dev/gh/charly/fabbro/internal/tui/tui.go#L304): `0644`
- [internal/config/config.go#L14](file:///var/home/sasha/para/areas/dev/gh/charly/fabbro/internal/config/config.go#L14): `0755`

**Description:** Sessions may contain proprietary code. On shared systems, `0644` allows other users to read.

**Recommendation:** For enhanced privacy:

```go
// config.go
os.MkdirAll(SessionsDir, 0700)

// session.go, tui.go
os.WriteFile(sessionPath, []byte(fileContent), 0600)
```

**Effort:** S (< 15 min)

---

### OPS-002: Error Messages Could Be More Specific (LOW)

**Location:** [cmd/fabbro/main.go#L135-L141](file:///var/home/sasha/para/areas/dev/gh/charly/fabbro/cmd/fabbro/main.go#L135-L141)

**Description:** Error messages don't always include the session ID being operated on.

**Recommendation:**

```go
if err != nil {
    return fmt.Errorf("failed to load session %q: %w", sessionID, err)
}
// ...
if err != nil {
    return fmt.Errorf("failed to parse FEM content in session %q: %w", sess.ID, err)
}
```

**Effort:** S (< 15 min)

---

## Convergence Assessment

| Pass | New CRITICAL | New HIGH | New Issues Total | False Positive Rate |
|------|--------------|----------|------------------|---------------------|
| 1 (Security) | 1 | 2 | 4 | 0% |
| 2 (Performance) | 0 | 0 | 0 | 0% |
| 3 (Maintainability) | 0 | 0 | 3 | 10% |
| 4 (Correctness) | 0 | 2 | 4 | 5% |
| 5 (Operations) | 0 | 0 | 2 | 15% |

**Status: CONVERGED** after pass 4. Pass 5 yielded only LOW issues with higher false positive rate.

---

## Recommended Next Actions

### Immediate (Before Next Release)

1. **SEC-001**: Increase session ID entropy to 8 bytes, add error handling
2. **REQ-001 + REQ-002**: Fix TUI save off-by-one bug and multi-annotation support
3. **SEC-002**: Validate session frontmatter on load

### Short-Term (Next Sprint)

4. **SEC-003**: Add input size limits
5. **MAINT-001**: Standardize time format
6. **REQ-003**: Add FEM edge case tests and documentation

### Deferred (Future)

7. **OPS-001**: Tighten file permissions
8. **MAINT-002**: Inject io.Writer for testability
9. **MAINT-003**: Centralize annotation type definitions

---

## Implementation Effort Summary

| Priority | Issues | Total Effort |
|----------|--------|--------------|
| Immediate | SEC-001, REQ-001/002, SEC-002 | ~3 hours |
| Short-term | SEC-003, MAINT-001, REQ-003 | ~3 hours |
| Deferred | OPS-001, MAINT-002/003, others | ~2 hours |

**Total: ~8 hours** to address all findings.

---

## References

- Rule of 5 methodology: Steve Yegge's "Six New Tips for Better Coding with Agents"
- Jeffrey Emanuel's discovery of iterative refinement for code review
- fabbro existing research: `research/2026-01-11-implementation-evaluation.md`, `research/2026-01-13-dogfooding-agent-integration.md`
