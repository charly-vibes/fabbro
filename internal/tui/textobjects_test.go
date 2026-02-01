package tui

import "testing"

func TestFindParagraph(t *testing.T) {
	tests := []struct {
		name          string
		lines         []string
		line          int
		wantStart     int
		wantEnd       int
	}{
		{
			name:      "single line document",
			lines:     []string{"hello"},
			line:      0,
			wantStart: 0,
			wantEnd:   0,
		},
		{
			name:      "first line of paragraph",
			lines:     []string{"first", "second", "third", "", "other"},
			line:      0,
			wantStart: 0,
			wantEnd:   2,
		},
		{
			name:      "middle line of paragraph",
			lines:     []string{"first", "second", "third", "", "other"},
			line:      1,
			wantStart: 0,
			wantEnd:   2,
		},
		{
			name:      "last line of paragraph",
			lines:     []string{"first", "second", "third", "", "other"},
			line:      2,
			wantStart: 0,
			wantEnd:   2,
		},
		{
			name:      "paragraph at end of file",
			lines:     []string{"intro", "", "final", "lines"},
			line:      3,
			wantStart: 2,
			wantEnd:   3,
		},
		{
			name:      "paragraph between blanks",
			lines:     []string{"", "middle", "paragraph", ""},
			line:      1,
			wantStart: 1,
			wantEnd:   2,
		},
		{
			name:      "on blank line returns that line",
			lines:     []string{"first", "", "third"},
			line:      1,
			wantStart: 1,
			wantEnd:   1,
		},
		{
			name:      "multiple blank lines before",
			lines:     []string{"", "", "content", "here"},
			line:      2,
			wantStart: 2,
			wantEnd:   3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			start, end := FindParagraph(tc.lines, tc.line)
			if start != tc.wantStart || end != tc.wantEnd {
				t.Errorf("FindParagraph() = (%d, %d), want (%d, %d)", start, end, tc.wantStart, tc.wantEnd)
			}
		})
	}
}

func TestFindCodeBlock(t *testing.T) {
	tests := []struct {
		name      string
		lines     []string
		line      int
		wantStart int
		wantEnd   int
	}{
		{
			name:      "not in code block",
			lines:     []string{"normal text", "more text"},
			line:      0,
			wantStart: -1,
			wantEnd:   -1,
		},
		{
			name:      "on opening fence",
			lines:     []string{"text", "```", "code", "```", "after"},
			line:      1,
			wantStart: 1,
			wantEnd:   3,
		},
		{
			name:      "inside code block",
			lines:     []string{"text", "```", "code", "```", "after"},
			line:      2,
			wantStart: 1,
			wantEnd:   3,
		},
		{
			name:      "on closing fence",
			lines:     []string{"text", "```", "code", "```", "after"},
			line:      3,
			wantStart: 1,
			wantEnd:   3,
		},
		{
			name:      "code block with language",
			lines:     []string{"```go", "func main() {}", "```"},
			line:      1,
			wantStart: 0,
			wantEnd:   2,
		},
		{
			name:      "multiple code blocks - first",
			lines:     []string{"```", "first", "```", "", "```", "second", "```"},
			line:      1,
			wantStart: 0,
			wantEnd:   2,
		},
		{
			name:      "multiple code blocks - second",
			lines:     []string{"```", "first", "```", "", "```", "second", "```"},
			line:      5,
			wantStart: 4,
			wantEnd:   6,
		},
		{
			name:      "outside between blocks",
			lines:     []string{"```", "first", "```", "between", "```", "second", "```"},
			line:      3,
			wantStart: -1,
			wantEnd:   -1,
		},
		{
			name:      "unclosed code block",
			lines:     []string{"text", "```", "unclosed code"},
			line:      2,
			wantStart: -1,
			wantEnd:   -1,
		},
		{
			name:      "indented fence",
			lines:     []string{"  ```", "code", "  ```"},
			line:      1,
			wantStart: 0,
			wantEnd:   2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			start, end := FindCodeBlock(tc.lines, tc.line)
			if start != tc.wantStart || end != tc.wantEnd {
				t.Errorf("FindCodeBlock() = (%d, %d), want (%d, %d)", start, end, tc.wantStart, tc.wantEnd)
			}
		})
	}
}

func TestFindSection(t *testing.T) {
	tests := []struct {
		name      string
		lines     []string
		line      int
		wantStart int
		wantEnd   int
	}{
		{
			name:      "no heading above - returns whole document",
			lines:     []string{"text", "more", "stuff"},
			line:      1,
			wantStart: 0,
			wantEnd:   2,
		},
		{
			name:      "on heading line",
			lines:     []string{"# Heading", "content", "more"},
			line:      0,
			wantStart: 0,
			wantEnd:   2,
		},
		{
			name:      "under heading to next same level",
			lines:     []string{"# One", "content", "# Two", "other"},
			line:      1,
			wantStart: 0,
			wantEnd:   1,
		},
		{
			name:      "section includes subheadings",
			lines:     []string{"# Main", "text", "## Sub", "subtext", "# Next"},
			line:      1,
			wantStart: 0,
			wantEnd:   3,
		},
		{
			name:      "subheading section to next same or higher level",
			lines:     []string{"# Main", "## Sub1", "content", "## Sub2", "other", "# Next"},
			line:      2,
			wantStart: 1,
			wantEnd:   2,
		},
		{
			name:      "last section extends to EOF",
			lines:     []string{"# First", "a", "# Last", "b", "c"},
			line:      4,
			wantStart: 2,
			wantEnd:   4,
		},
		{
			name:      "deeply nested heading",
			lines:     []string{"# H1", "## H2", "### H3", "content", "## H2b"},
			line:      3,
			wantStart: 2,
			wantEnd:   3,
		},
		{
			name:      "content before any heading",
			lines:     []string{"preamble", "# First"},
			line:      0,
			wantStart: 0,
			wantEnd:   0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			start, end := FindSection(tc.lines, tc.line)
			if start != tc.wantStart || end != tc.wantEnd {
				t.Errorf("FindSection() = (%d, %d), want (%d, %d)", start, end, tc.wantStart, tc.wantEnd)
			}
		})
	}
}
