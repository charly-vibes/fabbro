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

// annotationTypes defines the processing order (deterministic, not map iteration).
var annotationTypes = []string{"comment", "delete", "question", "expand", "keep", "unclear", "change"}

// patterns maps annotation type to its regex pattern.
var patterns = map[string]*regexp.Regexp{
	"comment":  regexp.MustCompile(`\{>>\s*(.*?)\s*<<\}`),
	"delete":   regexp.MustCompile(`\{--\s*(.*?)\s*--\}`),
	"question": regexp.MustCompile(`\{\?\?\s*(.*?)\s*\?\?\}`),
	"expand":   regexp.MustCompile(`\{!!\s*(.*?)\s*!!\}`),
	"keep":     regexp.MustCompile(`\{==\s*(.*?)\s*==\}`),
	"unclear":  regexp.MustCompile(`\{~~\s*(.*?)\s*~~\}`),
	"change":   regexp.MustCompile(`\{\+\+\s*(.*?)\s*\+\+\}`),
}

// openingMarkers are the opening delimiters for all annotation types.
// Used to detect nested markers (which are invalid).
var openingMarkers = []string{"{>>", "{--", "{??", "{!!", "{==", "{~~", "{++"}

// containsNestedMarker returns true if text contains any opening marker.
func containsNestedMarker(text string) bool {
	for _, marker := range openingMarkers {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return false
}

func Parse(content string) ([]Annotation, string, error) {
	lines := strings.Split(content, "\n")
	var annotations []Annotation
	var cleanLines []string

	for i, line := range lines {
		cleanLine := line

		for _, annotationType := range annotationTypes {
			pattern := patterns[annotationType]
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
