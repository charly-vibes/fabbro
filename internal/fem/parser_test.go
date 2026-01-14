package fem

import (
	"testing"
)

func TestParse_ExtractsSingleComment(t *testing.T) {
	content := "Hello {>> this is a comment <<} world"

	annotations, clean, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(annotations))
	}

	if annotations[0].Text != "this is a comment" {
		t.Errorf("expected Text='this is a comment', got %q", annotations[0].Text)
	}

	if annotations[0].Type != "comment" {
		t.Errorf("expected Type='comment', got %q", annotations[0].Type)
	}

	if clean != "Hello  world" {
		t.Errorf("expected clean='Hello  world', got %q", clean)
	}
}

func TestParse_ExtractsMultipleCommentsOnDifferentLines(t *testing.T) {
	content := `Line one {>> first comment <<}
Line two
Line three {>> second comment <<}`

	annotations, _, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if len(annotations) != 2 {
		t.Fatalf("expected 2 annotations, got %d", len(annotations))
	}

	if annotations[0].StartLine != 1 {
		t.Errorf("expected first annotation on line 1, got %d", annotations[0].StartLine)
	}

	if annotations[1].StartLine != 3 {
		t.Errorf("expected second annotation on line 3, got %d", annotations[1].StartLine)
	}
}

func TestParse_ReturnsCleanContent(t *testing.T) {
	content := `First line {>> comment <<}
Second line`

	_, clean, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	expected := `First line 
Second line`
	if clean != expected {
		t.Errorf("expected clean=%q, got %q", expected, clean)
	}
}

func TestParse_HandlesNoAnnotations(t *testing.T) {
	content := "Just plain text\nwith no annotations"

	annotations, clean, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if len(annotations) != 0 {
		t.Errorf("expected 0 annotations, got %d", len(annotations))
	}

	if clean != content {
		t.Errorf("expected clean to equal original content")
	}
}

func TestParse_AllAnnotationTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType string
		wantText string
	}{
		{
			name:     "comment",
			input:    "text {>> comment here <<}",
			wantType: "comment",
			wantText: "comment here",
		},
		{
			name:     "delete",
			input:    "text {-- DELETE: reason --}",
			wantType: "delete",
			wantText: "DELETE: reason",
		},
		{
			name:     "question",
			input:    "text {?? Why this? ??}",
			wantType: "question",
			wantText: "Why this?",
		},
		{
			name:     "expand",
			input:    "text {!! EXPAND: more detail !!}",
			wantType: "expand",
			wantText: "EXPAND: more detail",
		},
		{
			name:     "keep",
			input:    "text {== KEEP: good section ==}",
			wantType: "keep",
			wantText: "KEEP: good section",
		},
		{
			name:     "unclear",
			input:    "text {~~ UNCLEAR: confusing ~~}",
			wantType: "unclear",
			wantText: "UNCLEAR: confusing",
		},
		{
			name:     "change",
			input:    "text {++ [line 1] -> new content ++}",
			wantType: "change",
			wantText: "[line 1] -> new content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			annotations, clean, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if len(annotations) != 1 {
				t.Fatalf("expected 1 annotation, got %d", len(annotations))
			}

			if annotations[0].Type != tt.wantType {
				t.Errorf("expected Type=%q, got %q", tt.wantType, annotations[0].Type)
			}

			if annotations[0].Text != tt.wantText {
				t.Errorf("expected Text=%q, got %q", tt.wantText, annotations[0].Text)
			}

			if clean != "text " {
				t.Errorf("expected clean='text ', got %q", clean)
			}
		})
	}
}

func TestParse_MixedAnnotationTypes(t *testing.T) {
	content := `First line {>> comment <<}
Second line {-- delete --}
Third line {?? question ??}`

	annotations, _, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if len(annotations) != 3 {
		t.Fatalf("expected 3 annotations, got %d", len(annotations))
	}

	expected := []struct {
		typ       string
		text      string
		startLine int
	}{
		{"comment", "comment", 1},
		{"delete", "delete", 2},
		{"question", "question", 3},
	}

	for i, exp := range expected {
		if annotations[i].Type != exp.typ {
			t.Errorf("annotation %d: expected Type=%q, got %q", i, exp.typ, annotations[i].Type)
		}
		if annotations[i].Text != exp.text {
			t.Errorf("annotation %d: expected Text=%q, got %q", i, exp.text, annotations[i].Text)
		}
		if annotations[i].StartLine != exp.startLine {
			t.Errorf("annotation %d: expected StartLine=%d, got %d", i, exp.startLine, annotations[i].StartLine)
		}
	}
}

// Edge case tests for parser limitations

func TestParse_UnbalancedOpenMarkerIsPreserved(t *testing.T) {
	content := "text with {>> unbalanced marker"

	annotations, clean, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	// Should find no annotations (marker is not closed)
	if len(annotations) != 0 {
		t.Errorf("expected 0 annotations for unbalanced marker, got %d", len(annotations))
	}

	// Unbalanced marker should remain in clean content
	if clean != content {
		t.Errorf("expected unbalanced marker preserved, got %q", clean)
	}
}

func TestParse_UnbalancedCloseMarkerIsPreserved(t *testing.T) {
	content := "text with <<} orphan close"

	annotations, clean, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if len(annotations) != 0 {
		t.Errorf("expected 0 annotations for orphan close, got %d", len(annotations))
	}

	if clean != content {
		t.Errorf("expected orphan close preserved, got %q", clean)
	}
}

func TestParse_MultipleAnnotationsOnSameLine(t *testing.T) {
	content := "text {>> first <<} middle {>> second <<} end"

	annotations, clean, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if len(annotations) != 2 {
		t.Fatalf("expected 2 annotations on same line, got %d", len(annotations))
	}

	// Both should be on line 1
	for i, ann := range annotations {
		if ann.StartLine != 1 {
			t.Errorf("annotation %d: expected line 1, got %d", i, ann.StartLine)
		}
	}

	// Clean should have both removed
	if clean != "text  middle  end" {
		t.Errorf("expected clean='text  middle  end', got %q", clean)
	}
}

func TestParse_NestedMarkersUndefinedBehavior(t *testing.T) {
	// Nested markers have undefined behavior - the non-greedy regex matches
	// from the first {>> to the first <<}, which includes the nested open marker
	content := "text {>> outer {>> inner <<} still outer <<} end"

	annotations, clean, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	// Documents current behavior: matches "outer {>> inner" (first open to first close)
	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation for nested, got %d", len(annotations))
	}

	// The match includes the nested {>> because regex is non-greedy to first <<}
	if annotations[0].Text != "outer {>> inner" {
		t.Errorf("expected 'outer {>> inner', got %q", annotations[0].Text)
	}

	// Remaining text includes orphaned close marker
	if clean != "text  still outer <<} end" {
		t.Errorf("expected 'text  still outer <<} end', got %q", clean)
	}
}

func TestParse_EmptyAnnotationText(t *testing.T) {
	content := "text {>><<} with empty"

	annotations, clean, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(annotations))
	}

	if annotations[0].Text != "" {
		t.Errorf("expected empty text, got %q", annotations[0].Text)
	}

	if clean != "text  with empty" {
		t.Errorf("expected clean='text  with empty', got %q", clean)
	}
}

func TestParse_WhitespaceOnlyAnnotation(t *testing.T) {
	content := "text {>>   <<} with spaces"

	annotations, _, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(annotations))
	}

	// Whitespace is trimmed by the regex
	if annotations[0].Text != "" {
		t.Errorf("expected empty text after trim, got %q", annotations[0].Text)
	}
}
