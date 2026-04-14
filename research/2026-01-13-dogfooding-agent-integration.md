# Research: Dogfooding Fabbro with AI Coding Agents

**Date:** 2026-01-13
**Last revalidated:** 2026-04-14
**Status:** Revised
**Goal:** Discover the desired path for agents interacting with fabbro and humans using it

## Validation Scope

This document was revalidated against the current codebase and docs:

- `cmd/fabbro/main.go`
- `docs/cli.md`
- `docs/tui.md`
- `docs/fem.md`
- `README.md`

Unless otherwise noted, statements in **Current validated state** refer to the behavior implemented in the codebase as of the revalidation date above.

## Executive Summary

fabbro already has the core primitives needed for a useful agent workflow:

- non-interactive session creation
- structured JSON output for annotation extraction
- session listing and resume commands
- agent scaffolding via `fabbro init --agents`
- an AI-oriented primer via `fabbro prime --json`

The main insight still holds: **tools that work well for agents are structured, queryable, and designed for the agentic feedback loop** (agent action → human feedback → agent revision).

The strongest near-term path is:

1. use fabbro's existing CLI as the agent integration surface
2. validate the human-review loop with Claude Code, Amp, pi, and similar tools
3. add tighter integrations only where repeated friction is proven
4. defer MCP until the CLI-native path is clearly insufficient

---

## Part 1: Current Validated State

### Agent-Relevant Capabilities Already Implemented

Current fabbro capabilities relevant to agents:

- `fabbro review --stdin` — create a review session from piped content and open the TUI
- `fabbro review <file>` — create a review session from a file and open the TUI
- `fabbro review --stdin --no-interactive` — create a session without opening the TUI or editor; prints the session ID to stdout
- `fabbro review <file> --no-interactive` — same for file-based review
- `fabbro review --json` — output session creation info as JSON when running interactively
- `fabbro review --editor` — open the created session in `$EDITOR` instead of the TUI
- `fabbro apply <session-id> --json` — extract annotations as structured JSON
- `fabbro apply --file <path> --json` — find the latest session for a source file and extract annotations as JSON
- `fabbro session list --json` — list sessions as structured JSON
- `fabbro session resume <session-id>` — resume a session in the TUI
- `fabbro session resume <session-id> --editor` — open a session in `$EDITOR`
- `fabbro init --agents` — scaffold agent integration files and append a fabbro workflow section to `AGENTS.md`
- `fabbro prime --json` — output AI-optimized workflow context

### Current JSON Shape

`fabbro apply --json` currently returns a top-level object with:

- `sessionId`
- `sourceFile`
- `createdAt`
- `annotations`

Each annotation contains:

- `type`
- `startLine`
- `endLine`
- `text`

Example:

```json
{
  "sessionId": "abc12345",
  "sourceFile": "src/main.go",
  "createdAt": "2026-01-11T12:00:00Z",
  "annotations": [
    {
      "type": "comment",
      "text": "Consider error handling",
      "startLine": 5,
      "endLine": 5
    }
  ]
}
```

### Current Exit Code Contract

Documented CLI exit codes are currently simple:

- `0` = success
- `1` = error

Structured error JSON and semantic nonzero codes are still proposals, not current behavior.

### Remaining Gaps for Agents

Confirmed gaps that still matter for agents:

- no structured JSON error output contract yet
- no semantic exit code taxonomy yet
- no programmatic annotation injection command such as `fabbro annotate`
- no MCP server
- no formal versioned JSON schema doc or schema validation in CI
- no agent-specific workflow validation report comparing multiple agent environments side by side

---

## Part 2: Lessons from Beads

### Why Beads Works for Agents

Beads (`steveyegge/beads`) has become a strong reference point for AI-agent-friendly CLI design. Useful takeaways:

1. **Structured data, not markdown** — agents are more reliable with queryable structured state than with freeform plans.
2. **Git as database** — state stored alongside code creates good local ergonomics and auditability.
3. **Agent-optimized output** — JSON output, identifiers, and dependency-aware workflows reduce ambiguity.
4. **Lightweight CLI** — small composable commands make orchestration easy.
5. **Optional deeper integration** — MCP can help, but the CLI must already stand on its own.

### Beads Workflow Pattern

```text
Human: "What should we work on?"
Agent: [runs `bd ready`] → shows unblocked issues
Human: "Let's do bd-123"
Agent: [runs `bd update bd-123 --status=in_progress`]
       [does the work]
       [runs `bd close bd-123`]
Human: "What's next?"
Agent: [runs `bd ready`] → cycle continues
```

**Key insight:** the agent and human reason together using durable, queryable identifiers.

For fabbro, the equivalent durable identifiers are session IDs.

---

## Part 3: Agent-Friendly CLI Design Patterns

From the broader CLI-agent literature, the most relevant patterns for fabbro are:

### Pattern 1: Machine-Friendly Escape Hatches

Every important command should have:

- `--json` for structured output where appropriate
- consistent stdout vs stderr behavior
- documented exit codes

**Current fabbro assessment:**

- `fabbro apply --json` exists
- `fabbro session list --json` exists
- `fabbro prime --json` exists
- `fabbro review --json` exists for interactive session creation, while `--no-interactive` prints the raw session ID
- stderr warnings are used for source drift during `apply`
- semantic error payloads are not implemented yet

### Pattern 2: Output Formats as API Contracts

CLI outputs that agents rely on should be treated like APIs:

- document them explicitly
- avoid silent breaking changes
- validate them in tests or CI where possible

**Current fabbro assessment:**

- the JSON field names are stable and readable
- the schema is documented in prose, but not yet formalized as a versioned contract

### Pattern 3: MCP for Dynamic Discovery

MCP can help agents discover tools dynamically, but it should complement a strong CLI rather than compensate for a weak one.

**Current fabbro assessment:**

- no MCP server yet
- the current CLI is already strong enough to validate real workflows first

### Pattern 4: Minimal Friction in Human Handoff

When an agent creates something for human review, the handoff needs to be obvious and low-friction.

**Current fabbro assessment:**

- non-interactive session creation works well for agent orchestration
- `fabbro session resume <id>` is the human handoff command
- the TUI is still a separate context switch, so UX around handoff remains worth researching

---

## Part 4: Agent-Specific Integration Points

### Claude Code

Useful integration surfaces:

1. custom slash commands in `.claude/commands/`
2. `CLAUDE.md` or project context instructions
3. later, optional MCP

### Gemini CLI

Useful integration surfaces:

1. `GEMINI.md` or equivalent context file
2. shell-command orchestration
3. later, optional MCP

### Amp / AGENTS.md-driven agents

Useful integration surfaces:

1. `AGENTS.md`
2. generated command files from `fabbro init --agents`
3. shell-command workflows with JSON outputs

### pi.dev

Useful integration surfaces:

1. a pi extension that wraps the fabbro CLI
2. slash commands for review/session operations
3. custom tools that call `fabbro review`, `fabbro apply`, `fabbro session list`, and `fabbro prime`

**Important takeaway:** fabbro does not need deep integration with every agent from day one. The best common denominator is its CLI.

---

## Part 5: Proposed Workflow

This section describes the recommended workflow shape. Steps are labeled as current behavior when already supported.

### Phase 1: Human Reviews AI Output

#### Current supported flow

```text
AI generates content
  ↓
Agent runs: echo "$CONTENT" | fabbro review --stdin --no-interactive
  ↓
Agent captures the session ID from stdout
  ↓
Agent tells human: Run `fabbro session resume <session-id>`
  ↓
Human opens the session in the TUI
  ↓
Human adds annotations and saves with `w`
  ↓
Human exits via `Ctrl+C Ctrl+C` or the command palette
  ↓
Agent runs: fabbro apply <session-id> --json
  ↓
Agent revises content based on structured feedback
```

#### Current FEM annotation types

Canonical FEM syntax today:

- `{>> text <<}` — comment
- `{-- text --}` — delete
- `{?? text ??}` — question
- `{!! text !!}` — expand
- `{== text ==}` — keep
- `{~~ text ~~}` — unclear
- `{++ text ++}` — change

### Phase 2: AI Reviews Human Code

This is still mostly a proposal.

#### Proposed future flow

```text
Human: "Review my changes"
  ↓
Agent runs: git diff | fabbro review --stdin --no-interactive
  ↓
Agent captures session ID
  ↓
Agent adds annotations programmatically (future, requires fabbro annotate or equivalent)
  ↓
Human runs: fabbro session resume <session-id>
  ↓
Human reviews, edits, accepts, rejects, or extends feedback in TUI
```

This phase depends on capabilities fabbro does not yet have, especially programmatic annotation injection.

### Non-Interactive Mode for Automation

Current supported behavior:

```bash
# Create session without opening TUI
printf '%s' "$CONTENT" | fabbro review --stdin --no-interactive
# Output: <session-id>

# Later, apply feedback
fabbro apply <session-id> --json
```

---

## Part 6: Current Priorities

### Already Shipped

These items are done and should be treated as validated primitives, not future work:

1. non-interactive review session creation
2. session ID output suitable for capture by agents
3. file input for `review`
4. session listing in JSON
5. session resume command
6. agent scaffolding via `fabbro init --agents`
7. AI primer output via `fabbro prime --json`

### Next Up

These are the highest-value remaining items for agent integration:

1. **Structured JSON errors**
   - Goal: reliable machine-readable failure modes
   - Example proposal:
     ```json
     {"error":"session not found","code":"SESSION_NOT_FOUND","sessionId":"abc123"}
     ```

2. **Semantic exit codes**
   - Goal: distinguish invalid input, missing session, parse failure, and initialization problems

3. **Programmatic annotation injection**
   - Goal: enable AI-generated review suggestions before the human opens the TUI
   - Likely surface: `fabbro annotate <session-id> ...`

4. **Formal JSON schema contract**
   - Goal: document and test the output that agents depend on

5. **Cross-agent workflow validation**
   - Goal: compare Claude Code, Gemini CLI, Amp, and pi using the same baseline flow

### Deferred / Conditional

These should happen only if the CLI-native path proves insufficient:

1. MCP server
2. deeper agent-specific integrations beyond generated command files and docs
3. advanced diff-aware workflows with automatic line mapping

---

## Part 7: Agent Integration Scaffolding

### Current State: `fabbro init --agents`

This feature is already implemented.

Current behavior:

- creates `.agents/commands/fabbro-review.md`
- creates `.claude/commands/fabbro-review.md` if `.claude/` exists
- creates `.cursor/commands/fabbro-review.md` if `.cursor/` exists
- appends a `## fabbro workflow` section to `AGENTS.md` if it is not already present

This is a strong pattern because it uses file-based agent discovery rather than requiring custom runtime integration.

### Why This Matters

OpenSpec is still a useful reference here: agent tooling often works best when it scaffolds artifacts that agents already know how to consume.

For fabbro, the core lesson is:

> prefer project-local command and context artifacts first; add runtime protocols only when they solve a real validated problem.

---

## Part 8: Non-Obtrusive TUI UX

**Problem:** the agent workflow is already workable, but the human handoff into the TUI may still feel heavier than ideal.

### Questions Worth Researching

1. Should `fabbro session resume` remain full-screen only?
2. Is a popup-based terminal experience better in tmux or zellij?
3. How should agents notify users that a session is ready?
4. What is the fastest exit path that still feels safe?
5. Which parts of the handoff are universal across Claude Code, Amp, Gemini CLI, and pi?

### Reference Patterns

| Tool | Pattern |
|------|---------|
| `fzf` | Fast takeover and dismissal |
| `lazygit` | Full-screen but quick in and out |
| `gum` | Lightweight inline prompting |
| `charmbracelet/pop` | External notification |
| `tmux popup` | Floating pane inside terminal workflow |

### Candidate Approaches

1. tmux/zellij popup wrappers around `fabbro session resume`
2. tighter editor-based workflows via `--editor`
3. shell notifications after session creation
4. a lighter-weight resume mode if the current TUI feels too heavy in practice

---

## Part 9: Related Tooling References

| Tool | Relevant Lesson for Fabbro |
|------|-----------------------------|
| Beads | durable IDs and structured CLI workflows for agent coordination |
| OpenSpec | scaffold agent-readable project artifacts during init |
| Goose | treat protocols like MCP as optional orchestration layers, not starting points |
| OpenCode | useful reference if fabbro later adds MCP-facing tooling |
| pi | extension-based orchestration around existing CLIs rather than forced deep embedding |

---

## Part 10: Experiments to Run

### Experiment 1: Claude Code Baseline

**Goal:** validate the current CLI-native human-review loop.

**Steps:**
1. create or update a `.claude/commands/` command that runs:
   ```bash
   echo "$CONTENT" | fabbro review --stdin --no-interactive
   ```
2. confirm the agent captures the session ID correctly
3. confirm the human can run:
   ```bash
   fabbro session resume <session-id>
   ```
4. confirm the agent can apply annotations with:
   ```bash
   fabbro apply <session-id> --json
   ```

**Success criteria:**
- no command correction required during the loop
- human can complete review without confusion
- agent can reliably parse resulting JSON

### Experiment 2: Amp / AGENTS.md Baseline

**Goal:** validate that generated agent scaffolding is enough for good ergonomics.

**Steps:**
1. run `fabbro init --agents`
2. perform a real review loop through an AGENTS-driven agent
3. note where the user or agent had to improvise

**Success criteria:**
- generated files are enough to bootstrap the workflow
- any missing instructions are concrete and documentable

### Experiment 3: pi Extension Prototype

**Goal:** test whether a pi extension around the fabbro CLI is enough.

**Steps:**
1. create a pi extension that wraps:
   - `fabbro prime --json`
   - `fabbro review --stdin --no-interactive`
   - `fabbro apply <id> --json`
   - `fabbro session list --json`
2. validate a review loop for generated docs or plans

**Success criteria:**
- the extension does not need to embed the fabbro TUI
- the CLI-native workflow feels sufficient inside pi

### Experiment 4: MCP Spike Only If Needed

**Goal:** determine whether MCP solves a demonstrated problem rather than an imagined one.

**Steps:**
1. first gather friction data from the CLI-native experiments above
2. only then build a minimal `fabbro-mcp` spike if repeated integration pain remains

**Success criteria:**
- MCP scope is justified by observed failures or repeated manual glue

---

## Part 11: Open Questions

1. **Should agents add annotations directly, or only humans?**
   - Current state: humans annotate in the TUI
   - Proposal: support both if fabbro gains programmatic annotation injection

2. **What should the annotation-injection API look like?**
   - likely command surface: `fabbro annotate <session-id> ...`
   - unresolved: line vs range semantics, merge behavior, conflict behavior

3. **How should fabbro formalize machine-readable failures?**
   - structured JSON errors
   - semantic exit codes
   - stable error codes

4. **What is the minimum common denominator across agents?**
   - CLI only?
   - CLI + generated command files?
   - CLI + agent-specific context files?
   - when does it make sense to fork integrations by ecosystem?

5. **When is MCP actually warranted?**
   - after repeated real-world friction, not before

6. **What should happen with empty stdin in agent workflows?**
   - proposal: return a structured invalid-input error once structured errors exist

---

## Recommendation

**Recommended path:** treat fabbro's CLI as the primary integration surface.

Concretely:

1. standardize the current CLI-native review loop across agents
2. tighten machine-readable errors and contracts
3. prototype a pi extension and validate generated command files
4. postpone MCP until real usage demonstrates that the CLI approach is not enough

This path has the best cost/benefit ratio because it builds on capabilities fabbro already ships.

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
