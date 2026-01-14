package config

import "os"

const FabbroDir = ".fabbro"
const SessionsDir = ".fabbro/sessions"

func IsInitialized() bool {
	if _, err := os.Stat(FabbroDir); err != nil {
		return false
	}
	if _, err := os.Stat(SessionsDir); err != nil {
		return false
	}
	return true
}

func Init() error {
	return os.MkdirAll(SessionsDir, 0700)
}
