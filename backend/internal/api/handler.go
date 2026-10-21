// Package api exposes the calculator over a JSON REST interface.
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/MarianoMuri/sezzle-calculator/backend/internal/calculator"
)

const maxBodyBytes = 1 << 10 // 1 KiB is plenty for two numbers.

// CalculationRequest is the JSON body accepted by POST /api/v1/{operation}.
// Pointers distinguish "missing" from an explicit 0.
type CalculationRequest struct {
	A *float64 `json:"a"`
	B *float64 `json:"b,omitempty"`
}

// CalculationResponse is returned on success.
type CalculationResponse struct {
	Operation calculator.Operation `json:"operation"`
	A         float64              `json:"a"`
	B         *float64             `json:"b,omitempty"`
	Result    float64              `json:"result"`
}

// ErrorResponse is returned on any failure.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody carries a machine-readable code and a human-readable message.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type operationInfo struct {
	Name     calculator.Operation `json:"name"`
	Operands int                  `json:"operands"`
}

// NewRouter builds the HTTP handler with all routes and middleware.
func NewRouter(logger *slog.Logger, allowedOrigin string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealth)
	mux.HandleFunc("GET /api/v1/operations", handleOperations)
	mux.HandleFunc("POST /api/v1/{operation}", handleCalculate)

	// Method-less fallbacks: ServeMux prefers the more specific (method) pattern,
	// so these only catch wrong methods and answer with a JSON 405.
	mux.HandleFunc("/healthz", methodNotAllowed(http.MethodGet))
	mux.HandleFunc("/api/v1/operations", methodNotAllowed(http.MethodGet))
	mux.HandleFunc("/api/v1/{operation}", methodNotAllowed(http.MethodPost))

	return withCORS(allowedOrigin, withLogging(logger, mux))
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handleOperations(w http.ResponseWriter, _ *http.Request) {
	ops := calculator.Supported()
	out := make([]operationInfo, 0, len(ops))
	for _, op := range ops {
		n, _ := calculator.Arity(op) // ops come from Supported(), cannot fail
		out = append(out, operationInfo{Name: op, Operands: n})
	}
	writeJSON(w, http.StatusOK, map[string]any{"operations": out})
}

func handleCalculate(w http.ResponseWriter, r *http.Request) {
	op := calculator.Operation(r.PathValue("operation"))
	if _, err := calculator.Arity(op); err != nil {
		writeError(w, http.StatusNotFound, "unknown_operation", fmt.Sprintf("operation %q is not supported", op))
		return
	}

	req, err := decodeRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	result, err := calculator.Calculate(op, req.A, req.B)
	if err != nil {
		status, code := mapError(err)
		writeError(w, status, code, err.Error())
		return
	}

	resp := CalculationResponse{Operation: op, A: *req.A, Result: result}
	if n, _ := calculator.Arity(op); n == 2 {
		resp.B = req.B
	}
	writeJSON(w, http.StatusOK, resp)
}

// methodNotAllowed answers 405 with the standard JSON error shape and an
// Allow header, replacing ServeMux's plain-text default.
func methodNotAllowed(allow string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Allow", allow)
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed",
			fmt.Sprintf("method %s not allowed, use %s", r.Method, allow))
	}
}

func decodeRequest(w http.ResponseWriter, r *http.Request) (CalculationRequest, error) {
	var req CalculationRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		if errors.Is(err, io.EOF) {
			return req, errors.New("request body is empty")
		}
		var typeErr *json.UnmarshalTypeError
		if errors.As(err, &typeErr) {
			return req, fmt.Errorf("field %q must be a number", typeErr.Field)
		}
		return req, fmt.Errorf("malformed JSON: %v", err)
	}
	if dec.More() {
		return req, errors.New("request body must contain a single JSON object")
	}
	return req, nil
}

// mapError converts domain errors into HTTP status + stable error codes.
func mapError(err error) (int, string) {
	switch {
	case errors.Is(err, calculator.ErrDivisionByZero):
		return http.StatusUnprocessableEntity, "division_by_zero"
	case errors.Is(err, calculator.ErrNegativeSqrt):
		return http.StatusUnprocessableEntity, "negative_sqrt"
	case errors.Is(err, calculator.ErrUndefinedResult):
		return http.StatusUnprocessableEntity, "undefined_result"
	case errors.Is(err, calculator.ErrResultNotFinite):
		return http.StatusUnprocessableEntity, "result_out_of_range"
	case errors.Is(err, calculator.ErrMissingOperand):
		return http.StatusBadRequest, "missing_operand"
	case errors.Is(err, calculator.ErrNonFiniteOperand):
		return http.StatusBadRequest, "invalid_operand"
	default:
		return http.StatusInternalServerError, "internal_error"
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, ErrorResponse{Error: ErrorBody{Code: code, Message: msg}})
}