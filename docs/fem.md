# FEM (Fabbro Editing Markup) Syntax

FEM is a lightweight inline markup syntax for annotating text. fabbro uses FEM to embed review comments directly in source files.

## Comment Annotation

The primary annotation type is the comment:

```
{>> comment text <<}
```

Comments can be placed at the end of any line:

```go
func main() {
    fmt.Println("Hello") {>> Consider using log package <<}
}
```

## All Annotation Types

| Syntax | Type | Description |
|--------|------|-------------|
| `{>> text <<}` | comment | General comment |
| `{-- text --}` | delete | Mark for deletion |
| `{?? text ??}` | question | Ask a question |
| `{!! text !!}` | expand | Request more detail |
| `{== text ==}` | keep | Mark as good/keep |
| `{~~ text ~~}` | unclear | Mark as unclear |

## Syntax Rules

- Annotations start with opening marker and end with closing marker
- Whitespace inside the markers is trimmed
- Annotations should be placed at the end of lines, after the code
- Multiple annotations on the same line ARE supported

## Parsing

fabbro extracts annotations using regex pattern matching:

```
\{>>\s*(.*?)\s*<<\}
```

The parser returns:
- Line number (1-indexed)
- Annotation type
- The annotation text

## Parser Limitations

**Single-line only**: Annotations must be fully contained on a single line. Multi-line annotations are not supported.

**Unbalanced markers**: If an opening marker has no closing marker (or vice versa), it is left in the content unchanged.

```
text {>> unbalanced marker     → preserved as-is
text <<} orphan close          → preserved as-is
```

**Nested markers**: Nesting annotations is undefined behavior. The parser matches from the first opening marker to the first closing marker, which may produce unexpected results:

```
{>> outer {>> inner <<} still <<}
→ extracts "outer {>> inner" and leaves "still <<}" in content
```

**Empty annotations**: Empty annotations (`{>><<}`) are valid and produce an annotation with empty text.

## References

FEM is based on [CriticMarkup](https://criticmarkup.com/) with adaptations for code review workflows.
