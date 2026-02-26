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

// GetAPIKey returns the OpsGenie API key from env var or config file.
// Priority: OPSGENIE_API_KEY env var → ~/.opsgenie-cli-auth.json
func GetAPIKey() (string, error) {
	if key := os.Getenv("OPSGENIE_API_KEY"); key != "" {
		return key, nil
	}

	config, err := loadAuth()
	if err != nil {
		return "", fmt.Errorf("OPSGENIE_API_KEY not set and no config file found")
	}
	if config.APIKey != "" {
		return config.APIKey, nil
	}
	return "", fmt.Errorf("no valid authentication found")
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".opsgenie-cli-auth.json"), nil
}

func loadAuth() (*AuthConfig, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config AuthConfig
	return &config, json.Unmarshal(data, &config)
}

// SaveAuth writes authentication config to the config file.
func SaveAuth(config AuthConfig) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
