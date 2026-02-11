# fabbro justfile - unified local/CI workflow
#
# Same commands run locally and in CI for consistent diagnostics.
# Run `just` for default (fast unit tests), `just ci` for full pipeline.

set shell := ["bash", "-uc"]

# Default: run fast unit tests
default: test

# === Test Commands ===

# Run unit tests (Tier 1) - pre-push gate, <1s target
test:
    go test ./... -race -count=1

# Run unit tests with coverage
test-cover:
    go test ./... -race -count=1 -coverprofile=coverage.out -coverpkg=./...
    go tool cover -func=coverage.out

# Run integration tests (Tier 2) - real files, TUI program
test-integration:
    go test ./... -race -tags=integration -count=1

# Run fuzz tests (Tier 3) - 3s per fuzz target for CI
test-fuzz:
    go test ./... -tags=fuzz -fuzz=. -fuzztime=3s 2>/dev/null || true

# Run ALL test tiers - <15s target
test-all: test test-integration test-fuzz
    @echo "✅ All test tiers passed"

# === Coverage Commands ===

# Check coverage meets threshold (65% minimum)
check-coverage: test-cover
    #!/usr/bin/env bash
    set -euo pipefail
    pct=$(go tool cover -func=coverage.out | grep '^total:' | awk '{print substr($3, 1, length($3)-1)}')
    echo "Total coverage: ${pct}%"
    if (( $(echo "$pct < 65.0" | bc -l) )); then
        echo "❌ Coverage ${pct}% is below required 65%"
        exit 1
    fi
    echo "✅ Coverage ${pct}% meets threshold"

# Generate HTML coverage report
cover-html: test-cover
    go tool cover -html=coverage.out -o coverage.html
    @echo "📊 Open coverage.html in browser"

# === Lint Commands ===

# Run all linters
lint:
    go vet ./...
    @command -v staticcheck >/dev/null 2>&1 && staticcheck ./... || echo "⚠️  staticcheck not installed, skipping"

# Format all Go files
fmt:
    gofmt -w .

# Check formatting (no changes)
fmt-check:
    @test -z "$$(gofmt -l .)" || (echo "❌ Files need formatting:" && gofmt -l . && exit 1)

# === Build Commands ===

# Build the binary
build:
    go build -o bin/fabbro ./cmd/fabbro

# Build with version info
build-release version:
    go build -ldflags="-X main.version={{version}}" -o bin/fabbro ./cmd/fabbro

# Build and run with arguments (e.g., `just run init`, `just run review file.go`)
run *args:
    go run ./cmd/fabbro {{args}}

# Install locally to ~/go/bin for testing across directories (includes git commit)
install:
    go install -ldflags="-X main.version=dev-$(git rev-parse --short HEAD)" ./cmd/fabbro

# === CI Commands ===

# Full CI pipeline (what GitHub Actions runs)
ci: lint test-all check-coverage build
    @echo "✅ CI pipeline passed"

# Pre-push checks (fast gate for lefthook)
pre-push: lint test
    @echo "✅ Pre-push checks passed"

# === Release Commands ===

# Validate GoReleaser config
release-check:
    goreleaser check

# Local snapshot build (no publish)
release-snapshot:
    goreleaser release --snapshot --clean

# === Setup Commands ===

# Setup development environment
setup:
    go mod download
    @echo "Installing staticcheck..."
    go install honnef.co/go/tools/cmd/staticcheck@latest
    @echo "Installing lefthook..."
    go install github.com/evilmartians/lefthook@latest
    lefthook install
    @echo "✅ Development environment ready"
    @echo ""
    @echo "Run 'just test' to verify setup"

# Symlink .agents/commands to .claude/commands for Claude compatibility
setup-claude:
    mkdir -p .claude
    ln -sfn ../.agents/commands .claude/commands
    @echo "Symlinked .agents/commands → .claude/commands"

# === Utility Commands ===

# Clean build artifacts
clean:
    rm -rf bin/ coverage.out coverage.html

# Show available commands
help:
    @just --list

# === MCP Commands ===

# Directory for MCP servers
mcp_dir := ".mcp-servers"
mcp_tui_dir := mcp_dir / "mcp-tui-test"

# Setup mcp-tui-test from source (not published to PyPI)
setup-mcp-tui:
    #!/usr/bin/env bash
    set -euo pipefail
    if [ ! -d "{{mcp_tui_dir}}" ]; then
        mkdir -p {{mcp_dir}}
        git clone https://github.com/GeorgePearse/mcp-tui-test.git {{mcp_tui_dir}}
        # Fix setuptools flat-layout error (multiple top-level modules)
        if ! grep -q 'py-modules' "{{mcp_tui_dir}}/pyproject.toml"; then
            sed -i '/\[project.scripts\]/i [tool.setuptools]\npy-modules = ["server"]\n' "{{mcp_tui_dir}}/pyproject.toml"
        fi
    fi
    echo "✅ mcp-tui-test ready"

# Start Amp with mcp-tui-test for TUI testing
amp-tui *args: setup-mcp-tui
    amp --mcp-config '{"tui-test": {"command": "uv", "args": ["run", "--directory", "{{justfile_directory()}}/{{mcp_tui_dir}}", "python", "server.py"]}}' {{args}}

# Run TUI integration tests via MCP (starts Amp with skill loaded)
test-tui: setup-mcp-tui
    @echo "Starting TUI integration test session..."
    @echo "Use: /skill tui-test then ask to 'run all TUI integration tests'"
    amp --mcp-config '{"tui-test": {"command": "uv", "args": ["run", "--directory", "{{justfile_directory()}}/{{mcp_tui_dir}}", "python", "server.py"]}}'
