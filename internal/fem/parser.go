package fem

import (
	"fmt"
	"regexp"
	"strings"
)

type Annotation struct {
	Type      string `json:"type"`
	Text      string `json:"text"`
	StartLine int    `json:"startLine"`
	EndLine   int    `json:"endLine"`
}

// blockDeleteOpen matches a line that is only a block delete opener: {-- text --}
var blockDeleteOpen = regexp.MustCompile(`^\s*\{--\s*(.*?)\s*--\}\s*$`)

// blockDeleteClose matches a line that is only a block delete closer: {--/--}
var blockDeleteClose = regexp.MustCompile(`^\s*\{--/--\}\s*$`)

// Sentinels for escaped braces during parsing.
const escapeOpenBrace = "\x00ESC_OPEN\x00"
const escapeCloseBrace = "\x00ESC_CLOSE\x00"

func Parse(content string) ([]Annotation, string, error) {
	// Replace escaped braces with sentinels before parsing.
	content = strings.ReplaceAll(content, `\{`, escapeOpenBrace)
	content = strings.ReplaceAll(content, `\}`, escapeCloseBrace)

	lines := strings.Split(content, "\n")

	// Pre-pass: find block delete regions and mark lines to skip for inline parsing.
	type blockDelete struct {
		openLine  int    // 0-indexed
		closeLine int    // 0-indexed
		text      string // reason text from the opening marker
	}
	var blocks []blockDelete
	skipLines := make(map[int]bool)

	for i := 0; i < len(lines); i++ {
		if m := blockDeleteOpen.FindStringSubmatch(lines[i]); m != nil {
			// Look ahead for a matching {--/--}
			for j := i + 1; j < len(lines); j++ {
				if blockDeleteClose.MatchString(lines[j]) {
					blocks = append(blocks, blockDelete{
						openLine:  i,
						closeLine: j,
						text:      m[1],
					})
					skipLines[i] = true
					skipLines[j] = true
					i = j // advance past the block
					break
				}
			}
		}
	}

	var annotations []Annotation

	// Emit block delete annotations first.
	for _, b := range blocks {
		text := b.text
		// Strip "DELETE:" or "DELETE" prefix if present
		text = strings.TrimPrefix(text, "DELETE:")
		text = strings.TrimSpace(text)
		if text == "" {
			text = strings.TrimSpace(b.text)
		}
		startLine := b.openLine + 2 // first content line (1-indexed)
		endLine := b.closeLine      // last content line (1-indexed, line before closer)
		annotations = append(annotations, Annotation{
			Type:      "delete",
			Text:      text,
			StartLine: startLine,
			EndLine:   endLine,
		})
	}

	// Multi-line annotation pre-pass: find annotations spanning multiple lines.
	type multiLineAnnotation struct {
		at        AnnotationType
		openLine  int    // 0-indexed
		closeLine int    // 0-indexed
		prefix    string // text before the opener on openLine
		suffix    string // text after the closer on closeLine
		text      string // annotation text (with newlines)
	}
	var multiLines []multiLineAnnotation

	for i := 0; i < len(lines); i++ {
		if skipLines[i] {
			continue
		}
		for _, at := range AnnotationTypes {
			openIdx := strings.Index(lines[i], at.Open)
			if openIdx < 0 {
				continue
			}
			// Check if the closer is on the same line — if so, inline pass handles it.
			afterOpen := lines[i][openIdx+len(at.Open):]
			if strings.Contains(afterOpen, at.Close) {
				continue
			}
			// Look forward for the closer on a subsequent line.
			for j := i + 1; j < len(lines); j++ {
				if skipLines[j] {
					continue
				}
				closeIdx := strings.Index(lines[j], at.Close)
				if closeIdx >= 0 {
					// Build the annotation text from partial lines.
					var parts []string
					parts = append(parts, lines[i][openIdx+len(at.Open):])
					for k := i + 1; k < j; k++ {
						parts = append(parts, lines[k])
					}
					parts = append(parts, lines[j][:closeIdx])
					text := strings.TrimSpace(strings.Join(parts, "\n"))

					multiLines = append(multiLines, multiLineAnnotation{
						at:        at,
						openLine:  i,
						closeLine: j,
						prefix:    lines[i][:openIdx],
						suffix:    lines[j][closeIdx+len(at.Close):],
						text:      text,
					})
					// Mark intermediate lines as skip.
					for k := i + 1; k < j; k++ {
						skipLines[k] = true
					}
					// Mark open and close lines for special handling.
					skipLines[i] = true
					skipLines[j] = true
					break
				}
			}
		}
	}

	// Emit multi-line annotations.
	for _, ml := range multiLines {
		annotations = append(annotations, Annotation{
			Type:      ml.at.Name,
			Text:      ml.text,
			StartLine: ml.openLine + 1,
			EndLine:   ml.closeLine + 1,
		})
	}

	// Inline pass: parse each non-skipped line for inline annotations.
	var cleanLines []string
	for i, line := range lines {
		if skipLines[i] {
			// Check if this is a multi-line open/close line with remaining content.
			for _, ml := range multiLines {
				if i == ml.openLine {
					cleanLines = append(cleanLines, ml.prefix)
					goto nextLine
				}
				if i == ml.closeLine {
					cleanLines = append(cleanLines, ml.suffix)
					goto nextLine
				}
			}
			cleanLines = append(cleanLines, "")
			continue
		}

	nextLine:
		cleanLine := line
		if len(cleanLines) > i {
			// Already set by multi-line handling above.
			cleanLine = cleanLines[i]
		}

		for _, at := range AnnotationTypes {
			pattern := patterns[at.Name]
			matches := pattern.FindAllStringSubmatch(cleanLine, -1)

			for _, match := range matches {
				if len(match) >= 2 {
					if containsNestedMarker(match[1]) {
						continue
					}
					cleanLine = pattern.ReplaceAllString(cleanLine, "")
					lineNum := i + 1
					annotations = append(annotations, Annotation{
						Type:      at.Name,
						Text:      match[1],
						StartLine: lineNum,
						EndLine:   lineNum,
					})
				}
			}
		}
		// Check for unclosed opening markers on this line.
		// After removing matched annotations, any remaining opener without its closer is unclosed.
		for _, at := range AnnotationTypes {
			if strings.Contains(cleanLine, at.Open) && !strings.Contains(cleanLine, at.Close) {
				return nil, "", fmt.Errorf("unclosed %s marker on line %d", at.Open, i+1)
			}
		}

		if len(cleanLines) > i {
			cleanLines[i] = cleanLine
		} else {
			cleanLines = append(cleanLines, cleanLine)
		}
	}

	cleanContent := strings.Join(cleanLines, "\n")
	// Restore escaped braces to literal characters in clean output.
	cleanContent = strings.ReplaceAll(cleanContent, escapeOpenBrace, "{")
	cleanContent = strings.ReplaceAll(cleanContent, escapeCloseBrace, "}")
	return annotations, cleanContent, nil
}
