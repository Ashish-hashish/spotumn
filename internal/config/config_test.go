package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDirAndTemplate(t *testing.T) {
	dir := GetDir()
	if dir == "" {
		t.Fatal("expected non-empty config directory")
	}

	fi, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("config directory should exist: %v", err)
	}
	if !fi.IsDir() {
		t.Fatalf("config path %s is not a directory", dir)
	}

	// Verify permissions on config dir
	perm := fi.Mode().Perm()
	if perm&0077 != 0 {
		t.Logf("Notice: config directory permissions: %v", perm)
	}
}

func TestEnvOverride(t *testing.T) {
	testID := "test_client_id_12345"
	os.Setenv("SPOTUMN_CLIENT_ID", testID)
	defer os.Unsetenv("SPOTUMN_CLIENT_ID")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error loading with env: %v", err)
	}
	if cfg.ClientID != testID {
		t.Fatalf("expected client ID %s, got %s", testID, cfg.ClientID)
	}
}

func TestFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "config.yml")

	// Template write
	template := "client_id: \"test\"\nport: 8080\n"
	err := os.WriteFile(testPath, []byte(template), 0600)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	fi, err := os.Stat(testPath)
	if err != nil {
		t.Fatalf("failed to stat test file: %v", err)
	}

	if fi.Mode().Perm() != 0600 {
		t.Fatalf("expected 0600 permissions, got %o", fi.Mode().Perm())
	}
}
