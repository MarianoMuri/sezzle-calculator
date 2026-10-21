# Unit Test & Coverage Report

Generated on 2026-10-20 with Go 1.24 and Node 22. Re-generate anytime with `make coverage`.

## Summary

| Layer    | Framework                         | Tests                         | Statement coverage |
|----------|-----------------------------------|-------------------------------|--------------------|
| Backend  | `go test` (+ `-race`)             | 57 cases (table-driven subtests) | **100.0%**      |
| Frontend | Vitest + React Testing Library    | 50 tests in 3 files           | **100%** statements · 98.4% branches |

## Backend (`go test -race -coverprofile=coverage.out ./internal/...`)

```
ok   .../backend/internal/api          coverage: 100.0% of statements
ok   .../backend/internal/calculator   coverage: 100.0% of statements

internal/api/handler.go         NewRouter         100.0%
internal/api/handler.go         handleHealth      100.0%
internal/api/handler.go         handleOperations  100.0%
internal/api/handler.go         handleCalculate   100.0%
internal/api/handler.go         decodeRequest     100.0%
internal/api/handler.go         mapError          100.0%
internal/api/handler.go         writeJSON         100.0%
internal/api/handler.go         writeError        100.0%
internal/api/middleware.go      withCORS          100.0%
internal/api/middleware.go      WriteHeader       100.0%
internal/api/middleware.go      withLogging       100.0%
internal/calculator/calculator.go  Arity          100.0%
internal/calculator/calculator.go  Supported      100.0%
internal/calculator/calculator.go  Calculate      100.0%
internal/calculator/calculator.go  isFinite       100.0%
total:                                    (statements)  100.0%
```

`cmd/server` (process wiring: env vars, signal handling) is intentionally excluded; it contains no business logic. An HTML report is produced at `backend/coverage.html`.

**What is covered**

- `calculator`: every operation, negative numbers, decimals, `-0` normalisation, division by zero, `0^negative`, non-real powers (`(-8)^0.5`), negative square root, overflow to ±Inf, NaN/Inf operands, missing operands, unknown operations.
- `api` (real HTTP via `httptest.Server`): success for all 7 endpoints, explicit `0` operands, response shape (unary responses omit `b`), status/code mapping (400/404/405/422), malformed / empty / oversized JSON, wrong types, unknown fields, multiple JSON objects, CORS preflight and CORS disabled, `/healthz`, `/api/v1/operations`.

## Frontend (`npm run coverage`)

```
 Test Files  3 passed (3)
      Tests  50 passed (50)

File               | % Stmts | % Branch | % Funcs | % Lines |
-------------------|---------|----------|---------|---------|
All files          |     100 |    98.41 |     100 |     100 |
 src/api           |     100 |      100 |     100 |     100 |
 src/components    |     100 |      100 |     100 |     100 |
 src/domain        |     100 |      100 |     100 |     100 |
 src/hooks         |     100 |    93.75 |     100 |     100 |
```

The single uncovered branch is the `result === null` guard in `chainResult`, which the UI makes unreachable (the button only renders when a result exists). HTML/LCOV reports are written to `frontend/coverage/`.

**What is covered**

- `domain/validation`: accepted formats (sign, decimals, `.5`, `5.`, exponent, comma decimal, whitespace), rejected formats (`0x10`, `Infinity`, `NaN`, `1.2.3`, …), overflow, result formatting (`0.1 + 0.2 → 0.3`).
- `api/calculatorApi`: request URL/body, unary body omits `b`, backend error codes surfaced as `ApiError`, non-JSON error bodies, network failures, user-friendly messages.
- `components/Calculator` (user-centric tests with `user-event`): full calculation flow, Enter-to-submit, unary mode hides the second input, client-side validation blocks the API call and sets `aria-invalid`, backend error display, loading/disabled state, "use as next input", clear, history capped at 5.
