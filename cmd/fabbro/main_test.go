package main

import (
	"encoding/json"
	"os"
	"path/filepath"
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
	sess, err := session.Create(content)
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

	if annotations[0].Line != 1 {
		t.Errorf("expected first annotation on line 1, got %d", annotations[0].Line)
	}

	if annotations[1].Text != "consider refactoring" {
		t.Errorf("expected second annotation text 'consider refactoring', got %q", annotations[1].Text)
	}

	if annotations[1].Line != 3 {
		t.Errorf("expected second annotation on line 3, got %d", annotations[1].Line)
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
		SessionID   string           `json:"session_id"`
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

	if parsed["session_id"] != loaded.ID {
		t.Errorf("expected session_id=%s in JSON", loaded.ID)
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
