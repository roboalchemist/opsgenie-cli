package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// AuthConfig holds the authentication configuration.
type AuthConfig struct {
	APIKey string `json:"api_key"`
}

// ConfigPath returns the path to the auth config file (~/.opsgenie-cli-auth.json).
func ConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".opsgenie-cli-auth.json"
	}
	return filepath.Join(home, ".opsgenie-cli-auth.json")
}

// GetAPIKey returns the OpsGenie API key from env var or config file.
// Priority: OPSGENIE_API_KEY env var → ~/.opsgenie-cli-auth.json
func GetAPIKey() (string, error) {
	if key := os.Getenv("OPSGENIE_API_KEY"); key != "" {
		return key, nil
	}

	config, err := loadAuth()
	if err != nil {
		return "", fmt.Errorf("OPSGENIE_API_KEY not set and no config file found: set OPSGENIE_API_KEY or run 'opsgenie-cli auth login'")
	}
	if config.APIKey == "" {
		return "", fmt.Errorf("no valid authentication found: config file exists but api_key is empty")
	}
	return config.APIKey, nil
}

// SaveAPIKey writes the API key to the config file with mode 0600.
func SaveAPIKey(key string) error {
	return SaveAuth(AuthConfig{APIKey: key})
}

// SaveAuth writes authentication config to the config file with mode 0600.
func SaveAuth(config AuthConfig) error {
	path := ConfigPath()
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func loadAuth() (*AuthConfig, error) {
	path := ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config AuthConfig
	return &config, json.Unmarshal(data, &config)
}
