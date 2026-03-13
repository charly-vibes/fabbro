# Init Command Enhancements Plan

## Overview

Extend `fabbro init` with additional options: `--quiet`, `--agents`, and subdirectory detection.

**Spec**: `specs/01_initialization.feature`

## Current State

- `fabbro init` creates `.fabbro/sessions/` directory
- Already initialized projects show message and exit 0
- Missing: templates/, config.yaml, .gitignore, --quiet, --agents

## Desired End State

- `--quiet` suppresses output
- `--agents` scaffolds agent integration files
- Subdirectory detection warns about parent initialization
- Full `.fabbro/` structure: sessions/, templates/, config.yaml, .gitignore

---

## Phase 1: Complete .fabbro Structure (~30 min)

**Spec scenario**: "Initializing a new project" (currently @partial)

### Changes Required

**File: cmd/fabbro/main.go** or new **internal/init/init.go**

Create full structure:
```
.fabbro/
├── sessions/
├── templates/
├── config.yaml
└── .gitignore
```

**config.yaml** (minimal):
```yaml
# Fabbro configuration
# See https://github.com/charly-vibes/fabbro for documentation
version: 1
```

**.gitignore**:
```
sessions/
```

### Success Criteria

- [ ] `fabbro init` creates sessions/, templates/, config.yaml, .gitignore
- [ ] Existing files are not overwritten
- [ ] Test covers all created files

---

## Phase 2: Quiet Mode (~15 min)

**Spec scenario**: "Quiet initialization"

### Changes Required

**File: cmd/fabbro/main.go**
- Add `--quiet` flag to init command
- Suppress output when flag is set

### Success Criteria

- [ ] `fabbro init --quiet` produces no stdout output
- [ ] Exit code 0 on success
- [ ] Errors still go to stderr

---

## Phase 3: Subdirectory Detection (~30 min)

**Spec scenario**: "Initializing in a subdirectory of an initialized project"

### Changes Required

**File: internal/config/config.go** or new location
- Add `FindRootUp()` to walk parent directories looking for `.fabbro/`

**File: cmd/fabbro/main.go**
- Before creating `.fabbro/`, check parents
- If found, warn: "Warning: parent directory /path/to/parent is already initialized"
- Still create local `.fabbro/` (this is intentional per spec)

### Success Criteria

- [ ] Subdirectory init shows warning about parent
- [ ] Local `.fabbro/` is still created
- [ ] No warning when no parent is initialized

---

## Phase 4: Agent Integration Scaffolding (~1.5 hours)

**Spec scenarios**: "Initializing with agent integration scaffolding", "Initializing with agents updates AGENTS.md", "Agent scaffolding detects available agents"

### Changes Required

**File: internal/init/agents.go** (new)

Create agent command files:
```
.agents/commands/fabbro-review.md
.claude/commands/fabbro-review.md  (if .claude/ exists)
```

**Template content for fabbro-review.md**:
```markdown
# Fabbro Review Workflow

Use fabbro to review LLM-generated content with structured annotations.

## Usage

1. Generate content and pipe to fabbro:
   ```bash
   cat content.md | fabbro review --stdin
   ```

2. Annotate in the TUI:
   - `SPC` opens command palette
   - `c` comment, `d` delete, `e` expand, `q` question, `k` keep, `u` unclear
   - `w` saves and exits

3. Extract annotations:
   ```bash
   fabbro apply <session-id> --json
   ```

## Annotation Types

| Key | Type | Use for |
|-----|------|---------|
| c | comment | General feedback |
| d | delete | Mark for removal |
| e | expand | Request more detail |
| q | question | Ask clarifying question |
| k | keep | Mark as good |
| u | unclear | Flag confusion |
```

**AGENTS.md update logic**:
- If AGENTS.md exists, append fabbro section (if not already present)
- Use marker comments to detect existing section

### Detection Logic

- Always create `.agents/commands/`
- Create `.claude/commands/` only if `.claude/` directory exists
- Create `.cursor/commands/` only if `.cursor/` directory exists

### Success Criteria

- [ ] `fabbro init --agents` creates .agents/commands/fabbro-review.md
- [ ] Creates .claude/commands/ only if .claude/ exists
- [ ] Updates AGENTS.md with fabbro section (preserving existing content)
- [ ] Idempotent: running twice doesn't duplicate content

---

## Summary

| Phase | Deliverable | Time |
|-------|-------------|------|
| 1 | Complete .fabbro structure | 30m |
| 2 | `--quiet` flag | 15m |
| 3 | Subdirectory detection | 30m |
| 4 | `--agents` scaffolding | 1.5h |

**Total: ~2.75 hours**

## Dependencies

Phases are independent and can run in any order.
