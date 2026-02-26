package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

var version = "dev"

const (
	baseURLUS = "https://api.opsgenie.com"
	baseURLEU = "https://api.eu.opsgenie.com"
)

// Client is the OpsGenie API client.
type Client struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
	debug      func(string, ...interface{})
}

// NewClient creates a new OpsGenie API client.
// region should be "us" or "eu". The OPSGENIE_API_URL env var overrides the base URL.
func NewClient(apiKey, region, debugFlag string) *Client {
	baseURL := baseURLUS
	if region == "eu" {
		baseURL = baseURLEU
	}
	if override := os.Getenv("OPSGENIE_API_URL"); override != "" {
		baseURL = override
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey:  apiKey,
		baseURL: baseURL,
	}
}

// SetDebug sets the debug logging function.
func (c *Client) SetDebug(fn func(string, ...interface{})) {
	c.debug = fn
}

func (c *Client) debugLog(format string, args ...interface{}) {
	if c.debug != nil {
		c.debug(format, args...)
	}
}

func (c *Client) do(method, endpoint string, body, result interface{}) error {
	url := fmt.Sprintf("%s/%s", c.baseURL, endpoint)

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		c.debugLog("%s %s body=%s", method, url, string(jsonBody))
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		c.debugLog("%s %s", method, url)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "GenieKey "+c.apiKey)
	req.Header.Set("User-Agent", "opsgenie-cli/"+version)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	c.debugLog("Response status: %d", resp.StatusCode)
	c.debugLog("Response body: %s", truncate(string(respBody), 2000))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, truncate(string(respBody), 500))
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}
	}

	return nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
