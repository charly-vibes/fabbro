# fabbro

A local-first code review annotation tool with a terminal UI.

## Overview

fabbro lets you annotate code for review using [FEM (First Editor Markup)](docs/fem.md) syntax. It's designed to work with AI coding assistants like Claude Code, enabling structured feedback loops between human reviewers and AI.

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
```

## Commands

| Command | Description |
|---------|-------------|
| `fabbro init` | Initialize fabbro in the current directory (creates `.fabbro/`) |
| `fabbro review <file>` | Start a review session with content from a file |
| `fabbro review --stdin` | Start a review session, reading content from stdin |
| `fabbro apply <id>` | Show annotations from a session |
| `fabbro apply <id> --json` | Output annotations as JSON |
| `fabbro completion <shell>` | Generate shell completion scripts (bash, zsh, fish, powershell) |

See [CLI documentation](docs/cli.md) for full details.

## TUI Keybindings

| Key | Action |
|-----|--------|
| `j` / `↓` | Move cursor down |
| `k` / `↑` | Move cursor up |
| `v` | Toggle line selection |
| `c` | Add comment to selected line |
| `w` | Save session |
| `Q` | Quit |

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

fabbro uses First Editor Markup for annotations:

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
