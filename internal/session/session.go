package session

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/charly-vibes/fabbro/internal/config"
)

var validSessionID = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)
var reservedIDs = map[string]bool{"tutor": true, "_tutor_": true}

// ValidateSessionID checks that a custom session ID is valid.
func ValidateSessionID(id string) error {
	if !validSessionID.MatchString(id) {
		return fmt.Errorf("invalid session ID %q: must be 1-64 alphanumeric characters, dash, or underscore", id)
	}
	if reservedIDs[id] {
		return fmt.Errorf("session ID %q is reserved", id)
	}
	return nil
}

type Session struct {
	ID          string
	Content     string
	CreatedAt   time.Time
	SourceFile  string
	ContentHash string
}

func computeHash(content string) string {
	h := sha256.Sum256([]byte(content))
	return hex.EncodeToString(h[:])
}

// VerifySourceHash checks if the source file content still matches the hash stored at session creation.
// Returns true if the hash matches or if verification is not applicable (stdin sessions).
func (s *Session) VerifySourceHash() (bool, error) {
	if s.SourceFile == "" || s.ContentHash == "" {
		return true, nil
	}

	data, err := os.ReadFile(s.SourceFile)
	if err != nil {
		return false, fmt.Errorf("source file not found: %s", s.SourceFile)
	}

	return computeHash(string(data)) == s.ContentHash, nil
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

// CreateWithID creates a session with a specific custom ID.
func CreateWithID(id string, content string, sourceFile string) (*Session, error) {
	if err := ValidateSessionID(id); err != nil {
		return nil, err
	}

	sessionsDir, err := config.GetSessionsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	sessionPath := filepath.Join(sessionsDir, id+".fem")
	if _, err := os.Stat(sessionPath); err == nil {
		return nil, fmt.Errorf("session ID %q already exists", id)
	}

	normalizedSource := normalizeSourceFile(sourceFile)
	sess := &Session{
		ID:          id,
		Content:     content,
		CreatedAt:   time.Now().UTC(),
		SourceFile:  normalizedSource,
		ContentHash: computeHash(content),
	}

	return writeSession(sess, sessionPath, content)
}

func Create(content string, sourceFile string) (*Session, error) {
	sessionsDir, err := config.GetSessionsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	var session *Session
	var sessionPath string

	normalizedSource := normalizeSourceFile(sourceFile)

	for attempt := 0; attempt < maxCollisionRetries; attempt++ {
		id, err := generateID()
		if err != nil {
			return nil, err
		}

		sessionPath = filepath.Join(sessionsDir, id+".fem")

		if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
			session = &Session{
				ID:          id,
				Content:     content,
				CreatedAt:   time.Now().UTC(),
				SourceFile:  normalizedSource,
				ContentHash: computeHash(content),
			}
			break
		}
	}

	if session == nil {
		return nil, fmt.Errorf("failed to generate unique session ID after %d attempts", maxCollisionRetries)
	}

	return writeSession(session, sessionPath, content)
}

func writeSession(sess *Session, sessionPath string, content string) (*Session, error) {
	var sourceFileLine string
	if sess.SourceFile != "" {
		sourceFileLine = fmt.Sprintf("source_file: %s\n", quoteYAMLString(sess.SourceFile))
	}

	fileContent := fmt.Sprintf(`---
session_id: %s
created_at: %s
content_hash: %s
%s---

%s`, sess.ID, sess.CreatedAt.Format(time.RFC3339), sess.ContentHash, sourceFileLine, content)

	if err := os.WriteFile(sessionPath, []byte(fileContent), 0600); err != nil {
		return nil, fmt.Errorf("failed to write session file: %w", err)
	}

	return sess, nil
}

func Load(id string) (*Session, error) {
	sessionsDir, err := config.GetSessionsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}
	sessionPath := filepath.Join(sessionsDir, id+".fem")
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

	// Extract session_id, created_at, source_file, and content_hash from frontmatter
	var sessionID string
	var sourceFile string
	var contentHash string
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
		if strings.HasPrefix(line, "content_hash: ") {
			contentHash = strings.TrimPrefix(line, "content_hash: ")
		}
	}

	if sessionID == "" {
		return nil, fmt.Errorf("invalid session file: missing session_id")
	}
	if createdAt.IsZero() {
		return nil, fmt.Errorf("invalid session file: missing created_at")
	}

	return &Session{
		ID:          sessionID,
		Content:     body,
		CreatedAt:   createdAt,
		SourceFile:  sourceFile,
		ContentHash: contentHash,
	}, nil
}

// unquoteYAMLString removes single quotes and unescapes doubled quotes.
func unquoteYAMLString(s string) string {
	if len(s) >= 2 && s[0] == '\'' && s[len(s)-1] == '\'' {
		s = s[1 : len(s)-1]
	}
	return strings.ReplaceAll(s, "''", "'")
}

// List returns all sessions sorted by creation date (newest first).
func List() ([]*Session, error) {
	sessionsDir, err := config.GetSessionsDir()
	if err != nil {
		return []*Session{}, nil
	}
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Session{}, nil
		}
		return nil, fmt.Errorf("failed to read sessions directory: %w", err)
	}

	var sessions []*Session
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".fem") {
			continue
		}

		sessionID := strings.TrimSuffix(entry.Name(), ".fem")
		sess, err := Load(sessionID)
		if err != nil {
			continue
		}
		sessions = append(sessions, sess)
	}

	// Sort by creation date, newest first
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].CreatedAt.After(sessions[j].CreatedAt)
	})

	return sessions, nil
}

// LoadPartial loads a session by exact or prefix match on the ID.
// Returns an error if no match or multiple matches are found.
func LoadPartial(partial string) (*Session, error) {
	sessions, err := List()
	if err != nil {
		return nil, err
	}

	// Exact match first
	for _, sess := range sessions {
		if sess.ID == partial {
			return sess, nil
		}
	}

	// Prefix match
	var matches []string
	for _, sess := range sessions {
		if strings.HasPrefix(sess.ID, partial) {
			matches = append(matches, sess.ID)
		}
	}

	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("no session matching '%s'", partial)
	case 1:
		return Load(matches[0])
	default:
		return nil, fmt.Errorf("ambiguous session ID '%s' matches: %s",
			partial, strings.Join(matches, ", "))
	}
}

// Delete removes a session file by ID.
func Delete(id string) error {
	sessionsDir, err := config.GetSessionsDir()
	if err != nil {
		return fmt.Errorf("failed to find project root: %w", err)
	}
	sessionPath := filepath.Join(sessionsDir, id+".fem")
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		return fmt.Errorf("session not found: %s", id)
	}
	return os.Remove(sessionPath)
}

// FindBySourceFile finds the latest session created from the given source file.
// Returns an error if no matching session is found.
func FindBySourceFile(sourceFile string) (*Session, error) {
	normalizedQuery := normalizeSourceFile(sourceFile)
	if normalizedQuery == "" {
		return nil, fmt.Errorf("no session found for empty source file")
	}

	sessionsDir, err := config.GetSessionsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}
	entries, err := os.ReadDir(sessionsDir)
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
