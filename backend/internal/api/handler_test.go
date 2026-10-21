package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MarianoMuri/sezzle-calculator/backend/internal/calculator"
)

const testOrigin = "http://localhost:5173"

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := httptest.NewServer(NewRouter(logger, testOrigin))
	t.Cleanup(srv.Close)
	return srv
}

func post(t *testing.T, srv *httptest.Server, path, body string) (*http.Response, []byte) {
	t.Helper()
	resp, err := http.Post(srv.URL+path, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp, b
}

func TestCalculateSuccess(t *testing.T) {
	srv := newTestServer(t)
	tests := []struct {
		path, body string
		want       float64
	}{
		{"/api/v1/add", `{"a": 2, "b": 3}`, 5},
		{"/api/v1/subtract", `{"a": 2, "b": 3}`, -1},
		{"/api/v1/multiply", `{"a": 2.5, "b": 4}`, 10},
		{"/api/v1/divide", `{"a": 7, "b": 2}`, 3.5},
		{"/api/v1/power", `{"a": 2, "b": 8}`, 256},
		{"/api/v1/sqrt", `{"a": 16}`, 4},
		{"/api/v1/percentage", `{"a": 10, "b": 50}`, 5},
		{"/api/v1/add", `{"a": 0, "b": 0}`, 0}, // explicit zeros are valid
	}
	for _, tt := range tests {
		t.Run(tt.path+tt.body, func(t *testing.T) {
			resp, body := post(t, srv, tt.path, tt.body)
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status = %d, body = %s", resp.StatusCode, body)
			}
			if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q", ct)
			}
			var got CalculationResponse
			if err := json.Unmarshal(body, &got); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if got.Result != tt.want {
				t.Errorf("result = %v, want %v", got.Result, tt.want)
			}
		})
	}
}

func TestSqrtResponseOmitsB(t *testing.T) {
	srv := newTestServer(t)
	_, body := post(t, srv, "/api/v1/sqrt", `{"a": 9}`)
	if strings.Contains(string(body), `"b"`) {
		t.Errorf("unary response should not include b: %s", body)
	}
}

func TestCalculateErrors(t *testing.T) {
	srv := newTestServer(t)
	tests := []struct {
		name, path, body string
		status           int
		code             string
	}{
		{"division by zero", "/api/v1/divide", `{"a": 1, "b": 0}`, 422, "division_by_zero"},
		{"negative sqrt", "/api/v1/sqrt", `{"a": -4}`, 422, "negative_sqrt"},
		{"undefined power", "/api/v1/power", `{"a": -8, "b": 0.5}`, 422, "undefined_result"},
		{"overflow", "/api/v1/power", `{"a": 10, "b": 999}`, 422, "result_out_of_range"},
		{"missing b", "/api/v1/add", `{"a": 1}`, 400, "missing_operand"},
		{"missing a", "/api/v1/add", `{"b": 1}`, 400, "missing_operand"},
		{"string operand", "/api/v1/add", `{"a": "1", "b": 2}`, 400, "invalid_json"},
		{"malformed json", "/api/v1/add", `{"a": 1,`, 400, "invalid_json"},
		{"empty body", "/api/v1/add", ``, 400, "invalid_json"},
		{"unknown field", "/api/v1/add", `{"a": 1, "b": 2, "c": 3}`, 400, "invalid_json"},
		{"multiple objects", "/api/v1/add", `{"a": 1, "b": 2}{"a": 3}`, 400, "invalid_json"},
		{"body too large", "/api/v1/add", `{"a": 1, "b": ` + strings.Repeat("1", 2000) + `}`, 400, "invalid_json"},
		{"unknown operation", "/api/v1/modulo", `{"a": 1, "b": 2}`, 404, "unknown_operation"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := post(t, srv, tt.path, tt.body)
			if resp.StatusCode != tt.status {
				t.Fatalf("status = %d, want %d (body %s)", resp.StatusCode, tt.status, body)
			}
			var got ErrorResponse
			if err := json.Unmarshal(body, &got); err != nil {
				t.Fatalf("decode: %v (body %s)", err, body)
			}
			if got.Error.Code != tt.code {
				t.Errorf("code = %q, want %q", got.Error.Code, tt.code)
			}
			if got.Error.Message == "" {
				t.Error("message should not be empty")
			}
		})
	}
}

func TestOperationsEndpoint(t *testing.T) {
	srv := newTestServer(t)
	resp, err := http.Get(srv.URL + "/api/v1/operations")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var got struct {
		Operations []operationInfo `json:"operations"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got.Operations) != 7 {
		t.Fatalf("got %d operations, want 7", len(got.Operations))
	}
	for _, op := range got.Operations {
		if op.Name == "sqrt" && op.Operands != 1 {
			t.Errorf("sqrt operands = %d, want 1", op.Operands)
		}
	}
}

func TestHealth(t *testing.T) {
	srv := newTestServer(t)
	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d", resp.StatusCode)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	srv := newTestServer(t)
	resp, err := http.Get(srv.URL + "/api/v1/add")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", resp.StatusCode)
	}
}

func TestCORSPreflight(t *testing.T) {
	srv := newTestServer(t)
	req, _ := http.NewRequest(http.MethodOptions, srv.URL+"/api/v1/add", nil)
	req.Header.Set("Origin", testOrigin)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d, want 204", resp.StatusCode)
	}
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != testOrigin {
		t.Errorf("Allow-Origin = %q", got)
	}
}

func TestCORSDisabled(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := httptest.NewServer(NewRouter(logger, ""))
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected no CORS header, got %q", got)
	}
}

func TestMapError(t *testing.T) {
	tests := []struct {
		err    error
		status int
		code   string
	}{
		// JSON cannot encode NaN/Inf, so this path is only reachable from Go callers.
		{calculator.ErrNonFiniteOperand, http.StatusBadRequest, "invalid_operand"},
		{io.ErrUnexpectedEOF, http.StatusInternalServerError, "internal_error"},
	}
	for _, tt := range tests {
		status, code := mapError(tt.err)
		if status != tt.status || code != tt.code {
			t.Errorf("mapError(%v) = %d %s, want %d %s", tt.err, status, code, tt.status, tt.code)
		}
	}
}
