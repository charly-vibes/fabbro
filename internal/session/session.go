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

func generateID() string {
	bytes := make([]byte, 2)
	rand.Read(bytes)
	suffix := hex.EncodeToString(bytes)
	date := time.Now().Format("20060102")
	return date + "-" + suffix
}

func Create(content string) (*Session, error) {
	session := &Session{
		ID:        generateID(),
		Content:   content,
		CreatedAt: time.Now().UTC(),
	}

	fileContent := fmt.Sprintf(`---
session_id: %s
created_at: %s
---

%s`, session.ID, session.CreatedAt.Format(time.RFC3339), content)

	sessionPath := filepath.Join(config.SessionsDir, session.ID+".fem")
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

	for _, line := range strings.Split(frontmatter, "\n") {
		if strings.HasPrefix(line, "session_id: ") {
			sessionID = strings.TrimPrefix(line, "session_id: ")
		}
		if strings.HasPrefix(line, "created_at: ") {
			ts := strings.TrimPrefix(line, "created_at: ")
			createdAt, _ = time.Parse(time.RFC3339, ts)
		}
	}

	return &Session{
		ID:        sessionID,
		Content:   body,
		CreatedAt: createdAt,
	}, nil
}
