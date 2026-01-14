# Carapace Shell Completions Research

**Date**: 2026-01-14  
**Issue**: fabbro-p47  
**Goal**: Evaluate carapace as an alternative to Cobra's built-in completions for fabbro, with focus on nushell support.

## Executive Summary

**Recommendation: ADOPT**

Carapace is a well-maintained, actively developed shell completion library that significantly expands fabbro's shell support (from 4 to 11 shells) with minimal code changes. The integration is straightforward—a single `carapace.Gen(rootCmd)` call enables multi-shell completions. While nushell support has occasional version-specific quirks, it is production-ready and actively maintained.

## Carapace Overview

### What is Carapace?

Carapace is a multi-shell completion library and binary developed by Ralf Steube. It consists of two components:

1. **carapace** (library): A Go library for defining completions in Cobra-based CLI apps
2. **carapace-bin** (binary): A standalone completer providing completions for 500+ tools

### Key Features

| Feature | Description |
|---------|-------------|
| **Multi-shell support** | 11 shells: bash, zsh, fish, powershell, elvish, nushell, oil, ion, tcsh, xonsh, cmd |
| **Cobra integration** | Drop-in enhancement for Cobra commands |
| **Dynamic completions** | `ActionCallback` for runtime completion generation |
| **Caching** | Built-in disk cache for slow completions |
| **Concurrent batch** | Parallel completion generation |
| **Style/coloring** | Colored completions (zsh, elvish, powershell) |
| **Grouped tags** | Organize completions by category |
| **Bridge mode** | Can bridge completions from other frameworks |
| **Testing support** | `carapace.Test()` validates configuration at build time |

### Supported Shells

| Shell | Status | Notes |
|-------|--------|-------|
| Bash | ✅ Stable | Full support |
| Zsh | ✅ Stable | Full support with colored completions |
| Fish | ✅ Stable | Full support |
| PowerShell | ✅ Stable | Full support with colored completions |
| Elvish | ✅ Stable | Full support with colored completions |
| **Nushell** | ✅ Stable | Full support, actively maintained |
| Oil | ✅ Stable | Full support |
| Xonsh | ✅ Stable | Full support |
| Tcsh | ⚠️ Experimental | Basic support |
| Ion | ⚠️ Experimental | Basic support |
| Cmd (Windows) | ⚠️ Experimental | Requires clink |

## Nushell Support Assessment

### Current State

Nushell support in carapace is **production-ready**:

- First-class support since carapace 0.x
- Dedicated shell package: `internal/shell/nushell`
- Active maintenance with nushell version updates
- Used by major tools (kubectl, git completions via carapace-bin)

### Known Issues

1. **Version coupling**: Nushell's rapid development occasionally causes breaking changes. Recent example: nushell 0.104.0 → 0.105.1 required carapace updates due to `default` keyword changes.

2. **External completer setup**: Requires nushell configuration:
   ```nu
   # config.nu
   $env.config.completions.external = {
     enable: true
     completer: {|spans| carapace $spans.0 nushell ...$spans | from json }
   }
   ```

3. **Fallback behavior**: When carapace doesn't recognize a command, nushell may not fall back to file completions automatically.

### Nushell Integration Approach

For fabbro with carapace:

```nu
# User adds to config.nu
$env.config.completions.external.completer = {|spans|
  match $spans.0 {
    fabbro => { carapace fabbro nushell ...$spans | from json }
    _ => { null }  # fallback to default
  }
}
```

Or use carapace-bin for universal completions.

## Comparison: Carapace vs Cobra Completions

| Aspect | Cobra Built-in | Carapace |
|--------|---------------|----------|
| **Shells** | 4 (bash, zsh, fish, powershell) | 11 shells |
| **Nushell** | ❌ Not supported | ✅ Full support |
| **Code complexity** | More verbose | More concise |
| **Maintenance** | Part of Cobra (slower iteration) | Active, frequent releases |
| **Dependencies** | None (built-in) | +1 dependency |
| **Caching** | Manual | Built-in |
| **Testing** | Manual | Built-in `carapace.Test()` |
| **Coloring** | Limited | Full support (zsh, elvish, ps) |
| **Dynamic plugins** | Complex | `PreRun` hook |
| **Community** | Large (Cobra) | Growing (1.1k stars) |

### API Comparison

**Current fabbro (no completions defined):**
```go
// Would require with Cobra:
rootCmd.AddCommand(&cobra.Command{
    Use:   "completion [bash|zsh|fish|powershell]",
    Short: "Generate completion script",
    // ... ~50 lines of shell-specific code
})

// Per-flag completions:
cmd.RegisterFlagCompletionFunc("session", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    sessions, _ := session.List()
    return sessions, cobra.ShellCompDirectiveNoFileComp
})
```

**With Carapace:**
```go
import "github.com/carapace-sh/carapace"

func buildRootCmd() *cobra.Command {
    rootCmd := &cobra.Command{...}
    
    // Enable completions for all 11 shells
    carapace.Gen(rootCmd)
    
    return rootCmd
}

// Per-flag completions (in buildApplyCmd):
carapace.Gen(cmd).FlagCompletion(carapace.ActionMap{
    "session": carapace.ActionCallback(func(c carapace.Context) carapace.Action {
        sessions, _ := session.List()
        return carapace.ActionValues(sessions...)
    }),
})
```

## Integration Approach for Fabbro

### Phase 1: Basic Integration (Recommended First Step)

```go
// cmd/fabbro/main.go
import "github.com/carapace-sh/carapace"

func buildRootCmd(stdin io.Reader, stdout io.Writer) *cobra.Command {
    rootCmd := &cobra.Command{
        Use:     "fabbro",
        Short:   "A code review annotation tool",
        Version: version,
    }

    rootCmd.AddCommand(buildInitCmd(stdout))
    rootCmd.AddCommand(buildReviewCmd(stdin, stdout))
    rootCmd.AddCommand(buildApplyCmd(stdout))

    // Enable carapace completions
    carapace.Gen(rootCmd)

    return rootCmd
}
```

### Phase 2: Add Custom Completions

```go
func buildApplyCmd(stdout io.Writer) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "apply [session-id]",
        Short: "Apply annotations from a session",
        Args:  cobra.ExactArgs(1),
        // ...
    }

    // Complete session IDs
    carapace.Gen(cmd).PositionalCompletion(
        carapace.ActionCallback(func(c carapace.Context) carapace.Action {
            sessions, err := session.ListIDs()
            if err != nil {
                return carapace.ActionMessage("failed to list sessions: " + err.Error())
            }
            return carapace.ActionValues(sessions...)
        }),
    )

    return cmd
}

func buildReviewCmd(stdin io.Reader, stdout io.Writer) *cobra.Command {
    cmd := &cobra.Command{...}

    // Complete files for positional argument
    carapace.Gen(cmd).PositionalCompletion(
        carapace.ActionFiles(),
    )

    return cmd
}
```

### Phase 3: Add to go.mod

```bash
go get github.com/carapace-sh/carapace@latest
```

Current version: v1.11.0 (Dec 2025)

## Maintenance Burden Assessment

### Pros

1. **Minimal code changes**: Single `Gen()` call enables completions
2. **Active development**: 285 releases, last update Dec 2025
3. **Good documentation**: Comprehensive docs at carapace.sh
4. **Test support**: `carapace.Test(rootCmd)` catches issues at build time
5. **Single source**: Define completions once, works everywhere

### Cons

1. **New dependency**: Adds ~2MB to binary (with third_party vendored code)
2. **Version tracking**: May need updates when nushell releases break API
3. **Learning curve**: New API (though similar to Cobra)

### Dependency Analysis

```
github.com/carapace-sh/carapace v1.11.0
└── github.com/spf13/cobra (already have)
└── github.com/spf13/pflag (already have)
└── internal vendored packages (no external deps)
```

Carapace vendors its dependencies, minimizing supply chain risk.

## Recommendation

### Decision: **ADOPT**

**Rationale:**

1. **Nushell support is the killer feature**: Fabbro targets developers who use modern shells. Nushell is growing in popularity, and Cobra will likely never support it.

2. **Minimal integration cost**: Adding carapace requires ~10 lines of code changes.

3. **Future-proof**: As new shells emerge, carapace will likely support them before Cobra.

4. **Better developer experience**: Colored completions, caching, and grouped tags improve UX.

5. **Active maintenance**: More frequent updates than Cobra's completion code.

### Implementation Plan

1. **Add carapace dependency** (`go get github.com/carapace-sh/carapace`)
2. **Add `carapace.Gen(rootCmd)`** to enable basic completions
3. **Add session ID completion** for `fabbro apply` command
4. **Add file completion** for `fabbro review` command
5. **Update README** with nushell setup instructions
6. **Add carapace.Test()** to main_test.go

### Documentation Updates Needed

### Existing Dependencies (from go.mod)

Fabbro already has `mattn/go-isatty v0.0.20` as an indirect dependency (via Bubble Tea). This provides cross-platform TTY detection if needed for completion context awareness.

### User-Facing Completion Setup

After implementation, users will run:

```bash
# Generate and source completions
source <(fabbro _carapace bash)   # or zsh, fish, etc.

# Or add to shell config for persistence
fabbro _carapace bash >> ~/.bashrc
```

Update README.md shell completion section:

```markdown
## Shell Completion

Fabbro supports completions for 11 shells via carapace.

### Bash/Zsh/Fish/PowerShell
source <(fabbro _carapace)

### Nushell
Add to config.nu:
$env.config.completions.external.completer = {|spans|
  carapace $spans.0 nushell ...$spans | from json
}
```

## References

- [Carapace Library](https://github.com/carapace-sh/carapace) - 1.1k stars
- [Carapace Documentation](https://carapace.sh)
- [Carapace Go Docs](https://pkg.go.dev/github.com/carapace-sh/carapace)
- [Cobra Freeze Proposal](https://github.com/spf13/cobra/issues/2019) - Discussion on Cobra+Carapace integration
- [Nushell External Completers](https://www.nushell.sh/cookbook/external_completers.html)
