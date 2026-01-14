package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func TestLoad_ReturnsErrorForMissingSessionID(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	sessionFile := filepath.Join(config.SessionsDir, "no-id.fem")
	os.WriteFile(sessionFile, []byte("---\ncreated_at: 2026-01-14T10:00:00Z\n---\ncontent"), 0644)

	_, err := Load("no-id")
	if err == nil {
		t.Error("expected Load() to return error for missing session_id")
	}
	if !strings.Contains(err.Error(), "missing session_id") {
		t.Errorf("expected 'missing session_id' error, got: %v", err)
	}
}

func TestLoad_ReturnsErrorForMissingCreatedAt(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	sessionFile := filepath.Join(config.SessionsDir, "no-time.fem")
	os.WriteFile(sessionFile, []byte("---\nsession_id: test-123\n---\ncontent"), 0644)

	_, err := Load("no-time")
	if err == nil {
		t.Error("expected Load() to return error for missing created_at")
	}
	if !strings.Contains(err.Error(), "missing created_at") {
		t.Errorf("expected 'missing created_at' error, got: %v", err)
	}
}

func TestLoad_ReturnsErrorForInvalidCreatedAt(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	sessionFile := filepath.Join(config.SessionsDir, "bad-time.fem")
	os.WriteFile(sessionFile, []byte("---\nsession_id: test-123\ncreated_at: not-a-timestamp\n---\ncontent"), 0644)

	_, err := Load("bad-time")
	if err == nil {
		t.Error("expected Load() to return error for invalid created_at")
	}
	if !strings.Contains(err.Error(), "malformed created_at") {
		t.Errorf("expected 'malformed created_at' error, got: %v", err)
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
		id, err := generateID()
		if err != nil {
			t.Fatalf("generateID() returned error: %v", err)
		}
		if ids[id] {
			t.Errorf("generateID() returned duplicate ID: %s", id)
		}
		ids[id] = true
	}
}

func TestGenerateID_ReturnsDateBasedFormat(t *testing.T) {
	id, err := generateID()
	if err != nil {
		t.Fatalf("generateID() returned error: %v", err)
	}

	// Format should be: YYYYMMDD-xxxxxxxxxxxxxxxx (8 + 1 + 16 = 25 chars)
	if len(id) != 25 {
		t.Errorf("expected 25 char ID (YYYYMMDD-xxxxxxxxxxxxxxxx), got %d chars: %s", len(id), id)
	}

	// Should contain a hyphen at position 8
	if id[8] != '-' {
		t.Errorf("expected hyphen at position 8, got: %s", id)
	}

	// First 8 chars should be today's date in YYYYMMDD format
	datePrefix := id[:8]
	today := time.Now().Format("20060102")
	if datePrefix != today {
		t.Errorf("expected date prefix %s, got %s", today, datePrefix)
	}

	// Last 16 chars should be hex (8 bytes = 16 hex chars)
	suffix := id[9:]
	if len(suffix) != 16 {
		t.Errorf("expected 16 char suffix, got %d: %s", len(suffix), suffix)
	}
	for _, c := range suffix {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("expected hex chars in suffix, got: %s", suffix)
			break
		}
	}
}
