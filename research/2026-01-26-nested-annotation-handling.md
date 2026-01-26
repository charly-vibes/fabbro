# Research: Nested Annotation Handling in FEM

**Date**: 2026-01-26  
**Issue**: fabbro-zez  
**Status**: Complete

## Executive Summary

The FEM parser cannot correctly handle nested annotations because regex matching has no concept of balanced delimiters. For fabbro's code review use case, the recommended solution is to **explicitly forbid nesting** with detection and warnings, rather than implementing complex balanced parsing.

## 1. Analysis of Current Problems

### 1.1 Root Cause

The parser uses non-greedy regex patterns like:

```regex
\{>>\s*(.*?)\s*<<\}
```

On nested input `{>> outer {>> inner <<} still outer <<}`:
- First `{>>` opens
- First `<<}` closes (regex cannot track balance)
- Result: extracts `"outer {>> inner"`, leaves `"still outer <<}"` orphaned

### 1.2 Current Behavior (documented in tests)

```go
// TestParse_NestedMarkersUndefinedBehavior
content := "text {>> outer {>> inner <<} still outer <<} end"
// Extracts: "outer {>> inner"
// Clean output: "text  still outer <<} end"  // Corrupted!
```

### 1.3 Secondary Issues

1. **Non-deterministic ordering**: `patterns` is a Go map; iteration order varies between runs
2. **Regex duplication**: `markers.go` is "single source of truth" but patterns duplicate syntax
3. **Line vs cleanLine mismatch**: Matches found on original line, replacements on mutated cleanLine

## 2. Common Strategies for Multi-Annotation Handling

### Strategy A: Disallow Nesting (Most Common for Review Markup)

**Rule**: Annotation body must not contain any opening marker.

- Detect nested markers → leave text unchanged or warn
- CriticMarkup and similar systems typically take this approach
- Review markup prioritizes clarity over compositional structure

### Strategy B: Layering via Spans

Parse all markers into `[start, end)` spans, then apply policy:
- **Outer-wins**: Keep outer span, inner becomes literal text
- **Inner-wins**: Keep innermost spans
- **No-overlap**: Invalidate all overlapping spans

Common in syntax highlighters.

### Strategy C: Stack-Based Balanced Parsing

Replace regex with tokenizer + stack:
- Opening marker → push to stack
- Closing marker → pop if matches, else treat as literal
- Correctly handles balanced nesting

Requires defining semantics: tree vs flat list, cross-type nesting rules.

### Strategy D: Merging/Normalization

Merge overlapping annotations or split into non-overlapping.
Overkill for code review; causes surprising silent transformations.

### Strategy E: Escape Mechanism

Allow literal markers in annotation text via escaping:
- `\{>>` → literal `{>>`
- Pairs with Strategy A (forbid nesting, allow escaped literals)

## 3. Proposed Solutions

### Recommendation: Forbid Nesting with Detection

**Effort**: Small (1-3 hours)

**Implementation**:
1. After regex match, check if captured `Text` contains any opening marker prefix
2. If nested marker detected:
   - Do NOT strip from clean content (preserve original)
   - Skip emitting the annotation
   - Optionally emit warning
3. Document: "Nesting is invalid; use escaping for literal markers"

**Pros**:
- Keeps parser regex-based, single-pass
- Minimal code change, minimal risk
- Matches CriticMarkup expectations
- No UI changes required

**Cons**:
- Users cannot nest (but they can't now either—this makes failure predictable)
- Requires warning mechanism for good UX

**Impact on FEM Parsing/Rendering**:
- More reliable: malformed input won't corrupt clean output
- No partial removal leaving orphan markers
- Rendering unchanged

### Alternative: Stack-Based Tokenizer

**Effort**: Medium-Large (1-2 days)

**Only consider if**:
- Users need to comment on text containing FEM markers (meta-reviews)
- Range-based annotations needed (not just end-of-line)
- Multi-line annotations or structured threads planned

**Implementation**:
1. Tokenize line left-to-right for any opening/closing markers
2. Stack matching for balanced pairs
3. Build clean content by removing emitted spans

**Pros**:
- Correct balanced nesting support
- Deterministic behavior

**Cons**:
- More code and tests
- Must define nesting semantics (tree vs flat, cross-type rules)
- UI may need to represent nested structure

## 4. Recommendations

### Immediate Action

1. **Implement nesting detection** in `Parse()`:
   - Build list of opening tokens from `Markers` (avoid duplication)
   - Check captured text for any opening token
   - If found, skip annotation and preserve original line

2. **Make parsing deterministic**:
   - Iterate over fixed slice of types instead of map

3. **Add escape mechanism** (optional, low priority):
   - `\{>>` → literal `{>>`
   - Allows users to include markers in annotation text

### Future Consideration

Move to stack-based parsing only if:
- Meta-review workflows become common
- Range-based annotations needed
- Multi-line annotation support required

## 5. Implementation Sketch

```go
// openingMarkers built from Markers keys
var openingMarkers = []string{"{>>", "{--", "{??", "{!!", "{==", "{~~", "{++"}

func containsNestedMarker(text string) bool {
    for _, marker := range openingMarkers {
        if strings.Contains(text, marker) {
            return true
        }
    }
    return false
}

// In Parse(), after match:
if containsNestedMarker(match[1]) {
    // Skip this match, leave line unchanged
    continue
}
```

## 6. Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Silent skipping confuses users | Add warning mechanism or preserve with visual indicator |
| False positives (user wants literal `{>>`) | Add escape mechanism |
| Scope creep to "rich text editor" | Keep line-local, define strict policies upfront |

## Conclusion

For fabbro's local-first code review goals, **forbidding nesting with detection** is the highest-leverage solution. It makes failure predictable and non-destructive while keeping the parser simple. Stack-based balanced parsing should only be considered if concrete user needs emerge for nested or range-based annotations.
