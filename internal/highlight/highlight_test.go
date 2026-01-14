package highlight

import (
	"strings"
	"testing"
)

func TestNewWithFilename(t *testing.T) {
	h := New("main.go", "package main")
	if h == nil {
		t.Fatal("expected non-nil highlighter")
	}
	if h.lexer == nil {
		t.Error("expected lexer to be set")
	}
}

func TestNewWithContentAnalysis(t *testing.T) {
	goCode := `package main

import "fmt"

func main() {
	fmt.Println("hello")
}`
	h := New("", goCode)
	if h == nil {
		t.Fatal("expected non-nil highlighter")
	}
}

func TestHighlightLineReturnsTokens(t *testing.T) {
	h := New("test.go", "")
	tokens := h.HighlightLine("func main() {}")

	if len(tokens) == 0 {
		t.Error("expected at least one token")
	}

	var hasFunc bool
	for _, tok := range tokens {
		if strings.Contains(tok.Text, "func") {
			hasFunc = true
			break
		}
	}
	if !hasFunc {
		t.Error("expected to find 'func' in tokens")
	}
}

func TestRenderLineContainsANSI(t *testing.T) {
	h := New("test.go", "")
	rendered := h.RenderLine("func main() {}")

	if !strings.Contains(rendered, "\033[") {
		t.Error("expected ANSI escape codes in rendered output")
	}
	if !strings.Contains(rendered, "func") {
		t.Error("expected 'func' in rendered output")
	}
}

func TestRenderLinePlainText(t *testing.T) {
	h := New("", "just some plain text")
	rendered := h.RenderLine("hello world")

	if !strings.Contains(rendered, "hello world") {
		t.Error("expected plain text to be preserved")
	}
}

func TestAnsiColor(t *testing.T) {
	tests := []struct {
		hex  string
		want string
	}{
		{"#ff0000", "\033[38;2;255;0;0m"},
		{"#00ff00", "\033[38;2;0;255;0m"},
		{"#0000ff", "\033[38;2;0;0;255m"},
		{"invalid", ""},
	}

	for _, tt := range tests {
		got := ansiColor(tt.hex)
		if got != tt.want {
			t.Errorf("ansiColor(%q) = %q, want %q", tt.hex, got, tt.want)
		}
	}
}
