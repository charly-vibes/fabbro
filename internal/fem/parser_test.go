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

	if annotations[0].Line != 1 {
		t.Errorf("expected first annotation on line 1, got %d", annotations[0].Line)
	}

	if annotations[1].Line != 3 {
		t.Errorf("expected second annotation on line 3, got %d", annotations[1].Line)
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
