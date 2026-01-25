# Changelog

All notable changes to fabbro are documented in this file.

## [Unreleased]

### Added

- **Session Lookup by File** - `fabbro apply --file <path>` finds sessions by source file (2026-01-25)
- **Save Notification** - TUI shows confirmation when session is saved with auto-clear (2026-01-25)

### Fixed

- Viewport calculation now accounts for wrapped lines (2026-01-24)

## [0.1.0] - 2026-01-14

Initial feature-complete release with TUI, FEM parsing, and Claude Code integration.

### Added

#### TUI Features
- Syntax highlighting using Chroma
- Line wrap for long lines
- Visual indicator for annotated lines (gutter markers)
- Change annotation (`x`) for inline replacement suggestions
- Multi-line selection with `v` + `j/k`
- SPC command palette for annotations (always available)
- Context-aware hotkey bar
- Improved navigation: `Ctrl+d/u` (half-page), `gg/G` (first/last line)
- Annotation keybindings for all 6 types: comment, delete, question, expand, unclear, change

#### CLI Features
- `fabbro review <file>` - review from file argument
- `fabbro review --stdin` - review from stdin
- `fabbro apply <id> --json` - JSON output for agent integration
- `fabbro completion <shell>` - shell completions (bash, zsh, fish, powershell)
- `--version` flag

#### FEM Parser
- Full FEM (Fabbro Editing Markup) comment parser
- All 6 annotation type patterns
- Edge case handling and documented parser limitations

#### Session Management
- Date-based session naming format
- 64-bit entropy session IDs
- Frontmatter validation on load
- File permissions hardening (0600)

#### Integration
- TUI integration testing with mcp-tui-test MCP server
- Rule of 5 iterative review commands

### Fixed

- Off-by-one bug in annotation save
- 1-indexed line numbers for annotations
- `v` key consistently toggles selection off
- `w` key saves without quitting
- Error on conflicting `--stdin` and file args
- Session ID included in apply command error messages

### Security

- SEC-001: Increased session ID entropy to 64 bits
- SEC-002: Validate frontmatter on load
- SEC-003: Input size limits for review command
- SEC-004: Check sessions dir in IsInitialized
- OPS-001: Tightened file permissions for sessions

## [0.0.1] - 2026-01-11

Tracer bullet implementation proving end-to-end flow.

### Added

- `fabbro init` command - initialize `.fabbro/` directory
- Basic TUI with j/k navigation, line selection
- FEM comment parser (basic implementation)
- `fabbro apply` command with JSON output
- Integration test for full tracer bullet flow
- README, CONTRIBUTING, and initial documentation
- Gherkin specs for all planned features
