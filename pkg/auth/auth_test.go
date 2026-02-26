package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// setHome overrides HOME so ConfigPath() points to a temp directory.
// Returns a cleanup function that restores the original HOME.
func setHome(t *testing.T, dir string) {
	t.Helper()
	t.Setenv("HOME", dir)
}

func TestGetAPIKey_EnvVar(t *testing.T) {
	t.Setenv("OPSGENIE_API_KEY", "test-key-from-env")

	key, err := GetAPIKey()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "test-key-from-env" {
		t.Errorf("expected 'test-key-from-env', got %q", key)
	}
}

func TestGetAPIKey_ConfigFile(t *testing.T) {
	// Clear env var to ensure config file path is exercised.
	t.Setenv("OPSGENIE_API_KEY", "")

	tmp := t.TempDir()
	setHome(t, tmp)

	// Write a config file.
	cfg := AuthConfig{APIKey: "test-key-from-file"}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	cfgPath := filepath.Join(tmp, ".opsgenie-cli-auth.json")
	if err := os.WriteFile(cfgPath, data, 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	key, err := GetAPIKey()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "test-key-from-file" {
		t.Errorf("expected 'test-key-from-file', got %q", key)
	}
}

func TestGetAPIKey_NeitherFound(t *testing.T) {
	t.Setenv("OPSGENIE_API_KEY", "")

	tmp := t.TempDir()
	setHome(t, tmp)
	// No config file written — should error.

	_, err := GetAPIKey()
	if err == nil {
		t.Fatal("expected error when neither env var nor config file is present")
	}
}

func TestGetAPIKey_EnvVarTakesPrecedence(t *testing.T) {
	t.Setenv("OPSGENIE_API_KEY", "env-wins")

	tmp := t.TempDir()
	setHome(t, tmp)

	// Write a config file with a different key.
	cfg := AuthConfig{APIKey: "file-key"}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	cfgPath := filepath.Join(tmp, ".opsgenie-cli-auth.json")
	if err := os.WriteFile(cfgPath, data, 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	key, err := GetAPIKey()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "env-wins" {
		t.Errorf("expected env var to win, got %q", key)
	}
}

func TestSaveAPIKey_CreatesFileWithCorrectPermissions(t *testing.T) {
	t.Setenv("OPSGENIE_API_KEY", "")

	tmp := t.TempDir()
	setHome(t, tmp)

	if err := SaveAPIKey("my-saved-key"); err != nil {
		t.Fatalf("SaveAPIKey: %v", err)
	}

	cfgPath := filepath.Join(tmp, ".opsgenie-cli-auth.json")
	info, err := os.Stat(cfgPath)
	if err != nil {
		t.Fatalf("stat config file: %v", err)
	}

	// Verify permissions are 0600.
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("expected file permissions 0600, got %04o", perm)
	}

	// Verify the key was written correctly.
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read config file: %v", err)
	}
	var cfg AuthConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	if cfg.APIKey != "my-saved-key" {
		t.Errorf("expected 'my-saved-key', got %q", cfg.APIKey)
	}
}

func TestConfigPath_ReturnsExpectedPath(t *testing.T) {
	tmp := t.TempDir()
	setHome(t, tmp)

	path := ConfigPath()
	expected := filepath.Join(tmp, ".opsgenie-cli-auth.json")
	if path != expected {
		t.Errorf("expected %q, got %q", expected, path)
	}
}
