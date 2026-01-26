# Multi-Line Annotation Behavior Analysis

**Date:** 2026-01-26  
**Issue:** fabbro-9vc

## Current Behavior

When a user selects multiple lines (e.g., lines 5-10) and adds an annotation:

```go
// tui.go:448-455
for line := start; line <= end; line++ {
    m.annotations = append(m.annotations, fem.Annotation{
        StartLine: line + 1,
        EndLine:   line + 1,
        Type:      m.inputType,
        Text:      text,  // Same text for every line
    })
}
```

**Result:** 6 separate annotations created, each with `StartLine == EndLine`, all with identical text.

## Questions & Answers

### 1. Is this intentional for inline FEM markers?

**No, this appears to be a simplification, not intentional design.**

The FEM format is line-based—markers are embedded inline within the source text. The parser ([parser.go:30-47](file:///var/home/sasha/para/areas/dev/gh/charly/fabbro/internal/fem/parser.go#L30-L47)) extracts annotations per-line because that's how they appear in `.fem` files.

However, the `Annotation` struct already supports multi-line ranges:
```go
type Annotation struct {
    StartLine int `json:"startLine"`
    EndLine   int `json:"endLine"`  // Can differ from StartLine
    ...
}
```

### 2. Should we create a single annotation with StartLine=5, EndLine=10?

**Yes, for TUI-created annotations.**

| Approach | Pros | Cons |
|----------|------|------|
| **Single annotation (recommended)** | Semantic intent preserved; cleaner JSON output; easier for agents to interpret; matches user mental model | Requires deciding how to serialize multi-line to FEM format |
| **One per line (current)** | Simple implementation; direct FEM mapping | Loses semantic grouping; inflated annotation count; redundant text duplication |

### 3. How does this affect the apply command output?

Currently, `apply` handles multi-line annotations correctly ([main.go:203-206](file:///var/home/sasha/para/areas/dev/gh/charly/fabbro/cmd/fabbro/main.go#L203-L206)):

```go
if a.StartLine == a.EndLine {
    fmt.Fprintf(stdout, "  Line %d: [%s] %s\n", ...)
} else {
    fmt.Fprintf(stdout, "  Lines %d-%d: [%s] %s\n", ...)
}
```

The spec already anticipates this ([04_apply_feedback.feature:146-150](file:///var/home/sasha/para/areas/dev/gh/charly/fabbro/specs/04_apply_feedback.feature#L146-L150)):
```gherkin
@planned
Scenario: Multi-line annotations span correct range
  Given a session exists with an annotation spanning lines 42-50
  Then the annotation should have startLine 42 and endLine 50
```

## Recommendation

**Create a single annotation for multi-line selections in the TUI.**

### Implementation

Change `tui.go:448-455` from:
```go
for line := start; line <= end; line++ {
    m.annotations = append(m.annotations, fem.Annotation{...})
}
```

To:
```go
m.annotations = append(m.annotations, fem.Annotation{
    StartLine: start + 1,
    EndLine:   end + 1,
    Type:      m.inputType,
    Text:      text,
})
```

### FEM Serialization

For saving to `.fem` format, a multi-line annotation could be:

1. **Marker on first line only** — Place `{>> comment <<}` at end of line 5; store range in session metadata
2. **Block marker syntax** — Introduce `{>>START}` ... `{<<END}` for spans (breaking change)
3. **Metadata-only** — Don't embed multi-line in FEM text; store in session JSON only

**Recommendation:** Option 1 for now—minimal change, preserves FEM compatibility.

### Verification

- [ ] Existing tests pass after change
- [ ] Add test: multi-line selection creates single annotation
- [ ] Update spec `@planned` → `@implemented` for multi-line scenario
