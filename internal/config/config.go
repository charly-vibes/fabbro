package config

import (
	"errors"
	"os"
	"path/filepath"
)

const FabbroDir = ".fabbro"
const SessionsDir = ".fabbro/sessions"

var ErrNotInitialized = errors.New("not a fabbro project (no .fabbro directory found)")

func FindProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		sessionsPath := filepath.Join(dir, SessionsDir)
		if info, err := os.Stat(sessionsPath); err == nil && info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrNotInitialized
		}
		dir = parent
	}
}

func IsInitialized() bool {
	_, err := FindProjectRoot()
	return err == nil
}

func GetSessionsDir() (string, error) {
	root, err := FindProjectRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, SessionsDir), nil
}

func Init() error {
	return os.MkdirAll(SessionsDir, 0700)
}
