package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// ─── Test harness ─────────────────────────────────────────────────────────────

var binaryPath string

func TestMain(m *testing.M) {
	cmd := exec.Command("go", "build", "-o", "opsgenie-cli-test", ".")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic("Failed to build: " + err.Error())
	}
	binaryPath = "./opsgenie-cli-test"
	code := m.Run()
	os.Remove(binaryPath)
	os.Exit(code)
}

// runCLI runs the CLI binary with the given args, injecting the mock server URL
// and test API key. Returns stdout, stderr, and exit code.
func runCLI(t *testing.T, serverURL string, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = append(os.Environ(),
		"OPSGENIE_API_KEY=test-key",
		"OPSGENIE_API_URL="+serverURL,
		"NO_COLOR=1",
	)
	err := cmd.Run()
	exitCode := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	}
	return stdout.String(), stderr.String(), exitCode
}

// ─── Mock OpsGenie data ───────────────────────────────────────────────────────

var mockAlert = map[string]interface{}{
	"id":           "alert-id-123",
	"tinyId":       "42",
	"alias":        "test-alert",
	"message":      "Test alert message",
	"status":       "open",
	"acknowledged": false,
	"snoozed":      false,
	"isSeen":       false,
	"tags":         []string{"tag1", "tag2"},
	"count":        1,
	"source":       "test-source",
	"owner":        "owner@example.com",
	"priority":     "P3",
	"createdAt":    "2024-01-15T10:00:00Z",
	"updatedAt":    "2024-01-15T10:01:00Z",
	"closedAt":     "",
}

var mockTeam = map[string]interface{}{
	"id":          "team-id-456",
	"name":        "Test Team",
	"description": "A test team",
	"members":     []interface{}{},
}

var mockSchedule = map[string]interface{}{
	"id":          "schedule-id-789",
	"name":        "Test Schedule",
	"timezone":    "UTC",
	"enabled":     true,
	"description": "A test schedule",
}

var mockUser = map[string]interface{}{
	"id":        "user-id-001",
	"username":  "testuser@example.com",
	"fullName":  "Test User",
	"role":      map[string]interface{}{"id": "role-1", "name": "user"},
	"blocked":   false,
	"verified":  true,
	"createdAt": "2024-01-01T00:00:00Z",
}

var mockHeartbeat = map[string]interface{}{
	"name":          "test-heartbeat",
	"description":   "A test heartbeat",
	"interval":      10,
	"intervalUnit":  "minutes",
	"enabled":       true,
	"expired":       false,
	"lastPingAt":    "2024-01-15T09:55:00Z",
	"alertMessage":  "Heartbeat failed",
	"alertPriority": "P3",
}

var mockEscalation = map[string]interface{}{
	"id":          "escalation-id-001",
	"name":        "Test Escalation",
	"description": "A test escalation policy",
	"rules":       []interface{}{},
}

// ─── Mock server setup ────────────────────────────────────────────────────────

// newMockServer builds an httptest.Server that handles all OpsGenie v2 endpoints.
// It also records which HTTP methods were used per path for action command tests.
func newMockServer(t *testing.T) (*httptest.Server, *methodLog) {
	t.Helper()
	log := &methodLog{methods: make(map[string]string)}

	mux := http.NewServeMux()

	// ── alerts ────────────────────────────────────────────────────────────────

	// GET /v2/alerts → list
	// POST /v2/alerts → create
	mux.HandleFunc("/v2/alerts", func(w http.ResponseWriter, r *http.Request) {
		log.record(r.URL.Path, r.Method)
		if r.Method == http.MethodPost {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"result":    "Request will be processed",
				"requestId": "req-create-001",
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"data": []interface{}{mockAlert},
			"paging": map[string]interface{}{
				"first": "",
				"next":  "",
			},
		})
	})

	// GET /v2/alerts/count
	mux.HandleFunc("/v2/alerts/count", func(w http.ResponseWriter, r *http.Request) {
		log.record(r.URL.Path, r.Method)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"data": map[string]interface{}{"count": 7},
		})
	})

	// GET /v2/alerts/requests/{requestId} — async poll
	mux.HandleFunc("/v2/alerts/requests/", func(w http.ResponseWriter, r *http.Request) {
		log.record("/v2/alerts/requests/", r.Method)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"data": map[string]interface{}{
				"isSuccess": true,
				"status":    "completed",
				"alertId":   "alert-id-123",
			},
		})
	})

	// GET /v2/alerts/{id} and action endpoints
	mux.HandleFunc("/v2/alerts/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		log.record(path, r.Method)

		switch {
		case strings.HasSuffix(path, "/acknowledge"):
			// POST acknowledge — return 202 so the client polls
			writeJSON(w, http.StatusAccepted, map[string]interface{}{
				"requestId": "req-ack-001",
				"result":    "Request will be processed",
			})
		case strings.HasSuffix(path, "/close"):
			writeJSON(w, http.StatusAccepted, map[string]interface{}{
				"requestId": "req-close-001",
				"result":    "Request will be processed",
			})
		case strings.HasSuffix(path, "/snooze"):
			writeJSON(w, http.StatusAccepted, map[string]interface{}{
				"requestId": "req-snooze-001",
				"result":    "Request will be processed",
			})
		case strings.HasSuffix(path, "/escalate"):
			writeJSON(w, http.StatusAccepted, map[string]interface{}{
				"requestId": "req-escalate-001",
				"result":    "Request will be processed",
			})
		case strings.HasSuffix(path, "/assign"):
			writeJSON(w, http.StatusAccepted, map[string]interface{}{
				"requestId": "req-assign-001",
				"result":    "Request will be processed",
			})
		case strings.HasSuffix(path, "/notes"):
			writeJSON(w, http.StatusAccepted, map[string]interface{}{
				"requestId": "req-note-001",
				"result":    "Request will be processed",
			})
		case strings.HasSuffix(path, "/tags") && r.Method == http.MethodPost:
			writeJSON(w, http.StatusAccepted, map[string]interface{}{
				"requestId": "req-tags-001",
				"result":    "Request will be processed",
			})
		case strings.HasSuffix(path, "/tags") && r.Method == http.MethodDelete:
			writeJSON(w, http.StatusAccepted, map[string]interface{}{
				"requestId": "req-rmtags-001",
				"result":    "Request will be processed",
			})
		case r.Method == http.MethodDelete:
			writeJSON(w, http.StatusAccepted, map[string]interface{}{
				"requestId": "req-delete-001",
				"result":    "Request will be processed",
			})
		default:
			// GET /v2/alerts/{id}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"data": mockAlert,
			})
		}
	})

	// ── teams ─────────────────────────────────────────────────────────────────

	mux.HandleFunc("/v2/teams", func(w http.ResponseWriter, r *http.Request) {
		log.record(r.URL.Path, r.Method)
		if r.Method == http.MethodPost {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"data": mockTeam,
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"data": []interface{}{mockTeam},
		})
	})

	mux.HandleFunc("/v2/teams/", func(w http.ResponseWriter, r *http.Request) {
		log.record(r.URL.Path, r.Method)
		switch r.Method {
		case http.MethodDelete:
			writeJSON(w, http.StatusOK, map[string]interface{}{"result": "deleted"})
		case http.MethodPatch:
			writeJSON(w, http.StatusOK, map[string]interface{}{"data": mockTeam})
		default:
			writeJSON(w, http.StatusOK, map[string]interface{}{"data": mockTeam})
		}
	})

	// ── schedules ─────────────────────────────────────────────────────────────

	mux.HandleFunc("/v2/schedules", func(w http.ResponseWriter, r *http.Request) {
		log.record(r.URL.Path, r.Method)
		if r.Method == http.MethodPost {
			writeJSON(w, http.StatusOK, map[string]interface{}{"data": mockSchedule})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"data": []interface{}{mockSchedule},
		})
	})

	mux.HandleFunc("/v2/schedules/", func(w http.ResponseWriter, r *http.Request) {
		log.record(r.URL.Path, r.Method)
		switch r.Method {
		case http.MethodDelete:
			writeJSON(w, http.StatusOK, map[string]interface{}{"result": "deleted"})
		case http.MethodPatch:
			writeJSON(w, http.StatusOK, map[string]interface{}{"data": mockSchedule})
		default:
			writeJSON(w, http.StatusOK, map[string]interface{}{"data": mockSchedule})
		}
	})

	// ── users ─────────────────────────────────────────────────────────────────

	// users list uses ListAll which follows paging; return a single page with no next.
	mux.HandleFunc("/v2/users", func(w http.ResponseWriter, r *http.Request) {
		log.record(r.URL.Path, r.Method)
		if r.Method == http.MethodPost {
			writeJSON(w, http.StatusOK, map[string]interface{}{"data": mockUser})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"data":   []interface{}{mockUser},
			"paging": map[string]interface{}{"next": ""},
		})
	})

	mux.HandleFunc("/v2/users/", func(w http.ResponseWriter, r *http.Request) {
		log.record(r.URL.Path, r.Method)
		switch r.Method {
		case http.MethodDelete:
			writeJSON(w, http.StatusOK, map[string]interface{}{"result": "deleted"})
		case http.MethodPatch:
			writeJSON(w, http.StatusOK, map[string]interface{}{"data": mockUser})
		default:
			writeJSON(w, http.StatusOK, map[string]interface{}{"data": mockUser})
		}
	})

	// ── escalations ───────────────────────────────────────────────────────────

	mux.HandleFunc("/v2/escalations", func(w http.ResponseWriter, r *http.Request) {
		log.record(r.URL.Path, r.Method)
		if r.Method == http.MethodPost {
			writeJSON(w, http.StatusOK, map[string]interface{}{"data": mockEscalation})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"data": []interface{}{mockEscalation},
		})
	})

	mux.HandleFunc("/v2/escalations/", func(w http.ResponseWriter, r *http.Request) {
		log.record(r.URL.Path, r.Method)
		switch r.Method {
		case http.MethodDelete:
			writeJSON(w, http.StatusOK, map[string]interface{}{"result": "deleted"})
		case http.MethodPut:
			writeJSON(w, http.StatusOK, map[string]interface{}{"data": mockEscalation})
		default:
			writeJSON(w, http.StatusOK, map[string]interface{}{"data": mockEscalation})
		}
	})

	// ── heartbeats ────────────────────────────────────────────────────────────

	mux.HandleFunc("/v2/heartbeats", func(w http.ResponseWriter, r *http.Request) {
		log.record(r.URL.Path, r.Method)
		if r.Method == http.MethodPost {
			writeJSON(w, http.StatusOK, map[string]interface{}{"heartbeat": mockHeartbeat})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"data": []interface{}{mockHeartbeat},
		})
	})

	mux.HandleFunc("/v2/heartbeats/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		log.record(path, r.Method)

		switch {
		case strings.HasSuffix(path, "/enable"):
			writeJSON(w, http.StatusOK, map[string]interface{}{"heartbeat": mockHeartbeat})
		case strings.HasSuffix(path, "/disable"):
			writeJSON(w, http.StatusOK, map[string]interface{}{"heartbeat": mockHeartbeat})
		case strings.HasSuffix(path, "/ping"):
			writeJSON(w, http.StatusOK, map[string]interface{}{"result": "OK"})
		case r.Method == http.MethodDelete:
			writeJSON(w, http.StatusOK, map[string]interface{}{"result": "deleted"})
		case r.Method == http.MethodPatch:
			writeJSON(w, http.StatusOK, map[string]interface{}{"heartbeat": mockHeartbeat})
		default:
			writeJSON(w, http.StatusOK, map[string]interface{}{"data": mockHeartbeat})
		}
	})

	// ── 401 handler ───────────────────────────────────────────────────────────

	// used by the "invalid API key" test via a separate server
	_ = mux // used below via httptest.NewServer(mux)

	return httptest.NewServer(mux), log
}

// newUnauthorizedServer returns a mock that always responds 401.
func newUnauthorizedServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"message": "API key is invalid",
			"code":    401,
		})
	}))
}

// writeJSON marshals v and writes it as a JSON response.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// methodLog records the HTTP method last seen for each path.
type methodLog struct {
	methods map[string]string
}

func (l *methodLog) record(path, method string) {
	l.methods[path] = method
}

func (l *methodLog) lastMethod(path string) string {
	return l.methods[path]
}

// ─── Helper assertions ────────────────────────────────────────────────────────

func assertExitCode(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("exit code: got %d, want %d", got, want)
	}
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("expected output to contain %q\ngot: %q", substr, s)
	}
}

func assertNotContains(t *testing.T, s, substr string) {
	t.Helper()
	if strings.Contains(s, substr) {
		t.Errorf("expected output NOT to contain %q\ngot: %q", substr, s)
	}
}

func assertValidJSON(t *testing.T, s string) {
	t.Helper()
	if !json.Valid([]byte(strings.TrimSpace(s))) {
		t.Errorf("output is not valid JSON:\n%s", s)
	}
}

// ─── Built-in commands (no server needed) ─────────────────────────────────────

func TestIntegration_BuiltinCommands(t *testing.T) {
	// These commands require no API key or server.
	cases := []struct {
		name string
		args []string
	}{
		{"help flag", []string{"--help"}},
		{"version flag", []string{"--version"}},
		{"docs command", []string{"docs"}},
		{"completion bash", []string{"completion", "bash"}},
		{"completion zsh", []string{"completion", "zsh"}},
		{"skill print", []string{"skill", "print"}},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tc.args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			// These commands don't need env vars — pass clean minimal env
			cmd.Env = append(os.Environ(), "NO_COLOR=1")
			err := cmd.Run()
			if err != nil {
				t.Errorf("command %v failed: %v\nstdout: %s\nstderr: %s",
					tc.args, err, stdout.String(), stderr.String())
			}
		})
	}
}

// ─── Error handling tests ─────────────────────────────────────────────────────

func TestIntegration_NoAPIKey(t *testing.T) {
	cmd := exec.Command(binaryPath, "alerts", "list")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	// Intentionally omit OPSGENIE_API_KEY and use a minimal env
	cmd.Env = []string{"HOME=" + os.Getenv("HOME"), "PATH=" + os.Getenv("PATH")}
	_ = cmd.Run()

	combined := stdout.String() + stderr.String()
	if !strings.Contains(combined, "OPSGENIE_API_KEY") &&
		!strings.Contains(combined, "authentication") &&
		!strings.Contains(combined, "auth") {
		t.Errorf("expected authentication error, got stdout=%q stderr=%q",
			stdout.String(), stderr.String())
	}
}

func TestIntegration_InvalidAPIKey(t *testing.T) {
	srv := newUnauthorizedServer(t)
	defer srv.Close()

	_, stderr, exitCode := runCLI(t, srv.URL, "alerts", "list")
	if exitCode == 0 {
		t.Error("expected non-zero exit code on 401")
	}
	if !strings.Contains(stderr, "401") && !strings.Contains(stderr, "invalid") &&
		!strings.Contains(stderr, "Error") {
		t.Errorf("expected 401/invalid error in stderr, got: %q", stderr)
	}
}

// ─── alerts list ─────────────────────────────────────────────────────────────

func TestIntegration_AlertsList_DefaultTable(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "alerts", "list")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "alert-id-123")
	assertContains(t, stdout, "Test alert message")
}

func TestIntegration_AlertsList_JSON(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "alerts", "list", "--json")
	assertExitCode(t, exitCode, 0)
	assertValidJSON(t, stdout)
	assertContains(t, stdout, "alert-id-123")
	assertContains(t, stdout, "Test alert message")
}

func TestIntegration_AlertsList_Plaintext(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "alerts", "list", "--plaintext")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "alert-id-123")
	// Plaintext uses tabs between fields
	assertContains(t, stdout, "\t")
	// No ANSI codes
	assertNotContains(t, stdout, "\033[")
}

func TestIntegration_AlertsList_NoColor(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "alerts", "list", "--no-color")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "alert-id-123")
	assertNotContains(t, stdout, "\033[")
}

func TestIntegration_AlertsList_Fields(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "alerts", "list", "--json", "--fields", "id,message")
	assertExitCode(t, exitCode, 0)
	assertValidJSON(t, stdout)
	assertContains(t, stdout, "alert-id-123")
	assertContains(t, stdout, "Test alert message")
	// Fields not requested should not appear in output
	assertNotContains(t, stdout, "tinyId")
	assertNotContains(t, stdout, "source")
}

func TestIntegration_AlertsList_JQ(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "alerts", "list", "--json", "--jq", ".[0].message")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "Test alert message")
}

func TestIntegration_AlertsList_Debug(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	_, stderr, exitCode := runCLI(t, srv.URL, "alerts", "list", "--debug")
	assertExitCode(t, exitCode, 0)
	// Debug output goes to stderr
	if !strings.Contains(stderr, "DEBUG") && !strings.Contains(stderr, "debug") {
		t.Errorf("expected debug output in stderr, got: %q", stderr)
	}
}

func TestIntegration_AlertsList_WithLimit(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "alerts", "list", "--limit", "5", "--json")
	assertExitCode(t, exitCode, 0)
	assertValidJSON(t, stdout)
}

func TestIntegration_AlertsList_WithQuery(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "alerts", "list", "--query", "status:open", "--json")
	assertExitCode(t, exitCode, 0)
	assertValidJSON(t, stdout)
}

func TestIntegration_AlertsList_WithSort(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "alerts", "list", "--sort", "createdAt", "--json")
	assertExitCode(t, exitCode, 0)
	assertValidJSON(t, stdout)
}

// ─── alerts get ──────────────────────────────────────────────────────────────

func TestIntegration_AlertsGet_DefaultTable(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "alerts", "get", "alert-id-123")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "alert-id-123")
	assertContains(t, stdout, "Test alert message")
}

func TestIntegration_AlertsGet_JSON(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "alerts", "get", "alert-id-123", "--json")
	assertExitCode(t, exitCode, 0)
	assertValidJSON(t, stdout)
	assertContains(t, stdout, "alert-id-123")
}

func TestIntegration_AlertsGet_Plaintext(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "alerts", "get", "alert-id-123", "--plaintext")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "alert-id-123")
	assertNotContains(t, stdout, "\033[")
}

func TestIntegration_AlertsGet_NoColor(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "alerts", "get", "alert-id-123", "--no-color")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "alert-id-123")
	assertNotContains(t, stdout, "\033[")
}

// ─── alerts count ─────────────────────────────────────────────────────────────

func TestIntegration_AlertsCount_DefaultTable(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "alerts", "count")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "7")
}

func TestIntegration_AlertsCount_JSON(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "alerts", "count", "--json")
	assertExitCode(t, exitCode, 0)
	assertValidJSON(t, stdout)
	assertContains(t, stdout, "7")
}

func TestIntegration_AlertsCount_Plaintext(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "alerts", "count", "--plaintext")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "7")
	assertNotContains(t, stdout, "\033[")
}

func TestIntegration_AlertsCount_NoColor(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "alerts", "count", "--no-color")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "7")
	assertNotContains(t, stdout, "\033[")
}

func TestIntegration_AlertsCount_WithQuery(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "alerts", "count", "--query", "status:open", "--json")
	assertExitCode(t, exitCode, 0)
	assertValidJSON(t, stdout)
}

// ─── alerts action commands ───────────────────────────────────────────────────

func TestIntegration_AlertsAcknowledge_SendsPOST(t *testing.T) {
	srv, log := newMockServer(t)
	defer srv.Close()

	_, stderr, exitCode := runCLI(t, srv.URL, "alerts", "acknowledge", "alert-id-123")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stderr, "acknowledged")

	// Verify the acknowledge endpoint was called (may have trailing path params)
	found := false
	for path, method := range log.methods {
		if strings.Contains(path, "/acknowledge") && method == http.MethodPost {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected POST to /acknowledge endpoint, got methods: %v", log.methods)
	}
}

func TestIntegration_AlertsClose_SendsPOST(t *testing.T) {
	srv, log := newMockServer(t)
	defer srv.Close()

	_, stderr, exitCode := runCLI(t, srv.URL, "alerts", "close", "alert-id-123", "--note", "closing for test")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stderr, "closed")

	found := false
	for path, method := range log.methods {
		if strings.Contains(path, "/close") && method == http.MethodPost {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected POST to /close endpoint, got methods: %v", log.methods)
	}
}

func TestIntegration_AlertsSnooze_SendsPOST(t *testing.T) {
	srv, log := newMockServer(t)
	defer srv.Close()

	_, stderr, exitCode := runCLI(t, srv.URL, "alerts", "snooze", "alert-id-123",
		"--end-time", "2024-12-31T23:59:59Z")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stderr, "snoozed")

	found := false
	for path, method := range log.methods {
		if strings.Contains(path, "/snooze") && method == http.MethodPost {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected POST to /snooze endpoint, got methods: %v", log.methods)
	}
}

func TestIntegration_AlertsAssign_SendsPOST(t *testing.T) {
	srv, log := newMockServer(t)
	defer srv.Close()

	_, stderr, exitCode := runCLI(t, srv.URL, "alerts", "assign", "alert-id-123",
		"--owner", "newowner@example.com")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stderr, "assigned")

	found := false
	for path, method := range log.methods {
		if strings.Contains(path, "/assign") && method == http.MethodPost {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected POST to /assign endpoint, got methods: %v", log.methods)
	}
}

func TestIntegration_AlertsAddNote_SendsPOST(t *testing.T) {
	srv, log := newMockServer(t)
	defer srv.Close()

	_, stderr, exitCode := runCLI(t, srv.URL, "alerts", "add-note", "alert-id-123",
		"--note", "this is a test note")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stderr, "Note added")

	found := false
	for path, method := range log.methods {
		if strings.Contains(path, "/notes") && method == http.MethodPost {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected POST to /notes endpoint, got methods: %v", log.methods)
	}
}

func TestIntegration_AlertsAddTags_SendsPOST(t *testing.T) {
	srv, log := newMockServer(t)
	defer srv.Close()

	_, stderr, exitCode := runCLI(t, srv.URL, "alerts", "add-tags", "alert-id-123",
		"--tags", "production,critical")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stderr, "Tags added")

	found := false
	for path, method := range log.methods {
		if strings.Contains(path, "/tags") && method == http.MethodPost {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected POST to /tags endpoint, got methods: %v", log.methods)
	}
}

// ─── teams list ───────────────────────────────────────────────────────────────

func TestIntegration_TeamsList_DefaultTable(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "teams", "list")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "Test Team")
	assertContains(t, stdout, "team-id-456")
}

func TestIntegration_TeamsList_JSON(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "teams", "list", "--json")
	assertExitCode(t, exitCode, 0)
	assertValidJSON(t, stdout)
	assertContains(t, stdout, "Test Team")
	assertContains(t, stdout, "team-id-456")
}

func TestIntegration_TeamsList_Plaintext(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "teams", "list", "--plaintext")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "Test Team")
	assertContains(t, stdout, "\t")
	assertNotContains(t, stdout, "\033[")
}

func TestIntegration_TeamsList_NoColor(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "teams", "list", "--no-color")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "Test Team")
	assertNotContains(t, stdout, "\033[")
}

func TestIntegration_TeamsList_Fields(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "teams", "list", "--json", "--fields", "id,name")
	assertExitCode(t, exitCode, 0)
	assertValidJSON(t, stdout)
	assertContains(t, stdout, "team-id-456")
	assertContains(t, stdout, "Test Team")
	assertNotContains(t, stdout, "description")
}

func TestIntegration_TeamsList_JQ(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "teams", "list", "--json", "--jq", ".[0].name")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "Test Team")
}

func TestIntegration_TeamsList_Debug(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	_, stderr, exitCode := runCLI(t, srv.URL, "teams", "list", "--debug")
	assertExitCode(t, exitCode, 0)
	if !strings.Contains(stderr, "DEBUG") && !strings.Contains(stderr, "debug") {
		t.Errorf("expected debug output in stderr, got: %q", stderr)
	}
}

// ─── teams get ────────────────────────────────────────────────────────────────

func TestIntegration_TeamsGet_JSON(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "teams", "get", "team-id-456", "--json")
	assertExitCode(t, exitCode, 0)
	assertValidJSON(t, stdout)
	assertContains(t, stdout, "Test Team")
}

// ─── schedules list ───────────────────────────────────────────────────────────

func TestIntegration_SchedulesList_DefaultTable(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "schedules", "list")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "Test Schedule")
	assertContains(t, stdout, "schedule-id-789")
}

func TestIntegration_SchedulesList_JSON(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "schedules", "list", "--json")
	assertExitCode(t, exitCode, 0)
	assertValidJSON(t, stdout)
	assertContains(t, stdout, "Test Schedule")
	assertContains(t, stdout, "schedule-id-789")
}

func TestIntegration_SchedulesList_Plaintext(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "schedules", "list", "--plaintext")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "Test Schedule")
	assertContains(t, stdout, "\t")
	assertNotContains(t, stdout, "\033[")
}

func TestIntegration_SchedulesList_NoColor(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "schedules", "list", "--no-color")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "Test Schedule")
	assertNotContains(t, stdout, "\033[")
}

func TestIntegration_SchedulesList_Fields(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "schedules", "list", "--json", "--fields", "id,name")
	assertExitCode(t, exitCode, 0)
	assertValidJSON(t, stdout)
	assertContains(t, stdout, "schedule-id-789")
	assertNotContains(t, stdout, "timezone")
}

func TestIntegration_SchedulesList_JQ(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "schedules", "list", "--json", "--jq", ".[0].name")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "Test Schedule")
}

func TestIntegration_SchedulesList_Debug(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	_, stderr, exitCode := runCLI(t, srv.URL, "schedules", "list", "--debug")
	assertExitCode(t, exitCode, 0)
	if !strings.Contains(stderr, "DEBUG") && !strings.Contains(stderr, "debug") {
		t.Errorf("expected debug output in stderr, got: %q", stderr)
	}
}

// ─── users list ───────────────────────────────────────────────────────────────

func TestIntegration_UsersList_DefaultTable(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "users", "list")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "testuser@example.com")
	assertContains(t, stdout, "user-id-001")
}

func TestIntegration_UsersList_JSON(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "users", "list", "--json")
	assertExitCode(t, exitCode, 0)
	assertValidJSON(t, stdout)
	assertContains(t, stdout, "testuser@example.com")
	assertContains(t, stdout, "user-id-001")
}

func TestIntegration_UsersList_Plaintext(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "users", "list", "--plaintext")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "testuser@example.com")
	assertContains(t, stdout, "\t")
	assertNotContains(t, stdout, "\033[")
}

func TestIntegration_UsersList_NoColor(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "users", "list", "--no-color")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "testuser@example.com")
	assertNotContains(t, stdout, "\033[")
}

func TestIntegration_UsersList_Fields(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "users", "list", "--json", "--fields", "id,username")
	assertExitCode(t, exitCode, 0)
	assertValidJSON(t, stdout)
	assertContains(t, stdout, "testuser@example.com")
	assertNotContains(t, stdout, "fullName")
}

func TestIntegration_UsersList_JQ(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "users", "list", "--json", "--jq", ".[0].username")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "testuser@example.com")
}

func TestIntegration_UsersList_Debug(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	_, stderr, exitCode := runCLI(t, srv.URL, "users", "list", "--debug")
	assertExitCode(t, exitCode, 0)
	if !strings.Contains(stderr, "DEBUG") && !strings.Contains(stderr, "debug") {
		t.Errorf("expected debug output in stderr, got: %q", stderr)
	}
}

// ─── escalations list ─────────────────────────────────────────────────────────

func TestIntegration_EscalationsList_DefaultTable(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "escalations", "list")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "Test Escalation")
	assertContains(t, stdout, "escalation-id-001")
}

func TestIntegration_EscalationsList_JSON(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "escalations", "list", "--json")
	assertExitCode(t, exitCode, 0)
	assertValidJSON(t, stdout)
	assertContains(t, stdout, "Test Escalation")
}

func TestIntegration_EscalationsList_Plaintext(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "escalations", "list", "--plaintext")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "Test Escalation")
	assertContains(t, stdout, "\t")
	assertNotContains(t, stdout, "\033[")
}

func TestIntegration_EscalationsList_NoColor(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "escalations", "list", "--no-color")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "Test Escalation")
	assertNotContains(t, stdout, "\033[")
}

func TestIntegration_EscalationsList_Fields(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "escalations", "list", "--json", "--fields", "id,name")
	assertExitCode(t, exitCode, 0)
	assertValidJSON(t, stdout)
	assertContains(t, stdout, "escalation-id-001")
	assertNotContains(t, stdout, "description")
}

func TestIntegration_EscalationsList_JQ(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "escalations", "list", "--json", "--jq", ".[0].name")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "Test Escalation")
}

func TestIntegration_EscalationsList_Debug(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	_, stderr, exitCode := runCLI(t, srv.URL, "escalations", "list", "--debug")
	assertExitCode(t, exitCode, 0)
	if !strings.Contains(stderr, "DEBUG") && !strings.Contains(stderr, "debug") {
		t.Errorf("expected debug output in stderr, got: %q", stderr)
	}
}

// ─── heartbeats list ──────────────────────────────────────────────────────────

func TestIntegration_HeartbeatsList_DefaultTable(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "heartbeats", "list")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "test-heartbeat")
}

func TestIntegration_HeartbeatsList_JSON(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "heartbeats", "list", "--json")
	assertExitCode(t, exitCode, 0)
	assertValidJSON(t, stdout)
	assertContains(t, stdout, "test-heartbeat")
}

func TestIntegration_HeartbeatsList_Plaintext(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "heartbeats", "list", "--plaintext")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "test-heartbeat")
	assertContains(t, stdout, "\t")
	assertNotContains(t, stdout, "\033[")
}

func TestIntegration_HeartbeatsList_NoColor(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "heartbeats", "list", "--no-color")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "test-heartbeat")
	assertNotContains(t, stdout, "\033[")
}

func TestIntegration_HeartbeatsList_Fields(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "heartbeats", "list", "--json", "--fields", "name,enabled")
	assertExitCode(t, exitCode, 0)
	assertValidJSON(t, stdout)
	assertContains(t, stdout, "test-heartbeat")
	assertNotContains(t, stdout, "intervalUnit")
}

func TestIntegration_HeartbeatsList_JQ(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "heartbeats", "list", "--json", "--jq", ".[0].name")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "test-heartbeat")
}

func TestIntegration_HeartbeatsList_Debug(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	_, stderr, exitCode := runCLI(t, srv.URL, "heartbeats", "list", "--debug")
	assertExitCode(t, exitCode, 0)
	if !strings.Contains(stderr, "DEBUG") && !strings.Contains(stderr, "debug") {
		t.Errorf("expected debug output in stderr, got: %q", stderr)
	}
}

// ─── heartbeats get ───────────────────────────────────────────────────────────

func TestIntegration_HeartbeatsGet_DefaultTable(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "heartbeats", "get", "test-heartbeat")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "test-heartbeat")
}

func TestIntegration_HeartbeatsGet_JSON(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "heartbeats", "get", "test-heartbeat", "--json")
	assertExitCode(t, exitCode, 0)
	assertValidJSON(t, stdout)
	assertContains(t, stdout, "test-heartbeat")
}

func TestIntegration_HeartbeatsGet_Plaintext(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "heartbeats", "get", "test-heartbeat", "--plaintext")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "test-heartbeat")
	assertNotContains(t, stdout, "\033[")
}

func TestIntegration_HeartbeatsGet_NoColor(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "heartbeats", "get", "test-heartbeat", "--no-color")
	assertExitCode(t, exitCode, 0)
	assertContains(t, stdout, "test-heartbeat")
	assertNotContains(t, stdout, "\033[")
}

// ─── escalations get ─────────────────────────────────────────────────────────

func TestIntegration_EscalationsGet_JSON(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "escalations", "get", "escalation-id-001", "--json")
	assertExitCode(t, exitCode, 0)
	assertValidJSON(t, stdout)
	assertContains(t, stdout, "Test Escalation")
}

// ─── schedules get ────────────────────────────────────────────────────────────

func TestIntegration_SchedulesGet_JSON(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "schedules", "get", "schedule-id-789", "--json")
	assertExitCode(t, exitCode, 0)
	assertValidJSON(t, stdout)
	assertContains(t, stdout, "Test Schedule")
}

// ─── users get ────────────────────────────────────────────────────────────────

func TestIntegration_UsersGet_JSON(t *testing.T) {
	srv, _ := newMockServer(t)
	defer srv.Close()

	stdout, _, exitCode := runCLI(t, srv.URL, "users", "get", "user-id-001", "--json")
	assertExitCode(t, exitCode, 0)
	assertValidJSON(t, stdout)
	assertContains(t, stdout, "testuser@example.com")
}
