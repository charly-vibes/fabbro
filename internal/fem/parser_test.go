package fem

import (
	"strings"
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
			wantText: "-> new content",
		},
		{
			name:     "emphasize",
			input:    "text {** EMPHASIZE: This is the key takeaway **}",
			wantType: "emphasize",
			wantText: "EMPHASIZE: This is the key takeaway",
		},
		{
			name:     "section",
			input:    "text {## SECTION: This entire section needs rewriting ##}",
			wantType: "section",
			wantText: "SECTION: This entire section needs rewriting",
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

func TestParse_UnclosedMarkerReturnsError(t *testing.T) {
	content := "Content here. {>> This annotation is not closed"

	_, _, err := Parse(content)
	if err == nil {
		t.Fatal("expected error for unclosed marker, got nil")
	}

	// Error should indicate the line number
	errMsg := err.Error()
	if !strings.Contains(errMsg, "line 1") {
		t.Errorf("expected error to mention line 1, got %q", errMsg)
	}
	if !strings.Contains(errMsg, "unclosed") || !strings.Contains(errMsg, "{>>") {
		t.Errorf("expected error to mention unclosed marker type, got %q", errMsg)
	}
}

func TestParse_UnclosedMarkerOnLaterLine(t *testing.T) {
	content := `Line one is fine.
Line two is fine.
Line three has {-- unclosed delete`

	_, _, err := Parse(content)
	if err == nil {
		t.Fatal("expected error for unclosed marker, got nil")
	}

	if !strings.Contains(err.Error(), "line 3") {
		t.Errorf("expected error to mention line 3, got %q", err.Error())
	}
}

func TestParse_UnclosedMarkerWithValidAnnotations(t *testing.T) {
	content := `Valid line {>> good comment <<}
Bad line {?? unclosed question`

	_, _, err := Parse(content)
	if err == nil {
		t.Fatal("expected error for unclosed marker, got nil")
	}

	if !strings.Contains(err.Error(), "line 2") {
		t.Errorf("expected error to mention line 2, got %q", err.Error())
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

func TestParse_NestedMarkersAreSkipped(t *testing.T) {
	// Nested markers are detected and skipped - the line is preserved unchanged
	content := "text {>> outer {>> inner <<} still outer <<} end"

	annotations, clean, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	// Nested annotation is skipped entirely
	if len(annotations) != 0 {
		t.Fatalf("expected 0 annotations for nested markers, got %d", len(annotations))
	}

	// Original line is preserved unchanged (no corruption)
	if clean != content {
		t.Errorf("expected line preserved unchanged, got %q", clean)
	}
}

func TestParse_NestedDifferentTypes_InnerExtracted(t *testing.T) {
	// When different types are nested, the outer is skipped but inner is extracted
	content := "text {>> outer {-- delete inside --} comment <<} end"

	annotations, clean, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	// The outer comment contains {-- so it's skipped
	// But the inner delete annotation is valid and extracted
	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation (inner delete), got %d", len(annotations))
	}

	if annotations[0].Type != "delete" {
		t.Errorf("expected type 'delete', got %q", annotations[0].Type)
	}

	if annotations[0].Text != "delete inside" {
		t.Errorf("expected text 'delete inside', got %q", annotations[0].Text)
	}

	// The delete annotation is removed, but the malformed outer comment remains
	expected := "text {>> outer  comment <<} end"
	if clean != expected {
		t.Errorf("expected %q, got %q", expected, clean)
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

func TestParse_BlockDeleteWithReason(t *testing.T) {
	content := `Keep this line.
{-- DELETE: Too verbose --}
Delete this line.
And this line too.
{--/--}
Keep this line as well.`

	annotations, clean, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	// Should find one block delete annotation
	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(annotations))
	}

	ann := annotations[0]
	if ann.Type != "delete" {
		t.Errorf("expected type 'delete', got %q", ann.Type)
	}
	if ann.Text != "Too verbose" {
		t.Errorf("expected text 'Too verbose', got %q", ann.Text)
	}
	if ann.StartLine != 3 {
		t.Errorf("expected StartLine=3, got %d", ann.StartLine)
	}
	if ann.EndLine != 4 {
		t.Errorf("expected EndLine=4, got %d", ann.EndLine)
	}

	// Clean content: opening marker line and closing marker line removed,
	// content lines between them preserved
	expected := `Keep this line.

Delete this line.
And this line too.

Keep this line as well.`
	if clean != expected {
		t.Errorf("expected clean=\n%q\ngot\n%q", expected, clean)
	}
}

func TestParse_BlockDeleteWithoutReason(t *testing.T) {
	content := `Before.
{-- DELETE --}
Middle line.
{--/--}
After.`

	annotations, _, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(annotations))
	}

	ann := annotations[0]
	if ann.Type != "delete" {
		t.Errorf("expected type 'delete', got %q", ann.Type)
	}
	if ann.Text != "DELETE" {
		t.Errorf("expected text 'DELETE', got %q", ann.Text)
	}
	if ann.StartLine != 3 {
		t.Errorf("expected StartLine=3, got %d", ann.StartLine)
	}
	if ann.EndLine != 3 {
		t.Errorf("expected EndLine=3, got %d", ann.EndLine)
	}
}

func TestParse_BlockDeleteWithInlineAnnotations(t *testing.T) {
	content := `Start.
{-- DELETE: Remove this --}
Line one.
Line two. {>> also has a comment <<}
{--/--}
End.`

	annotations, _, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	// Should find block delete + inline comment
	if len(annotations) != 2 {
		t.Fatalf("expected 2 annotations, got %d", len(annotations))
	}

	// Block delete should come first (processed in pre-pass)
	if annotations[0].Type != "delete" {
		t.Errorf("expected first annotation type 'delete', got %q", annotations[0].Type)
	}
	if annotations[0].StartLine != 3 {
		t.Errorf("expected delete StartLine=3, got %d", annotations[0].StartLine)
	}
	if annotations[0].EndLine != 4 {
		t.Errorf("expected delete EndLine=4, got %d", annotations[0].EndLine)
	}

	if annotations[1].Type != "comment" {
		t.Errorf("expected second annotation type 'comment', got %q", annotations[1].Type)
	}
}

func TestParse_MultiLineAnnotationText(t *testing.T) {
	content := `{>> This is a multi-line
annotation that spans
multiple lines <<}`

	annotations, clean, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(annotations))
	}

	expectedText := "This is a multi-line\nannotation that spans\nmultiple lines"
	if annotations[0].Text != expectedText {
		t.Errorf("expected text=%q, got %q", expectedText, annotations[0].Text)
	}

	if annotations[0].StartLine != 1 {
		t.Errorf("expected StartLine=1, got %d", annotations[0].StartLine)
	}
	if annotations[0].EndLine != 3 {
		t.Errorf("expected EndLine=3, got %d", annotations[0].EndLine)
	}

	// All three lines should be cleaned
	expected := "\n\n"
	if clean != expected {
		t.Errorf("expected clean=%q, got %q", expected, clean)
	}
}

func TestParse_MultiLineAnnotationWithSurroundingContent(t *testing.T) {
	content := `Before text. {>> This spans
two lines <<} After text.`

	annotations, clean, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(annotations))
	}

	if annotations[0].Text != "This spans\ntwo lines" {
		t.Errorf("expected text='This spans\\ntwo lines', got %q", annotations[0].Text)
	}

	expected := "Before text. \n After text."
	if clean != expected {
		t.Errorf("expected clean=%q, got %q", expected, clean)
	}
}

func TestParse_EscapedMarkupNotParsed(t *testing.T) {
	content := `To add a comment, use the syntax \{>> comment <<\}`

	annotations, clean, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if len(annotations) != 0 {
		t.Errorf("expected 0 annotations for escaped markup, got %d", len(annotations))
	}

	// Clean content should have backslashes removed, preserving literal markers
	expected := `To add a comment, use the syntax {>> comment <<}`
	if clean != expected {
		t.Errorf("expected clean=%q, got %q", expected, clean)
	}
}

func TestParse_EscapedAndRealAnnotationOnSameLine(t *testing.T) {
	content := `Use \{>> like this <<\} for literal. {>> real comment <<}`

	annotations, clean, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(annotations))
	}

	if annotations[0].Type != "comment" {
		t.Errorf("expected type 'comment', got %q", annotations[0].Type)
	}
	if annotations[0].Text != "real comment" {
		t.Errorf("expected text 'real comment', got %q", annotations[0].Text)
	}

	expected := `Use {>> like this <<} for literal. `
	if clean != expected {
		t.Errorf("expected clean=%q, got %q", expected, clean)
	}
}

func TestParse_NestedBracesInAnnotationText(t *testing.T) {
	content := `Code example. {>> Use {curly braces} in the output <<}`

	annotations, clean, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(annotations))
	}

	if annotations[0].Text != "Use {curly braces} in the output" {
		t.Errorf("expected text 'Use {curly braces} in the output', got %q", annotations[0].Text)
	}

	if clean != "Code example. " {
		t.Errorf("expected clean='Code example. ', got %q", clean)
	}
}

func TestParse_SidecarLineReference(t *testing.T) {
	content := `Original content here.

{>> [line 1] This needs clarification <<}`

	annotations, _, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(annotations))
	}

	ann := annotations[0]
	if ann.Type != "comment" {
		t.Errorf("expected type 'comment', got %q", ann.Type)
	}
	if ann.StartLine != 1 {
		t.Errorf("expected StartLine=1 (from line ref), got %d", ann.StartLine)
	}
	if ann.EndLine != 1 {
		t.Errorf("expected EndLine=1 (from line ref), got %d", ann.EndLine)
	}
	if ann.Text != "This needs clarification" {
		t.Errorf("expected text='This needs clarification', got %q", ann.Text)
	}
}

func TestParse_SidecarLineRangeReference(t *testing.T) {
	content := `Line one.
Line two.
Line three.

{>> [lines 1-2] These lines need work <<}`

	annotations, _, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(annotations))
	}

	ann := annotations[0]
	if ann.StartLine != 1 {
		t.Errorf("expected StartLine=1, got %d", ann.StartLine)
	}
	if ann.EndLine != 2 {
		t.Errorf("expected EndLine=2, got %d", ann.EndLine)
	}
	if ann.Text != "These lines need work" {
		t.Errorf("expected text='These lines need work', got %q", ann.Text)
	}
}

func TestParse_SidecarWithMultipleAnnotationTypes(t *testing.T) {
	content := `Some code here.
More code.

{?? [line 1] Why this approach? ??}
{!! [line 2] EXPAND: Add error handling !!}`

	annotations, _, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if len(annotations) != 2 {
		t.Fatalf("expected 2 annotations, got %d", len(annotations))
	}

	if annotations[0].Type != "question" || annotations[0].StartLine != 1 {
		t.Errorf("first annotation: expected question on line 1, got %s on line %d", annotations[0].Type, annotations[0].StartLine)
	}
	if annotations[1].Type != "expand" || annotations[1].StartLine != 2 {
		t.Errorf("second annotation: expected expand on line 2, got %s on line %d", annotations[1].Type, annotations[1].StartLine)
	}
}

func TestParse_SidecarDoesNotAffectInlineAnnotations(t *testing.T) {
	content := `Line one {>> inline comment <<}`

	annotations, _, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(annotations))
	}

	// No [line N] prefix, so StartLine should be the physical line
	if annotations[0].StartLine != 1 {
		t.Errorf("expected StartLine=1, got %d", annotations[0].StartLine)
	}
	if annotations[0].Text != "inline comment" {
		t.Errorf("expected text='inline comment', got %q", annotations[0].Text)
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
