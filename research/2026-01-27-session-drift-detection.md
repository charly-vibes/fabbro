# Session-to-File Drift Detection Research

**Date:** 2026-01-27  
**Issue:** fabbro-3e7  
**Question:** How should fabbro detect when a source file has changed after a session was created?

## Problem Statement

When a user creates a review session from a file, fabbro stores a snapshot of that content. If the original file is later edited (outside fabbro), the session's annotations may reference incorrect line numbers.

**Scenario:**
1. `fabbro review file.go` → creates session with 100-line snapshot
2. User adds 10 lines at the top of `file.go` outside fabbro
3. `fabbro apply <session-id>` → annotations point to wrong lines

## Approaches

### Option A: Content Hash (SHA256)

Store a cryptographic hash of the original file content in session frontmatter.

#### Implementation

```yaml
---
session_id: 20260127-abc123
created_at: 2026-01-27T10:00:00Z
source_file: 'internal/handler.go'
content_hash: 'dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f'
---
```

**Go implementation:**
```go
import (
    "crypto/sha256"
    "encoding/hex"
)

func hashContent(content []byte) string {
    hash := sha256.Sum256(content)
    return hex.EncodeToString(hash[:])
}

func (s *Session) IsDrifted(currentContent []byte) bool {
    return s.ContentHash != "" && s.ContentHash != hashContent(currentContent)
}
```

#### Advantages
- **No external dependencies**: Works everywhere, no git required
- **Fast**: SHA256 of even large files is <1ms
- **Simple**: Easy to implement and test
- **Deterministic**: Same content always produces same hash
- **Works with stdin**: Can hash content from any source

#### Disadvantages
- **Binary detection only**: Can only say "changed" or "not changed"
- **No line remapping**: Cannot adjust annotation line numbers
- **No change visibility**: Cannot show what changed

#### Use Cases
- Quick "is this stale?" check
- Warn user before applying annotations
- Require `--force` flag if drift detected

---

### Option B: Git Commit SHA

Store the git commit hash when the file was last modified, enabling richer drift detection.

#### Implementation

```yaml
---
session_id: 20260127-abc123
created_at: 2026-01-27T10:00:00Z
source_file: 'internal/handler.go'
content_hash: 'dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f'
git_commit: 'e36e68e4a2b1c3d4e5f6a7b8c9d0e1f2a3b4c5d6'
---
```

**Detection commands:**
```bash
# Get commit when session was created
git log -1 --format="%H" -- path/to/file

# Check if file changed since that commit
git diff --quiet <stored-commit> HEAD -- path/to/file

# Get unified diff for line remapping
git diff -U0 <stored-commit> HEAD -- path/to/file
```

**Go implementation:**
```go
import "os/exec"

type GitInfo struct {
    Commit     string
    IsDirty    bool   // uncommitted changes
    RepoRoot   string
}

func GetGitInfo(filePath string) (*GitInfo, error) {
    // Check if in git repo
    cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
    if err := cmd.Run(); err != nil {
        return nil, nil // Not a git repo, gracefully skip
    }
    
    // Get last commit for this file
    cmd = exec.Command("git", "log", "-1", "--format=%H", "--", filePath)
    out, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    
    return &GitInfo{Commit: strings.TrimSpace(string(out))}, nil
}

func (s *Session) GetDrift(filePath string) (*DriftInfo, error) {
    if s.GitCommit == "" {
        // Fall back to content hash
        return s.checkContentHash(filePath)
    }
    
    // Check if file changed since stored commit
    cmd := exec.Command("git", "diff", "--quiet", s.GitCommit, "HEAD", "--", filePath)
    if cmd.Run() == nil {
        return &DriftInfo{Drifted: false}, nil
    }
    
    // Get diff for potential line remapping
    cmd = exec.Command("git", "diff", "-U0", s.GitCommit, "HEAD", "--", filePath)
    diff, _ := cmd.Output()
    
    return &DriftInfo{
        Drifted: true,
        Diff:    string(diff),
    }, nil
}
```

#### Advantages
- **Rich diff information**: Can show exactly what changed
- **Line remapping possible**: Can parse unified diff to adjust line numbers
- **Commit-based tracking**: Aligns with developer workflow
- **History-aware**: Can detect changes across multiple commits

#### Disadvantages
- **Git dependency**: Only works in git repositories
- **Dirty tree complexity**: Uncommitted changes add edge cases
- **Subprocess overhead**: Shells out to git CLI
- **Path resolution**: Must handle relative vs absolute paths correctly

#### Line Remapping Algorithm

Given a unified diff with `-U0` (no context lines):

```diff
@@ -16,3 +16,4 @@ type Session struct
-       ID        string
+       ID         string
+       SourceFile string
```

Parse hunks to build offset map:
- Lines 1-15: offset = 0
- Lines 16+: offset = +1 (one line added)

Apply to annotation line numbers before displaying/applying.

---

## Comparison Matrix

| Criterion | Content Hash | Git Commit |
|-----------|--------------|------------|
| **Dependency** | None | Git CLI |
| **Detection speed** | <1ms | ~10-50ms |
| **Binary changed?** | ✓ | ✓ |
| **What changed?** | ✗ | ✓ (diff) |
| **Line remapping** | ✗ | ✓ (possible) |
| **Works with stdin** | ✓ | ✗ |
| **Works outside repo** | ✓ | ✗ |
| **Implementation complexity** | Low | Medium |

---

## Recommendation: Layered Approach

Implement **both** as complementary layers:

### Layer 1: Content Hash (Always)
- Store `content_hash` in every session
- Fast, universal "has it changed?" check
- Works everywhere, including stdin and non-git contexts

### Layer 2: Git Info (When Available)
- Store `git_commit` when file is in a git repo
- Enables richer "what changed?" information
- Enables future line remapping feature
- Gracefully degrades to content hash when git unavailable

### API Design

```go
type DriftStatus int

const (
    DriftNone     DriftStatus = iota // No changes detected
    DriftDetected                     // Changed, details unknown
    DriftWithDiff                     // Changed, diff available
)

type DriftInfo struct {
    Status      DriftStatus
    ContentHash string // Current file hash (for comparison)
    Diff        string // Git diff if available
    LinesAdded  int    // Summary stats
    LinesRemoved int
}

// Check returns drift information for a session's source file.
// Falls back gracefully: git diff → content hash → error.
func (s *Session) CheckDrift() (*DriftInfo, error)
```

### CLI Integration

```bash
# Show warning on apply if drifted
$ fabbro apply session-123
⚠️  Source file has changed since session was created.
    3 lines added, 1 line removed.
    Use --force to apply anyway, or --show-diff to see changes.

# Force apply despite drift
$ fabbro apply session-123 --force

# Show what changed
$ fabbro apply session-123 --show-diff
```

### JSON Output (for agents)

```json
{
  "session_id": "20260127-abc123",
  "drift": {
    "status": "drifted",
    "lines_added": 3,
    "lines_removed": 1,
    "has_diff": true
  },
  "annotations": [...]
}
```

---

## Edge Cases

### 1. File Deleted
- Content hash check fails (file not found)
- Return clear error: "Source file no longer exists"

### 2. Git Not Installed
- `exec.LookPath("git")` returns error
- Fall back to content hash only

### 3. Not a Git Repository
- `git rev-parse` fails
- Fall back to content hash only

### 4. Uncommitted Changes (Dirty Tree)
- File modified but not committed
- Content hash will detect drift
- Git diff shows uncommitted changes vs stored commit

### 5. File Renamed/Moved
- Content hash still works (check current path)
- Git can detect renames with `git log --follow`
- Consider: store content hash as primary, path as secondary

### 6. Binary Files
- SHA256 works on any content
- Git diff may not be useful
- Likely rare for code review tool

---

## Implementation Phases

### Phase 1: Content Hash (1-2 hours)
1. Add `ContentHash` field to `Session` struct
2. Compute hash in `session.Create()`
3. Add `session.IsDrifted(filePath)` method
4. Add frontmatter parsing for `content_hash`

### Phase 2: CLI Warning (1 hour)
1. Call `IsDrifted()` in `apply` command
2. Show warning if drifted
3. Add `--force` flag to proceed anyway

### Phase 3: Git Integration (2-3 hours)
1. Add `GitCommit` field to `Session` struct
2. Detect git repo and get commit in `Create()`
3. Add `GetDiff()` method for richer info
4. Add `--show-diff` flag to `apply` command

### Phase 4: Line Remapping (Future)
1. Parse unified diff hunks
2. Build line offset map
3. Adjust annotation line numbers on apply
4. Mark as stretch goal—may not be needed if users re-review

---

## Decision Points for Human Review

1. **Should drift block apply by default?** 
   - Option A: Warn but proceed (current files win)
   - Option B: Block, require `--force` (session data protected)

2. **Include diff in JSON output?**
   - Pro: Agents can make informed decisions
   - Con: Large diffs bloat output

3. **Implement line remapping?**
   - Pro: Seamless experience when files change
   - Con: Complex, may introduce subtle bugs

4. **Store both hashes or just content hash?**
   - Both: Maximum flexibility, slightly larger frontmatter
   - Content only: Simpler, git as pure runtime check

---

## References

- [Git diff documentation](https://git-scm.com/docs/git-diff)
- [Unified diff format](https://en.wikipedia.org/wiki/Diff#Unified_format)
- [Go crypto/sha256](https://pkg.go.dev/crypto/sha256)
- Issue fabbro-3e7: Explore session-to-file sync using git
- Issue fabbro-ltb: File-based session lookup (completed prerequisite)
