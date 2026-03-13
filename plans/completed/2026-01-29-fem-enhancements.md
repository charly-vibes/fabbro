# FEM Parser Enhancements Plan

## Overview

Extend FEM parser to support block markers, escaping, error reporting, and additional annotation types.

**Spec**: `specs/06_fem_markup.feature`

## Current State

- Inline markers work: `{>> <<}`, `{-- --}`, `{?? ??}`, `{!! !!}`, `{== ==}`, `{~~ ~~}`, `{++ ++}`
- No block delete syntax (`{-- --}...{--/--}`)
- No escaping support
- No syntax error reporting
- No emphasize (`{** **}`) or section (`{## ##}`) markers

## Desired End State

- Block delete spans multiple lines
- Escaped markers `\{>>` are not parsed
- Unclosed markers produce parse errors with line numbers
- Emphasize and section annotation types supported
- Nested braces in annotation text work correctly

---

## Phase 1: Block Delete Syntax (~1.5 hours)

**Spec scenario**: "Block delete with reason"

### Syntax

```
{-- DELETE: Too verbose --}
Delete this line.
And this line too.
{--/--}
```

The opening marker contains the annotation text. Lines between markers are the affected range.

### Changes Required

**File: internal/fem/parser.go**

Add block parsing:
```go
type blockState struct {
    inBlock     bool
    blockType   string
    startLine   int
    text        string
}

func (p *Parser) parseBlocks(lines []string) []Annotation {
    var annotations []Annotation
    var block blockState
    
    for i, line := range lines {
        if !block.inBlock {
            // Check for block start: {-- DELETE: text --}
            if match := blockStartPattern.FindStringSubmatch(line); match != nil {
                block.inBlock = true
                block.blockType = "delete"
                block.startLine = i + 1  // next line is start of content
                block.text = match[1]
            }
        } else {
            // Check for block end: {--/--}
            if blockEndPattern.MatchString(line) {
                annotations = append(annotations, Annotation{
                    Type:      block.blockType,
                    StartLine: block.startLine,
                    EndLine:   i - 1,
                    Text:      block.text,
                })
                block.inBlock = false
            }
        }
    }
    return annotations
}
```

### Success Criteria

- [ ] Block delete spans correct lines
- [ ] Annotation text comes from opening marker
- [ ] Works with existing inline annotations

---

## Phase 2: Escape Sequences (~45 min)

**Spec scenario**: "Escaped markup is not parsed"

### Syntax

```
To add a comment, use \{>> comment <<\}
```

### Changes Required

**File: internal/fem/parser.go**

Pre-process escape sequences:
```go
const escapePlaceholder = "\x00ESCAPED_BRACE\x00"

func (p *Parser) preprocess(content string) string {
    // Replace \{ with placeholder
    content = strings.ReplaceAll(content, `\{`, escapePlaceholder+"L")
    content = strings.ReplaceAll(content, `\}`, escapePlaceholder+"R")
    return content
}

func (p *Parser) postprocess(content string) string {
    // Restore escaped braces as literal characters
    content = strings.ReplaceAll(content, escapePlaceholder+"L", "{")
    content = strings.ReplaceAll(content, escapePlaceholder+"R", "}")
    return content
}
```

### Success Criteria

- [ ] `\{>>` is not parsed as annotation start
- [ ] Escaped braces appear as literal `{` `}` in output
- [ ] Existing annotations still work

---

## Phase 3: Parse Error Reporting (~1 hour)

**Spec scenarios**: "Applying session with malformed FEM", "Unclosed annotation marker"

### Changes Required

**File: internal/fem/parser.go**

Add error tracking:
```go
type ParseError struct {
    Line    int
    Column  int
    Message string
}

type ParseResult struct {
    Annotations []Annotation
    Errors      []ParseError
}

func (p *Parser) ParseWithErrors(content string) ParseResult {
    // Track unclosed markers (both inline and block)
    // Return errors with line numbers
}
```

**Error cases to detect:**
1. Unclosed inline marker: `{>>` without `<<}`
2. Unclosed block marker: `{-- ... --}` without `{--/--}`
3. Mismatched markers: `{>>` closed with `--}`
4. Orphaned block end: `{--/--}` without opening block

**File: cmd/fabbro/main.go**

On parse errors, output:
```
Error: malformed FEM syntax
  line 15: unclosed annotation marker {>>
  line 42: unclosed block delete (opened at line 38)
```

### Success Criteria

- [ ] Unclosed inline markers produce errors with line numbers
- [ ] Unclosed block markers produce errors referencing opening line
- [ ] Mismatched markers reported
- [ ] `fabbro apply` exits with code 1 on parse errors

---

## Phase 4: Emphasize and Section Markers (~30 min)

**Spec scenarios**: "Emphasize syntax", "Section annotation"

### Syntax

```
{** EMPHASIZE: Key takeaway **}
{## SECTION: Needs rewriting ##}
```

### Changes Required

**File: internal/fem/parser.go**

Add patterns:
```go
"emphasize": regexp.MustCompile(`\{\*\*\s*(.*?)\s*\*\*\}`),
"section":   regexp.MustCompile(`\{##\s*(.*?)\s*##\}`),
```

**File: internal/tui/tui.go**

Add to annotation types and markers map.

### Success Criteria

- [ ] Emphasize annotation parsed correctly
- [ ] Section annotation parsed correctly
- [ ] TUI can create both types

---

## Phase 5: Nested Braces in Annotation Text (~45 min)

**Spec scenario**: "Nested braces in annotation text"

### Problem

Current regex `\{>>\s*(.*?)\s*<<\}` fails with:
```
{>> Use {curly braces} in the output <<}
```

Regex-based solutions (`.+?`, lookahead, etc.) cannot reliably handle arbitrary nesting because regex cannot count brace depth.

### Solution

**Use a custom parser with brace counting**, not regex:

```go
func parseAnnotation(content string, startIdx int) (text string, endIdx int, ok bool) {
    depth := 0
    var textStart, textEnd int
    
    // Find opening marker and track where text starts
    // Count { and } to find matching closing marker
    for i := startIdx; i < len(content); i++ {
        switch content[i] {
        case '{':
            depth++
        case '}':
            depth--
            if depth == 0 {
                // Found matching close - check for marker suffix
            }
        }
    }
    // ...
}
```

**Alternative approach**: Greedy match to last `<<}` on same line, then validate:
```go
// Match greedily to last <<} on line
pattern := regexp.MustCompile(`\{>>\s*(.+)\s*<<\}`)
```

This works for single-line annotations but fails for multi-line. Recommend custom parser.

### Success Criteria

- [ ] `{>> Use {curly braces} in output <<}` parses correctly
- [ ] Annotation text is `Use {curly braces} in output`
- [ ] Nested markers like `{>> Use {>> inner <<} in output <<}` handled gracefully (error or outer-only)

---

## Phase 6: Multi-line Annotation Text (~30 min)

**Spec scenario**: "Newlines in annotation text"

### Problem

Current parser works line-by-line, can't handle:
```
{>> This is a multi-line
annotation that spans
multiple lines <<}
```

### Solution

Parse full content as string, not line-by-line:
```go
// Use (?s) flag for single-line mode (. matches newlines)
pattern := regexp.MustCompile(`(?s)\{>>\s*(.*?)\s*<<\}`)
```

Track line numbers via byte offset.

### Success Criteria

- [ ] Multi-line annotations parsed correctly
- [ ] Newlines preserved in annotation text
- [ ] Line number reported correctly (first line of marker)

---

## Summary

| Phase | Deliverable | Time |
|-------|-------------|------|
| 1 | Block delete syntax | 1.5h |
| 2 | Escape sequences | 45m |
| 3 | Parse error reporting (inline + block) | 1.25h |
| 4 | Emphasize/section markers | 30m |
| 5 | Nested braces handling (custom parser) | 45m |
| 6 | Multi-line annotation text | 30m |

**Total: ~5.25 hours**

## Dependencies

- Phase 1-6 are independent
- Phase 3 (error reporting) affects `fabbro apply` command
