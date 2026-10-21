package calculator

import (
	"errors"
	"math"
	"testing"
)

func f(v float64) *float64 { return &v }

func TestCalculate(t *testing.T) {
	tests := []struct {
		name    string
		op      Operation
		a, b    *float64
		want    float64
		wantErr error
	}{
		{"add", Add, f(2), f(3), 5, nil},
		{"add negatives", Add, f(-2.5), f(-0.5), -3, nil},
		{"subtract", Subtract, f(10), f(4), 6, nil},
		{"subtract to negative", Subtract, f(4), f(10), -6, nil},
		{"multiply", Multiply, f(3), f(-4), -12, nil},
		{"multiply by zero gives positive zero", Multiply, f(-3), f(0), 0, nil},
		{"divide", Divide, f(10), f(4), 2.5, nil},
		{"divide by zero", Divide, f(1), f(0), 0, ErrDivisionByZero},
		{"power", Power, f(2), f(10), 1024, nil},
		{"power negative exponent", Power, f(2), f(-1), 0.5, nil},
		{"power zero to negative", Power, f(0), f(-1), 0, ErrDivisionByZero},
		{"power undefined", Power, f(-8), f(0.5), 0, ErrUndefinedResult},
		{"power overflow", Power, f(10), f(400), 0, ErrResultNotFinite},
		{"sqrt", Sqrt, f(81), nil, 9, nil},
		{"sqrt ignores b", Sqrt, f(4), f(100), 2, nil},
		{"sqrt negative", Sqrt, f(-1), nil, 0, ErrNegativeSqrt},
		{"percentage", Percentage, f(15), f(200), 30, nil},
		{"multiply overflow", Multiply, f(math.MaxFloat64), f(10), 0, ErrResultNotFinite},
		{"unknown op", Operation("modulo"), f(1), f(2), 0, ErrUnknownOperation},
		{"missing a", Add, nil, f(2), 0, ErrMissingOperand},
		{"missing b", Add, f(1), nil, 0, ErrMissingOperand},
		{"missing a unary", Sqrt, nil, nil, 0, ErrMissingOperand},
		{"NaN operand", Add, f(math.NaN()), f(1), 0, ErrNonFiniteOperand},
		{"Inf operand", Add, f(1), f(math.Inf(1)), 0, ErrNonFiniteOperand},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Calculate(tt.op, tt.a, tt.b)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
			if math.Signbit(got) && got == 0 {
				t.Errorf("got negative zero")
			}
		})
	}
}

func TestArity(t *testing.T) {
	for _, op := range Supported() {
		n, err := Arity(op)
		if err != nil {
			t.Fatalf("Arity(%s) error: %v", op, err)
		}
		want := 2
		if op == Sqrt {
			want = 1
		}
		if n != want {
			t.Errorf("Arity(%s) = %d, want %d", op, n, want)
		}
	}
	if _, err := Arity("nope"); !errors.Is(err, ErrUnknownOperation) {
		t.Errorf("expected ErrUnknownOperation, got %v", err)
	}
}

func TestSupportedMatchesRegistry(t *testing.T) {
	if len(Supported()) != len(operations) {
		t.Fatalf("Supported() has %d ops, registry has %d", len(Supported()), len(operations))
	}
}
