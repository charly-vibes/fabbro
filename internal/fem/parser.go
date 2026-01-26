package fem

import (
	"strings"
)

type Annotation struct {
	Type      string `json:"type"`
	Text      string `json:"text"`
	StartLine int    `json:"startLine"`
	EndLine   int    `json:"endLine"`
}

func Parse(content string) ([]Annotation, string, error) {
	lines := strings.Split(content, "\n")
	var annotations []Annotation
	var cleanLines []string

	for i, line := range lines {
		cleanLine := line

		for _, at := range AnnotationTypes {
			pattern := patterns[at.Name]
			matches := pattern.FindAllStringSubmatch(line, -1)

			for _, match := range matches {
				if len(match) >= 2 {
					// Skip annotations with nested markers (invalid syntax)
					if containsNestedMarker(match[1]) {
						continue
					}
					// Only remove from cleanLine if we're accepting this annotation
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

		cleanLines = append(cleanLines, cleanLine)
	}

	cleanContent := strings.Join(cleanLines, "\n")
	return annotations, cleanContent, nil
}
