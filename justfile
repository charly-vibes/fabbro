# fabbro justfile

# Symlink .agents/commands to .claude/commands for Claude compatibility
setup-claude:
    mkdir -p .claude
    ln -sfn ../".agents/commands" .claude/commands
    @echo "Symlinked .agents/commands → .claude/commands"
