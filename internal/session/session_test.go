package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charly-vibes/fabbro/internal/config"
)

func TestCreate_CreatesSessionFile(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Initialize fabbro first
	config.Init()

	content := "Hello, world!"
	session, err := Create(content)
	if err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}

	if session.ID == "" {
		t.Error("expected session ID to be set")
	}

	if session.Content != content {
		t.Errorf("expected Content=%q, got %q", content, session.Content)
	}

	// Check file exists
	sessionFile := filepath.Join(config.SessionsDir, session.ID+".fem")
	if _, err := os.Stat(sessionFile); err != nil {
		t.Fatalf("expected session file %s to exist: %v", sessionFile, err)
	}
}

func TestCreate_FileContainsFrontmatterAndContent(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	content := "Line one\nLine two\nLine three"
	session, _ := Create(content)

	sessionFile := filepath.Join(config.SessionsDir, session.ID+".fem")
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		t.Fatalf("failed to read session file: %v", err)
	}

	fileContent := string(data)

	// Check frontmatter markers
	if !strings.HasPrefix(fileContent, "---\n") {
		t.Error("expected file to start with ---")
	}

	if !strings.Contains(fileContent, "session_id: "+session.ID) {
		t.Error("expected file to contain session_id in frontmatter")
	}

	if !strings.Contains(fileContent, "created_at:") {
		t.Error("expected file to contain created_at in frontmatter")
	}

	// Check content after frontmatter
	if !strings.Contains(fileContent, content) {
		t.Error("expected file to contain original content")
	}
}

func TestLoad_LoadsExistingSession(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	content := "Test content for loading"
	created, _ := Create(content)

	loaded, err := Load(created.ID)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if loaded.ID != created.ID {
		t.Errorf("expected ID=%s, got %s", created.ID, loaded.ID)
	}

	if loaded.Content != content {
		t.Errorf("expected Content=%q, got %q", content, loaded.Content)
	}
}

func TestLoad_ReturnsErrorForNonexistentSession(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	_, err := Load("nonexistent")
	if err == nil {
		t.Error("expected Load() to return error for nonexistent session")
	}
}

func TestLoad_ReturnsErrorForMissingFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	// Write a file without frontmatter
	sessionFile := filepath.Join(config.SessionsDir, "bad.fem")
	os.WriteFile(sessionFile, []byte("no frontmatter here"), 0644)

	_, err := Load("bad")
	if err == nil {
		t.Error("expected Load() to return error for missing frontmatter")
	}
	if !strings.Contains(err.Error(), "missing frontmatter") {
		t.Errorf("expected 'missing frontmatter' error, got: %v", err)
	}
}

func TestLoad_ReturnsErrorForMalformedFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	// Write a file with incomplete frontmatter (only one ---)
	sessionFile := filepath.Join(config.SessionsDir, "malformed.fem")
	os.WriteFile(sessionFile, []byte("---\nsession_id: test\n"), 0644)

	_, err := Load("malformed")
	if err == nil {
		t.Error("expected Load() to return error for malformed frontmatter")
	}
	if !strings.Contains(err.Error(), "malformed frontmatter") {
		t.Errorf("expected 'malformed frontmatter' error, got: %v", err)
	}
}

func TestCreate_ReturnsErrorWhenSessionsDirMissing(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Don't initialize - sessions dir doesn't exist
	_, err := Create("content")
	if err == nil {
		t.Error("expected Create() to return error when sessions dir missing")
	}
}

func TestGenerateID_ReturnsUniqueIDs(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := generateID()
		if ids[id] {
			t.Errorf("generateID() returned duplicate ID: %s", id)
		}
		ids[id] = true
	}
}

func TestGenerateID_Returns8CharHex(t *testing.T) {
	id := generateID()
	if len(id) != 8 {
		t.Errorf("expected 8 char ID, got %d chars: %s", len(id), id)
	}
	for _, c := range id {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("expected hex chars only, got: %s", id)
			break
		}
	}
}
