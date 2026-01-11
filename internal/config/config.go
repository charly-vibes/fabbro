package config

import "os"

const FabbroDir = ".fabbro"
const SessionsDir = ".fabbro/sessions"

func IsInitialized() bool {
	_, err := os.Stat(FabbroDir)
	return err == nil
}

func Init() error {
	return os.MkdirAll(SessionsDir, 0755)
}
