package session

import "time"

type Session struct {
	ID        string
	Content   string
	CreatedAt time.Time
}

func Create(content string) (*Session, error) {
	return nil, nil
}

func Load(id string) (*Session, error) {
	return nil, nil
}
