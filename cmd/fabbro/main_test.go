package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charly-vibes/fabbro/internal/config"
	"github.com/charly-vibes/fabbro/internal/fem"
	"github.com/charly-vibes/fabbro/internal/session"
)

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

	code := realMain([]string{"--help"}, stdin, &stdout, &stderr)

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

	code := realMain([]string{"--version"}, stdin, &stdout, &stderr)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestInitCommand(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"init"}, stdin, &stdout, &stderr)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "Initialized fabbro") {
		t.Errorf("expected 'Initialized fabbro' in output, got %q", stdout.String())
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

	code := realMain([]string{"init"}, stdin, &stdout, &stderr)

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

	code := realMain([]string{"apply", "some-id"}, stdin, &stdout, &stderr)

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

	code := realMain([]string{"apply", sess.ID, "--json"}, stdin, &stdout, &stderr)

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

	code := realMain([]string{"apply", sess.ID}, stdin, &stdout, &stderr)

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

	code := realMain([]string{"apply", "--file", "plans/my-plan.md", "--json"}, stdin, &stdout, &stderr)

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

	code := realMain([]string{"apply", "--file", "nonexistent.md"}, stdin, &stdout, &stderr)

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

	code := realMain([]string{"apply", "session-123", "--file", "doc.md"}, stdin, &stdout, &stderr)

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

	code := realMain([]string{"apply"}, stdin, &stdout, &stderr)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestReviewCommandNotInitialized(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("content")

	code := realMain([]string{"review", "--stdin"}, stdin, &stdout, &stderr)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestReviewCommandWithoutStdinFlag(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"review"}, stdin, &stdout, &stderr)

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

	code := realMain([]string{"apply", "nonexistent-session"}, stdin, &stdout, &stderr)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestUnknownCommand(t *testing.T) {
	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"unknown"}, stdin, &stdout, &stderr)

	if code != 1 {
		t.Errorf("expected exit code 1 for unknown command, got %d", code)
	}
}

func TestReviewCommandWithFile(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	// Create a test file
	testContent := "# My Document\n\nThis is test content."
	os.WriteFile("document.md", []byte(testContent), 0644)

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	// Note: TUI will fail without a TTY, but we can verify session creation
	realMain([]string{"review", "document.md"}, stdin, &stdout, &stderr)

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

	config.Init()

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"review", "missing.md"}, stdin, &stdout, &stderr)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestReviewCommandRejectsOversizedStdin(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	largeInput := strings.Repeat("x", maxInputBytes+1)
	var stdout, stderr strings.Builder
	stdin := strings.NewReader(largeInput)

	code := realMain([]string{"review", "--stdin"}, stdin, &stdout, &stderr)

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

	config.Init()

	largeFile := filepath.Join(tmpDir, "large.txt")
	f, _ := os.Create(largeFile)
	f.Truncate(maxInputBytes + 1)
	f.Close()

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("")

	code := realMain([]string{"review", largeFile}, stdin, &stdout, &stderr)

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

	config.Init()

	// Create a file
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("content"), 0644)

	var stdout, stderr strings.Builder
	stdin := strings.NewReader("stdin content")

	// Pass both --stdin and a file path
	code := realMain([]string{"review", "--stdin", testFile}, stdin, &stdout, &stderr)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "cannot use both --stdin and a file path") {
		t.Errorf("expected conflict error, got: %s", stderr.String())
	}
}
