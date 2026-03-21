package app

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if os.Getenv(logLevelEnvVar) == "" {
		_ = os.Setenv(logLevelEnvVar, "ERROR")
	}
	m.Run()
}
