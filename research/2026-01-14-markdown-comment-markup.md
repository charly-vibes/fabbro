# Markdown Comment Markup: Alternative to FEM

**Date**: 2026-01-14  
**Status**: Exploration  
**Priority**: Low (P3)  
**Epic**: fabbro-ejx  
**Author**: AI-assisted

## Context: How Fabbro Works

Fabbro reviews **any text content** (plans, diffs, markdown docs, code) piped via stdin:

```bash
cat plan.md | fabbro review --stdin          # Review a plan
git diff HEAD~1 | fabbro review --stdin      # Review a diff
```

The primary use case is **reviewing LLM-generated plans and markdown documents**, not necessarily source code files. This changes the calculus significantly—HTML comments are valid in markdown!

### Current Flow

1. Human uses TUI to add FEM annotations → saved as `.fem` session file
2. `fabbro apply <id> --json` extracts annotations as structured JSON
3. AI consumes the JSON to understand feedback

### Future Flow (AI-generated reviews)

1. AI reviews content and produces annotated output
2. Fabbro parses annotations from AI output
3. Human reviews in TUI, AI consumes structured feedback

**Key question**: Which annotation format do AI models produce more reliably?

## Hypothesis

Embedding fabbro annotations inside HTML comments (`<!-- ... -->`) instead of custom FEM syntax (`{>> ... <<}`) could:

1. Produce valid markdown output (better tooling compatibility)
2. Improve AI model comprehension (trained on markdown, not CriticMarkup)
3. Enable A/B testing across models without retraining

## Current FEM Syntax

```markdown
## Phase 1: Setup

Create the project structure. {>> This timeline seems aggressive <<}

### Dependencies
- cobra for CLI {?? Why not urfave/cli? ??}
```

## Proposed Markdown Comment Syntax

### Option A: Inline Comments

```markdown
## Phase 1: Setup

Create the project structure. <!-- fabbro:comment: This timeline seems aggressive -->

### Dependencies
- cobra for CLI <!-- fabbro:question: Why not urfave/cli? -->
```

### Option B: Block Range Comments

For multi-line annotations, specify a range:

```markdown
<!-- fabbro:comment:lines=5-10: This entire section needs more detail -->

## Phase 2: Implementation

Step 1: Do the thing
Step 2: Do more things
Step 3: Profit
```

### Option C: JSON-in-Comments (Structured)

```markdown
Create the project structure. <!-- {"fabbro":"comment","text":"Timeline seems aggressive"} -->
```

Or for blocks:

```markdown
<!-- fabbro:block
{
  "type": "comment",
  "range": {"start": 5, "end": 10},
  "text": "This entire section needs more detail"
}
-->
```

## Annotation Type Mapping

| FEM Syntax | Markdown Comment |
|------------|------------------|
| `{>> text <<}` | `<!-- fabbro:comment: text -->` |
| `{-- text --}` | `<!-- fabbro:delete: text -->` |
| `{?? text ??}` | `<!-- fabbro:question: text -->` |
| `{!! text !!}` | `<!-- fabbro:expand: text -->` |
| `{== text ==}` | `<!-- fabbro:keep: text -->` |
| `{~~ text ~~}` | `<!-- fabbro:unclear: text -->` |

## Trade-offs

### Pros

- **Valid Markdown**: Renders correctly in GitHub, editors, etc.
- **AI Familiarity**: Models see HTML comments constantly in training data
- **Tooling**: Syntax highlighters, linters won't complain
- **Extensible**: JSON variant allows arbitrary metadata

### Cons

- **Verbosity**: `<!-- fabbro:comment: -->` is longer than `{>> <<}`
- **Parsing Complexity**: HTML comment parsing is more nuanced
- **Language Conflicts**: HTML comments in actual HTML/JSX files could collide
- **Visual Noise**: More characters = harder to scan

## Quantifying Model Performance

### Test Harness Design

1. **Generate Test Cases**: Create N annotated files with known ground truth
2. **Input Variants**: Same content with FEM vs Markdown-comment markup
3. **Model Tasks**:
   - Extract all annotations (precision/recall)
   - Generate annotations for a diff (quality scoring)
   - Respond to annotations appropriately (semantic eval)

### Metrics

| Metric | Description |
|--------|-------------|
| **Extraction Accuracy** | % of annotations correctly parsed from output |
| **Position Accuracy** | % of annotations attached to correct line |
| **Type Accuracy** | % of annotation types correctly identified |
| **Generation Quality** | Human eval / LLM-as-judge on generated reviews |
| **Format Adherence** | % of outputs using valid syntax |

### Test Case Categories

1. **Simple**: Single annotation per file
2. **Dense**: Multiple annotations, various types
3. **Edge Cases**: Nested quotes, special chars, multi-line
4. **Language Variety**: Go, Python, JS, Rust, etc.
5. **Adversarial**: Intentionally confusing patterns

### Evaluation Script Sketch

```bash
#!/bin/bash
# eval-markup-formats.sh

MODELS=("gpt-4o" "claude-sonnet" "gemini-pro")
FORMATS=("fem" "markdown-comment")
TESTCASES="test/markup-eval/*.md"

for model in "${MODELS[@]}"; do
  for format in "${FORMATS[@]}"; do
    fabbro eval --model "$model" --format "$format" --cases "$TESTCASES" \
      >> results/"$model-$format.json"
  done
done

# Aggregate results
fabbro eval-report results/
```

## Next Steps

Tracked in beads (all P3/low priority):

1. [ ] `fabbro-gt1` Build small test corpus (20-50 cases)
2. [ ] `fabbro-nts` Token count analysis
3. [ ] `fabbro-kvs` Create evaluation harness
4. [ ] Run comparative tests across 3+ models
5. [ ] `fabbro-ejx` Implement pluggable parser architecture

## Scope: Markdown Files Only

This proposal applies to **markdown-based content**:
- LLM-generated plans (`plans/*.md`)
- Design documents
- Git diffs (which are plain text, comments work inline)

For **source code files** (.go, .py, .rs), HTML comments are invalid syntax. Options:
1. Keep FEM for code files, markdown-comments for .md files
2. Use language-specific comment syntax: `// fabbro:comment: ...` for Go
3. Out of scope—fabbro primary use case is markdown review

## Open Questions

- Should we support **both** formats? (Parse either, output configurable)
- How do we handle markdown that already contains HTML comments?
- Is the `fabbro:` prefix sufficient namespacing?
- Should block annotations use line numbers or markers?
- For code files: use FEM, language comments, or declare out of scope?

## Related

- [FEM Spec](../docs/fem.md)
- [CriticMarkup](https://criticmarkup.com/)
