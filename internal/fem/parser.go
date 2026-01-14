package fem

import (
	"regexp"
	"strings"
)

type Annotation struct {
	Type      string `json:"type"`
	Text      string `json:"text"`
	StartLine int    `json:"startLine"`
	EndLine   int    `json:"endLine"`
}

var patterns = map[string]*regexp.Regexp{
	"comment":  regexp.MustCompile(`\{>>\s*(.*?)\s*<<\}`),
	"delete":   regexp.MustCompile(`\{--\s*(.*?)\s*--\}`),
	"question": regexp.MustCompile(`\{\?\?\s*(.*?)\s*\?\?\}`),
	"expand":   regexp.MustCompile(`\{!!\s*(.*?)\s*!!\}`),
	"keep":     regexp.MustCompile(`\{==\s*(.*?)\s*==\}`),
	"unclear":  regexp.MustCompile(`\{~~\s*(.*?)\s*~~\}`),
	"change":   regexp.MustCompile(`\{\+\+\s*(.*?)\s*\+\+\}`),
}

func Parse(content string) ([]Annotation, string, error) {
	lines := strings.Split(content, "\n")
	var annotations []Annotation
	var cleanLines []string

	for i, line := range lines {
		cleanLine := line

		for annotationType, pattern := range patterns {
			matches := pattern.FindAllStringSubmatch(line, -1)
			cleanLine = pattern.ReplaceAllString(cleanLine, "")

			for _, match := range matches {
				if len(match) >= 2 {
					lineNum := i + 1
					annotations = append(annotations, Annotation{
						Type:      annotationType,
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
