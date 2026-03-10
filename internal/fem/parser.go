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

	// Inline pass: parse each non-skipped line for inline annotations.
	var cleanLines []string
	for i, line := range lines {
		if skipLines[i] {
			cleanLines = append(cleanLines, "")
			continue
		}

		cleanLine := line
		for _, at := range AnnotationTypes {
			pattern := patterns[at.Name]
			matches := pattern.FindAllStringSubmatch(line, -1)

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

		cleanLines = append(cleanLines, cleanLine)
	}

	cleanContent := strings.Join(cleanLines, "\n")
	// Restore escaped braces to literal characters in clean output.
	cleanContent = strings.ReplaceAll(cleanContent, escapeOpenBrace, "{")
	cleanContent = strings.ReplaceAll(cleanContent, escapeCloseBrace, "}")
	return annotations, cleanContent, nil
}
