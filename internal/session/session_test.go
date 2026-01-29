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
	session, err := Create(content, "")
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
	session, _ := Create(content, "")

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
	created, _ := Create(content, "")

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

	// Prevent FindProjectRoot from walking up to the real project
	t.Setenv("FABBRO_PROJECT_ROOT_STOP", tmpDir)

	// Don't initialize - sessions dir doesn't exist
	_, err := Create("content", "")
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

func TestCreate_StoresSourceFileInFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	content := "Test content"
	sourceFile := "plans/my-plan.md"
	sess, err := Create(content, sourceFile)
	if err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}

	if sess.SourceFile != sourceFile {
		t.Errorf("expected SourceFile=%q, got %q", sourceFile, sess.SourceFile)
	}

	// Check file contains source_file in frontmatter
	sessionFile := filepath.Join(config.SessionsDir, sess.ID+".fem")
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		t.Fatalf("failed to read session file: %v", err)
	}

	if !strings.Contains(string(data), "source_file: 'plans/my-plan.md'") {
		t.Errorf("expected frontmatter to contain source_file, got:\n%s", string(data))
	}
}

func TestCreate_OmitsSourceFileWhenEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	sess, _ := Create("content", "")

	sessionFile := filepath.Join(config.SessionsDir, sess.ID+".fem")
	data, _ := os.ReadFile(sessionFile)

	if strings.Contains(string(data), "source_file:") {
		t.Errorf("expected frontmatter to NOT contain source_file for stdin, got:\n%s", string(data))
	}
}

func TestLoad_ParsesSourceFile(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	sourceFile := "docs/readme.md"
	created, _ := Create("content", sourceFile)

	loaded, err := Load(created.ID)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if loaded.SourceFile != sourceFile {
		t.Errorf("expected SourceFile=%q, got %q", sourceFile, loaded.SourceFile)
	}
}

func TestLoad_HandlesLegacySessionsWithoutSourceFile(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	// Write legacy session file without source_file
	sessionFile := filepath.Join(config.SessionsDir, "legacy-session.fem")
	os.WriteFile(sessionFile, []byte("---\nsession_id: legacy-session\ncreated_at: 2026-01-14T10:00:00Z\n---\ncontent"), 0644)

	loaded, err := Load("legacy-session")
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if loaded.SourceFile != "" {
		t.Errorf("expected empty SourceFile for legacy session, got %q", loaded.SourceFile)
	}
}

func TestFindBySourceFile_ReturnsMatchingSession(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	sourceFile := "plans/test-plan.md"
	created, _ := Create("content", sourceFile)

	found, err := FindBySourceFile(sourceFile)
	if err != nil {
		t.Fatalf("FindBySourceFile() returned error: %v", err)
	}

	if found.ID != created.ID {
		t.Errorf("expected ID=%s, got %s", created.ID, found.ID)
	}
}

func TestFindBySourceFile_ReturnsLatestSession(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	sourceFile := "doc.md"

	// Create older session
	older, _ := Create("old content", sourceFile)
	time.Sleep(1100 * time.Millisecond) // Ensure different timestamps (RFC3339 is second-precision)

	// Create newer session
	newer, _ := Create("new content", sourceFile)

	found, err := FindBySourceFile(sourceFile)
	if err != nil {
		t.Fatalf("FindBySourceFile() returned error: %v", err)
	}

	if found.ID == older.ID {
		t.Errorf("expected latest session %s, got older %s", newer.ID, found.ID)
	}
	if found.ID != newer.ID {
		t.Errorf("expected ID=%s, got %s", newer.ID, found.ID)
	}
}

func TestFindBySourceFile_ReturnsErrorWhenNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	_, err := FindBySourceFile("nonexistent.md")
	if err == nil {
		t.Error("expected FindBySourceFile() to return error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "no session found") {
		t.Errorf("expected 'no session found' error, got: %v", err)
	}
}

func TestFindBySourceFile_NormalizesPath(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	// Create session with one path format
	Create("content", "docs/readme.md")

	// Find with slightly different format
	found, err := FindBySourceFile("./docs/readme.md")
	if err != nil {
		t.Fatalf("FindBySourceFile() should normalize path, got error: %v", err)
	}

	if found.SourceFile != "docs/readme.md" {
		t.Errorf("expected SourceFile='docs/readme.md', got %q", found.SourceFile)
	}
}

func TestList_ReturnsEmptyForNoSessions(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	sessions, err := List()
	if err != nil {
		t.Fatalf("List() returned error: %v", err)
	}

	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(sessions))
	}
}

func TestList_ReturnsAllSessions(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	Create("content1", "file1.go")
	Create("content2", "file2.go")
	Create("content3", "")

	sessions, err := List()
	if err != nil {
		t.Fatalf("List() returned error: %v", err)
	}

	if len(sessions) != 3 {
		t.Errorf("expected 3 sessions, got %d", len(sessions))
	}
}

func TestList_SortsNewestFirst(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	older, _ := Create("old", "old.go")
	time.Sleep(1100 * time.Millisecond)
	newer, _ := Create("new", "new.go")

	sessions, err := List()
	if err != nil {
		t.Fatalf("List() returned error: %v", err)
	}

	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}

	if sessions[0].ID != newer.ID {
		t.Errorf("expected newest first, got %s (expected %s)", sessions[0].ID, newer.ID)
	}
	if sessions[1].ID != older.ID {
		t.Errorf("expected oldest second, got %s (expected %s)", sessions[1].ID, older.ID)
	}
}

func TestList_SkipsMalformedSessions(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	config.Init()

	Create("valid", "valid.go")

	// Write malformed session file
	sessionFile := filepath.Join(config.SessionsDir, "malformed.fem")
	os.WriteFile(sessionFile, []byte("no frontmatter"), 0644)

	sessions, err := List()
	if err != nil {
		t.Fatalf("List() returned error: %v", err)
	}

	if len(sessions) != 1 {
		t.Errorf("expected 1 valid session, got %d", len(sessions))
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
