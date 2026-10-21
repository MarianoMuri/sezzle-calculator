// Package calculator contains the pure arithmetic domain logic.
//
// It has no knowledge of HTTP or JSON so it can be tested in isolation and
// reused by any transport (REST, gRPC, CLI...).
package calculator

import (
	"errors"
	"fmt"
	"math"
)

// Operation identifies a supported arithmetic operation.
type Operation string

const (
	Add        Operation = "add"
	Subtract   Operation = "subtract"
	Multiply   Operation = "multiply"
	Divide     Operation = "divide"
	Power      Operation = "power"
	Sqrt       Operation = "sqrt"
	Percentage Operation = "percentage"
)

// Domain errors. Callers can match them with errors.Is to map them to
// transport-specific responses (e.g. HTTP 422).
var (
	ErrDivisionByZero   = errors.New("division by zero")
	ErrNegativeSqrt     = errors.New("square root of a negative number is not a real number")
	ErrUnknownOperation = errors.New("unknown operation")
	ErrMissingOperand   = errors.New("missing operand")
	ErrNonFiniteOperand = errors.New("operands must be finite numbers")
	ErrResultNotFinite  = errors.New("result is out of range")
	ErrUndefinedResult  = errors.New("result is undefined")
)

// spec describes an operation: how many operands it takes and how to compute it.
type spec struct {
	arity int
	fn    func(a, b float64) (float64, error)
}

var operations = map[Operation]spec{
	Add:      {2, func(a, b float64) (float64, error) { return a + b, nil }},
	Subtract: {2, func(a, b float64) (float64, error) { return a - b, nil }},
	Multiply: {2, func(a, b float64) (float64, error) { return a * b, nil }},
	Divide: {2, func(a, b float64) (float64, error) {
		if b == 0 {
			return 0, ErrDivisionByZero
		}
		return a / b, nil
	}},
	Power: {2, func(a, b float64) (float64, error) {
		if a == 0 && b < 0 {
			return 0, ErrDivisionByZero
		}
		r := math.Pow(a, b)
		if math.IsNaN(r) {
			// e.g. (-8)^0.5 has no real result.
			return 0, ErrUndefinedResult
		}
		return r, nil
	}},
	Sqrt: {1, func(a, _ float64) (float64, error) {
		if a < 0 {
			return 0, ErrNegativeSqrt
		}
		return math.Sqrt(a), nil
	}},
	// Percentage returns a% of b (e.g. 15 % of 200 = 30).
	Percentage: {2, func(a, b float64) (float64, error) { return a / 100 * b, nil }},
}

// Arity returns how many operands op requires, or an error if op is unknown.
func Arity(op Operation) (int, error) {
	s, ok := operations[op]
	if !ok {
		return 0, fmt.Errorf("%w: %q", ErrUnknownOperation, op)
	}
	return s.arity, nil
}

// Supported returns the list of supported operations in a stable order.
func Supported() []Operation {
	return []Operation{Add, Subtract, Multiply, Divide, Power, Sqrt, Percentage}
}

// Calculate validates the operands and executes op.
// For unary operations (sqrt) b must be nil.
func Calculate(op Operation, a, b *float64) (float64, error) {
	s, ok := operations[op]
	if !ok {
		return 0, fmt.Errorf("%w: %q", ErrUnknownOperation, op)
	}
	if a == nil || (s.arity == 2 && b == nil) {
		return 0, fmt.Errorf("%w: %s requires %d operand(s)", ErrMissingOperand, op, s.arity)
	}

	var bv float64
	if s.arity == 2 {
		bv = *b
	}
	if !isFinite(*a) || !isFinite(bv) {
		return 0, ErrNonFiniteOperand
	}

	r, err := s.fn(*a, bv)
	if err != nil {
		return 0, err
	}
	if !isFinite(r) {
		return 0, ErrResultNotFinite
	}
	// Avoid returning "-0" to clients.
	if r == 0 {
		r = 0
	}
	return r, nil
}

func isFinite(f float64) bool { return !math.IsNaN(f) && !math.IsInf(f, 0) }
