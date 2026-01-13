# Research: Dogfooding Fabbro with AI Coding Agents

**Date:** 2026-01-13
**Status:** Revised (Rule of 5 reviewed)
**Goal:** Discover the "desired path" for agents interacting with fabbro and humans using it

## Executive Summary

This document explores how fabbro should integrate with AI coding agents (Claude Code, Gemini CLI, Amp) based on research into existing tools like Steve Yegge's Beads and broader CLI-agent design patterns. The key insight: **tools that work well for agents are structured, queryable, and designed for the agentic feedback loop** (the cycle of agent action → human feedback → agent revision).

### Fabbro Today

Current agent-relevant capabilities:
- `fabbro review --stdin` — Pipe content for review, opens TUI
- `fabbro apply <session-id> --json` — Extract annotations as structured JSON
- JSON schema uses camelCase: `sessionId`, `startLine`, `endLine`, `type`, `content`

What's missing for agents:
- Non-interactive session creation
- Session listing/management commands
- Programmatic annotation injection
- MCP server for dynamic discovery
- Agent integration scaffolding (`fabbro init --agents`)

---

## Part 1: Lessons from Beads

### Why Beads Works for Agents

Beads (steveyegge/beads) has become the de facto issue tracker for AI coding agents. Key insights from Steve Yegge:

1. **Structured data, not markdown** — Agents struggle with parsing markdown plans. They need queryable, structured data (JSONL).

2. **Git as database** — JSONL stored in `.beads/` directory, versioned with code. Self-healing through git history.

3. **Agent-optimized output** — JSON output, dependency tracking, auto-ready task detection.

4. **Lightweight CLI** — Single binary, simple commands (`bd list`, `bd ready`, `bd show`, `bd close`).

5. **MCP server for deep integration** — `beads-mcp` allows agents to discover and use beads tools dynamically.

### Beads Workflow Pattern

```
Human: "What should we work on?"
Agent: [runs `bd ready`] → Shows unblocked issues
Human: "Let's do bd-123"
Agent: [runs `bd update bd-123 --status=in_progress`]
       [does the work]
       [runs `bd close bd-123`]
Human: "What's next?"
Agent: [runs `bd ready`] → Cycle continues
```

**Key insight:** The agent and human reason about work together using issue IDs as shared reference points.

---

## Part 2: Agent-Friendly CLI Design Patterns

From InfoQ's "Keep the Terminal Relevant: Patterns for AI Agent Driven CLIs":

### Pattern 1: Machine-Friendly Escape Hatches

Every command needs:
- `--json` flag for structured output
- Semantic exit codes (0=success, 1=error, specific codes for specific failures)
- Consistent output to stdout (data) vs stderr (diagnostics)

**Status legend:** ✅ Done | 🔶 Partial | ❌ Not started

**Fabbro status:** ✅ `fabbro apply --json` exists. ❌ Need more commands with JSON output.

### Pattern 2: Output Formats as API Contracts

CLI outputs are versioned APIs:
- Breaking changes require major version bumps
- Define explicit JSON schemas
- CI should validate output schemas

**Fabbro status:** ✅ Just canonicalized JSON schema to camelCase. Need to document schema formally.

### Pattern 3: MCP for Dynamic Discovery

Model Context Protocol allows agents to:
- Discover tool capabilities dynamically
- Execute CLIs through constrained, versioned schemas
- Avoid brittleness from output format changes

**Fabbro status:** ❌ No MCP server yet. This is a priority for agent integration.

### Pattern 4: Real-Time Feedback

Long-running tasks need progress reporting:
- Event streaming for agents to detect failures early
- Graceful termination (handle SIGTERM)
- Consistent output streams

**Fabbro status:** 🔶 TUI is interactive, but `fabbro apply` is fast enough to not need streaming.

---

## Part 3: Agent-Specific Integration Points

### Claude Code

**Integration options:**
1. **Custom slash commands** — `.claude/commands/review.md` for `/review` command
2. **CLAUDE.md instructions** — Tell Claude about fabbro workflow
3. **MCP server** — Deepest integration, allows dynamic tool discovery

**Desired workflow:**
```
User: Review this output
Claude: [creates session] → echo "$CONTENT" | fabbro review --stdin --no-interactive
        [prints session ID to stdout, e.g., "session-abc123"]
        [tells human]: "Session created. Run: fabbro resume session-abc123"
Human: [runs fabbro resume session-abc123]
        [adds annotations in TUI, saves with 'w', exits with 'q']
Claude: [runs fabbro apply session-abc123 --json]
        [parses annotations, makes code changes]
```

### Gemini CLI

**Integration options:**
1. **GEMINI.md context file** — Similar to CLAUDE.md
2. **MCP server** — Gemini CLI supports MCP
3. **Tool definitions** — Define fabbro as available tool

**Considerations:**
- Gemini CLI has built-in Google service tools
- May prefer non-interactive mode for automation

### Amp (Sourcegraph)

**Integration options:**
1. **AGENTS.md** — Already using this pattern
2. **MCP server** — `beads-mcp` pattern
3. **Skills** — Amp has built-in skill system

**Considerations:**
- Amp already uses beads successfully
- Can follow same patterns: issue-driven workflow, JSON output, MCP

---

## Part 4: Proposed Agent Workflow for Fabbro

### The Desired Path

#### Phase 1: Human Reviews AI Output

```
┌─────────────────────────────────────────────────────────────┐
│ AI generates long-form content (docs, code, explanations)  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ Human: "Review this with fabbro"                            │
│ Agent: echo "$CONTENT" | fabbro review --stdin --no-interactive │
│        → Creates session, prints session ID to stdout       │
│        → Tells human: "Run: fabbro resume <session-id>"     │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ Human: fabbro resume <session-id>                           │
│        → Opens TUI, adds annotations using FEM syntax:      │
│                                                             │
│   FEM Annotation Syntax:                                    │
│   - {>> comment <<} — General feedback                      │
│   - {-- deletion --} — Mark for removal                     │
│   - {++ addition ++} — Mark for inclusion                   │
│   - {~~ old ~> new ~~} — Suggest replacement                │
│   - {== highlight ==} — Highlight important section         │
│   - {?? question ??} — Ask for clarification                │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ Human saves (w) and exits                                   │
│ Session ID printed to stdout                                │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ Agent: fabbro apply <session-id> --json                     │
│        → Parses annotations, returns structured feedback    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ Agent processes each annotation:                            │
│   - "comment" → Consider feedback, may revise               │
│   - "delete" → Remove or shorten section                    │
│   - "question" → Answer in next revision                    │
│   - "expand" → Add more detail                              │
│   - "keep" → Preserve this section                          │
└─────────────────────────────────────────────────────────────┘
```

#### Phase 2: AI Reviews Human Code

```
┌─────────────────────────────────────────────────────────────┐
│ Human: "Review my changes"                                  │
│ Agent: git diff | fabbro review --stdin --no-interactive    │
│        → Creates session, gets session ID                   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ Agent adds annotations programmatically:                     │
│   fabbro annotate <session-id> \                            │
│     --line=42 --type=comment --text="Consider error handling"│
│   fabbro annotate <session-id> \                            │
│     --line=55 --type=question --text="Why not use constants?"│
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ Agent: "Review ready. Run: fabbro resume <session-id>"      │
│ Human opens session in TUI, sees AI feedback                │
│ Human can respond, ask clarifications, accept/reject        │
└─────────────────────────────────────────────────────────────┘
```

### Non-Interactive Mode for Automation

For CI/CD and automated workflows:

```bash
# Create session without opening TUI
echo "$CONTENT" | fabbro review --stdin --no-interactive
# Output: session-id-abc123

# Later, apply feedback
fabbro apply session-id-abc123 --json
```

---

## Part 5: Implementation Priorities

### Must Have (for dogfooding)

1. **`--no-interactive` flag for `review`** — Create session without opening TUI
2. **Session ID output to stdout** — So agents can capture it
3. **Robust JSON schema** — ✅ Done (sessionId, startLine, endLine)
4. **AGENTS.md documentation** — ✅ Already exists
5. **Error output format** — Structured JSON errors for agent error handling:
   ```json
   {"error": "session not found", "code": "SESSION_NOT_FOUND", "sessionId": "abc123"}
   ```
   Exit codes: 0=success, 1=general error, 2=session not found, 3=invalid input

### Should Have (for agent integration)

6. **MCP server** — `fabbro-mcp` for dynamic tool discovery
7. **File input for `review`** — `fabbro review document.md`
8. **`fabbro sessions --json`** — List sessions as structured data
9. **Session resume** — `fabbro resume <session-id>`
10. **`fabbro annotate`** — Programmatic annotation injection for AI reviewers:
    ```bash
    fabbro annotate <session-id> --line=42 --type=comment --text="..."
    fabbro annotate <session-id> --range=10-15 --type=question --text="..."
    ```

### Nice to Have (for power users)

11. **Annotation templates** — Common patterns (e.g., "needs more examples")
12. **Diff integration** — `git diff | fabbro review` with line mapping
13. **Multi-line annotation support** — JSON schema with `startLine` != `endLine`

---

## Part 6: Agent Integration Scaffolding (OpenSpec/Goose Pattern)

### Lesson from OpenSpec

OpenSpec's `openspec init` command scaffolds per-agent integration files:
- Detects which agents are available (Claude Code, Cursor, Amp, etc.)
- Creates custom slash commands in `.claude/commands/`, `.cursor/commands/`, etc.
- Writes managed `AGENTS.md` stub at project root
- Agents discover fabbro naturally through their existing context mechanisms

**Key insight:** Tools don't need to "hook into" agents — they create **file-based artifacts** that agents discover.

### Lesson from Block's Goose

Goose takes the opposite approach — it's the **orchestrator** that wraps other tools:
- MCP-first architecture (everything is an extension)
- Can use Claude Code, Cursor, Gemini CLI as backend providers
- Recipes (YAML workflows) for automation
- Deeplinks for one-click extension install

### Proposed: `fabbro init --agents`

```bash
# Initialize fabbro with agent integration files
fabbro init --agents

# What it creates:
.fabbro/
├── sessions/
├── config.yaml
└── .gitignore

.agents/commands/
└── fabbro-review.md      # Custom command for Amp

.claude/commands/
└── fabbro-review.md      # Custom command for Claude Code

.cursor/commands/
└── fabbro-review.md      # Custom command for Cursor

# Also appends to AGENTS.md:
# ## Fabbro Review Workflow
# Use `fabbro review --stdin --no-interactive` to create sessions...
```

### Command File Template

```markdown
---
description: Review content with fabbro TUI
---

Create a fabbro review session for the provided content.

1. Run: `echo "$CONTENT" | fabbro review --stdin --no-interactive`
2. Capture the session ID from stdout
3. Tell the user: "Session created. Run: fabbro resume <session-id>"
4. After user annotates, run: `fabbro apply <session-id> --json`
5. Process annotations and revise content accordingly
```

---

## Part 7: Non-Obtrusive TUI UX (Open Research)

**Problem:** When agents invoke fabbro, the TUI shouldn't disrupt terminal flow.

### Questions to Research

1. **Launch mode:** Should fabbro open inline, in new pane, or floating overlay?
2. **Session handoff:** Agent creates session → how does human seamlessly enter TUI?
3. **Notification:** How to signal "session ready for review" without interrupting?
4. **Fast exit:** How to make review feel like a quick aside, not a context switch?

### Reference Patterns

| Tool | Pattern |
|------|---------|
| **fzf** | Floating overlay, instant dismiss |
| **lazygit** | Full-screen but launches fast, `q` exits instantly |
| **gum** | Inline prompts, no screen takeover |
| **charmbracelet/pop** | Desktop notifications from CLI |
| **tmux popup** | Floating pane within terminal multiplexer |

### Possible Approaches

1. **tmux/zellij popup** — `fabbro resume` opens in floating pane
2. **$EDITOR pattern** — Like `git commit`, opens TUI and waits
3. **Background + notify** — Session created in background, notify when ready
4. **Inline mode** — Minimal TUI that doesn't clear screen

**Status:** See issue `fabbro-d3d` for research task.

---

## Part 8: Related TUI Tools

Tools researched for patterns applicable to fabbro:

| Tool | Key Pattern for Fabbro |
|------|------------------------|
| **Elia** (darrenburns/elia) | Keyboard-centric UI, conversation persistence → model for session management |
| **Ralph TUI** (subsy/ralph-tui) | Agent loop orchestration with prd.json (product requirements doc) and Beads → completion detection via `<promise>COMPLETE</promise>` token |
| **OpenCode** (sst/opencode) | Open-source Claude Code alternative with MCP → reference for `fabbro-mcp` implementation |
| **OpenSpec** (Fission-AI/OpenSpec) | Agent scaffolding via `init` command → model for `fabbro init --agents` |
| **Goose** (block/goose) | MCP-first orchestrator, CLI providers, recipes → reference for `fabbro-mcp` |

---

## Part 9: Experiments to Run

### Experiment 1: Claude Code Integration

**Goal:** Validate human-reviews-AI workflow with slash command.

**Steps:**
1. Create `.claude/commands/review.md`:
   ```markdown
   ---
   description: Create a fabbro review session from content
   ---
   
   Create a fabbro review session for the provided content.
   Use `echo "$CONTENT" | fabbro review --stdin --no-interactive`.
   Tell the user the session ID and instruct them to run `fabbro resume <session-id>`.
   After the user saves annotations, run `fabbro apply <session-id> --json` to get structured feedback.
   ```

2. Test workflow with generated documentation

**Success criteria:**
- Agent creates session without blocking
- Human can resume and annotate
- Agent parses JSON output and acts on feedback

### Experiment 2: Gemini CLI Integration

**Goal:** Test cross-agent compatibility.

**Steps:**
1. Create `GEMINI.md` with fabbro instructions
2. Test: "Generate a README, then let me review it with fabbro"

**Success criteria:**
- Same workflow works with different agent
- JSON output is parsed correctly

### Experiment 3: Amp Integration (current)

**Goal:** Document friction points in current dogfooding.

Already using fabbro with Amp via AGENTS.md. Document findings:
- What works well?
- Where are friction points?
- What's missing?

**Success criteria:**
- Documented list of improvements needed
- At least one friction point addressed

### Experiment 4: MCP Server Prototype

**Goal:** Enable dynamic tool discovery.

Build minimal `fabbro-mcp` with:
- `fabbro_review` tool — Create session
- `fabbro_apply` tool — Get annotations
- `fabbro_annotate` tool — Add annotations programmatically
- `fabbro_sessions` resource — List sessions

**Success criteria:**
- Agent can discover fabbro tools via MCP
- All tools work without AGENTS.md instructions

---

## Part 10: Open Questions

1. **Should agents add annotations directly?** Or only humans?
   - Pro: AI code review could add annotations via `fabbro annotate`
   - Con: Loses the "human-in-the-loop" value prop
   - **Proposed answer:** Support both. Phase 1 = human annotates AI output. Phase 2 = AI annotates human code. Human always has final say via TUI.

2. **How to handle multi-line annotations?** 
   - Current: startLine == endLine (single line only)
   - Needed: Block annotations spanning ranges
   - **Proposed answer:** Add `--range=START-END` flag to `fabbro annotate`. JSON schema already supports `startLine` != `endLine`.

3. **Session lifecycle?**
   - When should sessions be auto-cleaned?
   - How long should annotations persist?
   - **Proposed answer:** Sessions persist until explicitly deleted. Add `fabbro sessions --cleanup --older-than=7d` for garbage collection.

4. **Integration depth?**
   - Light: CLI only, agent reads JSON output
   - Medium: Custom slash commands
   - Deep: MCP server with full tool discovery
   - **Proposed answer:** Start with Medium (slash commands), graduate to Deep (MCP) once workflow is validated.

5. **Concurrent session handling?**
   - What if multiple agents create sessions on same content?
   - **Proposed answer:** Sessions are independent by design. Each gets unique ID. No collision possible.

6. **Empty input handling?**
   - What should `echo "" | fabbro review --stdin --no-interactive` do?
   - **Proposed answer:** Return error with exit code 3 (invalid input): `{"error": "empty input", "code": "EMPTY_INPUT"}`

---

## References

| Source | URL | Accessed |
|--------|-----|----------|
| Beads GitHub | https://github.com/steveyegge/beads | 2026-01-13 |
| The Beads Revolution (Steve Yegge) | https://steve-yegge.medium.com/the-beads-revolution-how-i-built-the-todo-system-that-ai-agents-actually-want-to-use-228a5f9be2a9 | 2026-01-13 |
| Beads Best Practices (Steve Yegge) | https://steve-yegge.medium.com/beads-best-practices-2db636b9760c | 2026-01-13 |
| Keep the Terminal Relevant (InfoQ) | https://www.infoq.com/articles/ai-agent-cli/ | 2026-01-13 |
| Building your own CLI Coding Agent (Martin Fowler) | https://martinfowler.com/articles/build-own-coding-agent.html | 2026-01-13 |
| Claude Code Slash Commands | https://code.claude.com/docs/en/slash-commands | 2026-01-13 |
| Awesome Agentic Patterns | https://github.com/nibzard/awesome-agentic-patterns | 2026-01-13 |
