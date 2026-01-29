# fabbro

> *"For you, il miglior fabbro"*
> — after T.S. Eliot, The Waste Land

A local-first code review annotation tool with a terminal UI.

## Overview

fabbro lets you annotate code for review using [FEM (Fabbro Editing Markup)](docs/fem.md) syntax. It's designed to work with AI coding assistants like Claude Code, enabling structured feedback loops between human reviewers and AI.

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

## Commands

| Command | Description |
|---------|-------------|
| `fabbro init` | Initialize fabbro in the current directory (creates `.fabbro/`) |
| `fabbro review <file>` | Start a review session with content from a file |
| `fabbro review --stdin` | Start a review session, reading content from stdin |
| `fabbro apply <id>` | Show annotations from a session |
| `fabbro apply <id> --json` | Output annotations as JSON |
| `fabbro apply --file <path>` | Find and apply latest session for a source file |
| `fabbro session list` | List all editing sessions |
| `fabbro session resume <id>` | Resume a previous editing session |
| `fabbro tutor` | Start the interactive tutorial (like vimtutor) |
| `fabbro prime` | Output AI-optimized workflow context |
| `fabbro completion <shell>` | Generate shell completion scripts (bash, zsh, fish, powershell) |

See [CLI documentation](docs/cli.md) for full details.

## TUI Keybindings

| Key | Action |
|-----|--------|
| `j` / `↓` | Move cursor down |
| `k` / `↑` | Move cursor up |
| `gg` / `G` | Jump to first/last line |
| `Ctrl+d` / `Ctrl+u` | Scroll half page down/up |
| `/` | Search (fuzzy match) |
| `n` / `p` | Next/previous search match |
| `Esc` | Clear selection/search |
| `v` | Toggle line selection |
| `Space` | Open annotation palette (when selected) |
| `c` | Comment (when selected) |
| `d` | Delete annotation (when selected) |
| `q` | Question annotation (when selected) |
| `e` | Expand annotation (when selected) |
| `u` | Unclear annotation (when selected) |
| `r` | Change/replacement annotation (when selected) |
| `w` | Save session |
| `Ctrl+C Ctrl+C` | Quit (with confirmation) |

See [TUI documentation](docs/tui.md) for full details.

## Shell Completion

Enable tab completion for your shell:

```bash
# Bash (add to ~/.bashrc)
source <(fabbro completion bash)

# Zsh (add to ~/.zshrc)
source <(fabbro completion zsh)

# Fish (add to ~/.config/fish/config.fish)
fabbro completion fish | source
```

## FEM Syntax

fabbro uses Fabbro Editing Markup for annotations:

```
{>> This is a comment <<}
```

See [FEM documentation](docs/fem.md) for the full syntax reference.

## Claude Code Integration

fabbro is designed to integrate with Claude Code for AI-assisted code review. See [integration guide](docs/claude-code.md).

## Development

```bash
# Run the tool locally (builds and runs in one step)
just run init
just run review --stdin < file.go

# Install to ~/go/bin for testing across directories
just install

# Run tests
just test

# Run full CI pipeline
just ci

# See all available commands
just help
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the development workflow.

## License

MIT
