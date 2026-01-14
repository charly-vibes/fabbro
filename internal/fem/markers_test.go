package fem

import "testing"

func TestAnnotationTypes_AllDefined(t *testing.T) {
	expected := []string{"comment", "delete", "question", "expand", "keep", "unclear"}

	for _, typ := range expected {
		if _, ok := Markers[typ]; !ok {
			t.Errorf("Markers missing type %q", typ)
		}
		if _, ok := Prompts[typ]; !ok {
			t.Errorf("Prompts missing type %q", typ)
		}
	}
}

func TestAnnotationTypes_MarkersAndPromptsMatch(t *testing.T) {
	for typ := range Markers {
		if _, ok := Prompts[typ]; !ok {
			t.Errorf("Markers has %q but Prompts does not", typ)
		}
	}
	for typ := range Prompts {
		if _, ok := Markers[typ]; !ok {
			t.Errorf("Prompts has %q but Markers does not", typ)
		}
	}
}

func TestValidAnnotationType(t *testing.T) {
	tests := []struct {
		typ   string
		valid bool
	}{
		{"comment", true},
		{"delete", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		got := ValidAnnotationType(tt.typ)
		if got != tt.valid {
			t.Errorf("ValidAnnotationType(%q) = %v, want %v", tt.typ, got, tt.valid)
		}
	}
}
