package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvLoadsVariables(t *testing.T) {
	const key = "TAKE_HOME_EBANX_TEST_TOKEN"
	restoreEnvironment(t, key)
	path := writeEnvFile(t, key+"=test-token\n")

	if err := LoadEnv(path); err != nil {
		t.Fatalf("LoadEnv returned an unexpected error: %v", err)
	}
	if got := os.Getenv(key); got != "test-token" {
		t.Fatalf("expected loaded value %q, got %q", "test-token", got)
	}
}

func TestLoadEnvDoesNotReplaceExistingVariable(t *testing.T) {
	const key = "TAKE_HOME_EBANX_EXISTING_TOKEN"
	t.Setenv(key, "exported-token")
	path := writeEnvFile(t, key+"=file-token\n")

	if err := LoadEnv(path); err != nil {
		t.Fatalf("LoadEnv returned an unexpected error: %v", err)
	}
	if got := os.Getenv(key); got != "exported-token" {
		t.Fatalf("expected existing value to be preserved, got %q", got)
	}
}

func TestLoadEnvAcceptsMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.env")

	if err := LoadEnv(path); err != nil {
		t.Fatalf("expected missing file to be accepted, got %v", err)
	}
}

func TestLoadEnvRejectsMalformedEntry(t *testing.T) {
	path := writeEnvFile(t, "invalid-entry\n")

	if err := LoadEnv(path); err == nil {
		t.Fatal("expected malformed entry to return an error")
	}
}

func writeEnvFile(t *testing.T, contents string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write environment file: %v", err)
	}
	return path
}

func restoreEnvironment(t *testing.T, key string) {
	t.Helper()

	previousValue, existed := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset environment variable: %v", err)
	}
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(key, previousValue)
			return
		}
		_ = os.Unsetenv(key)
	})
}
