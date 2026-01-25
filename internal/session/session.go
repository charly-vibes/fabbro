package session

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charly-vibes/fabbro/internal/config"
)

type Session struct {
	ID         string
	Content    string
	CreatedAt  time.Time
	SourceFile string
}

func generateID() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	suffix := hex.EncodeToString(bytes)
	date := time.Now().Format("20060102")
	return date + "-" + suffix, nil
}

const maxCollisionRetries = 3

// normalizeSourceFile cleans and normalizes the source file path for consistent storage and lookup.
func normalizeSourceFile(path string) string {
	if path == "" {
		return ""
	}
	cleaned := filepath.Clean(path)
	return filepath.ToSlash(cleaned)
}

// quoteYAMLString escapes a string for safe YAML storage using single quotes.
func quoteYAMLString(s string) string {
	escaped := strings.ReplaceAll(s, "'", "''")
	return "'" + escaped + "'"
}

func Create(content string, sourceFile string) (*Session, error) {
	var session *Session
	var sessionPath string

	normalizedSource := normalizeSourceFile(sourceFile)

	for attempt := 0; attempt < maxCollisionRetries; attempt++ {
		id, err := generateID()
		if err != nil {
			return nil, err
		}

		sessionPath = filepath.Join(config.SessionsDir, id+".fem")

		if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
			session = &Session{
				ID:         id,
				Content:    content,
				CreatedAt:  time.Now().UTC(),
				SourceFile: normalizedSource,
			}
			break
		}
	}

	if session == nil {
		return nil, fmt.Errorf("failed to generate unique session ID after %d attempts", maxCollisionRetries)
	}

	// Build frontmatter with optional source_file
	var sourceFileLine string
	if normalizedSource != "" {
		sourceFileLine = fmt.Sprintf("source_file: %s\n", quoteYAMLString(normalizedSource))
	}

	fileContent := fmt.Sprintf(`---
session_id: %s
created_at: %s
%s---

%s`, session.ID, session.CreatedAt.Format(time.RFC3339), sourceFileLine, content)

	if err := os.WriteFile(sessionPath, []byte(fileContent), 0600); err != nil {
		return nil, fmt.Errorf("failed to write session file: %w", err)
	}

	return session, nil
}

func Load(id string) (*Session, error) {
	sessionPath := filepath.Join(config.SessionsDir, id+".fem")
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	content := string(data)

	// Parse frontmatter
	if !strings.HasPrefix(content, "---\n") {
		return nil, fmt.Errorf("invalid session file: missing frontmatter")
	}

	parts := strings.SplitN(content, "---\n", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid session file: malformed frontmatter")
	}

	frontmatter := parts[1]
	body := strings.TrimPrefix(parts[2], "\n")

	// Extract session_id, created_at, and source_file from frontmatter
	var sessionID string
	var sourceFile string
	var createdAt time.Time
	var parseErr error

	for _, line := range strings.Split(frontmatter, "\n") {
		if strings.HasPrefix(line, "session_id: ") {
			sessionID = strings.TrimPrefix(line, "session_id: ")
		}
		if strings.HasPrefix(line, "created_at: ") {
			ts := strings.TrimPrefix(line, "created_at: ")
			createdAt, parseErr = time.Parse(time.RFC3339, ts)
			if parseErr != nil {
				return nil, fmt.Errorf("invalid session file: malformed created_at: %w", parseErr)
			}
		}
		if strings.HasPrefix(line, "source_file: ") {
			sourceFile = unquoteYAMLString(strings.TrimPrefix(line, "source_file: "))
		}
	}

	if sessionID == "" {
		return nil, fmt.Errorf("invalid session file: missing session_id")
	}
	if createdAt.IsZero() {
		return nil, fmt.Errorf("invalid session file: missing created_at")
	}

	return &Session{
		ID:         sessionID,
		Content:    body,
		CreatedAt:  createdAt,
		SourceFile: sourceFile,
	}, nil
}

// unquoteYAMLString removes single quotes and unescapes doubled quotes.
func unquoteYAMLString(s string) string {
	if len(s) >= 2 && s[0] == '\'' && s[len(s)-1] == '\'' {
		s = s[1 : len(s)-1]
	}
	return strings.ReplaceAll(s, "''", "'")
}

// FindBySourceFile finds the latest session created from the given source file.
// Returns an error if no matching session is found.
func FindBySourceFile(sourceFile string) (*Session, error) {
	normalizedQuery := normalizeSourceFile(sourceFile)
	if normalizedQuery == "" {
		return nil, fmt.Errorf("no session found for empty source file")
	}

	entries, err := os.ReadDir(config.SessionsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read sessions directory: %w", err)
	}

	var latestSession *Session

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".fem") {
			continue
		}

		sessionID := strings.TrimSuffix(entry.Name(), ".fem")
		sess, err := Load(sessionID)
		if err != nil {
			continue // Skip malformed sessions
		}

		if sess.SourceFile != normalizedQuery {
			continue
		}

		if latestSession == nil || sess.CreatedAt.After(latestSession.CreatedAt) {
			latestSession = sess
		}
	}

	if latestSession == nil {
		return nil, fmt.Errorf("no session found for file: %s", sourceFile)
	}

	return latestSession, nil
}
