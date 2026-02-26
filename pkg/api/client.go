package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var version = "dev"

const (
	baseURLUS       = "https://api.opsgenie.com"
	baseURLEU       = "https://api.eu.opsgenie.com"
	defaultTimeout  = 30 * time.Second
	maxRetries      = 3
	pollInterval    = time.Second
	maxPollDuration = 30 * time.Second
)

// Client is the OpsGenie API client.
type Client struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
	debug      bool
}

// NewClient creates a new OpsGenie API client.
// region should be "us" or "eu". The OPSGENIE_API_URL env var overrides the base URL.
func NewClient(apiKey, region string, debug bool) *Client {
	baseURL := baseURLUS
	if region == "eu" {
		baseURL = baseURLEU
	}
	if override := os.Getenv("OPSGENIE_API_URL"); override != "" {
		baseURL = override
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		apiKey:  apiKey,
		baseURL: baseURL,
		debug:   debug,
	}
}

func (c *Client) debugLog(format string, args ...interface{}) {
	if c.debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] "+format+"\n", args...)
	}
}

// buildURL constructs the full request URL. path should start with /v2/... or /v1/...
func (c *Client) buildURL(path string) string {
	return c.baseURL + path
}

// doRequest performs a single HTTP request with auth headers and returns the raw response.
func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, []byte, error) {
	fullURL := c.buildURL(path)

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, nil, fmt.Errorf("marshal request: %w", err)
		}
		c.debugLog("%s %s body=%s", method, fullURL, string(jsonBody))
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		c.debugLog("%s %s", method, fullURL)
	}

	req, err := http.NewRequest(method, fullURL, reqBody)
	if err != nil {
		return nil, nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "GenieKey "+c.apiKey)
	req.Header.Set("User-Agent", "opsgenie-cli/"+version)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read response: %w", err)
	}

	c.debugLog("Response status: %d", resp.StatusCode)
	c.debugLog("X-RateLimit-Remaining: %s / %s",
		resp.Header.Get("X-RateLimit-Remaining"),
		resp.Header.Get("X-RateLimit-Limit"))
	c.debugLog("Response body: %s", truncate(string(respBody), 2000))

	return resp, respBody, nil
}

// do executes an HTTP request with rate-limit retry logic.
func (c *Client) do(method, path string, body, result interface{}) error {
	var lastErr error
	backoff := time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			c.debugLog("Retry %d/%d after %s (rate limited)", attempt, maxRetries, backoff)
			time.Sleep(backoff)
			backoff *= 2
		}

		resp, respBody, err := c.doRequest(method, path, body)
		if err != nil {
			return err
		}

		// Rate limited — backoff and retry
		if resp.StatusCode == http.StatusTooManyRequests {
			lastErr = fmt.Errorf("rate limited (429)")
			continue
		}

		// Async accepted — poll for completion
		if resp.StatusCode == http.StatusAccepted {
			return c.pollRequestResult(respBody, result)
		}

		// Error response
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return parseErrorResponse(resp.StatusCode, respBody)
		}

		// Success — no body expected
		if result == nil || len(respBody) == 0 {
			return nil
		}

		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}
		return nil
	}

	return fmt.Errorf("exceeded %d retries: %w", maxRetries, lastErr)
}

// Get performs a GET request and decodes the response into result.
func (c *Client) Get(path string, result interface{}) error {
	return c.do(http.MethodGet, path, nil, result)
}

// Post performs a POST request with a body and decodes the response into result.
func (c *Client) Post(path string, body, result interface{}) error {
	return c.do(http.MethodPost, path, body, result)
}

// Put performs a PUT request with a body and decodes the response into result.
func (c *Client) Put(path string, body, result interface{}) error {
	return c.do(http.MethodPut, path, body, result)
}

// Patch performs a PATCH request with a body and decodes the response into result.
func (c *Client) Patch(path string, body, result interface{}) error {
	return c.do(http.MethodPatch, path, body, result)
}

// Delete performs a DELETE request and decodes the response into result (may be nil).
func (c *Client) Delete(path string, result interface{}) error {
	return c.do(http.MethodDelete, path, nil, result)
}

// ListAll follows offset-based pagination and returns combined raw data pages.
// It calls the provided path repeatedly, following paging.next until exhausted.
// Each page's raw "data" JSON is appended to a combined JSON array in result.
// result must be a pointer to a json.RawMessage or slice that can accept unmarshalled arrays.
func (c *Client) ListAll(path string, params url.Values, result interface{}) error {
	if params == nil {
		params = url.Values{}
	}

	// Set a reasonable default page size if not provided
	if params.Get("limit") == "" {
		params.Set("limit", "100")
	}

	type pageEnvelope struct {
		Data   json.RawMessage `json:"data"`
		Paging *Paging         `json:"paging,omitempty"`
	}

	var allItems []json.RawMessage
	nextPath := path + "?" + params.Encode()

	for nextPath != "" {
		c.debugLog("ListAll fetching: %s", nextPath)

		resp, respBody, err := c.doRequest(http.MethodGet, nextPath, nil)
		if err != nil {
			return err
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			return fmt.Errorf("rate limited during pagination")
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return parseErrorResponse(resp.StatusCode, respBody)
		}

		var page pageEnvelope
		if err := json.Unmarshal(respBody, &page); err != nil {
			return fmt.Errorf("parse page: %w", err)
		}

		// page.Data may be an array or a single object
		if len(page.Data) > 0 {
			// If it's an array, unwrap and extend; otherwise append as single item
			if page.Data[0] == '[' {
				var items []json.RawMessage
				if err := json.Unmarshal(page.Data, &items); err != nil {
					return fmt.Errorf("unmarshal page data: %w", err)
				}
				allItems = append(allItems, items...)
			} else {
				allItems = append(allItems, page.Data)
			}
		}

		// Follow next page if available
		nextPath = ""
		if page.Paging != nil && page.Paging.Next != "" {
			// next is an absolute URL; extract just the path+query
			parsed, err := url.Parse(page.Paging.Next)
			if err == nil {
				nextPath = parsed.Path + "?" + parsed.RawQuery
			}
		}
	}

	// Marshal combined array and unmarshal into result
	combined, err := json.Marshal(allItems)
	if err != nil {
		return fmt.Errorf("marshal combined results: %w", err)
	}
	if err := json.Unmarshal(combined, result); err != nil {
		return fmt.Errorf("decode combined results: %w", err)
	}
	return nil
}

// pollRequestResult extracts a requestId from a 202 response body and polls until completion.
func (c *Client) pollRequestResult(body []byte, result interface{}) error {
	var asyncResp struct {
		RequestID string `json:"requestId"`
		Result    string `json:"result,omitempty"`
	}
	if err := json.Unmarshal(body, &asyncResp); err != nil || asyncResp.RequestID == "" {
		// No requestId — treat as success, try to decode body into result if provided
		if result != nil && len(body) > 0 {
			_ = json.Unmarshal(body, result)
		}
		return nil
	}

	c.debugLog("Async request accepted, polling requestId=%s", asyncResp.RequestID)

	deadline := time.Now().Add(maxPollDuration)
	pollPath := "/v2/alerts/requests/" + asyncResp.RequestID

	for time.Now().Before(deadline) {
		time.Sleep(pollInterval)

		var statusEnvelope struct {
			Data RequestResult `json:"data"`
		}
		resp, respBody, err := c.doRequest(http.MethodGet, pollPath, nil)
		if err != nil {
			return fmt.Errorf("poll request %s: %w", asyncResp.RequestID, err)
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return parseErrorResponse(resp.StatusCode, respBody)
		}

		if err := json.Unmarshal(respBody, &statusEnvelope); err != nil {
			return fmt.Errorf("parse poll response: %w", err)
		}

		status := statusEnvelope.Data
		c.debugLog("Poll result: isSuccess=%v status=%s", status.IsSuccess, status.Status)

		if status.IsSuccess {
			if result != nil {
				// Decode the final result if caller wants it
				_ = json.Unmarshal(respBody, result)
			}
			return nil
		}

		// Terminal failure states
		switch status.Status {
		case "failed", "cancelled":
			return fmt.Errorf("async request %s %s", asyncResp.RequestID, status.Status)
		}
	}

	return fmt.Errorf("timed out waiting for async request %s after %s", asyncResp.RequestID, maxPollDuration)
}

// parseErrorResponse constructs a structured error from an API error response body.
func parseErrorResponse(statusCode int, body []byte) error {
	var errResp ErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Message != "" {
		errResp.Code = statusCode
		return &errResp
	}
	return fmt.Errorf("API error (status %d): %s", statusCode, truncate(string(body), 500))
}

// RateLimitInfo parses rate limit headers from an HTTP response.
type RateLimitInfo struct {
	Limit     int
	Remaining int
}

// ParseRateLimit reads X-RateLimit-* headers from a response.
func ParseRateLimit(resp *http.Response) RateLimitInfo {
	info := RateLimitInfo{}
	if v := resp.Header.Get("X-RateLimit-Limit"); v != "" {
		info.Limit, _ = strconv.Atoi(v)
	}
	if v := resp.Header.Get("X-RateLimit-Remaining"); v != "" {
		info.Remaining, _ = strconv.Atoi(v)
	}
	return info
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
