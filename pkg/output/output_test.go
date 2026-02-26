package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

// captureStdout replaces os.Stdout with a pipe and returns the captured bytes.
func captureStdout(fn func()) (string, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	old := os.Stdout
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// captureStderr replaces os.Stderr with a pipe and returns the captured bytes.
func captureStderr(fn func()) (string, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	old := os.Stderr
	os.Stderr = w

	fn()

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// --- RenderTable tests ---

func TestRenderTable_TableMode(t *testing.T) {
	headers := []string{"ID", "NAME", "STATUS"}
	rows := [][]string{
		{"1", "alpha", "active"},
		{"2", "beta", "inactive"},
	}
	rawData := []map[string]string{
		{"id": "1", "name": "alpha", "status": "active"},
		{"id": "2", "name": "beta", "status": "inactive"},
	}
	opts := Options{Mode: ModeTable, NoColor: true}

	out, err := captureStdout(func() {
		if err := RenderTable(headers, rows, rawData, opts); err != nil {
			t.Errorf("RenderTable error: %v", err)
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	// Table output should contain header and row values
	if !strings.Contains(out, "ID") {
		t.Errorf("expected output to contain header 'ID', got:\n%s", out)
	}
	if !strings.Contains(out, "alpha") {
		t.Errorf("expected output to contain 'alpha', got:\n%s", out)
	}
	if !strings.Contains(out, "inactive") {
		t.Errorf("expected output to contain 'inactive', got:\n%s", out)
	}
}

func TestRenderTable_JSONMode(t *testing.T) {
	headers := []string{"ID", "NAME"}
	rows := [][]string{{"1", "alpha"}}
	rawData := []map[string]string{{"id": "1", "name": "alpha"}}
	opts := Options{Mode: ModeJSON}

	out, err := captureStdout(func() {
		if err := RenderTable(headers, rows, rawData, opts); err != nil {
			t.Errorf("RenderTable error: %v", err)
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	var result []map[string]string
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %s", err, out)
	}
	if len(result) != 1 || result[0]["id"] != "1" {
		t.Errorf("unexpected JSON result: %v", result)
	}
}

func TestRenderTable_PlaintextMode(t *testing.T) {
	headers := []string{"ID", "NAME"}
	rows := [][]string{
		{"1", "alpha"},
		{"2", "beta"},
	}
	rawData := []map[string]string{}
	opts := Options{Mode: ModePlaintext}

	out, err := captureStdout(func() {
		if err := RenderTable(headers, rows, rawData, opts); err != nil {
			t.Errorf("RenderTable error: %v", err)
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (header + 2 rows), got %d:\n%s", len(lines), out)
	}
	if lines[0] != "ID\tNAME" {
		t.Errorf("expected header line 'ID\\tNAME', got %q", lines[0])
	}
	if lines[1] != "1\talpha" {
		t.Errorf("expected row '1\\talpha', got %q", lines[1])
	}
}

// --- RenderJSON tests ---

func TestRenderJSON_Simple(t *testing.T) {
	data := map[string]string{"key": "value"}
	opts := Options{Mode: ModeJSON}

	out, err := captureStdout(func() {
		if err := RenderJSON(data, opts); err != nil {
			t.Errorf("RenderJSON error: %v", err)
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %s", err, out)
	}
	if result["key"] != "value" {
		t.Errorf("expected key=value, got %v", result)
	}
}

func TestRenderJSON_Array(t *testing.T) {
	data := []map[string]interface{}{
		{"id": "1", "name": "alpha"},
		{"id": "2", "name": "beta"},
	}
	opts := Options{Mode: ModeJSON}

	out, err := captureStdout(func() {
		if err := RenderJSON(data, opts); err != nil {
			t.Errorf("RenderJSON error: %v", err)
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %s", err, out)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 elements, got %d", len(result))
	}
}

// --- Field selection tests ---

func TestRenderJSON_FieldsFilter_Map(t *testing.T) {
	data := map[string]interface{}{
		"id":     "1",
		"name":   "alpha",
		"status": "active",
		"secret": "hidden",
	}
	opts := Options{Mode: ModeJSON, Fields: []string{"id", "name"}}

	out, err := captureStdout(func() {
		if err := RenderJSON(data, opts); err != nil {
			t.Errorf("RenderJSON error: %v", err)
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %s", err, out)
	}
	if _, ok := result["secret"]; ok {
		t.Errorf("filtered field 'secret' should not appear in output")
	}
	if _, ok := result["status"]; ok {
		t.Errorf("filtered field 'status' should not appear in output")
	}
	if result["id"] != "1" {
		t.Errorf("expected id=1, got %v", result["id"])
	}
	if result["name"] != "alpha" {
		t.Errorf("expected name=alpha, got %v", result["name"])
	}
}

func TestRenderJSON_FieldsFilter_Array(t *testing.T) {
	data := []map[string]interface{}{
		{"id": "1", "name": "alpha", "secret": "s1"},
		{"id": "2", "name": "beta", "secret": "s2"},
	}
	opts := Options{Mode: ModeJSON, Fields: []string{"id", "name"}}

	out, err := captureStdout(func() {
		if err := RenderJSON(data, opts); err != nil {
			t.Errorf("RenderJSON error: %v", err)
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %s", err, out)
	}
	for _, item := range result {
		if _, ok := item["secret"]; ok {
			t.Errorf("filtered field 'secret' should not appear in array output")
		}
		if _, ok := item["id"]; !ok {
			t.Errorf("field 'id' should appear in array output")
		}
	}
}

func TestRenderTable_JSONMode_WithFields(t *testing.T) {
	headers := []string{"ID", "NAME", "SECRET"}
	rows := [][]string{{"1", "alpha", "s1"}}
	rawData := []map[string]interface{}{
		{"id": "1", "name": "alpha", "secret": "s1"},
	}
	opts := Options{Mode: ModeJSON, Fields: []string{"id", "name"}}

	out, err := captureStdout(func() {
		if err := RenderTable(headers, rows, rawData, opts); err != nil {
			t.Errorf("RenderTable error: %v", err)
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %s", err, out)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 element, got %d", len(result))
	}
	if _, ok := result[0]["secret"]; ok {
		t.Errorf("'secret' field should have been filtered out")
	}
}

// --- JQ filtering tests ---

func TestRenderJSON_JQFilter_Identity(t *testing.T) {
	data := map[string]interface{}{"id": "1", "name": "alpha"}
	opts := Options{Mode: ModeJSON, JQExpr: "."}

	out, err := captureStdout(func() {
		if err := RenderJSON(data, opts); err != nil {
			t.Errorf("RenderJSON error: %v", err)
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %s", err, out)
	}
	if result["id"] != "1" {
		t.Errorf("identity jq should preserve data, got %v", result)
	}
}

func TestRenderJSON_JQFilter_FieldExtract(t *testing.T) {
	data := map[string]interface{}{"id": "42", "name": "test"}
	opts := Options{Mode: ModeJSON, JQExpr: ".id"}

	out, err := captureStdout(func() {
		if err := RenderJSON(data, opts); err != nil {
			t.Errorf("RenderJSON error: %v", err)
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	trimmed := strings.TrimSpace(out)
	if trimmed != `"42"` {
		t.Errorf("expected output to be \"42\", got %q", trimmed)
	}
}

func TestRenderJSON_JQFilter_ArrayMap(t *testing.T) {
	data := []map[string]interface{}{
		{"id": "1", "name": "alpha"},
		{"id": "2", "name": "beta"},
	}
	opts := Options{Mode: ModeJSON, JQExpr: ".[].name"}

	out, err := captureStdout(func() {
		if err := RenderJSON(data, opts); err != nil {
			t.Errorf("RenderJSON error: %v", err)
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	// Multiple outputs get wrapped in array
	var result []interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %s", err, out)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 names, got %d", len(result))
	}
	if result[0] != "alpha" {
		t.Errorf("expected first name 'alpha', got %v", result[0])
	}
}

func TestRenderJSON_JQFilter_Invalid(t *testing.T) {
	data := map[string]interface{}{"id": "1"}
	opts := Options{Mode: ModeJSON, JQExpr: "invalid jq {{{"}

	err := renderJSONTo(io.Discard, data, opts)
	if err == nil {
		t.Error("expected error for invalid jq expression, got nil")
	}
}

func TestRenderJSON_JQFilter_WithFields(t *testing.T) {
	// Fields filtering happens before JQ
	data := map[string]interface{}{"id": "1", "name": "alpha", "secret": "hidden"}
	opts := Options{Mode: ModeJSON, Fields: []string{"id", "name"}, JQExpr: ".name"}

	out, err := captureStdout(func() {
		if err := RenderJSON(data, opts); err != nil {
			t.Errorf("RenderJSON error: %v", err)
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	trimmed := strings.TrimSpace(out)
	if trimmed != `"alpha"` {
		t.Errorf("expected output to be \"alpha\", got %q", trimmed)
	}
}

// --- Error / Success message tests ---

func TestError_WithColor(t *testing.T) {
	out, err := captureStderr(func() {
		Error("something went wrong")
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "something went wrong") {
		t.Errorf("expected error message in output, got: %q", out)
	}
}

func TestError_NoColor(t *testing.T) {
	out, err := captureStderr(func() {
		Error("no color error", Options{NoColor: true})
	})
	if err != nil {
		t.Fatal(err)
	}
	expected := "Error: no color error\n"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestSuccess_WithColor(t *testing.T) {
	out, err := captureStderr(func() {
		Success("operation completed")
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "operation completed") {
		t.Errorf("expected success message in output, got: %q", out)
	}
}

func TestSuccess_NoColor(t *testing.T) {
	out, err := captureStderr(func() {
		Success("all good", Options{NoColor: true})
	})
	if err != nil {
		t.Fatal(err)
	}
	expected := "OK: all good\n"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

// --- NoColor mode tests ---

func TestRenderTable_NoColor(t *testing.T) {
	headers := []string{"NAME", "STATUS"}
	rows := [][]string{{"alpha", "active"}}
	rawData := []map[string]string{{"name": "alpha", "status": "active"}}
	opts := Options{Mode: ModeTable, NoColor: true}

	out, err := captureStdout(func() {
		if err := RenderTable(headers, rows, rawData, opts); err != nil {
			t.Errorf("RenderTable error: %v", err)
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	// With NoColor, output should not contain ANSI escape codes
	if strings.Contains(out, "\x1b[") {
		t.Errorf("NoColor output should not contain ANSI codes, got:\n%s", out)
	}
	if !strings.Contains(out, "NAME") {
		t.Errorf("expected 'NAME' in output, got:\n%s", out)
	}
}

// --- filterFields unit tests ---

func TestFilterFields_Map(t *testing.T) {
	data := map[string]interface{}{
		"a": 1,
		"b": 2,
		"c": 3,
	}
	result, err := filterFields(data, []string{"a", "c"})
	if err != nil {
		t.Fatal(err)
	}
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if _, ok := m["b"]; ok {
		t.Error("field 'b' should be filtered out")
	}
	if m["a"] == nil {
		t.Error("field 'a' should be present")
	}
}

func TestFilterFields_EmptyFields(t *testing.T) {
	data := map[string]interface{}{"a": 1, "b": 2}
	result, err := filterFields(data, nil)
	if err != nil {
		t.Fatal(err)
	}
	// Empty fields list means no filtering
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if len(m) != 2 {
		t.Errorf("expected 2 fields with no filtering, got %d", len(m))
	}
}

func TestFilterFields_Struct(t *testing.T) {
	type item struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Pass string `json:"pass"`
	}
	data := item{ID: "1", Name: "test", Pass: "secret"}
	result, err := filterFields(data, []string{"id", "name"})
	if err != nil {
		t.Fatal(err)
	}
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if _, ok := m["pass"]; ok {
		t.Error("'pass' should be filtered out")
	}
	if m["id"] != "1" {
		t.Errorf("expected id=1, got %v", m["id"])
	}
}

// --- applyJQ unit tests ---

func TestApplyJQ_Identity(t *testing.T) {
	data := map[string]interface{}{"x": float64(1)}
	result, err := applyJQ(data, ".")
	if err != nil {
		t.Fatal(err)
	}
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if m["x"] != float64(1) {
		t.Errorf("expected x=1, got %v", m["x"])
	}
}

func TestApplyJQ_FieldAccess(t *testing.T) {
	data := map[string]interface{}{"name": "hello"}
	result, err := applyJQ(data, ".name")
	if err != nil {
		t.Fatal(err)
	}
	if result != "hello" {
		t.Errorf("expected 'hello', got %v", result)
	}
}

func TestApplyJQ_ParseError(t *testing.T) {
	data := map[string]interface{}{}
	_, err := applyJQ(data, "{{invalid")
	if err == nil {
		t.Error("expected parse error for invalid jq")
	}
}

func TestApplyJQ_NoOutput(t *testing.T) {
	data := map[string]interface{}{"a": 1}
	result, err := applyJQ(data, "empty")
	if err != nil {
		t.Fatal(err)
	}
	if result != nil {
		t.Errorf("expected nil for empty output, got %v", result)
	}
}

// --- renderJSONTo tests (internal, but testable) ---

func TestRenderJSONTo_WritesJSON(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"hello": "world"}
	opts := Options{}

	if err := renderJSONTo(&buf, data, opts); err != nil {
		t.Fatal(err)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if result["hello"] != "world" {
		t.Errorf("expected hello=world, got %v", result)
	}
}

// Ensure Error and Success with plain strings (no Options arg) still work
func TestError_DefaultNoOptions(t *testing.T) {
	out, err := captureStderr(func() {
		// Set NO_COLOR to avoid ANSI in test output
		orig := os.Getenv("NO_COLOR")
		os.Setenv("NO_COLOR", "1")
		defer func() { os.Setenv("NO_COLOR", orig) }()
		Error("test error")
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "test error") {
		t.Errorf("expected 'test error' in stderr, got: %q", out)
	}
}

func TestSuccess_DefaultNoOptions(t *testing.T) {
	out, err := captureStderr(func() {
		orig := os.Getenv("NO_COLOR")
		os.Setenv("NO_COLOR", "1")
		defer func() { os.Setenv("NO_COLOR", orig) }()
		Success("test success")
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "test success") {
		t.Errorf("expected 'test success' in stderr, got: %q", out)
	}
}

// Prevent unused import warning
var _ = fmt.Sprintf
