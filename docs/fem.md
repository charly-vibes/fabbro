# FEM (First Editor Markup) Syntax

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

## Syntax Rules

- Comments start with `{>>` and end with `<<}`
- Whitespace inside the markers is trimmed
- Comments should be placed at the end of lines, after the code
- Multiple comments on the same line are not supported

## Parsing

fabbro extracts annotations using regex pattern matching:

```
\{>>\s*(.*?)\s*<<\}
```

The parser returns:
- Line number (1-indexed)
- Annotation type (`comment`)
- The comment text

## Future Annotation Types

The FEM specification includes additional annotation types not yet implemented in fabbro:

| Syntax | Type | Description |
|--------|------|-------------|
| `{>> text <<}` | comment | Add a comment |
| `{-- text --}` | delete | Mark for deletion |
| `{++ text ++}` | insert | Mark for insertion |
| `{~~ old ~> new ~~}` | replace | Mark for replacement |
| `{== text ==}` | highlight | Highlight text |

## References

FEM is based on [CriticMarkup](https://criticmarkup.com/) with adaptations for code review workflows.
