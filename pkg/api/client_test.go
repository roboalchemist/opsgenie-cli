package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
)

// --- Helpers ---

// newTestClient creates a Client pointed at the given test server URL.
func newTestClient(t *testing.T, serverURL string) *Client {
	t.Helper()
	t.Setenv("OPSGENIE_API_URL", serverURL)
	return NewClient("test-key", "us", false)
}

// jsonEncode encodes v to JSON, panicking on error (test helper only).
func jsonEncode(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("jsonEncode: %v", err))
	}
	return b
}

// --- NewClient construction tests ---

func TestNewClient_USRegion(t *testing.T) {
	// Ensure env override is cleared so region logic is exercised.
	t.Setenv("OPSGENIE_API_URL", "")

	c := NewClient("key", "us", false)
	if c.baseURL != baseURLUS {
		t.Errorf("expected baseURL=%q for US region, got %q", baseURLUS, c.baseURL)
	}
}

func TestNewClient_EURegion(t *testing.T) {
	t.Setenv("OPSGENIE_API_URL", "")

	c := NewClient("key", "eu", false)
	if c.baseURL != baseURLEU {
		t.Errorf("expected baseURL=%q for EU region, got %q", baseURLEU, c.baseURL)
	}
}

func TestNewClient_UnknownRegionDefaultsToUS(t *testing.T) {
	t.Setenv("OPSGENIE_API_URL", "")

	c := NewClient("key", "ap", false)
	if c.baseURL != baseURLUS {
		t.Errorf("expected US base URL for unknown region, got %q", c.baseURL)
	}
}

func TestNewClient_EnvOverride(t *testing.T) {
	override := "http://custom.example.com"
	t.Setenv("OPSGENIE_API_URL", override)

	c := NewClient("key", "us", false)
	if c.baseURL != override {
		t.Errorf("expected baseURL=%q from env override, got %q", override, c.baseURL)
	}
}

func TestNewClient_APIKeyStored(t *testing.T) {
	t.Setenv("OPSGENIE_API_URL", "")

	c := NewClient("my-secret-key", "us", false)
	if c.apiKey != "my-secret-key" {
		t.Errorf("expected apiKey=%q, got %q", "my-secret-key", c.apiKey)
	}
}

// --- Auth header tests ---

func TestGet_SendsCorrectAuthHeader(t *testing.T) {
	var gotAuth string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jsonEncode(map[string]interface{}{"result": "ok"}))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	var result map[string]interface{}
	if err := c.Get("/v2/alerts", &result); err != nil {
		t.Fatalf("Get error: %v", err)
	}

	expected := "GenieKey test-key"
	if gotAuth != expected {
		t.Errorf("expected Authorization header %q, got %q", expected, gotAuth)
	}
}

func TestGet_SendsUserAgentHeader(t *testing.T) {
	var gotUA string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jsonEncode(map[string]string{"result": "ok"}))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	var result map[string]string
	_ = c.Get("/v2/test", &result)

	if !strings.HasPrefix(gotUA, "opsgenie-cli/") {
		t.Errorf("expected User-Agent to start with 'opsgenie-cli/', got %q", gotUA)
	}
}

// --- POST tests ---

func TestPost_SendsBodyAndAuthHeader(t *testing.T) {
	var gotBody []byte
	var gotAuth string
	var gotContentType string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")

		var err error
		gotBody, err = readAll(r)
		if err != nil {
			t.Errorf("reading body: %v", err)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jsonEncode(map[string]string{"requestId": "abc"}))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	payload := map[string]string{"message": "test alert"}
	var result map[string]string
	if err := c.Post("/v2/alerts", payload, &result); err != nil {
		t.Fatalf("Post error: %v", err)
	}

	if gotAuth != "GenieKey test-key" {
		t.Errorf("expected correct auth header, got %q", gotAuth)
	}
	if gotContentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", gotContentType)
	}

	var decoded map[string]string
	if err := json.Unmarshal(gotBody, &decoded); err != nil {
		t.Fatalf("body is not valid JSON: %v — body: %s", err, gotBody)
	}
	if decoded["message"] != "test alert" {
		t.Errorf("expected body message='test alert', got %q", decoded["message"])
	}
}

// --- Error response parsing ---

func TestGet_4xxReturnsErrorResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write(jsonEncode(map[string]interface{}{
			"message": "Alert not found",
			"code":    404,
		}))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	err := c.Get("/v2/alerts/nonexistent", nil)
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}

	errResp, ok := err.(*ErrorResponse)
	if !ok {
		t.Fatalf("expected *ErrorResponse, got %T: %v", err, err)
	}
	if errResp.Message != "Alert not found" {
		t.Errorf("expected message 'Alert not found', got %q", errResp.Message)
	}
	if errResp.Code != 404 {
		t.Errorf("expected code 404, got %d", errResp.Code)
	}
}

func TestGet_401ReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write(jsonEncode(map[string]interface{}{
			"message": "Unauthorized",
			"code":    401,
		}))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	err := c.Get("/v2/alerts", nil)
	if err == nil {
		t.Fatal("expected error for 401, got nil")
	}
}

func TestGet_ErrorResponseWithDetails(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(jsonEncode(map[string]interface{}{
			"message": "Validation error",
			"code":    400,
			"errors":  []string{"field 'message' is required"},
		}))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	err := c.Get("/v2/alerts", nil)
	if err == nil {
		t.Fatal("expected error for 400, got nil")
	}

	errResp, ok := err.(*ErrorResponse)
	if !ok {
		t.Fatalf("expected *ErrorResponse, got %T", err)
	}
	if len(errResp.Errors) == 0 {
		t.Error("expected Errors slice to be populated")
	}
	// Verify the Error() string format with details
	errStr := errResp.Error()
	if !strings.Contains(errStr, "details:") {
		t.Errorf("expected error string to contain 'details:', got %q", errStr)
	}
}

func TestGet_NonJSONErrorBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error (plain text)"))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	err := c.Get("/v2/alerts", nil)
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}
	// Should fall back to generic error message
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected error to mention status 500, got %q", err.Error())
	}
}

// --- Rate limiting retry ---

func TestGet_RateLimitRetries(t *testing.T) {
	var callCount int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&callCount, 1)
		if n <= 2 {
			// First two calls: rate limited
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write(jsonEncode(map[string]interface{}{
				"message": "Too Many Requests",
				"code":    429,
			}))
			return
		}
		// Third call: success
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jsonEncode(map[string]interface{}{"result": "ok"}))
	}))
	defer ts.Close()

	// Use a client with very short backoff by temporarily overriding sleep via direct approach.
	// Since we can't easily mock time.Sleep, we just verify the retry happens.
	// The test may be slightly slow (1s backoff) but is correct.
	// We use a small trick: replace the test server to avoid actual sleeping by
	// overriding maxRetries; instead we verify the call count.
	c := newTestClient(t, ts.URL)

	var result map[string]interface{}
	err := c.Get("/v2/alerts", &result)
	// With maxRetries=3, two 429s + one success should succeed.
	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if atomic.LoadInt32(&callCount) != 3 {
		t.Errorf("expected 3 calls (2 rate-limited + 1 success), got %d", callCount)
	}
}

func TestGet_RateLimitExceedsMaxRetries(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always rate-limit
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"message":"rate limited","code":429}`))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	err := c.Get("/v2/alerts", nil)
	if err == nil {
		t.Fatal("expected error after exceeding max retries, got nil")
	}
	if !strings.Contains(err.Error(), "retries") {
		t.Errorf("expected error to mention retries, got %q", err.Error())
	}
}

// --- Async polling (202 responses) ---

func TestPost_AsyncPollingSuccess(t *testing.T) {
	const requestID = "req-12345"
	var pollCount int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Initial POST returns 202 with requestId
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write(jsonEncode(map[string]string{
				"requestId": requestID,
				"result":    "Request will be processed",
			}))
			return
		}

		// Poll endpoint
		if strings.Contains(r.URL.Path, "/v2/alerts/requests/") {
			n := atomic.AddInt32(&pollCount, 1)
			if n < 2 {
				// Still processing
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(jsonEncode(map[string]interface{}{
					"data": map[string]interface{}{
						"isSuccess": false,
						"status":    "Processing",
					},
				}))
				return
			}
			// Done
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(jsonEncode(map[string]interface{}{
				"data": map[string]interface{}{
					"isSuccess": true,
					"status":    "Processed",
					"alertId":   "alert-999",
				},
			}))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	payload := map[string]string{"message": "test"}
	err := c.Post("/v2/alerts", payload, nil)
	if err != nil {
		t.Fatalf("expected async poll to succeed, got: %v", err)
	}

	if atomic.LoadInt32(&pollCount) < 2 {
		t.Errorf("expected at least 2 poll calls, got %d", pollCount)
	}
}

func TestPost_AsyncPollingFailed(t *testing.T) {
	const requestID = "req-fail-999"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write(jsonEncode(map[string]string{"requestId": requestID}))
			return
		}
		if strings.Contains(r.URL.Path, "/v2/alerts/requests/") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(jsonEncode(map[string]interface{}{
				"data": map[string]interface{}{
					"isSuccess": false,
					"status":    "failed",
				},
			}))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	err := c.Post("/v2/alerts", map[string]string{"message": "test"}, nil)
	if err == nil {
		t.Fatal("expected error for failed async request, got nil")
	}
	if !strings.Contains(err.Error(), "failed") {
		t.Errorf("expected error to contain 'failed', got %q", err.Error())
	}
}

func TestPost_AsyncPollingCancelled(t *testing.T) {
	const requestID = "req-cancel-888"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write(jsonEncode(map[string]string{"requestId": requestID}))
			return
		}
		if strings.Contains(r.URL.Path, "/v2/alerts/requests/") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(jsonEncode(map[string]interface{}{
				"data": map[string]interface{}{
					"isSuccess": false,
					"status":    "cancelled",
				},
			}))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	err := c.Post("/v2/alerts", map[string]string{"message": "test"}, nil)
	if err == nil {
		t.Fatal("expected error for cancelled async request, got nil")
	}
	if !strings.Contains(err.Error(), "cancelled") {
		t.Errorf("expected error to mention 'cancelled', got %q", err.Error())
	}
}

func TestPost_Async_NoRequestID_TreatedAsSuccess(t *testing.T) {
	// 202 with no requestId body → treat as immediate success
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"result":"accepted with no requestId"}`))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	err := c.Post("/v2/alerts", map[string]string{"message": "test"}, nil)
	if err != nil {
		t.Fatalf("expected success when no requestId in 202, got: %v", err)
	}
}

// --- ListAll pagination ---

func TestListAll_SinglePage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jsonEncode(map[string]interface{}{
			"data": []map[string]string{
				{"id": "1", "name": "alpha"},
				{"id": "2", "name": "beta"},
			},
			// No paging.next → single page
		}))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	var results []map[string]string
	if err := c.ListAll("/v2/alerts", nil, &results); err != nil {
		t.Fatalf("ListAll error: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
	if results[0]["id"] != "1" {
		t.Errorf("expected first id='1', got %q", results[0]["id"])
	}
}

func TestListAll_MultiplePages(t *testing.T) {
	var callCount int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&callCount, 1)

		w.WriteHeader(http.StatusOK)
		if n == 1 {
			// First page — provide a next link pointing to /v2/alerts?offset=2
			nextURL := fmt.Sprintf("http://%s/v2/alerts?offset=2&limit=2", r.Host)
			_, _ = w.Write(jsonEncode(map[string]interface{}{
				"data": []map[string]string{
					{"id": "1", "name": "alpha"},
					{"id": "2", "name": "beta"},
				},
				"paging": map[string]string{"next": nextURL},
			}))
		} else {
			// Second page — no next
			_, _ = w.Write(jsonEncode(map[string]interface{}{
				"data": []map[string]string{
					{"id": "3", "name": "gamma"},
				},
			}))
		}
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	var results []map[string]string
	if err := c.ListAll("/v2/alerts", nil, &results); err != nil {
		t.Fatalf("ListAll error: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results across 2 pages, got %d", len(results))
	}
	if results[2]["id"] != "3" {
		t.Errorf("expected third id='3', got %q", results[2]["id"])
	}
	if atomic.LoadInt32(&callCount) != 2 {
		t.Errorf("expected 2 page fetches, got %d", callCount)
	}
}

func TestListAll_EmptyDataField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jsonEncode(map[string]interface{}{
			"data": []map[string]string{},
		}))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	var results []map[string]string
	if err := c.ListAll("/v2/alerts", nil, &results); err != nil {
		t.Fatalf("ListAll error: %v", err)
	}
	// Empty array is valid
	if results == nil {
		results = []map[string]string{}
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestListAll_SetsDefaultLimit(t *testing.T) {
	var gotURL string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jsonEncode(map[string]interface{}{"data": []map[string]string{}}))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	var results []map[string]string
	_ = c.ListAll("/v2/alerts", nil, &results)

	if !strings.Contains(gotURL, "limit=100") {
		t.Errorf("expected default limit=100 in query, got %q", gotURL)
	}
}

func TestListAll_CustomLimitPreserved(t *testing.T) {
	var gotURL string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jsonEncode(map[string]interface{}{"data": []map[string]string{}}))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	params := url.Values{}
	params.Set("limit", "50")
	var results []map[string]string
	_ = c.ListAll("/v2/alerts", params, &results)

	if !strings.Contains(gotURL, "limit=50") {
		t.Errorf("expected limit=50 to be preserved, got %q", gotURL)
	}
}

func TestListAll_RateLimitedReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"message":"rate limited","code":429}`))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	var results []map[string]string
	err := c.ListAll("/v2/alerts", nil, &results)
	if err == nil {
		t.Fatal("expected error for rate limit during pagination, got nil")
	}
	if !strings.Contains(err.Error(), "rate limited") {
		t.Errorf("expected 'rate limited' in error, got %q", err.Error())
	}
}

func TestListAll_ErrorResponseReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write(jsonEncode(map[string]interface{}{
			"message": "Forbidden",
			"code":    403,
		}))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	var results []map[string]string
	err := c.ListAll("/v2/alerts", nil, &results)
	if err == nil {
		t.Fatal("expected error for 403, got nil")
	}
}

// --- GetWithParams single page ---

func TestGetWithParams_UnwrapsDataField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jsonEncode(map[string]interface{}{
			"data": map[string]string{
				"id":   "sched-1",
				"name": "My Schedule",
			},
		}))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	params := url.Values{}
	params.Set("identifierType", "id")
	var result map[string]string
	if err := c.GetWithParams("/v2/schedules/sched-1", params, &result); err != nil {
		t.Fatalf("GetWithParams error: %v", err)
	}

	if result["id"] != "sched-1" {
		t.Errorf("expected id='sched-1', got %q", result["id"])
	}
	if result["name"] != "My Schedule" {
		t.Errorf("expected name='My Schedule', got %q", result["name"])
	}
}

func TestGetWithParams_NoParams(t *testing.T) {
	var gotQuery string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jsonEncode(map[string]interface{}{
			"data": map[string]string{"id": "1"},
		}))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	var result map[string]string
	if err := c.GetWithParams("/v2/test", nil, &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// No params → no query string appended
	if gotQuery != "" {
		t.Errorf("expected empty query string with nil params, got %q", gotQuery)
	}
}

func TestGetWithParams_WithParams(t *testing.T) {
	var gotQuery string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jsonEncode(map[string]interface{}{"data": map[string]string{"id": "1"}}))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	params := url.Values{}
	params.Set("timezone", "US/Eastern")
	var result map[string]string
	_ = c.GetWithParams("/v2/schedules/on-call", params, &result)

	if !strings.Contains(gotQuery, "timezone") {
		t.Errorf("expected 'timezone' in query string, got %q", gotQuery)
	}
}

func TestGetWithParams_EmptyDataField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Response with no data field
		_, _ = w.Write(jsonEncode(map[string]interface{}{"result": "ok"}))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	var result map[string]string
	// Should return nil error even with empty/missing data field
	err := c.GetWithParams("/v2/test", nil, &result)
	if err != nil {
		t.Fatalf("expected no error with empty data field, got: %v", err)
	}
}

func TestGetWithParams_RateLimitRetries(t *testing.T) {
	var callCount int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&callCount, 1)
		if n <= 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"message":"rate limited","code":429}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jsonEncode(map[string]interface{}{"data": map[string]string{"id": "1"}}))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	var result map[string]string
	err := c.GetWithParams("/v2/test", nil, &result)
	if err != nil {
		t.Fatalf("expected success after retry, got: %v", err)
	}
}

func TestGetWithParams_RateLimitExceedsRetries(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"message":"rate limited","code":429}`))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	var result map[string]string
	err := c.GetWithParams("/v2/test", nil, &result)
	if err == nil {
		t.Fatal("expected error after exceeding retries")
	}
}

// --- HTTP verb methods (Put, Patch, Delete) ---

func TestPut_SendsCorrectMethod(t *testing.T) {
	var gotMethod string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jsonEncode(map[string]string{"result": "ok"}))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	var result map[string]string
	if err := c.Put("/v2/alerts/test-id/priority", map[string]string{"priority": "P1"}, &result); err != nil {
		t.Fatalf("Put error: %v", err)
	}
	if gotMethod != http.MethodPut {
		t.Errorf("expected PUT method, got %q", gotMethod)
	}
}

func TestPatch_SendsCorrectMethod(t *testing.T) {
	var gotMethod string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jsonEncode(map[string]string{"result": "ok"}))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	var result map[string]string
	if err := c.Patch("/v2/alerts/test-id", map[string]string{"message": "updated"}, &result); err != nil {
		t.Fatalf("Patch error: %v", err)
	}
	if gotMethod != http.MethodPatch {
		t.Errorf("expected PATCH method, got %q", gotMethod)
	}
}

func TestDelete_SendsCorrectMethod(t *testing.T) {
	var gotMethod string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jsonEncode(map[string]string{"requestId": "del-req-1"}))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	var result map[string]string
	if err := c.Delete("/v2/alerts/test-id", &result); err != nil {
		t.Fatalf("Delete error: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("expected DELETE method, got %q", gotMethod)
	}
}

func TestDelete_NilResult(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jsonEncode(map[string]string{"requestId": "del-req-1"}))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	// nil result should be fine
	if err := c.Delete("/v2/alerts/test-id", nil); err != nil {
		t.Fatalf("Delete with nil result: %v", err)
	}
}

// --- ParseRateLimit ---

func TestParseRateLimit_Full(t *testing.T) {
	h := http.Header{}
	h.Set("X-RateLimit-Limit", "600")
	h.Set("X-RateLimit-Remaining", "580")
	resp := &http.Response{Header: h}
	info := ParseRateLimit(resp)
	if info.Limit != 600 {
		t.Errorf("expected Limit=600, got %d", info.Limit)
	}
	if info.Remaining != 580 {
		t.Errorf("expected Remaining=580, got %d", info.Remaining)
	}
}

func TestParseRateLimit_EmptyHeaders(t *testing.T) {
	resp := &http.Response{Header: http.Header{}}
	info := ParseRateLimit(resp)
	if info.Limit != 0 {
		t.Errorf("expected Limit=0 for missing header, got %d", info.Limit)
	}
	if info.Remaining != 0 {
		t.Errorf("expected Remaining=0 for missing header, got %d", info.Remaining)
	}
}

func TestParseRateLimit_InvalidValues(t *testing.T) {
	h := http.Header{}
	h.Set("X-RateLimit-Limit", "not-a-number")
	h.Set("X-RateLimit-Remaining", "also-not-a-number")
	resp := &http.Response{Header: h}
	// Should not panic; returns zero values for invalid headers
	info := ParseRateLimit(resp)
	if info.Limit != 0 {
		t.Errorf("expected Limit=0 for invalid header, got %d", info.Limit)
	}
}

// --- ErrorResponse.Error() formatting ---

func TestErrorResponse_Error_WithoutDetails(t *testing.T) {
	e := &ErrorResponse{Code: 403, Message: "Forbidden"}
	s := e.Error()
	if !strings.Contains(s, "403") {
		t.Errorf("expected code 403 in error string, got %q", s)
	}
	if !strings.Contains(s, "Forbidden") {
		t.Errorf("expected 'Forbidden' in error string, got %q", s)
	}
	if strings.Contains(s, "details:") {
		t.Errorf("'details:' should not appear when Errors is empty, got %q", s)
	}
}

func TestErrorResponse_Error_WithDetails(t *testing.T) {
	e := &ErrorResponse{
		Code:    400,
		Message: "Bad Request",
		Errors:  []string{"field required", "invalid format"},
	}
	s := e.Error()
	if !strings.Contains(s, "details:") {
		t.Errorf("expected 'details:' in error string, got %q", s)
	}
	if !strings.Contains(s, "field required") {
		t.Errorf("expected error detail in string, got %q", s)
	}
}

// --- truncate ---

func TestTruncate_ShortString(t *testing.T) {
	s := truncate("hello", 10)
	if s != "hello" {
		t.Errorf("expected 'hello', got %q", s)
	}
}

func TestTruncate_LongString(t *testing.T) {
	s := truncate("hello world", 5)
	if s != "hello..." {
		t.Errorf("expected 'hello...', got %q", s)
	}
}

func TestTruncate_ExactLength(t *testing.T) {
	s := truncate("exact", 5)
	if s != "exact" {
		t.Errorf("expected 'exact' (no truncation at exact length), got %q", s)
	}
}

// --- buildURL ---

func TestBuildURL(t *testing.T) {
	t.Setenv("OPSGENIE_API_URL", "")
	c := NewClient("key", "us", false)
	got := c.buildURL("/v2/alerts")
	expected := baseURLUS + "/v2/alerts"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

// --- debugLog (coverage only — no assertion on output) ---

func TestDebugLog_NoOutput_WhenDisabled(t *testing.T) {
	t.Setenv("OPSGENIE_API_URL", "")
	c := NewClient("key", "us", false) // debug=false
	// Should not panic or produce visible output; just exercise the code path.
	c.debugLog("test message %s", "arg")
}

func TestDebugLog_WithDebugEnabled(t *testing.T) {
	t.Setenv("OPSGENIE_API_URL", "")
	c := NewClient("key", "us", true) // debug=true
	// Exercise the debug path; we don't capture stderr here, just ensure no panic.
	c.debugLog("debug: %d items", 5)
}

// --- Get_NilResult (no body parsing) ---

func TestGet_NilResult_SuccessNoParsing(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Return a non-empty body that would normally be parsed
		_, _ = w.Write([]byte(`{"result":"ok","took":0.1}`))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	// nil result means caller doesn't care about response body
	if err := c.Get("/v2/account", nil); err != nil {
		t.Fatalf("expected no error with nil result, got: %v", err)
	}
}

// --- do: unmarshal error on success body ---

func TestGet_UnmarshalErrorOnSuccessBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Not valid JSON for the expected type
		_, _ = w.Write([]byte(`not json at all`))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	var result map[string]string
	err := c.Get("/v2/test", &result)
	if err == nil {
		t.Fatal("expected unmarshal error for invalid JSON body, got nil")
	}
	if !strings.Contains(err.Error(), "parse response") {
		t.Errorf("expected 'parse response' in error, got %q", err.Error())
	}
}

// --- GetWithParams: envelope parse error ---

func TestGetWithParams_InvalidEnvelopeReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not valid json`))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	var result map[string]string
	err := c.GetWithParams("/v2/test", nil, &result)
	if err == nil {
		t.Fatal("expected error for invalid JSON envelope, got nil")
	}
	if !strings.Contains(err.Error(), "envelope") {
		t.Errorf("expected 'envelope' in error, got %q", err.Error())
	}
}

// --- GetWithParams: data unmarshal error ---

func TestGetWithParams_DataUnmarshalError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Valid envelope but data is not compatible with []int target
		_, _ = w.Write(jsonEncode(map[string]interface{}{
			"data": map[string]string{"id": "abc"},
		}))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	// Target is an int slice — incompatible with a JSON object
	var result []int
	err := c.GetWithParams("/v2/test", nil, &result)
	if err == nil {
		t.Fatal("expected error when data is incompatible with result type, got nil")
	}
	if !strings.Contains(err.Error(), "data") {
		t.Errorf("expected 'data' in error, got %q", err.Error())
	}
}

// --- ListAll: invalid JSON page ---

func TestListAll_InvalidPageJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not json`))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	var results []map[string]string
	err := c.ListAll("/v2/alerts", nil, &results)
	if err == nil {
		t.Fatal("expected error for invalid page JSON, got nil")
	}
	if !strings.Contains(err.Error(), "parse page") {
		t.Errorf("expected 'parse page' in error, got %q", err.Error())
	}
}

// --- ListAll: decode combined results error ---

func TestListAll_DecodeCombinedResultsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Returns array of items
		_, _ = w.Write(jsonEncode(map[string]interface{}{
			"data": []map[string]string{{"id": "1"}},
		}))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	// Target is incompatible — string slice can't hold JSON objects
	var result string
	err := c.ListAll("/v2/alerts", nil, &result)
	if err == nil {
		t.Fatal("expected error when combined results can't be decoded, got nil")
	}
	if !strings.Contains(err.Error(), "decode combined") {
		t.Errorf("expected 'decode combined' in error, got %q", err.Error())
	}
}

// --- pollRequestResult: no requestId but with result pointer ---

func TestPost_Async_NoRequestID_WithResult(t *testing.T) {
	// 202 with no requestId and with a result pointer — body should be decoded into result
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write(jsonEncode(map[string]string{
			"status": "immediate",
		}))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	var result map[string]string
	// Pass result so the body-decode path is exercised
	if err := c.Post("/v2/alerts", map[string]string{"msg": "t"}, &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// result["status"] may or may not be set (body has no requestId so treated as immediate success)
}

// --- pollRequestResult: success with non-nil result ---

func TestPost_AsyncPollingSuccess_WithResult(t *testing.T) {
	const requestID = "req-with-result"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write(jsonEncode(map[string]string{"requestId": requestID}))
			return
		}
		if strings.Contains(r.URL.Path, "/v2/alerts/requests/") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(jsonEncode(map[string]interface{}{
				"data": map[string]interface{}{
					"isSuccess": true,
					"status":    "Processed",
					"alertId":   "final-alert-id",
				},
			}))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	// Pass a non-nil result to exercise the result decode path in pollRequestResult
	var result map[string]interface{}
	err := c.Post("/v2/alerts", map[string]string{"message": "test"}, &result)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

// --- pollRequestResult: bad JSON in poll response ---

func TestPost_AsyncPoll_BadPollJSON(t *testing.T) {
	const requestID = "req-bad-json"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write(jsonEncode(map[string]string{"requestId": requestID}))
			return
		}
		if strings.Contains(r.URL.Path, "/v2/alerts/requests/") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`not valid json`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	err := c.Post("/v2/alerts", map[string]string{"message": "test"}, nil)
	if err == nil {
		t.Fatal("expected error for bad poll JSON, got nil")
	}
	if !strings.Contains(err.Error(), "parse poll response") {
		t.Errorf("expected 'parse poll response' in error, got %q", err.Error())
	}
}

// --- ListAll with single-object data field ---

func TestListAll_SingleObjectData(t *testing.T) {
	// Some OpsGenie endpoints return a single object (not array) in data
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// data is a JSON object, not array
		_, _ = w.Write(jsonEncode(map[string]interface{}{
			"data": map[string]string{"id": "obj-1", "name": "object"},
		}))
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	var results []interface{}
	if err := c.ListAll("/v2/test", nil, &results); err != nil {
		t.Fatalf("ListAll error: %v", err)
	}
	// Single object is wrapped in a 1-element array
	if len(results) != 1 {
		t.Errorf("expected 1 element for single-object data, got %d", len(results))
	}
}

// --- pollRequestResult: poll error response ---

func TestPost_AsyncPoll_ErrorResponse(t *testing.T) {
	const requestID = "req-poll-err"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write(jsonEncode(map[string]string{"requestId": requestID}))
			return
		}
		// Poll returns an error status
		if strings.Contains(r.URL.Path, "/v2/alerts/requests/") {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write(jsonEncode(map[string]interface{}{
				"message": "Request not found",
				"code":    404,
			}))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	c := newTestClient(t, ts.URL)
	err := c.Post("/v2/alerts", map[string]string{"message": "test"}, nil)
	if err == nil {
		t.Fatal("expected error when poll returns error status, got nil")
	}
}

// --- doRequest: marshal error ---

func TestPost_MarshalError(t *testing.T) {
	t.Setenv("OPSGENIE_API_URL", "http://localhost:1") // unreachable, but we fail before connecting
	c := NewClient("key", "us", false)

	// A channel cannot be marshaled to JSON
	type unmarshalable struct {
		Ch chan int
	}
	var result interface{}
	err := c.Post("/v2/test", unmarshalable{Ch: make(chan int)}, &result)
	if err == nil {
		t.Fatal("expected marshal error for unmarshalable body, got nil")
	}
	if !strings.Contains(err.Error(), "marshal") {
		t.Errorf("expected error to mention marshal, got %q", err.Error())
	}
}

// --- readAll helper ---

func readAll(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	buf := make([]byte, 0, 512)
	tmp := make([]byte, 512)
	for {
		n, err := r.Body.Read(tmp)
		buf = append(buf, tmp[:n]...)
		if err != nil {
			break
		}
	}
	return buf, nil
}
