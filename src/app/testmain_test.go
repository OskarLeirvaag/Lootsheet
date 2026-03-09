package app

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if os.Getenv(logLevelEnvVar) == "" {
		os.Setenv(logLevelEnvVar, "ERROR")
	}
	os.Exit(m.Run())
}
