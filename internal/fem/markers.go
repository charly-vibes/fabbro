package fem

import (
	"regexp"
	"strings"
)

// AnnotationType defines a single annotation type with its delimiters and prompt.
type AnnotationType struct {
	Name   string
	Open   string // Opening delimiter without spaces, e.g. "{>>"
	Close  string // Closing delimiter without spaces, e.g. "<<}"
	Prompt string
}

// AnnotationTypes is the single source of truth for all FEM annotation types.
// Order matters for deterministic parsing.
var AnnotationTypes = []AnnotationType{
	{Name: "comment", Open: "{>>", Close: "<<}", Prompt: "Comment:"},
	{Name: "delete", Open: "{--", Close: "--}", Prompt: "Reason for deletion:"},
	{Name: "question", Open: "{??", Close: "??}", Prompt: "Question:"},
	{Name: "expand", Open: "{!!", Close: "!!}", Prompt: "What to expand:"},
	{Name: "keep", Open: "{==", Close: "==}", Prompt: "Reason to keep:"},
	{Name: "unclear", Open: "{~~", Close: "~~}", Prompt: "What's unclear:"},
	{Name: "change", Open: "{++", Close: "++}", Prompt: "Replacement text:"},
}

// Markers maps annotation type to opening and closing delimiters (with spaces for rendering).
// Derived from AnnotationTypes for backward compatibility.
var Markers = func() map[string][2]string {
	m := make(map[string][2]string)
	for _, at := range AnnotationTypes {
		m[at.Name] = [2]string{at.Open + " ", " " + at.Close}
	}
	return m
}()

// Prompts maps annotation type to input prompt text.
// Derived from AnnotationTypes for backward compatibility.
var Prompts = func() map[string]string {
	m := make(map[string]string)
	for _, at := range AnnotationTypes {
		m[at.Name] = at.Prompt
	}
	return m
}()

// patterns maps annotation type to its compiled regex pattern.
// Generated from AnnotationTypes.
var patterns = func() map[string]*regexp.Regexp {
	m := make(map[string]*regexp.Regexp)
	for _, at := range AnnotationTypes {
		open := regexp.QuoteMeta(at.Open)
		close := regexp.QuoteMeta(at.Close)
		pattern := open + `\s*(.*?)\s*` + close
		m[at.Name] = regexp.MustCompile(pattern)
	}
	return m
}()

// openingMarkers contains all opening delimiters for nested marker detection.
// Generated from AnnotationTypes.
var openingMarkers = func() []string {
	markers := make([]string, len(AnnotationTypes))
	for i, at := range AnnotationTypes {
		markers[i] = at.Open
	}
	return markers
}()

// containsNestedMarker returns true if text contains any opening marker.
func containsNestedMarker(text string) bool {
	for _, marker := range openingMarkers {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return false
}

// ValidAnnotationType returns true if typ is a known annotation type.
func ValidAnnotationType(typ string) bool {
	_, ok := Markers[typ]
	return ok
}
