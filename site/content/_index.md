---
title: "fabbro"
type: docs
---

# fabbro

> *"For you, il miglior fabbro"*
> — after T.S. Eliot, The Waste Land

A local-first code review annotation tool with a terminal UI.

## Overview

fabbro lets you annotate code for review using [FEM (Fabbro Editing Markup)]({{< relref "/docs/fem" >}}) syntax. It's designed to work with AI coding assistants like Claude Code, enabling structured feedback loops between human reviewers and AI.

## Installation

```bash
go install github.com/charly-vibes/fabbro/cmd/fabbro@latest
```

Or build from source:

```bash
git clone https://github.com/charly-vibes/fabbro.git
cd fabbro
go build -o fabbro ./cmd/fabbro
```

## Quick Start

```bash
# Initialize fabbro in your project
fabbro init

# Start a review session with content from stdin
cat file.go | fabbro review --stdin

# In the TUI: navigate with j/k, select with v, comment with c, save with w

# Extract annotations as JSON
fabbro apply <session-id> --json

# Or find session by source file (useful for agents)
fabbro apply --file plans/my-plan.md --json
```
