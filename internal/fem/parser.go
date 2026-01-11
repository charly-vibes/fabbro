package fem

import (
	"regexp"
	"strings"
)

type Annotation struct {
	Type string
	Text string
	Line int
}

var commentPattern = regexp.MustCompile(`\{>>\s*(.*?)\s*<<\}`)

func Parse(content string) ([]Annotation, string, error) {
	lines := strings.Split(content, "\n")
	var annotations []Annotation
	var cleanLines []string

	for i, line := range lines {
		matches := commentPattern.FindAllStringSubmatch(line, -1)
		cleanLine := commentPattern.ReplaceAllString(line, "")

		for _, match := range matches {
			if len(match) >= 2 {
				annotations = append(annotations, Annotation{
					Type: "comment",
					Text: match[1],
					Line: i + 1,
				})
			}
		}

		cleanLines = append(cleanLines, cleanLine)
	}

	cleanContent := strings.Join(cleanLines, "\n")
	return annotations, cleanContent, nil
}
