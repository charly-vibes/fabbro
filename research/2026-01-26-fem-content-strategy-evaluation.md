# FEM File Content Strategy Evaluation

**Date:** 2026-01-26  
**Issue:** fabbro-3kq  
**Question:** Should `.fem` files contain all original content or only annotated regions?

## Current Implementation

The current implementation stores **full content**:

```
---
session_id: 20260126-abc123
created_at: 2026-01-26T10:00:00Z
source_file: '/path/to/file.go'
---

func main() {
    fmt.Println("Hello") {>> Consider using log package <<}
    // ... all other lines preserved ...
}
```

Annotations are inline using FEM markup (e.g., `{>> comment <<}`).

## Option A: Full Content (Current)

Store the entire original file with inline annotations.

### Advantages
- **Self-contained**: Session file is readable standalone
- **No reconstruction needed**: Content is immediately usable
- **Git-friendly**: Clean diffs show exactly what changed
- **Simple loading**: Just parse frontmatter + body
- **Editor-friendly**: Can open in any text editor

### Disadvantages
- **Storage overhead**: Duplicates original file content
- **Staleness risk**: If source file changes, session content diverges
- **Larger files**: A 1000-line file with 1 annotation = 1000+ lines stored

### Storage Example
- Original: 500 lines × 80 chars = ~40 KB
- With 5 annotations: ~40.5 KB (negligible overhead from markup)

## Option B: Annotated Regions Only

Store only annotation metadata with line references.

```yaml
---
session_id: 20260126-abc123
created_at: 2026-01-26T10:00:00Z
source_file: '/path/to/file.go'
source_hash: sha256:abcdef...
---
annotations:
  - line: 2
    type: comment
    text: "Consider using log package"
  - line: 45
    type: question
    text: "Why not use a map here?"
```

### Advantages
- **Minimal storage**: Only metadata, ~100 bytes per annotation
- **Smaller files**: 5 annotations = ~500 bytes vs ~40 KB
- **Source tracking**: Hash ensures we detect source changes

### Disadvantages
- **Reconstruction required**: Must read original file + merge annotations
- **Source dependency**: Broken if source file is deleted/moved
- **Line drift**: Source file edits invalidate line numbers
- **Complex loading**: Read session → read source → validate hash → merge
- **Not standalone**: Session file meaningless without source

### Storage Example
- 5 annotations: ~500 bytes
- Savings: ~39.5 KB per session

## Analysis

### 1. Storage Size Implications

| Scenario | Full Content | Regions Only | Savings |
|----------|--------------|--------------|---------|
| 100-line file, 2 annotations | ~8 KB | ~200 B | 97.5% |
| 500-line file, 5 annotations | ~40 KB | ~500 B | 98.8% |
| 1000-line file, 10 annotations | ~80 KB | ~1 KB | 98.8% |
| 10 sessions on 500-line file | ~400 KB | ~5 KB | 98.8% |

**Verdict:** Regions-only saves 97-99% storage, but absolute numbers are small (KB, not MB). Storage is cheap; complexity is not.

### 2. Performance Impact

| Operation | Full Content | Regions Only |
|-----------|--------------|--------------|
| Save | O(n) write | O(a) write (a = annotations) |
| Load | O(n) parse | O(a) parse + O(n) read source |
| Display | Immediate | Requires merge step |
| Source missing | Works | Fails |
| Source modified | Content preserved | Line numbers invalid |

**Verdict:** Full content is faster and more resilient for all read operations.

### 3. Reconstruction Complexity

With regions-only, reconstruction requires:

1. Read session file (annotation metadata)
2. Locate source file (may have moved)
3. Read source file
4. Validate hash matches (detect changes)
5. If hash differs: attempt fuzzy line matching? Fail? Warn?
6. Merge annotations into content at correct positions

This introduces significant edge cases:
- **Source deleted**: Session unusable
- **Source renamed**: Session unusable (unless we track renames)
- **Source edited**: Lines may shift; annotations attach to wrong content
- **Git checkout**: Different branch = different source = mismatched annotations

Handling these requires either:
- Complex heuristics (fuzzy matching, content anchoring)
- Strict requirements (immutable source during review)
- Hybrid approach (store context around annotated lines)

**Verdict:** Reconstruction complexity outweighs storage savings for a local-first tool.

### 4. Use Case Alignment

Fabbro's core use cases:

| Use Case | Full Content | Regions Only |
|----------|--------------|--------------|
| Review stdin (piped content) | ✓ Works | ✗ No source file |
| Review file, continue later | ✓ Works | ⚠ Source must be unchanged |
| Share session file | ✓ Self-contained | ✗ Requires source |
| Archive old reviews | ✓ Standalone | ✗ Source may be gone |
| Git-track sessions | ✓ Diff-friendly | ✓ Small diffs |

**Verdict:** Full content aligns better with local-first, offline-capable design.

### 5. Future Considerations

- **Session-to-file sync** (fabbro-3e7): Easier with full content—write directly
- **Export formats**: Full content simplifies conversion to other formats
- **Multi-file reviews**: Regions-only would need per-file source tracking

## Recommendation

**Keep full content in `.fem` files.**

### Rationale

1. **Simplicity**: No reconstruction logic, no source dependencies
2. **Resilience**: Works offline, works if source changes, works for stdin
3. **Alignment**: Matches local-first philosophy—sessions are independent artifacts
4. **Storage cost is acceptable**: Even 100 sessions × 50 KB = 5 MB (trivial)
5. **Git-friendly**: Full content shows meaningful diffs

### When Regions-Only Makes Sense

Consider regions-only if:
- Reviewing very large files (>1 MB) becomes common
- Server-side storage becomes a concern
- Multi-file review sessions need optimization

These scenarios are not current priorities.

### Implementation Notes

The current implementation is correct. No changes needed.

If storage ever becomes a concern, consider:
- Compression (gzip `.fem.gz`)
- Optional "compact mode" for large files
- Content deduplication across sessions

## Conclusion

**Full content storage is the right choice for fabbro's design goals.** The storage overhead is negligible for typical use cases, while the simplicity and resilience benefits are significant. Regions-only would add complexity without meaningful user benefit.
