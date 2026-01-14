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
	ID        string
	Content   string
	CreatedAt time.Time
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

func Create(content string) (*Session, error) {
	var session *Session
	var sessionPath string

	for attempt := 0; attempt < maxCollisionRetries; attempt++ {
		id, err := generateID()
		if err != nil {
			return nil, err
		}

		sessionPath = filepath.Join(config.SessionsDir, id+".fem")

		if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
			session = &Session{
				ID:        id,
				Content:   content,
				CreatedAt: time.Now().UTC(),
			}
			break
		}
	}

	if session == nil {
		return nil, fmt.Errorf("failed to generate unique session ID after %d attempts", maxCollisionRetries)
	}

	fileContent := fmt.Sprintf(`---
session_id: %s
created_at: %s
---

%s`, session.ID, session.CreatedAt.Format(time.RFC3339), content)

	if err := os.WriteFile(sessionPath, []byte(fileContent), 0644); err != nil {
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

	// Extract session_id and created_at from frontmatter
	var sessionID string
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
	}

	if sessionID == "" {
		return nil, fmt.Errorf("invalid session file: missing session_id")
	}
	if createdAt.IsZero() {
		return nil, fmt.Errorf("invalid session file: missing created_at")
	}

	return &Session{
		ID:        sessionID,
		Content:   body,
		CreatedAt: createdAt,
	}, nil
}
