package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsInitialized_ReturnsFalseWhenNoFabbroDir(t *testing.T) {
	// Use a temp directory as working directory
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	if IsInitialized() {
		t.Error("expected IsInitialized() to return false when .fabbro does not exist")
	}
}

func TestIsInitialized_ReturnsTrueWhenFullyInitialized(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Create both .fabbro and .fabbro/sessions directories
	os.MkdirAll(SessionsDir, 0755)

	if !IsInitialized() {
		t.Error("expected IsInitialized() to return true when fully initialized")
	}
}

func TestIsInitialized_ReturnsFalseWhenSessionsDirMissing(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Create only .fabbro but not sessions
	os.Mkdir(FabbroDir, 0755)

	if IsInitialized() {
		t.Error("expected IsInitialized() to return false when sessions dir missing")
	}
}

func TestInit_CreatesFabbroSessionsDir(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	err := Init()
	if err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}

	// Check .fabbro/sessions exists
	sessionsPath := filepath.Join(FabbroDir, "sessions")
	info, err := os.Stat(sessionsPath)
	if err != nil {
		t.Fatalf("expected %s to exist, got error: %v", sessionsPath, err)
	}
	if !info.IsDir() {
		t.Errorf("expected %s to be a directory", sessionsPath)
	}
}

func TestInit_IsIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// First init
	err := Init()
	if err != nil {
		t.Fatalf("first Init() returned error: %v", err)
	}

	// Second init should not error
	err = Init()
	if err != nil {
		t.Fatalf("second Init() returned error: %v", err)
	}
}

func TestInit_CreatesPrivateDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	Init()

	// Sessions dir should be 0700 (owner only)
	info, _ := os.Stat(SessionsDir)
	mode := info.Mode().Perm()
	if mode != 0700 {
		t.Errorf("expected SessionsDir permissions 0700, got %04o", mode)
	}
}
