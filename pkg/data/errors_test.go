package data

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestSyntaxError(t *testing.T) {
	tests := []struct {
		name string
		err  *SyntaxError
		want string
	}{
		{"with position", &SyntaxError{Format: "json", Offset: 12, Line: 2, Col: 5, Msg: "bad"},
			"json: syntax error at line 2, column 5 (offset 12): bad"},
		{"offset only", &SyntaxError{Format: "csv", Offset: 3, Msg: "bad"},
			"csv: syntax error at offset 3: bad"},
		{"no format", &SyntaxError{Msg: "x"}, "data: syntax error at offset 0: x"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
			wrapped := fmt.Errorf("parse: %w", tt.err)
			if !errors.Is(wrapped, ErrSyntax) || errors.Is(wrapped, ErrLimit) {
				t.Error("Is mismatch")
			}
			var se *SyntaxError
			if !errors.As(wrapped, &se) || se != tt.err {
				t.Error("As failed")
			}
		})
	}
}

func TestLimitError(t *testing.T) {
	e := &LimitError{Limit: LimitDepth, Max: 10, Got: 11}
	if got, want := e.Error(), "limit exceeded: depth (max 10, got 11)"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	wrapped := fmt.Errorf("parse: %w", e)
	if !errors.Is(wrapped, ErrLimit) || errors.Is(wrapped, ErrSyntax) {
		t.Error("Is mismatch")
	}
	var le *LimitError
	if !errors.As(wrapped, &le) || le.Got != 11 {
		t.Error("As failed")
	}
}

func TestContextErrorsPassThrough(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := NewGuard(Limits{Context: ctx}).Check()
	if err != context.Canceled {
		t.Errorf("err = %v, want context.Canceled itself", err)
	}
	if errors.Is(err, ErrLimit) || errors.Is(err, ErrSyntax) {
		t.Error("context error matched a data sentinel")
	}

	dctx, dcancel := context.WithTimeout(context.Background(), 0)
	defer dcancel()
	if err := NewGuard(Limits{Context: dctx}).Check(); err != context.DeadlineExceeded {
		t.Errorf("err = %v, want DeadlineExceeded itself", err)
	}
}
