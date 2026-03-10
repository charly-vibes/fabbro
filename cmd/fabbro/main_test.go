package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charly-vibes/fabbro/internal/config"
	"github.com/charly-vibes/fabbro/internal/fem"
	"github.com/charly-vibes/fabbro/internal/session"
)

func noopTUI(tea.Model) error { return nil }

func TestTracerBullet(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// 1. Init
	err := config.Init()
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if !config.IsInitialized() {
		t.Fatal("expected fabbro to be initialized")
	}

	// 2. Create session with known content
	content := "First line\nSecond line\nThird line"
	sess, err := session.Create(content, "")
	if err != nil {
		t.Fatalf("Create session failed: %v", err)
	}

	// 3. Manually add FEM markup to session file
	femContent := `---
session_id: ` + sess.ID + `
created_at: 2026-01-11T22:00:00Z
---

First line {>> needs review <<}
Second line
Third line {>> consider refactoring <<}`

	sessionPath := filepath.Join(config.SessionsDir, sess.ID+".fem")
	err = os.WriteFile(sessionPath, []byte(femContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write FEM content: %v", err)
	}

	// 4. Load session and parse
	loaded, err := session.Load(sess.ID)
	if err != nil {
		t.Fatalf("Load session failed: %v", err)
	}

	annotations, cleanContent, err := fem.Parse(loaded.Content)
	if err != nil {
		t.Fatalf("Parse FEM failed: %v", err)
	}

	// 5. Verify JSON output structure
	if len(annotations) != 2 {
		t.Fatalf("expected 2 annotations, got %d", len(annotations))
	}

	if annotations[0].Text != "needs review" {
		t.Errorf("expected first annotation text 'needs review', got %q", annotations[0].Text)
	}

	if annotations[0].StartLine != 1 {
		t.Errorf("expected first annotation on line 1, got %d", annotations[0].StartLine)
	}

	if annotations[1].Text != "consider refactoring" {
		t.Errorf("expected second annotation text 'consider refactoring', got %q", annotations[1].Text)
	}

	if annotations[1].StartLine != 3 {
		t.Errorf("expected second annotation on line 3, got %d", annotations[1].StartLine)
	}

	// Verify clean content strips markers
	expectedClean := `First line 
Second line
Third line `
	if cleanContent != expectedClean {
		t.Errorf("expected clean content:\n%q\ngot:\n%q", expectedClean, cleanContent)
	}

	// Verify JSON marshaling works
	output := struct {
		SessionID   string           `json:"sessionId"`
		Annotations []fem.Annotation `json:"annotations"`
	}{
		SessionID:   loaded.ID,
		Annotations: annotations,
	}

	jsonBytes, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	var parsed map[string]interface{}
	err = json.Unmarshal(jsonBytes, &parsed)
	if err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if parsed["sessionId"] != loaded.ID {
		t.Errorf("expected sessionId=%s in JSON", loaded.ID)
	}

	annots := parsed["annotations"].([]interface{})
	if len(annots) != 2 {
		t.Errorf("expected 2 annotations in JSON, got %d", len(annots))
	}
}

func TestBuildProducesWorkingBinary(t *testing.T) {
	// This test just verifies the package compiles
	// The actual binary test is done via go build in CI
	t.Log("Build test passed - package compiles successfully")
}

func TestRealMainHelp(t *testing.T) {
	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"--help"}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "fabbro") {
		t.Error("expected help output to contain 'fabbro'")
	}
}

func TestRealMainVersion(t *testing.T) {
	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"--version"}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestInitCommand(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Prevent FindProjectRoot from walking up to the real project
	t.Setenv("FABBRO_PROJECT_ROOT_STOP", tmpDir)

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"init"}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "Initialized fabbro") {
		t.Errorf("expected 'Initialized fabbro' in output, got %q", stdout.String())
	}
}

func TestInitCommandQuiet(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	t.Setenv("FABBRO_PROJECT_ROOT_STOP", tmpDir)

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"init", "--quiet"}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr.String())
	}
	if stdout.String() != "" {
		t.Errorf("expected no stdout output with --quiet, got %q", stdout.String())
	}
	// Verify it still initialized
	if !config.IsInitialized() {
		t.Error("expected fabbro to be initialized after init --quiet")
	}
}

func TestInitCommandQuietAlreadyInitialized(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"init", "--quiet"}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr.String())
	}
	if stdout.String() != "" {
		t.Errorf("expected no stdout output with --quiet, got %q", stdout.String())
	}
}

func TestInitCommandWithAgents(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	t.Setenv("FABBRO_PROJECT_ROOT_STOP", tmpDir)

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"init", "--agents"}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr.String())
	}

	// Verify .fabbro/sessions was created
	if !config.IsInitialized() {
		t.Error("expected fabbro to be initialized")
	}

	// Verify .agents/commands/fabbro-review.md exists
	agentsPath := filepath.Join(tmpDir, ".agents", "commands", "fabbro-review.md")
	agentsContent, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("expected .agents/commands/fabbro-review.md to exist: %v", err)
	}
	if !strings.Contains(string(agentsContent), "fabbro") {
		t.Error("expected .agents/commands/fabbro-review.md to contain fabbro workflow instructions")
	}

	// Verify .claude/commands/fabbro-review.md exists
	claudePath := filepath.Join(tmpDir, ".claude", "commands", "fabbro-review.md")
	claudeContent, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("expected .claude/commands/fabbro-review.md to exist: %v", err)
	}
	if !strings.Contains(string(claudeContent), "fabbro") {
		t.Error("expected .claude/commands/fabbro-review.md to contain fabbro workflow instructions")
	}
}

func TestInitCommandSubdirectoryWarnsAboutParent(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	// Initialize in the parent directory
	os.Chdir(tmpDir)
	config.Init()

	// Create and cd into a subdirectory
	subDir := filepath.Join(tmpDir, "subproject")
	os.MkdirAll(subDir, 0755)
	os.Chdir(subDir)

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"init"}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr.String())
	}

	output := stdout.String()
	// Should warn about parent
	if !strings.Contains(output, "warning") && !strings.Contains(output, "Warning") {
		t.Errorf("expected warning about parent initialization, got %q", output)
	}
	// Should still initialize in cwd
	if _, err := os.Stat(filepath.Join(subDir, ".fabbro", "sessions")); err != nil {
		t.Errorf("expected .fabbro/sessions to be created in subdirectory: %v", err)
	}
}

func TestInitCommandAlreadyInitialized(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"init"}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "already initialized") {
		t.Errorf("expected 'already initialized' in output, got %q", stdout.String())
	}
}

func TestApplyCommandNotInitialized(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"apply", "some-id"}, stdin, &stdout, &stderr, noopTUI)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestApplyCommandWithJSON(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	// Create a session with FEM content
	sess, _ := session.Create("Test content", "")
	femContent := `---
session_id: ` + sess.ID + `
created_at: 2026-01-11T22:00:00Z
---

Test content {>> a comment <<}`

	sessionPath := filepath.Join(config.SessionsDir, sess.ID+".fem")
	os.WriteFile(sessionPath, []byte(femContent), 0644)

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"apply", sess.ID, "--json"}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout.String()), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if result["sessionId"] != sess.ID {
		t.Errorf("expected sessionId=%s, got %v", sess.ID, result["sessionId"])
	}
}

func TestApplyCommandJSONCompleteness(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	sess, _ := session.Create("Test content", "")
	femContent := `---
session_id: ` + sess.ID + `
created_at: 2026-01-11T22:00:00Z
content_hash: abc
---

Test content {>> a comment <<}`

	sessionPath := filepath.Join(config.SessionsDir, sess.ID+".fem")
	os.WriteFile(sessionPath, []byte(femContent), 0644)

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"apply", sess.ID, "--json"}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout.String()), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	// Must have createdAt field
	if _, ok := result["createdAt"]; !ok {
		t.Error("JSON output missing 'createdAt' field")
	}

	// Must have sourceFile field (even if empty for stdin)
	if _, ok := result["sourceFile"]; !ok {
		t.Error("JSON output missing 'sourceFile' field")
	}

	// Must have sessionId
	if _, ok := result["sessionId"]; !ok {
		t.Error("JSON output missing 'sessionId' field")
	}
}

func TestApplyCommandCompactJSON(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	sess, _ := session.Create("Test content", "")
	femContent := `---
session_id: ` + sess.ID + `
created_at: 2026-01-11T22:00:00Z
---

Test content {>> a comment <<}`

	sessionPath := filepath.Join(config.SessionsDir, sess.ID+".fem")
	os.WriteFile(sessionPath, []byte(femContent), 0644)

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"apply", sess.ID, "--json", "--compact"}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr.String())
	}

	output := strings.TrimSpace(stdout.String())

	// Compact JSON should be a single line
	if strings.Count(output, "\n") != 0 {
		t.Errorf("expected single-line output, got:\n%s", output)
	}

	// Should still be valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("failed to parse compact JSON: %v", err)
	}

	if result["sessionId"] != sess.ID {
		t.Errorf("expected sessionId=%s, got %v", sess.ID, result["sessionId"])
	}
}

func TestApplyCommandWithoutJSON(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	sess, _ := session.Create("Test content", "")
	femContent := `---
session_id: ` + sess.ID + `
created_at: 2026-01-11T22:00:00Z
---

Test content {>> a comment <<}`

	sessionPath := filepath.Join(config.SessionsDir, sess.ID+".fem")
	os.WriteFile(sessionPath, []byte(femContent), 0644)

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"apply", sess.ID}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	if !strings.Contains(stdout.String(), "Session:") {
		t.Error("expected output to contain 'Session:'")
	}
	if !strings.Contains(stdout.String(), "Annotations:") {
		t.Error("expected output to contain 'Annotations:'")
	}
}

func TestApplyCommandByFile(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	// Create a session with source file
	sess, _ := session.Create("Test content", "plans/my-plan.md")
	femContent := `---
session_id: ` + sess.ID + `
created_at: 2026-01-11T22:00:00Z
source_file: 'plans/my-plan.md'
---

Test content {>> a comment <<}`

	sessionPath := filepath.Join(config.SessionsDir, sess.ID+".fem")
	os.WriteFile(sessionPath, []byte(femContent), 0644)

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"apply", "--file", "plans/my-plan.md", "--json"}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d. stderr: %s", code, stderr.String())
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout.String()), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if result["sessionId"] != sess.ID {
		t.Errorf("expected sessionId=%s, got %v", sess.ID, result["sessionId"])
	}
	if result["sourceFile"] != "plans/my-plan.md" {
		t.Errorf("expected sourceFile=plans/my-plan.md, got %v", result["sourceFile"])
	}
}

func TestApplyCommandByFileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"apply", "--file", "nonexistent.md"}, stdin, &stdout, &stderr, noopTUI)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestApplyCommandMutualExclusivity(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"apply", "session-123", "--file", "doc.md"}, stdin, &stdout, &stderr, noopTUI)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestApplyCommandNoInput(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"apply"}, stdin, &stdout, &stderr, noopTUI)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestApplyCommandContentHashWarning(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	// Create a source file
	sourceFile := filepath.Join(tmpDir, "document.md")
	os.WriteFile(sourceFile, []byte("original content"), 0644)

	// Create session from file
	sess, _ := session.Create("original content", sourceFile)

	// Modify the source file after session creation
	os.WriteFile(sourceFile, []byte("modified content"), 0644)

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"apply", sess.ID}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	// Should still output annotations
	if !strings.Contains(stdout.String(), "Session:") {
		t.Error("expected output to contain session info")
	}

	// Should warn about changed source
	if !strings.Contains(stderr.String(), "source file has changed") {
		t.Errorf("expected warning about changed source file, got stderr: %q", stderr.String())
	}
}

func TestApplyCommandContentHashNoWarningForStdin(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	// Create session from stdin (no source file)
	sess, _ := session.Create("some content", "")

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"apply", sess.ID}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	// Should NOT warn for stdin sessions
	if strings.Contains(stderr.String(), "source file has changed") {
		t.Error("should not warn about source hash for stdin sessions")
	}
}

func TestReviewCommandNotInitialized(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	t.Setenv("FABBRO_PROJECT_ROOT_STOP", tmpDir)

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("content")

	code := realMain([]string{"review", "--stdin"}, stdin, &stdout, &stderr, noopTUI)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestReviewCommandWithCustomID(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	t.Setenv("FABBRO_PROJECT_ROOT_STOP", tmpDir)
	config.Init()

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("test content")

	code := realMain([]string{"review", "--stdin", "--id", "my-review"}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr.String())
	}

	if !strings.Contains(stdout.String(), "my-review") {
		t.Errorf("expected output to contain custom ID 'my-review', got: %s", stdout.String())
	}

	// Verify session file exists with custom name
	sessionPath := filepath.Join(config.SessionsDir, "my-review.fem")
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		t.Error("expected session file 'my-review.fem' to exist")
	}
}

func TestReviewCommandWithInvalidID(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	t.Setenv("FABBRO_PROJECT_ROOT_STOP", tmpDir)
	config.Init()

	tests := []struct {
		name string
		id   string
	}{
		{"spaces", "my review"},
		{"slashes", "my/review"},
		{"too long", strings.Repeat("a", 65)},
		{"reserved tutor", "tutor"},
		{"reserved _tutor_", "_tutor_"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr strings.Builder
			stdin := strings.NewReader("test content")

			code := realMain([]string{"review", "--stdin", "--id", tt.id}, stdin, &stdout, &stderr, noopTUI)

			if code != 1 {
				t.Errorf("expected exit code 1 for invalid ID %q, got %d", tt.id, code)
			}
		})
	}
}

func TestReviewCommandWithEditor(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	t.Setenv("FABBRO_PROJECT_ROOT_STOP", tmpDir)
	t.Setenv("EDITOR", "true") // no-op command
	config.Init()

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("editor test content")

	code := realMain([]string{"review", "--stdin", "--editor"}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr.String())
	}

	// Session should have been created
	if !strings.Contains(stdout.String(), "Created session:") {
		t.Errorf("expected session creation message, got: %s", stdout.String())
	}
}

func TestReviewCommandWithNoInteractive(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	t.Setenv("FABBRO_PROJECT_ROOT_STOP", tmpDir)
	config.Init()

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("non-interactive content")

	code := realMain([]string{"review", "--stdin", "--no-interactive"}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr.String())
	}

	output := strings.TrimSpace(stdout.String())

	// Should print session ID
	if output == "" {
		t.Error("expected session ID to be printed to stdout")
	}

	// Session file should exist
	sessionPath := filepath.Join(config.SessionsDir, output+".fem")
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		t.Errorf("expected session file %q to exist", sessionPath)
	}
}

func TestReviewCommandWithDuplicateID(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	t.Setenv("FABBRO_PROJECT_ROOT_STOP", tmpDir)
	config.Init()

	// Create first session
	var stdout1, stderr1 strings.Builder
	stdin1 := strings.NewReader("first content")
	realMain([]string{"review", "--stdin", "--id", "dup-test"}, stdin1, &stdout1, &stderr1, noopTUI)

	// Try to create duplicate
	var stdout2, stderr2 strings.Builder
	stdin2 := strings.NewReader("second content")
	code := realMain([]string{"review", "--stdin", "--id", "dup-test"}, stdin2, &stdout2, &stderr2, noopTUI)

	if code != 1 {
		t.Errorf("expected exit code 1 for duplicate ID, got %d", code)
	}
}

func TestReviewCommandWithoutStdinFlag(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	t.Setenv("FABBRO_PROJECT_ROOT_STOP", tmpDir)
	config.Init()

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"review"}, stdin, &stdout, &stderr, noopTUI)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestApplyCommandSessionNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"apply", "nonexistent-session"}, stdin, &stdout, &stderr, noopTUI)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestUnknownCommand(t *testing.T) {
	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"unknown"}, stdin, &stdout, &stderr, noopTUI)

	if code != 1 {
		t.Errorf("expected exit code 1 for unknown command, got %d", code)
	}
}

func TestReviewCommandWithFile(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	t.Setenv("FABBRO_PROJECT_ROOT_STOP", tmpDir)
	config.Init()

	// Create a test file
	testContent := "# My Document\n\nThis is test content."
	os.WriteFile("document.md", []byte(testContent), 0644)

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	// Note: TUI will fail without a TTY, but we can verify session creation
	realMain([]string{"review", "document.md"}, stdin, &stdout, &stderr, noopTUI)

	// Verify the session was created (message appears before TUI attempt)
	if !strings.Contains(stdout.String(), "Created session:") {
		t.Errorf("expected 'Created session:' in output, got %q", stdout.String())
	}

	// Verify session file exists
	files, err := os.ReadDir(config.SessionsDir)
	if err != nil {
		t.Fatalf("failed to read sessions dir: %v", err)
	}
	if len(files) == 0 {
		t.Error("expected session file to be created")
	}

	// Verify file content was captured in session
	sessionPath := filepath.Join(config.SessionsDir, files[0].Name())
	sessionContent, _ := os.ReadFile(sessionPath)
	if !strings.Contains(string(sessionContent), "This is test content") {
		t.Errorf("expected session to contain original file content")
	}
}

func TestReviewCommandWithNonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	t.Setenv("FABBRO_PROJECT_ROOT_STOP", tmpDir)
	config.Init()

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"review", "missing.md"}, stdin, &stdout, &stderr, noopTUI)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestReviewCommandRejectsOversizedStdin(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	t.Setenv("FABBRO_PROJECT_ROOT_STOP", tmpDir)
	config.Init()

	largeInput := strings.Repeat("x", maxInputBytes+1)
	var stdout, stderr strings.Builder
	stdin := strings.NewReader(largeInput)

	code := realMain([]string{"review", "--stdin"}, stdin, &stdout, &stderr, noopTUI)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "input too large") {
		t.Errorf("expected 'input too large' error, got: %s", stderr.String())
	}
}

func TestReviewCommandRejectsOversizedFile(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	t.Setenv("FABBRO_PROJECT_ROOT_STOP", tmpDir)
	config.Init()

	largeFile := filepath.Join(tmpDir, "large.txt")
	f, _ := os.Create(largeFile)
	f.Truncate(maxInputBytes + 1)
	f.Close()

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"review", largeFile}, stdin, &stdout, &stderr, noopTUI)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "file too large") {
		t.Errorf("expected 'file too large' error, got: %s", stderr.String())
	}
}

func TestReviewCommandRejectsStdinAndFile(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	t.Setenv("FABBRO_PROJECT_ROOT_STOP", tmpDir)
	config.Init()

	// Create a file
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("content"), 0644)

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("stdin content")

	// Pass both --stdin and a file path
	code := realMain([]string{"review", "--stdin", testFile}, stdin, &stdout, &stderr, noopTUI)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "cannot use both --stdin and a file path") {
		t.Errorf("expected conflict error, got: %s", stderr.String())
	}
}

func TestReviewCommandJSONOutput(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	t.Setenv("FABBRO_PROJECT_ROOT_STOP", tmpDir)
	config.Init()

	// Create a test file
	testContent := "# Test\n\nContent for JSON test."
	os.WriteFile("test.md", []byte(testContent), 0644)

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	// Note: TUI will fail without a TTY, but session creation and JSON output happen first
	realMain([]string{"review", "--json", "test.md"}, stdin, &stdout, &stderr, noopTUI)

	output := stdout.String()

	// Verify JSON output format
	if !strings.HasPrefix(output, "{") {
		t.Errorf("expected JSON output to start with '{', got %q", output)
	}
	if !strings.Contains(output, `"sessionId"`) {
		t.Errorf("expected JSON to contain 'sessionId' key, got %q", output)
	}

	// Parse and validate JSON structure
	var result map[string]string
	// Find the JSON line (first line before any TUI errors)
	lines := strings.Split(output, "\n")
	if len(lines) == 0 {
		t.Fatal("expected at least one line of output")
	}
	if err := json.Unmarshal([]byte(lines[0]), &result); err != nil {
		t.Errorf("expected valid JSON, got parse error: %v for output: %q", err, lines[0])
	}
	if result["sessionId"] == "" {
		t.Error("expected sessionId to be non-empty")
	}
}

func TestCompletionCommand(t *testing.T) {
	shells := []string{"bash", "zsh", "fish", "powershell"}

	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			var stdout, stderr strings.Builder
			stdin := strings.NewReader("")

			code := realMain([]string{"completion", shell}, stdin, &stdout, &stderr, noopTUI)

			if code != 0 {
				t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr.String())
			}
			if stdout.Len() == 0 {
				t.Error("expected completion output, got empty")
			}
		})
	}
}

func TestCompletionCommandInvalidShell(t *testing.T) {
	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"completion", "invalid"}, stdin, &stdout, &stderr, noopTUI)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestSessionListShowsAnnotationCount(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	// Create a session with FEM annotations
	sess, _ := session.Create("Test content", "main.go")
	femContent := `---
session_id: ` + sess.ID + `
created_at: 2026-01-11T22:00:00Z
source_file: 'main.go'
---

Test content {>> a comment <<} {?? a question ??}`

	sessionPath := filepath.Join(config.SessionsDir, sess.ID+".fem")
	os.WriteFile(sessionPath, []byte(femContent), 0644)

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"session", "list"}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "2") {
		t.Errorf("expected annotation count '2' in output, got %q", output)
	}
}

func TestSessionListJSONIncludesAnnotations(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	sess, _ := session.Create("Test content", "main.go")
	femContent := `---
session_id: ` + sess.ID + `
created_at: 2026-01-11T22:00:00Z
source_file: 'main.go'
---

Test content {>> a comment <<}`

	sessionPath := filepath.Join(config.SessionsDir, sess.ID+".fem")
	os.WriteFile(sessionPath, []byte(femContent), 0644)

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"session", "list", "--json"}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr.String())
	}

	var result []map[string]interface{}
	if err := json.Unmarshal([]byte(stdout.String()), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 session, got %d", len(result))
	}

	annotations, ok := result[0]["annotations"]
	if !ok {
		t.Fatal("expected 'annotations' key in JSON output")
	}
	if int(annotations.(float64)) != 1 {
		t.Errorf("expected annotations=1, got %v", annotations)
	}
}

func TestSessionListEmptyShowsHelpfulMessage(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"session", "list"}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	output := stdout.String()
	if !strings.Contains(output, "No sessions found") {
		t.Errorf("expected 'No sessions found' in output, got %q", output)
	}
	if !strings.Contains(output, "fabbro review") {
		t.Errorf("expected 'fabbro review' suggestion in output, got %q", output)
	}
}

func TestSessionExportToStdout(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	sess, _ := session.Create("exported content", "export.go")
	femContent := fmt.Sprintf("---\nsession_id: %s\ncreated_at: 2026-01-11T22:00:00Z\nsource_file: 'export.go'\n---\n\nexported content {>> nice <<}", sess.ID)
	os.WriteFile(filepath.Join(config.SessionsDir, sess.ID+".fem"), []byte(femContent), 0644)

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"session", "export", sess.ID}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "exported content") {
		t.Errorf("expected session content in stdout, got %q", output)
	}
	if !strings.Contains(output, "{>> nice <<}") {
		t.Errorf("expected FEM markers in output, got %q", output)
	}
}

func TestSessionExportToFile(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	sess, _ := session.Create("file export content", "export.go")

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	outputFile := filepath.Join(tmpDir, "review.fem")
	code := realMain([]string{"session", "export", sess.ID, "--output", outputFile}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr.String())
	}

	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("expected output file to exist: %v", err)
	}

	if !strings.Contains(string(data), "file export content") {
		t.Errorf("expected content in output file, got %q", string(data))
	}
}

func TestSessionExportNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"session", "export", "nonexistent"}, stdin, &stdout, &stderr, noopTUI)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestSessionCleanDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	// Create an old session (fake old timestamp)
	sess, _ := session.Create("old content", "old.go")
	oldContent := fmt.Sprintf("---\nsession_id: %s\ncreated_at: 2025-01-01T00:00:00Z\nsource_file: 'old.go'\n---\n\nold content", sess.ID)
	os.WriteFile(filepath.Join(config.SessionsDir, sess.ID+".fem"), []byte(oldContent), 0644)

	// Create a recent session
	recent, _ := session.Create("recent content", "recent.go")

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"session", "clean", "--older-than", "7d", "--dry-run"}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, sess.ID) {
		t.Errorf("expected old session ID in dry-run output, got %q", output)
	}
	if strings.Contains(output, recent.ID) {
		t.Errorf("expected recent session to NOT appear in dry-run output, got %q", output)
	}

	// Both files should still exist (dry run)
	if _, err := os.Stat(filepath.Join(config.SessionsDir, sess.ID+".fem")); err != nil {
		t.Error("expected old session file to still exist after dry-run")
	}
	if _, err := os.Stat(filepath.Join(config.SessionsDir, recent.ID+".fem")); err != nil {
		t.Error("expected recent session file to still exist after dry-run")
	}
}

func TestSessionCleanWithForce(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	// Create an old session
	sess, _ := session.Create("old content", "old.go")
	oldContent := fmt.Sprintf("---\nsession_id: %s\ncreated_at: 2025-01-01T00:00:00Z\nsource_file: 'old.go'\n---\n\nold content", sess.ID)
	os.WriteFile(filepath.Join(config.SessionsDir, sess.ID+".fem"), []byte(oldContent), 0644)

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"session", "clean", "--older-than", "7d", "--force"}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Deleted") {
		t.Errorf("expected 'Deleted' in output, got %q", output)
	}

	// Old session should be gone
	if _, err := os.Stat(filepath.Join(config.SessionsDir, sess.ID+".fem")); !os.IsNotExist(err) {
		t.Error("expected old session file to be deleted")
	}
}

func TestSessionCleanSafetyLimit(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"session", "clean", "--older-than", "0d"}, stdin, &stdout, &stderr, noopTUI)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestSessionCleanNoMatches(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	// Create a very recent session
	session.Create("recent content", "recent.go")

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"session", "clean", "--older-than", "7d", "--force"}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "No sessions") {
		t.Errorf("expected 'No sessions' message, got %q", output)
	}
}

func TestSessionDeleteWithForce(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	sess, _ := session.Create("content to delete", "delete-me.go")

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"session", "delete", sess.ID, "--force"}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Deleted session") {
		t.Errorf("expected 'Deleted session' in output, got %q", output)
	}

	// Verify session is gone
	sessionPath := filepath.Join(config.SessionsDir, sess.ID+".fem")
	if _, err := os.Stat(sessionPath); !os.IsNotExist(err) {
		t.Error("expected session file to be deleted")
	}
}

func TestSessionDeleteNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"session", "delete", "nonexistent", "--force"}, stdin, &stdout, &stderr, noopTUI)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestSessionDeleteWithConfirmation(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	sess, _ := session.Create("content to delete", "delete-me.go")

	// Simulate user typing "y\n"
	var stdout, stderr strings.Builder
	stdin := strings.NewReader("y\n")

	code := realMain([]string{"session", "delete", sess.ID}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr.String())
	}

	// Verify session is gone
	sessionPath := filepath.Join(config.SessionsDir, sess.ID+".fem")
	if _, err := os.Stat(sessionPath); !os.IsNotExist(err) {
		t.Error("expected session file to be deleted after confirmation")
	}
}

func TestSessionDeleteAborted(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	sess, _ := session.Create("content to keep", "keep-me.go")

	// Simulate user typing "n\n"
	var stdout, stderr strings.Builder
	stdin := strings.NewReader("n\n")

	code := realMain([]string{"session", "delete", sess.ID}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Aborted") {
		t.Errorf("expected 'Aborted' in output, got %q", output)
	}

	// Verify session still exists
	sessionPath := filepath.Join(config.SessionsDir, sess.ID+".fem")
	if _, err := os.Stat(sessionPath); err != nil {
		t.Error("expected session file to still exist after abort")
	}
}

func TestSessionResumeNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"session", "resume", "nonexistent"}, stdin, &stdout, &stderr, noopTUI)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestSessionResumeWithEditor(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()
	// Use 'true' as editor — it's a no-op command that exits 0
	t.Setenv("EDITOR", "true")

	sess, _ := session.Create("Test content", "main.go")

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"session", "resume", sess.ID, "--editor"}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Opening") {
		t.Errorf("expected 'Opening' in output for editor mode, got %q", output)
	}
}

func TestSessionResumeWithEditorNoEditorSet(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()
	t.Setenv("EDITOR", "")
	t.Setenv("VISUAL", "")

	sess, _ := session.Create("Test content", "main.go")

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"session", "resume", sess.ID, "--editor"}, stdin, &stdout, &stderr, noopTUI)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestSessionShowDisplaysDetails(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	sess, _ := session.Create("Test content", "main.go")
	femContent := `---
session_id: ` + sess.ID + `
created_at: 2026-01-11T22:00:00Z
source_file: 'main.go'
---

Test content {>> a comment <<} {?? a question ??} {-- remove this --}`

	sessionPath := filepath.Join(config.SessionsDir, sess.ID+".fem")
	os.WriteFile(sessionPath, []byte(femContent), 0644)

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"session", "show", sess.ID}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Session ID:") {
		t.Errorf("expected 'Session ID:' in output, got %q", output)
	}
	if !strings.Contains(output, sess.ID) {
		t.Errorf("expected session ID in output, got %q", output)
	}
	if !strings.Contains(output, "Source:") {
		t.Errorf("expected 'Source:' in output, got %q", output)
	}
	if !strings.Contains(output, "main.go") {
		t.Errorf("expected 'main.go' in output, got %q", output)
	}
	if !strings.Contains(output, "Annotations (3 total):") {
		t.Errorf("expected 'Annotations (3 total):' in output, got %q", output)
	}
	if !strings.Contains(output, "comment:") {
		t.Errorf("expected 'comment:' in annotation breakdown, got %q", output)
	}
	if !strings.Contains(output, "question:") {
		t.Errorf("expected 'question:' in annotation breakdown, got %q", output)
	}
	if !strings.Contains(output, "delete:") {
		t.Errorf("expected 'delete:' in annotation breakdown, got %q", output)
	}
}

func TestSessionShowNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"session", "show", "nonexistent"}, stdin, &stdout, &stderr, noopTUI)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestSessionShowStdinSource(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	sess, _ := session.Create("Test content", "")
	femContent := `---
session_id: ` + sess.ID + `
created_at: 2026-01-11T22:00:00Z
---

Test content`

	sessionPath := filepath.Join(config.SessionsDir, sess.ID+".fem")
	os.WriteFile(sessionPath, []byte(femContent), 0644)

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"session", "show", sess.ID}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "(stdin)") {
		t.Errorf("expected '(stdin)' for source, got %q", output)
	}
}

func TestPrimeCommand(t *testing.T) {
	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"prime"}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr.String())
	}

	output := stdout.String()

	// Verify essential sections are present
	expectedSections := []string{
		"fabbro",
		"review",
		"apply",
		"session",
		"FEM",
		"{>>",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("expected output to contain %q", section)
		}
	}
}

func TestPrimeCommandJSON(t *testing.T) {
	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"prime", "--json"}, stdin, &stdout, &stderr, noopTUI)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d; stderr: %s", code, stderr.String())
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout.String()), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	// Verify JSON structure
	if _, ok := result["purpose"]; !ok {
		t.Error("expected 'purpose' key in JSON output")
	}
	if _, ok := result["commands"]; !ok {
		t.Error("expected 'commands' key in JSON output")
	}
	if _, ok := result["femSyntax"]; !ok {
		t.Error("expected 'femSyntax' key in JSON output")
	}
}
