# fabbro: Human Review Tool for LLM Outputs

**Design Document v1.0**  
*Learning from beads, optimized for human-AI collaboration*

---

## Executive Summary

**fabbro** is a review tool for humans to markup and comment on LLM-generated content without losing their place while reading. Inspired by literary editors like Ezra Pound ("il miglior fabbro" - the better craftsman), it enables linear reading with inline feedback.

**Core Problem**: When LLMs generate long responses (500+ lines), users must scroll back and forth to add comments, losing context and flow.

**Solution**: Two-tier approach
1. **Quick inline markup** for short conversations (<500 lines)
2. **Editor-based review sessions** for long documents (500+ lines)

**Primary Inspiration**: [beads](https://github.com/steveyegge/beads) by Steve Yegge - issue tracker for AI agents that successfully integrates with Claude Code via CLI-first design and SessionStart hooks.

---

## Table of Contents

1. [Learning from beads](#learning-from-beads)
2. [The Two Workflows](#the-two-workflows)
3. [fabbro CLI Design](#fabbro-cli-design)
4. [Markup Language (FEM)](#markup-language-fem)
5. [Integration with Claude Code](#integration-with-claude-code)
6. [Implementation Phases](#implementation-phases)
7. [Example Session Flows](#example-session-flows)
8. [Architecture Decisions](#architecture-decisions)
9. [Comparison: beads vs fabbro](#comparison-beads-vs-fabbro)
10. [Next Steps](#next-steps)

---

## Learning from beads

### What is beads?

beads (`bd`) is an issue tracker designed for AI coding agents, created by Steve Yegge. It solves the "50 First Dates" problem where agents wake up with no memory of previous work.

**Key Innovations**:
- **Git-backed storage**: JSONL files committed to git (source of truth)
- **SQLite cache**: Local fast queries, auto-syncs with JSONL
- **Hash-based IDs**: Collision-resistant (bd-a1b2, bd-f14c) for multi-agent workflows
- **Dependency graph**: Four types (blocks, related, parent-child, discovered-from)
- **CLI-first**: No GUI required, works via bash commands
- **SessionStart hooks**: Auto-injects context at conversation start

### How beads integrates with Claude Code

**Installation Pattern**:
```bash
# Quick install
curl -fsSL https://raw.githubusercontent.com/steveyegge/beads/main/scripts/install.sh | bash

# Or via npm (for Claude Code Web)
npm install -g @beads/bd

# Initialize in project
bd init
```

**SessionStart Hook** (`.claude/hooks/session-start.sh`):
```bash
#!/bin/bash
# Install bd if not present
if ! command -v bd &> /dev/null; then
    npm install -g @beads/bd
fi

# Initialize if needed
if [ ! -d .beads ]; then
    bd init --quiet
fi

echo "✓ bd ready. Use 'bd ready' to see work."
```

**Claude Code Integration Methods**:
1. **CLI commands via bash_tool** - Direct execution (recommended)
2. **MCP server** - For environments without shell access
3. **Plugin/slash commands** - `/beads:ready`, `/beads:create`, etc.

**Why It Works**:
- **Context efficient**: ~1-2K tokens vs 10-50K for MCP tool schemas
- **Low latency**: Direct CLI, no protocol overhead
- **Universal**: Works with any editor with shell access
- **Lightweight**: Single Go binary, no daemon required for basic use

### Key Takeaways for fabbro

1. ✅ **CLI-first design** - Not editor-first, not web-first
2. ✅ **SessionStart hooks** - Auto-inject context at conversation start
3. ✅ **Git-backed storage** - Persist sessions across conversations
4. ✅ **Simple data format** - Parseable markup, not binary
5. ✅ **No context switching** - Everything in conversation flow
6. ✅ **Bash integration** - Use bash_tool for editor operations
7. ✅ **npm packaging** - Easy install for Claude Code Web

---

## The Two Workflows

### Workflow 1: Quick Inline Edits

**Use Case**: Short feedback on recent responses

**Pattern**: Lightweight markup directly in conversation

**Example**:
```
You: Explain dependency injection

Claude: [generates 200 line explanation]

You: {-- DELETE: lines 10-30, too much history --}
     {>> The examples section is great <<}
     {!! EXPAND: Show how DI works with testing !!}

Claude: I see your feedback:
        - Delete historical context ✓
        - Examples section is good ✓
        - Add testing examples ✓
        
        Here's the revision...
```

**Implementation**: Claude parses markup inline, no external tool

**Advantages**:
- Zero latency
- No file creation
- Stays in conversation
- Works for simple feedback

**Limitations**:
- Only for recent messages (<500 lines)
- Hard to review long documents linearly
- No persistent session

**When to Use**: 
- Recent messages
- Simple edits (delete, comment, question)
- Quick iterations

---

### Workflow 2: Long Async Edits

**Use Case**: Detailed review of long documents

**Pattern**: Create review session, open editor, parse feedback

**Example**:
```
You: Write a comprehensive guide to React hooks

Claude: [generates 2000 lines]

You: /fabbro review

Claude: Opening your editor for detailed review...
        [calls fabbro CLI]

[Editor opens with markup file]
[You read linearly, add markup inline]
[Save and close]

Claude: I see your feedback:
        - Delete: Historical context section ✓
        - Expand: useEffect cleanup examples ✓
        - Question: Why useCallback vs useMemo? ✓
        
        Here's the revised guide...
        
        Regarding your question about useCallback vs useMemo:
        [detailed explanation]
```

**Implementation**: fabbro CLI creates .fem file, opens $EDITOR, parses results

**Advantages**:
- Linear reading (no scrolling back/forth)
- Use your preferred editor (vim, vscode, emacs)
- Persistent sessions (resume later)
- Complex feedback (multiple types on same paragraph)

**Limitations**:
- Requires tool installation
- Slightly higher latency (file I/O)
- Need to understand markup syntax

**When to Use**:
- Long documents (>500 lines)
- Need careful review
- Complex feedback
- Multiple feedback types

---

## fabbro CLI Design

### Installation

Following beads pattern for consistency:

```bash
# Quick install (Unix)
curl -fsSL https://raw.githubusercontent.com/username/fabbro/main/install.sh | bash

# Quick install (Windows)
irm https://raw.githubusercontent.com/username/fabbro/main/install.ps1 | iex

# Via npm (for Claude Code Web)
npm install -g fabbro

# Via Homebrew
brew tap username/fabbro
brew install fabbro

# From source
git clone https://github.com/username/fabbro
cd fabbro
go build -o fabbro ./cmd/fabbro
```

### Core Commands

```bash
# Initialize fabbro in project
fabbro init                             # Creates .fabbro/ directory

# Create review session
fabbro review [file]                    # Review file
fabbro review --stdin                   # Review from stdin (pipe)
fabbro review --last                    # Review last Claude response
fabbro review --last --format minimal   # Use minimal template

# Resume interrupted review
fabbro resume <session-id>              # Continue editing
fabbro sessions                         # List active sessions
fabbro show <session-id>                # Show session details

# Apply feedback (parse and format for Claude)
fabbro apply <session-id>               # Human-readable output
fabbro apply <session-id> --json        # JSON for Claude

# Session management
fabbro list                             # List all sessions
fabbro delete <session-id>              # Delete session
fabbro clean --older-than 7d            # Clean old sessions

# Configuration
fabbro config set editor "code --wait"  # Set preferred editor
fabbro config set template standard     # Set default template
fabbro config list                      # Show all config
```

### File Structure

```
.fabbro/
├── sessions/                           # Review sessions
│   ├── session-abc123.fem              # Markup file (opened in editor)
│   ├── session-abc123.json             # Session metadata
│   ├── session-abc123.original.txt     # Original content (backup)
│   └── session-abc123.feedback.json    # Parsed feedback
├── config.yaml                         # User preferences
├── templates/                          # Markup templates
│   ├── standard.md                     # Full syntax guide
│   ├── minimal.md                      # Compact reference
│   └── custom/                         # User templates
└── .gitignore                          # Ignore sessions/
```

**What gets committed to git**:
- `.fabbro/config.yaml` (optional, user choice)
- `.fabbro/templates/custom/` (if user creates custom templates)
- Sessions are NOT committed (temporary, conversation-specific)

### Session Metadata Schema

```json
{
  "session_id": "abc123",
  "created_at": "2026-01-09T10:30:00Z",
  "status": "editing|completed|abandoned",
  "original_source": "claude_response|file|stdin",
  "original_length": 2000,
  "editor_opened_at": "2026-01-09T10:30:05Z",
  "editor_closed_at": "2026-01-09T10:45:30Z",
  "template_used": "standard",
  "feedback_count": 12,
  "feedback_types": {
    "delete": 3,
    "comment": 5,
    "question": 2,
    "expand": 2
  }
}
```

---

## Markup Language (FEM)

**FEM = Fabbro Edit Markup**

### Design Principles

1. **Human-readable** - Clear even in plain text
2. **Unambiguous** - Clear start/end markers
3. **Inline-friendly** - Works mid-paragraph
4. **Cursor-friendly** - Easy to type while reading
5. **Parse-friendly** - Simple regex patterns

### Core Syntax

```markdown
{-- DELETE: reason --}
Text to remove
{--/--}

{>> COMMENT: your feedback <<}

{?? QUESTION: ask for clarification ??}

{!! EXPAND: request more detail !!}

{~~ UNCLEAR: mark confusing sections ~~}

{== KEEP: mark excellent sections ==}

{** EMPHASIZE: highlight key point **}

{## SECTION: organize feedback ##}
```

### Rationale for Each Marker

| Marker | Purpose | Visual Cue | Example |
|--------|---------|------------|---------|
| `{-- --}` | **Delete** | Strikethrough metaphor | Remove verbose intro |
| `{>> <<}` | **Comment** | Speech bubble direction | General feedback |
| `{?? ??}` | **Question** | Question marks | Ask for clarification |
| `{!! !!}` | **Expand** | Exclamation = emphasis | Need more examples |
| `{~~ ~~}` | **Unclear** | Wavy = confused | Confusing explanation |
| `{== ==}` | **Keep** | Equals = same/good | Mark excellent work |
| `{** **}` | **Emphasize** | Bold metaphor | Key insight |
| `{## ##}` | **Section** | Heading metaphor | Organize feedback |

### Detailed Examples

#### Delete Block

```markdown
{-- DELETE: Historical context not needed for tutorial --}
The concept of dependency injection has roots in the 1990s...
(5 paragraphs of history)
{--/--}
```

**Parsed output**:
```json
{
  "type": "delete",
  "start_line": 10,
  "end_line": 35,
  "reason": "Historical context not needed for tutorial",
  "deleted_text": "The concept of dependency injection..."
}
```

#### Multiple Feedback Types on Same Paragraph

```markdown
## The useEffect Hook

{>> Good introduction but needs concrete examples <<}

{?? What's the difference between useEffect and useLayoutEffect? ??}

The useEffect hook lets you perform side effects in function components.
{!! EXPAND: Show cleanup function example with event listeners !!}
It's similar to componentDidMount and componentDidUpdate in class components.
{~~ UNCLEAR: "similar to" is vague, explain the actual differences ~~}

{== KEEP: The basic syntax example is perfect ==}
```

**Parsed output**:
```json
[
  {
    "type": "comment",
    "line": 12,
    "text": "Good introduction but needs concrete examples"
  },
  {
    "type": "question",
    "line": 14,
    "text": "What's the difference between useEffect and useLayoutEffect?"
  },
  {
    "type": "expand",
    "line": 17,
    "text": "Show cleanup function example with event listeners"
  },
  {
    "type": "unclear",
    "line": 18,
    "text": "'similar to' is vague, explain the actual differences"
  },
  {
    "type": "keep",
    "line": 20,
    "text": "The basic syntax example is perfect"
  }
]
```

### Syntax Variations

#### Short Form (for quick inline comments)

```markdown
This explanation is great {>> 👍 <<}
This part is confusing {~~ unclear ~~}
```

#### Nested Markup (handled by parser)

```markdown
{-- DELETE: Too verbose --}
This section {>> which is actually good <<} should be removed.
{--/--}
```

**Parser behavior**: Outer delete takes precedence, inner comment is noted as "user wanted to keep this part but it's in deleted block"

### Template Files

#### Standard Template (`.fabbro/templates/standard.md`)

```markdown
# Reviewing Claude's Response
# Session ID: {{session_id}}
# Created: {{timestamp}}

Use the markup syntax below to add feedback while reading.
Save and close when done.

---

{{content}}

---

# Markup Syntax Reference

## Delete Section
{-- DELETE: reason --}
Text to remove
{--/--}

## Add Comment
{>> COMMENT: your feedback <<}

## Ask Question  
{?? QUESTION: ask for clarification ??}

## Request Expansion
{!! EXPAND: request more detail !!}

## Mark Unclear
{~~ UNCLEAR: mark confusing sections ~~}

## Mark Excellent
{== KEEP: mark good sections ==}

## Emphasize Key Point
{** EMPHASIZE: highlight important point **}

## Section Boundary
{## SECTION: organize your feedback ##}

---
Tips:
- Read linearly, add markup as you go
- Multiple feedback types can be on same paragraph
- Use {== KEEP ==} to mark sections you especially like
- Questions will be answered in the revision
```

#### Minimal Template (`.fabbro/templates/minimal.md`)

```markdown
# Review Session {{session_id}}

{{content}}

---
Quick syntax: {-- delete --} {>> comment <<} {?? question ??} {!! expand !!} {~~ unclear ~~} {== keep ==}
```

---

## Integration with Claude Code

### Method 1: Skill Definition (Recommended)

**Location**: `/mnt/skills/user/fabbro/SKILL.md`

```markdown
# Fabbro - Review Long Responses

## Purpose
Enable humans to review and markup long AI responses using their preferred editor.

## When to Use
- User says "/fabbro review", "/review", "open in editor", "let me markup that"
- Response is >200 lines
- User wants to add inline comments/questions/deletions
- User requests detailed feedback workflow

## Trigger Patterns
- Explicit: "/fabbro review", "/review", "fabbro review"
- Implicit: "let me review that", "open in editor", "I want to markup this", "let me add comments"

## How It Works

### Step 1: Detect Review Request

Listen for trigger patterns in user's message.

### Step 2: Create Review File

Use `bash_tool` to create the review file:

```bash
# Get the last assistant message and create review session
fabbro review --last --format standard --output /tmp/fabbro-session-{{random_id}}.fem
```

This creates a `.fem` file containing:
- Header with syntax guide
- The last assistant response
- Footer with quick reference

### Step 3: Open Editor

Use `bash_tool` to open user's preferred editor:

```bash
# Open editor and wait for it to close
${EDITOR:-vim} /tmp/fabbro-session-{{session_id}}.fem
```

**CRITICAL**: This command blocks until the editor closes. Claude should not continue until the user saves and exits.

### Step 4: Parse Feedback

After editor closes, use `bash_tool` to parse the markup:

```bash
fabbro apply {{session_id}} --json
```

**Example output**:
```json
{
  "session_id": "abc123",
  "feedback": [
    {
      "type": "comment",
      "line": 15,
      "text": "Good intro but jump to examples faster"
    },
    {
      "type": "delete",
      "start_line": 20,
      "end_line": 45,
      "reason": "Historical context not needed",
      "deleted_text": "The concept of dependency..."
    },
    {
      "type": "question",
      "line": 60,
      "text": "How does this work with threads?"
    },
    {
      "type": "expand",
      "line": 75,
      "text": "Need concrete examples of mutable borrows"
    },
    {
      "type": "keep",
      "start_line": 90,
      "end_line": 95,
      "text": "This section is perfect"
    }
  ]
}
```

### Step 5: Generate Revision

Process the feedback naturally:

1. **Acknowledge each piece of feedback** explicitly
2. **Group related feedback** (e.g., all questions together)
3. **Generate revised content** addressing all points
4. **Answer questions** inline in the revision
5. **Preserve sections marked as "keep"** exactly

**Example acknowledgment**:
```
I see your feedback:
1. You want me to jump to examples faster ✓
2. Historical context removed ✓
3. The ownership rules section is perfect - keeping as-is ✓
4. You asked about threading - let me explain... ✓
5. You want concrete borrow examples - adding now ✓

Here's the revised guide:
[revised content]
```

## Important Notes

1. **Always show the markup syntax** in the header - users may not remember
2. **Wait for editor to close** - don't continue until file is saved
3. **Acknowledge all feedback** - mention each marked section
4. **Preserve KEEP sections** exactly as they were
5. **Answer questions** inline in the revised content
6. **Handle empty markup** - if user closes without changes, ask if they want to continue
7. **Support multiple reviews** - user can invoke /fabbro review on revisions

## Edge Cases

### User Closes Without Saving

If the parsed feedback is empty:

```
I see you closed the editor without making changes. Would you like to:
1. Continue with the original response
2. Open the editor again
3. Provide feedback directly in chat?
```

### Syntax Errors in Markup

If markup is malformed, show what was successfully parsed:

```
I parsed these feedback items:
- Comment at line 15: "Good intro"
- Question at line 60: "How does this work with threads?"

I couldn't parse some sections. Could you clarify:
- Line 100: Unclosed {-- DELETE marker
```

### Very Large Responses (>1000 lines)

Before opening editor:

```
This is a large response (1247 lines). Would you like to:
1. Review full document in editor
2. Review specific sections only (which sections?)
3. Provide feedback in chat instead
```

## Configuration

User can configure fabbro behavior:

```bash
# Set preferred editor
fabbro config set editor "code --wait"

# Set default template
fabbro config set template minimal

# Auto-save sessions
fabbro config set autosave true
```

## Troubleshooting

### Editor doesn't open

```
It looks like fabbro couldn't open your editor. Let me check:
[runs: fabbro config get editor]

Your editor is set to: {{editor}}

Would you like to:
1. Set a different editor: fabbro config set editor "vim"
2. Provide feedback directly in chat instead
```

### fabbro not installed

```
It looks like fabbro isn't installed. I can help you install it:

For Unix/Linux/Mac:
curl -fsSL https://raw.githubusercontent.com/username/fabbro/main/install.sh | bash

For Windows (PowerShell):
irm https://raw.githubusercontent.com/username/fabbro/main/install.ps1 | iex

Or via npm:
npm install -g fabbro

Would you like me to walk you through the installation?
```

## Examples

### Example 1: Simple Review

User: Write a guide to Python decorators
Claude: [generates 300 line guide]
User: /review

Claude: I'll open that in your editor for review.
[creates .fem file, opens editor]

[User marks up:]
- {-- DELETE: Historical context --}
- {>> Good examples <<}
- {?? How do class decorators differ? ??}

[Saves and closes]

Claude: I see your feedback:
- Removed historical context ✓
- Examples section is good ✓
- About class decorators: [detailed explanation]

Here's the revised guide...

### Example 2: Multiple Review Cycles

User: Write about microservices architecture
Claude: [generates 800 lines]
User: /review

[First review]
- User marks sections to delete/expand
- Claude revises

User: Better, let me review again
Claude: Opening the revised version...

[Second review]
- User adds more specific feedback
- Claude refines further

User: Perfect, let me review one more time to verify
Claude: Opening the final version for verification...
```

### Method 2: SessionStart Hook

**File**: `.claude/hooks/session-start.sh`

```bash
#!/bin/bash

# Install fabbro if not present
if ! command -v fabbro &> /dev/null; then
    echo "📝 Installing fabbro review tool..."
    npm install -g fabbro --silent
fi

# Initialize in project if needed
if [ ! -d .fabbro ]; then
    fabbro init --quiet
fi

echo "✓ fabbro ready. Use '/review' to review long responses."
```

**For Claude Code Web** (`.claude/hooks/session-start.sh`):

```bash
#!/bin/bash
set -e

echo "📝 Setting up fabbro review tool..."

# Try npm first
if ! command -v fabbro &> /dev/null; then
    if npm install -g fabbro --silent 2>/dev/null && command -v fabbro &> /dev/null; then
        echo "✓ Installed via npm"
    elif command -v go &> /dev/null; then
        echo "npm failed, trying go install..."
        go install github.com/username/fabbro/cmd/fabbro@latest
        export PATH="$PATH:$HOME/go/bin"
        echo "✓ Installed via go"
    else
        echo "✗ Installation failed"
        exit 1
    fi
fi

# Initialize if needed
if [ ! -d .fabbro ]; then
    fabbro init --quiet
fi

echo "✓ fabbro ready. Use '/review' for detailed feedback."
```

### Method 3: User Preferences Integration

**File**: `.claude/settings.local.json`

```json
{
  "instructions": "When I say '/review', use fabbro to open my editor for detailed feedback. Run: fabbro review --last"
}
```

This injects the instruction into every conversation automatically.

---

## Implementation Phases

### Phase 1: MVP (Week 1)

**Goal**: Prove the concept with minimal functionality

**Deliverables**:
- `fabbro init` - Create `.fabbro/` directory
- `fabbro review --stdin` - Accept content from pipe
- Basic FEM parser (delete, comment, question only)
- Manual workflow test (no Claude integration)

**Implementation**:
```bash
# Test manually
echo "Long content here..." | fabbro review --stdin
# Opens editor
# Parse feedback
fabbro apply session-xyz
```

**Success Criteria**: Can create review session, markup in editor, parse feedback

**Time Estimate**: 3-4 days

---

### Phase 2: Editor Integration (Week 2)

**Goal**: Seamless editor workflow

**Deliverables**:
- Auto-detect `$EDITOR` (fallback to vim)
- Template system (standard, minimal)
- Session persistence (`.fabbro/sessions/`)
- `fabbro resume <session-id>` command
- `fabbro sessions` list command

**Implementation**:
```python
# editor.py
import os
import subprocess

def open_editor(filepath: str) -> bool:
    editor = os.environ.get('EDITOR', 'vim')
    result = subprocess.run([editor, filepath])
    return result.returncode == 0
```

**Success Criteria**: Editor opens, waits for close, handles interruptions

**Time Estimate**: 4-5 days

---

### Phase 3: Claude Code Skill (Week 3)

**Goal**: Integrate with Claude Code via skill

**Deliverables**:
- Complete `SKILL.md` for `/mnt/skills/user/fabbro/`
- SessionStart hook script
- Test with Claude Code CLI
- Document in `AGENTS.md` pattern
- Handle edge cases (no editor, parse errors)

**Implementation**:
- Write comprehensive SKILL.md (see Method 1 above)
- Test with real Claude Code sessions
- Iterate based on feedback

**Success Criteria**: Claude can invoke `/review`, open editor, parse feedback, generate revision

**Time Estimate**: 5-6 days

---

### Phase 4: Advanced Features (Week 4)

**Goal**: Polish and extend

**Deliverables**:
- Git integration (commit review sessions optionally)
- All feedback types (expand, unclear, keep, emphasize, section)
- Session history and diff
- `fabbro diff session-1 session-2` command
- Web UI for session visualization (optional)

**Implementation**:
- Complete parser for all markup types
- Add git commands (optional)
- Build simple web UI with live reload

**Success Criteria**: Full feature set working, documented, tested

**Time Estimate**: 6-7 days

---

### Phase 5: Polish & Documentation (Week 5)

**Goal**: Production-ready release

**Deliverables**:
- Complete documentation (README, SKILL.md, examples)
- Installation scripts (bash, PowerShell, npm)
- Test suite (unit tests, integration tests)
- Example sessions (with screenshots/recordings)
- Community templates

**Success Criteria**: Anyone can install and use fabbro with Claude Code

**Time Estimate**: 4-5 days

---

**Total Timeline**: ~4-5 weeks to production-ready v1.0

---

## Example Session Flows

### Flow 1: Quick Python Tutorial Review

```
User: Explain Python list comprehensions comprehensively

Claude: [generates 600 line guide with examples, edge cases, performance notes]

User: /review

Claude: Opening in your editor for detailed review...
[runs: fabbro review --last --format standard]
[opens: $EDITOR /tmp/fabbro-session-abc123.fem]

[User's markup in editor:]
---
# Python List Comprehensions: Complete Guide

## Introduction

{>> Good overview but too formal - make it more casual <<}

List comprehensions provide a concise way to create lists...

## Basic Syntax

{== KEEP: This explanation is perfect ==}

[expression for item in iterable]

## Filtering Examples

{-- DELETE: These examples are too similar --}
# Example 1
numbers = [x for x in range(10) if x % 2 == 0]

# Example 2  
evens = [num for num in range(10) if num % 2 == 0]
{--/--}

{!! EXPAND: Show nested list comprehension example !!}

## Performance

{?? How does this compare to map() and filter()? ??}

## Edge Cases

{~~ UNCLEAR: The explanation of the walrus operator here is confusing ~~}
---

[User saves and closes]

Claude: I see your feedback:
1. Make the tone more casual ✓
2. Keep the basic syntax section exactly ✓
3. Remove duplicate filtering examples ✓
4. Add nested list comprehension example ✓
5. Explain comparison to map()/filter() ✓
6. Clarify walrus operator usage ✓

Here's the revised guide:

# Python List Comprehensions: Complete Guide

## Introduction

Let's learn list comprehensions - they're awesome! You'll use them all the time...

## Basic Syntax

[expression for item in iterable]

## Filtering Examples

# Filter even numbers
evens = [x for x in range(10) if x % 2 == 0]  # [0, 2, 4, 6, 8]

## Nested List Comprehensions

Here's how to flatten a matrix:

matrix = [[1, 2, 3], [4, 5, 6], [7, 8, 9]]
flat = [num for row in matrix for num in row]  # [1,2,3,4,5,6,7,8,9]

## Performance: List Comp vs map()/filter()

List comprehensions are generally faster than map()/filter() because:
- They're optimized at the bytecode level
- No function call overhead
- More readable for complex logic

Benchmark example:
[code showing timing comparison]

## The Walrus Operator (:=)

In Python 3.8+, you can assign and use values in one expression:

# Without walrus - processing twice
filtered = [expensive_func(x) for x in items if expensive_func(x) > 10]

# With walrus - processing once
filtered = [y for x in items if (y := expensive_func(x)) > 10]

This is especially useful when the operation is expensive.

Would you like me to add more examples or clarify anything?
```

---

### Flow 2: Research Paper Review (Complex Feedback)

```
User: Write a research paper outline on quantum computing

Claude: [generates 1200 line outline with abstract, intro, methodology, results, discussion]

User: /review

[User's comprehensive markup:]
---
# Quantum Computing: A Comprehensive Survey

## Abstract

{>> Too technical for a general audience - simplify <<}
{~~ UNCLEAR: Acronyms not defined (NISQ, QEC) ~~}

We present a comprehensive survey of quantum computing...

## 1. Introduction

{== KEEP: Historical context is excellent ==}

### 1.1 Background

{?? Why focus on superconducting qubits vs trapped ions? ??}

### 1.2 Motivation

{!! EXPAND: Add concrete applications (drug discovery, cryptography) !!}

## 2. Theoretical Foundation

{-- DELETE: Dirac notation deep dive is too advanced --}
The mathematics of quantum mechanics...
(8 pages of equations)
{--/--}

### 2.1 Superposition

{** EMPHASIZE: This is the key concept - make it prominent **}

### 2.2 Entanglement

{>> Good explanation but needs visual analogy <<}

## 3. Hardware Implementations

{## SECTION: Hardware needs significant expansion ##}

{!! EXPAND: Compare architectures (IBM, Google, IonQ) !!}

### 3.1 Superconducting Qubits

{== KEEP: Technical details are spot-on ==}

### 3.2 Error Correction

{?? What's the current state of error rates? ??}
{?? Timeline for fault-tolerant quantum computing? ??}

## 4. Algorithms

{~~ UNCLEAR: Jump from Shor's to VQE is too abrupt ~~}

### 4.1 Shor's Algorithm

{!! EXPAND: Step-by-step walkthrough with small example (factor 15) !!}

## 5. Conclusion

{-- DELETE: Speculation about AGI is out of scope --}
Some researchers believe quantum computing could enable...
{--/--}

{>> Needs stronger conclusion tying back to intro <<}
---

[Saves and closes]

Claude: Comprehensive feedback received! Let me address each point:

**Abstract**:
- Simplified technical language ✓
- Defined acronyms (NISQ, QEC) ✓

**Introduction**:
- Historical context kept exactly as-is ✓
- Superconducting vs trapped ions: I'll explain the tradeoffs ✓
- Added drug discovery and cryptography examples ✓

**Theoretical Foundation**:
- Removed advanced Dirac notation section ✓
- Emphasized superposition as key concept ✓
- Added visual analogy for entanglement (Bell's telephone) ✓

**Hardware**:
- Expanded section with IBM vs Google vs IonQ comparison ✓
- Current error rates: ~0.1% for best systems ✓
- Fault tolerance timeline: 2030-2035 estimated ✓

**Algorithms**:
- Smoothed transition from Shor's to VQE ✓
- Added step-by-step Shor's example (factor 15) ✓

**Conclusion**:
- Removed AGI speculation ✓
- Stronger conclusion added ✓

Here's the revised outline:

# Quantum Computing: A Comprehensive Survey

## Abstract

We present a survey of quantum computing for technical readers from adjacent fields. We cover hardware (superconducting and trapped-ion systems), algorithms (Shor's, VQE), and near-term applications. NISQ (Noisy Intermediate-Scale Quantum) devices show promise, but QEC (Quantum Error Correction) remains challenging...

[Complete revision addressing all 15 feedback items]

Would you like me to expand any specific section further?
```

---

### Flow 3: Code Review Session

```
User: Generate a production-ready REST API for user management

Claude: [generates 800 lines: routes, models, middleware, tests]

User: /review

[User's markup focuses on code quality:]
---
# User Management API

## Routes (routes/users.js)

```javascript
const express = require('express');
const router = express.Router();

{>> Good structure but missing rate limiting <<}

router.post('/users', async (req, res) => {
  {-- DELETE: Validation is incomplete --}
  if (!req.body.email) {
    return res.status(400).json({ error: 'Email required' });
  }
  {--/--}
  
  {!! EXPAND: Use express-validator for full validation !!}
  
  {?? Should we hash passwords here or in model? ??}
  
  const user = await User.create(req.body);
  res.json(user);
});

router.get('/users/:id', async (req, res) => {
  {== KEEP: Error handling is good ==}
  try {
    const user = await User.findById(req.params.id);
    if (!user) return res.status(404).json({ error: 'Not found' });
    res.json(user);
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});
```

## Models (models/User.js)

{~~ UNCLEAR: Why is password stored as plain text? ~~}

{!! EXPAND: Add bcrypt hashing, salt rounds config !!}

## Tests (tests/users.test.js)

{-- DELETE: Mocked tests don't test actual database --}
jest.mock('../models/User');
{--/--}

{!! EXPAND: Add integration tests with test database !!}

{?? How do we handle auth tokens in tests? ??}
---

[Saves and closes]

Claude: I see your code review feedback:

**Routes**:
- Added express-rate-limit middleware ✓
- Removed incomplete validation ✓
- Added express-validator schema ✓
- Password hashing: I'll handle it in pre-save hook (best practice) ✓
- Kept error handling as-is ✓

**Models**:
- Added bcrypt password hashing ✓
- Configurable salt rounds (10 default) ✓

**Tests**:
- Removed mocked tests ✓
- Added supertest integration tests ✓
- Auth tokens: Using JWT in Authorization header ✓

Here's the revised code:

```javascript
// routes/users.js
const express = require('express');
const router = express.Router();
const rateLimit = require('express-rate-limit');
const { body, validationResult } = require('express-validator');

// Rate limiting: 100 requests per 15 minutes
const limiter = rateLimit({
  windowMs: 15 * 60 * 1000,
  max: 100
});

router.use(limiter);

router.post('/users',
  // Validation rules
  body('email').isEmail().normalizeEmail(),
  body('password').isLength({ min: 8 }),
  body('name').trim().notEmpty(),
  
  async (req, res) => {
    // Check validation
    const errors = validationResult(req);
    if (!errors.isEmpty()) {
      return res.status(400).json({ errors: errors.array() });
    }
    
    try {
      // Password hashing happens in pre-save hook
      const user = await User.create(req.body);
      res.status(201).json(user);
    } catch (err) {
      res.status(500).json({ error: err.message });
    }
  }
);

// ... other routes
```

```javascript
// models/User.js
const mongoose = require('mongoose');
const bcrypt = require('bcrypt');

const SALT_ROUNDS = process.env.BCRYPT_SALT_ROUNDS || 10;

const userSchema = new mongoose.Schema({
  email: { type: String, required: true, unique: true },
  password: { type: String, required: true },
  name: String
});

// Hash password before saving
userSchema.pre('save', async function(next) {
  if (!this.isModified('password')) return next();
  
  this.password = await bcrypt.hash(this.password, SALT_ROUNDS);
  next();
});

// Method to compare passwords
userSchema.methods.comparePassword = function(candidatePassword) {
  return bcrypt.compare(candidatePassword, this.password);
};

module.exports = mongoose.model('User', userSchema);
```

```javascript
// tests/users.integration.test.js
const request = require('supertest');
const app = require('../app');
const User = require('../models/User');
const jwt = require('jsonwebtoken');

describe('User API Integration Tests', () => {
  beforeEach(async () => {
    await User.deleteMany({});
  });
  
  describe('POST /users', () => {
    it('creates user with hashed password', async () => {
      const res = await request(app)
        .post('/users')
        .send({
          email: 'test@example.com',
          password: 'password123',
          name: 'Test User'
        });
      
      expect(res.status).toBe(201);
      expect(res.body.email).toBe('test@example.com');
      
      // Verify password is hashed
      const user = await User.findById(res.body._id);
      expect(user.password).not.toBe('password123');
      expect(await user.comparePassword('password123')).toBe(true);
    });
    
    it('validates email format', async () => {
      const res = await request(app)
        .post('/users')
        .send({
          email: 'invalid-email',
          password: 'password123'
        });
      
      expect(res.status).toBe(400);
      expect(res.body.errors).toBeDefined();
    });
  });
  
  describe('GET /users/:id', () => {
    it('returns user by id with valid token', async () => {
      const user = await User.create({
        email: 'test@example.com',
        password: 'password123',
        name: 'Test'
      });
      
      const token = jwt.sign({ id: user._id }, process.env.JWT_SECRET);
      
      const res = await request(app)
        .get(`/users/${user._id}`)
        .set('Authorization', `Bearer ${token}`);
      
      expect(res.status).toBe(200);
      expect(res.body.email).toBe('test@example.com');
    });
  });
});
```

The code is now production-ready with:
- Rate limiting
- Full validation
- Password hashing with bcrypt
- Integration tests with real database
- JWT authentication handling

Would you like me to add more tests or review any specific part?
```

---

## Architecture Decisions

### Why CLI-first?

Following beads' successful pattern:
1. **Lower latency** - Direct execution vs MCP overhead
2. **Context efficient** - No need for large tool schemas
3. **Universal** - Works with any tool that has bash access
4. **Simpler** - Fewer moving parts, easier to debug
5. **Familiar** - Unix philosophy (do one thing well)

### Why Git-backed sessions?

1. **Persistence** - Sessions survive conversation compaction
2. **Audit trail** - History of all reviews
3. **Collaboration** - Share sessions with team (optional)
4. **No server** - Local-first, works offline
5. **Familiar** - Git is already universal

### Why Custom Markup (FEM)?

**Alternatives considered**:

| Alternative | Why Not |
|-------------|---------|
| CriticMarkup | Breaks with code (orthogonal to Markdown) |
| Git diff | Too verbose, not human-friendly for writing |
| Google Docs | Not CLI-friendly, requires web |
| Word tracked changes | Binary format, not parseable |
| HTML comments | Clunky syntax, not intuitive |

**FEM advantages**:
- Human-readable in plain text
- Easy to type while reading
- Unambiguous delimiters
- Simple regex parsing
- Works in any editor
- No special tooling required

### Why SessionStart Hook?

1. **Auto-install** - User doesn't need to remember setup
2. **Zero config** - Works out of the box
3. **Portable** - Travels with the git repo
4. **Transparent** - User sees installation in terminal
5. **Following beads** - Proven pattern

### Why Two Workflows?

**Quick inline** for:
- Recent messages
- Simple feedback
- Fast iterations
- Low friction

**Editor session** for:
- Long documents
- Complex feedback
- Linear reading
- Thoughtful review

Users choose based on context, not forced into one pattern.

### Why No Persistent Daemon?

Unlike beads (which needs daemon for auto-sync), fabbro:
- Operates on conversation scope only
- No background sync needed
- Sessions are temporary
- Simpler architecture
- Less resource usage

---

## Comparison: beads vs fabbro

### Similarities

| Aspect | beads | fabbro |
|--------|-------|--------|
| **Integration** | CLI via bash_tool | CLI via bash_tool |
| **Installation** | SessionStart hook | SessionStart hook |
| **Data format** | JSONL (parseable) | FEM markup (parseable) |
| **Storage** | Git-backed | Git-backed (optional) |
| **Philosophy** | Unix tools | Unix tools |
| **Target** | Claude Code | Claude Code |

### Differences

| Aspect | beads | fabbro |
|--------|-------|--------|
| **Domain** | Issue/task tracking | Document review |
| **Persistence** | Cross-session (permanent) | Within conversation (temporary) |
| **Data structure** | Structured (SQLite) | Unstructured (markup) |
| **Primary user** | AI agent | Human |
| **Workflow** | Ongoing task management | One-time review cycles |
| **Daemon** | Yes (for auto-sync) | No (not needed) |
| **Context size** | 1-2K tokens (primer) | Full document (variable) |
| **Integration depth** | Core to workflow | On-demand feature |
| **State** | Stateful (database) | Stateless (sessions) |

### Complementary Use

beads and fabbro solve different problems:

- **beads**: "What should I work on?" (agent memory)
- **fabbro**: "Let me review what you wrote" (human feedback)

They can be used together:

```
User: /beads ready
Claude: Here are ready tasks...
User: Work on bd-a1b2
Claude: [implements feature, generates 500 line explanation]
User: /fabbro review
[User reviews implementation]
Claude: [revises based on feedback]
User: Looks good, /beads close bd-a1b2
```

---

## Next Steps

### Immediate Actions

1. **Validate design** with potential users
   - Does this solve the "scrolling problem"?
   - Is the markup syntax intuitive?
   - Is the workflow too heavy?

2. **Build MVP** (Phase 1)
   - Focus on core functionality
   - Test with real Claude Code sessions
   - Get feedback early

3. **Iterate** based on usage
   - What features get used?
   - What's missing?
   - What's confusing?

### Development Roadmap

**Week 1**: MVP (init, review, basic parser)
**Week 2**: Editor integration (templates, sessions)
**Week 3**: Claude Code skill (SKILL.md, testing)
**Week 4**: Advanced features (all markup types, git integration)
**Week 5**: Polish (docs, tests, examples)

### Open Questions

1. **Should sessions be committed to git by default?**
   - Pro: Persistent audit trail
   - Con: Clutters repo with temporary files
   - Proposal: User choice via config

2. **Should fabbro support collaborative review?**
   - Multiple humans reviewing same document
   - Merge feedback from different reviewers
   - Proposal: Phase 6 feature

3. **Should there be a web UI?**
   - For non-CLI users
   - For visualizing session history
   - Proposal: Community contribution

4. **How to handle very large documents (10K+ lines)?**
   - Section-based review?
   - Progressive loading?
   - Proposal: Warn user, offer section split

5. **Integration with other LLM CLIs?**
   - Aider, Cursor, Cline, etc.
   - Generic skill format?
   - Proposal: Phase 7, after Claude Code is proven

### Success Metrics

**Technical**:
- Installation success rate >95%
- Editor opening success rate >99%
- Parse accuracy >99.5%
- Average session time <5 minutes

**User Experience**:
- Users prefer fabbro over scrolling for >200 line reviews
- Users successfully complete reviews on first try
- Feedback is accurately interpreted by Claude
- Revision quality improves vs no review

**Adoption**:
- Used in at least 50 projects within 3 months
- Community contributions (templates, integrations)
- Integration requests from other LLM CLIs
- Positive feedback from Claude Code community

---

## Appendix A: Parser Implementation

### Python Implementation (Reference)

```python
# fabbro/parser.py
import re
from dataclasses import dataclass
from typing import List, Optional

@dataclass
class Feedback:
    type: str  # delete, comment, question, expand, unclear, keep, emphasize, section
    location: int  # line number or start line
    end_location: Optional[int]  # for delete blocks
    content: str
    context: str  # surrounding lines for LLM understanding

class FEMParser:
    """Parser for Fabbro Edit Markup"""
    
    PATTERNS = {
        'delete': r'\{--\s*DELETE:([^}]+)--\}(.*?)\{--/--\}',
        'comment': r'\{>>\s*(?:COMMENT:)?([^<]+)<<\}',
        'question': r'\{\?\?\s*(?:QUESTION:)?([^?]+)\?\?\}',
        'expand': r'\{!!\s*(?:EXPAND:)?([^!]+)!!\}',
        'unclear': r'\{~~\s*(?:UNCLEAR:)?([^~]+)~~\}',
        'keep': r'\{==\s*(?:KEEP:)?([^=]+)==\}',
        'emphasize': r'\{\*\*\s*(?:EMPHASIZE:)?([^*]+)\*\*\}',
        'section': r'\{##\s*(?:SECTION:)?([^#]+)##\}'
    }
    
    def parse(self, fem_content: str) -> List[Feedback]:
        """Parse FEM markup into structured feedback"""
        feedback_list = []
        lines = fem_content.split('\n')
        
        # Parse multi-line delete blocks
        feedback_list.extend(self._parse_delete_blocks(fem_content, lines))
        
        # Parse inline feedback
        for line_num, line in enumerate(lines, 1):
            for fb_type, pattern in self.PATTERNS.items():
                if fb_type == 'delete':
                    continue  # Already handled
                    
                matches = re.finditer(pattern, line)
                for match in matches:
                    feedback_list.append(Feedback(
                        type=fb_type,
                        location=line_num,
                        end_location=None,
                        content=match.group(1).strip(),
                        context=self._get_context(lines, line_num)
                    ))
        
        # Sort by location
        return sorted(feedback_list, key=lambda f: f.location)
    
    def _parse_delete_blocks(self, content: str, lines: List[str]) -> List[Feedback]:
        """Parse multi-line delete blocks"""
        pattern = r'\{--\s*DELETE:([^}]+)--\}(.*?)\{--/--\}'
        matches = re.finditer(pattern, content, re.DOTALL)
        
        feedback = []
        for match in matches:
            reason = match.group(1).strip()
            deleted_text = match.group(2).strip()
            
            # Find line numbers
            start_pos = match.start()
            end_pos = match.end()
            
            start_line = content[:start_pos].count('\n') + 1
            end_line = content[:end_pos].count('\n') + 1
            
            feedback.append(Feedback(
                type='delete',
                location=start_line,
                end_location=end_line,
                content=reason,
                context=deleted_text[:200]  # First 200 chars
            ))
        
        return feedback
    
    def _get_context(self, lines: List[str], line_num: int, context_size: int = 2) -> str:
        """Get surrounding lines for context"""
        start = max(0, line_num - context_size - 1)
        end = min(len(lines), line_num + context_size)
        return '\n'.join(lines[start:end])
    
    def to_json(self, feedback_list: List[Feedback]) -> dict:
        """Convert feedback to JSON for Claude"""
        return {
            'total_feedback': len(feedback_list),
            'types': {
                fb_type: len([f for f in feedback_list if f.type == fb_type])
                for fb_type in set(f.type for f in feedback_list)
            },
            'feedback': [
                {
                    'type': f.type,
                    'line': f.location,
                    'end_line': f.end_location,
                    'text': f.content,
                    'context': f.context
                }
                for f in feedback_list
            ]
        }
```

### Go Implementation (Production)

```go
// parser.go
package fabbro

import (
    "regexp"
    "strings"
)

type Feedback struct {
    Type        string `json:"type"`
    Location    int    `json:"line"`
    EndLocation *int   `json:"end_line,omitempty"`
    Content     string `json:"text"`
    Context     string `json:"context"`
}

type Parser struct {
    patterns map[string]*regexp.Regexp
}

func NewParser() *Parser {
    return &Parser{
        patterns: map[string]*regexp.Regexp{
            "delete":    regexp.MustCompile(`\{--\s*DELETE:([^}]+)--\}(.*?)\{--/--\}`),
            "comment":   regexp.MustCompile(`\{>>\s*(?:COMMENT:)?([^<]+)<<\}`),
            "question":  regexp.MustCompile(`\{\?\?\s*(?:QUESTION:)?([^?]+)\?\?\}`),
            "expand":    regexp.MustCompile(`\{!!\s*(?:EXPAND:)?([^!]+)!!\}`),
            "unclear":   regexp.MustCompile(`\{~~\s*(?:UNCLEAR:)?([^~]+)~~\}`),
            "keep":      regexp.MustCompile(`\{==\s*(?:KEEP:)?([^=]+)==\}`),
            "emphasize": regexp.MustCompile(`\{\*\*\s*(?:EMPHASIZE:)?([^*]+)\*\*\}`),
            "section":   regexp.MustCompile(`\{##\s*(?:SECTION:)?([^#]+)##\}`),
        },
    }
}

func (p *Parser) Parse(content string) ([]*Feedback, error) {
    lines := strings.Split(content, "\n")
    feedback := []*Feedback{}
    
    // Parse delete blocks
    feedback = append(feedback, p.parseDeleteBlocks(content, lines)...)
    
    // Parse inline feedback
    for lineNum, line := range lines {
        for fbType, pattern := range p.patterns {
            if fbType == "delete" {
                continue
            }
            
            matches := pattern.FindAllStringSubmatch(line, -1)
            for _, match := range matches {
                feedback = append(feedback, &Feedback{
                    Type:     fbType,
                    Location: lineNum + 1,
                    Content:  strings.TrimSpace(match[1]),
                    Context:  p.getContext(lines, lineNum),
                })
            }
        }
    }
    
    return feedback, nil
}

func (p *Parser) parseDeleteBlocks(content string, lines []string) []*Feedback {
    pattern := p.patterns["delete"]
    matches := pattern.FindAllStringSubmatchIndex(content, -1)
    
    feedback := []*Feedback{}
    for _, match := range matches {
        startPos := match[0]
        endPos := match[1]
        
        startLine := strings.Count(content[:startPos], "\n") + 1
        endLine := strings.Count(content[:endPos], "\n") + 1
        
        reason := strings.TrimSpace(content[match[2]:match[3]])
        deletedText := strings.TrimSpace(content[match[4]:match[5]])
        
        // Truncate context if too long
        if len(deletedText) > 200 {
            deletedText = deletedText[:200] + "..."
        }
        
        feedback = append(feedback, &Feedback{
            Type:        "delete",
            Location:    startLine,
            EndLocation: &endLine,
            Content:     reason,
            Context:     deletedText,
        })
    }
    
    return feedback
}

func (p *Parser) getContext(lines []string, lineNum int) string {
    contextSize := 2
    start := max(0, lineNum-contextSize)
    end := min(len(lines), lineNum+contextSize+1)
    return strings.Join(lines[start:end], "\n")
}
```

---

## Appendix B: Template Files

### Standard Template

```markdown
# Reviewing Claude's Response
# Session ID: {{session_id}}
# Created: {{timestamp}}

Use the markup syntax below to add feedback while reading linearly.
Save and close this file when done.

---

{{content}}

---

# Markup Syntax Reference

## Delete Section
Remove unwanted content with reason:

{-- DELETE: reason for deletion --}
Text to remove (can span multiple lines)
{--/--}

Example:
{-- DELETE: Too much historical context for tutorial --}
In the 1990s, the concept emerged...
{--/--}

## Add Comment
General feedback on any section:

{>> COMMENT: your feedback here <<}

Example:
{>> This explanation is clear but needs concrete examples <<}

## Ask Question
Ask for clarification or additional information:

{?? QUESTION: what you want to know ??}

Example:
{?? How does this work in multi-threaded environments? ??}

## Request Expansion
Ask for more detail on specific topics:

{!! EXPAND: what needs more detail !!}

Example:
{!! EXPAND: Show step-by-step example of error handling !!}

## Mark Unclear
Flag confusing or ambiguous sections:

{~~ UNCLEAR: why it's confusing ~~}

Example:
{~~ UNCLEAR: The relationship between these two concepts is not explained ~~}

## Mark Excellent
Highlight sections that are particularly good:

{== KEEP: what makes it good ==}

Example:
{== KEEP: This example perfectly illustrates the concept ==}

## Emphasize Key Point
Mark important concepts that should be prominent:

{** EMPHASIZE: why it's important **}

Example:
{** EMPHASIZE: This is the core principle - make it stand out **}

## Section Boundary
Organize feedback into logical groups:

{## SECTION: section description ##}

Example:
{## SECTION: Authentication needs complete rewrite ##}

---

## Tips for Effective Review

1. **Read linearly** - Don't jump around, add markup as you go
2. **Be specific** - "Unclear" is less helpful than "The causal relationship is unclear"
3. **Combine types** - Multiple feedback types can apply to the same paragraph
4. **Use KEEP** - Mark good sections so Claude knows what to preserve
5. **Ask questions** - Questions will be answered in the revision
6. **Be concise** - Brief feedback is easier to process

## Common Patterns

**Too verbose**: {-- DELETE: Remove for conciseness --}
**Need examples**: {!! EXPAND: Add 2-3 concrete examples !!}
**Confusing flow**: {~~ UNCLEAR: Transition between topics is abrupt ~~}
**Great work**: {== KEEP: Perfect explanation ==}
**Answer needed**: {?? What about edge case X? ??}

---

Save this file and close your editor when done.
fabbro will parse your feedback and Claude will generate a revision.
```

### Minimal Template

```markdown
# Review: {{session_id}}

{{content}}

---
Quick Reference:
{-- DELETE: reason --} ... {--/--}  Remove section
{>> comment <<}                     General feedback
{?? question ??}                    Ask question
{!! expand !!}                      Need more detail
{~~ unclear ~~}                     Mark confusing
{== keep ==}                        Mark excellent
```

### Code Review Template

```markdown
# Code Review: {{session_id}}

{{content}}

---
Markup Reference:

{-- DELETE: reason --} ... {--/--}  Remove code
{>> comment <<}                     Code feedback
{?? question ??}                    Ask about implementation
{!! expand !!}                      Need tests/docs
{~~ UNCLEAR: security ~~}           Flag security concern
{== KEEP: good pattern ==}          Mark good code

Common patterns:
- Missing validation: {!! EXPAND: Add input validation !!}
- Security issue: {~~ UNCLEAR: Is this SQL injection safe? ~~}
- Good code: {== KEEP: Error handling is excellent ==}
- Need tests: {!! EXPAND: Add unit tests for edge cases !!}
```

---

## Appendix C: Installation Scripts

### Unix/Linux/Mac Install Script

```bash
#!/bin/bash
# install.sh - Install fabbro

set -e

REPO="username/fabbro"
VERSION="${FABBRO_VERSION:-latest}"

echo "📝 Installing fabbro..."

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "❌ Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Download URL
if [ "$VERSION" = "latest" ]; then
    URL="https://github.com/$REPO/releases/latest/download/fabbro-$OS-$ARCH"
else
    URL="https://github.com/$REPO/releases/download/$VERSION/fabbro-$OS-$ARCH"
fi

# Install directory
INSTALL_DIR="${FABBRO_INSTALL_DIR:-/usr/local/bin}"

# Download
echo "Downloading from $URL..."
curl -fsSL "$URL" -o /tmp/fabbro

# Install
chmod +x /tmp/fabbro
sudo mv /tmp/fabbro "$INSTALL_DIR/fabbro"

# Verify
if command -v fabbro &> /dev/null; then
    VERSION=$(fabbro --version)
    echo "✅ fabbro installed successfully: $VERSION"
    echo ""
    echo "Get started:"
    echo "  fabbro init              # Initialize in project"
    echo "  fabbro review --stdin    # Review from pipe"
    echo ""
    echo "Integration with Claude Code:"
    echo "  Create .claude/hooks/session-start.sh and add:"
    echo "  fabbro init --quiet"
else
    echo "❌ Installation failed"
    exit 1
fi
```

### Windows Install Script

```powershell
# install.ps1 - Install fabbro on Windows

$ErrorActionPreference = "Stop"

$Repo = "username/fabbro"
$Version = if ($env:FABBRO_VERSION) { $env:FABBRO_VERSION } else { "latest" }

Write-Host "📝 Installing fabbro..." -ForegroundColor Cyan

# Download URL
if ($Version -eq "latest") {
    $Url = "https://github.com/$Repo/releases/latest/download/fabbro-windows-amd64.exe"
} else {
    $Url = "https://github.com/$Repo/releases/download/$Version/fabbro-windows-amd64.exe"
}

# Install directory
$InstallDir = if ($env:FABBRO_INSTALL_DIR) {
    $env:FABBRO_INSTALL_DIR
} else {
    "$env:USERPROFILE\AppData\Local\Microsoft\WindowsApps"
}

# Create directory if it doesn't exist
if (!(Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

# Download
Write-Host "Downloading from $Url..." -ForegroundColor Gray
$TempFile = "$env:TEMP\fabbro.exe"
Invoke-WebRequest -Uri $Url -OutFile $TempFile

# Install
$DestFile = "$InstallDir\fabbro.exe"
Move-Item -Path $TempFile -Destination $DestFile -Force

# Verify
if (Get-Command fabbro -ErrorAction SilentlyContinue) {
    $InstalledVersion = fabbro --version
    Write-Host "✅ fabbro installed successfully: $InstalledVersion" -ForegroundColor Green
    Write-Host ""
    Write-Host "Get started:" -ForegroundColor Cyan
    Write-Host "  fabbro init              # Initialize in project"
    Write-Host "  fabbro review --stdin    # Review from pipe"
} else {
    Write-Host "❌ Installation failed" -ForegroundColor Red
    exit 1
}
```

---

## Appendix D: FAQ

### General Questions

**Q: Why not just use CriticMarkup?**  
A: CriticMarkup is orthogonal to Markdown syntax, which means it breaks when used with code. FEM is designed specifically for reviewing any content type (prose, code, data).

**Q: Can I use fabbro without Claude Code?**  
A: Yes! fabbro works standalone. The Claude Code integration is optional.

**Q: Does fabbro work with other LLM CLIs?**  
A: The design is portable. Phase 7 will add support for Aider, Cursor, Cline, etc.

**Q: Why not a web UI?**  
A: Terminal-first is more efficient for developers. A web UI is possible as a community contribution.

### Technical Questions

**Q: How does fabbro handle large files (10K+ lines)?**  
A: Future feature: section-based review. For now, fabbro warns and asks if you want to proceed.

**Q: Can multiple people review the same document?**  
A: Not yet. Phase 6 will add collaborative review with feedback merging.

**Q: What editors are supported?**  
A: Any editor that can be launched from command line: vim, emacs, nano, vscode, sublime, etc.

**Q: Does fabbro require git?**  
A: No, but git integration is recommended for session persistence.

### Workflow Questions

**Q: When should I use quick inline vs editor session?**  
A: Quick inline for <500 lines, editor session for detailed review of long documents.

**Q: Can I resume an interrupted review?**  
A: Yes: `fabbro resume <session-id>`

**Q: How do I see my review history?**  
A: `fabbro sessions` lists all sessions. `fabbro show <session-id>` shows details.

**Q: Can I create custom templates?**  
A: Yes! Add `.md` files to `.fabbro/templates/custom/` and use: `fabbro review --template custom/mytemplate`

### Integration Questions

**Q: Does fabbro work in Claude Code for Web?**  
A: Yes, via npm package with SessionStart hook (see Method 2).

**Q: Can I use fabbro in CI/CD?**  
A: Not designed for that use case. fabbro is for human review, not automated checks.

**Q: Does fabbro integrate with beads?**  
A: They're complementary but independent. Use beads for task tracking, fabbro for document review.

---

**Document Version**: 1.0  
**Last Updated**: 2026-01-09  
**Author**: Designed based on beads learnings  
**Status**: Ready for implementation

---

**End of Document**
