package data

import (
	"errors"
	"math"
	"strings"
	"testing"
)

func TestParseNumber(t *testing.T) {
	huge := strings.Repeat("9", 100)
	tests := []struct {
		text string
		ok   bool
		kind Kind
	}{
		{"0", true, KindInteger},
		{"-0", true, KindInteger},
		{"7", true, KindInteger},
		{"-12", true, KindInteger},
		{huge, true, KindInteger},
		{"1.0", true, KindDecimal},
		{"-0.5", true, KindDecimal},
		{"1e3", true, KindDecimal},
		{"1E+3", true, KindDecimal},
		{"1.5e-3", true, KindDecimal},
		{"1e400", true, KindDecimal},
		{"", false, 0},
		{"-", false, 0},
		{"+1", false, 0},
		{"01", false, 0},
		{"-01", false, 0},
		{".5", false, 0},
		{"1.", false, 0},
		{"1e", false, 0},
		{"1e+", false, 0},
		{"1.e3", false, 0},
		{" 1", false, 0},
		{"1 ", false, 0},
		{"0x1F", false, 0},
		{"0o17", false, 0},
		{"1_000", false, 0},
		{".inf", false, 0},
		{"NaN", false, 0},
		{"Infinity", false, 0},
		{"1.2.3", false, 0},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			n, err := ParseNumber(tt.text)
			if tt.ok {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if n.Text != tt.text || n.String() != tt.text {
					t.Errorf("text changed: %q", n.Text)
				}
				if n.Kind() != tt.kind {
					t.Errorf("kind = %v, want %v", n.Kind(), tt.kind)
				}
				return
			}
			if !errors.Is(err, ErrNumberSyntax) {
				t.Errorf("err = %v, want ErrNumberSyntax", err)
			}
		})
	}
}

func TestNumberInt64(t *testing.T) {
	tests := []struct {
		text string
		want int64
		err  error
	}{
		{"0", 0, nil},
		{"-0", 0, nil},
		{"9223372036854775807", math.MaxInt64, nil},
		{"-9223372036854775808", math.MinInt64, nil},
		{"9223372036854775808", 0, ErrNumberRange},
		{"-9223372036854775809", 0, ErrNumberRange},
		{strings.Repeat("9", 100), 0, ErrNumberRange},
		{"1.0", 0, ErrNotInteger},
		{"1e3", 0, ErrNotInteger},
		{"abc", 0, ErrNumberSyntax},
		{"", 0, ErrNumberSyntax},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got, err := Number{Text: tt.text}.Int64()
			if !errors.Is(err, tt.err) || (tt.err == nil && err != nil) {
				t.Fatalf("err = %v, want %v", err, tt.err)
			}
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestNumberFloat64(t *testing.T) {
	tests := []struct {
		text string
		want float64
		err  error
	}{
		{"1.5", 1.5, nil},
		{"-0.5", -0.5, nil},
		{"1e3", 1000, nil},
		{"42", 42, nil},
		{"1e-400", 0, nil},
		{"1e400", math.Inf(1), ErrNumberRange},
		{"-1e400", math.Inf(-1), ErrNumberRange},
		{"Infinity", 0, ErrNumberSyntax},
		{"0x1p-2", 0, ErrNumberSyntax},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got, err := Number{Text: tt.text}.Float64()
			if !errors.Is(err, tt.err) || (tt.err == nil && err != nil) {
				t.Fatalf("err = %v, want %v", err, tt.err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
	if v, _ := (Number{Text: "-0"}).Float64(); !math.Signbit(v) {
		t.Error("-0 lost its sign")
	}
}

func TestNumberBig(t *testing.T) {
	huge := strings.Repeat("9", 100)
	tests := []struct {
		text string
		want string // Rat.RatString
		err  error
	}{
		{huge, huge, nil},
		{"0.25", "1/4", nil},
		{"1e3", "1000", nil},
		{"-5E-1", "-1/2", nil},
		{"1e400", "1" + strings.Repeat("0", 400), nil},
		{"1e999999999", "", ErrNumberRange},
		{"1e-999999999", "", ErrNumberRange},
		{"1e99999999999999999999", "", ErrNumberRange},
		{"x", "", ErrNumberSyntax},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got, err := Number{Text: tt.text}.Big()
			if !errors.Is(err, tt.err) || (tt.err == nil && err != nil) {
				t.Fatalf("err = %v, want %v", err, tt.err)
			}
			if err == nil && got.RatString() != tt.want {
				t.Errorf("got %s, want %s", got.RatString(), tt.want)
			}
		})
	}
}

func TestNumberBigInt(t *testing.T) {
	huge := strings.Repeat("9", 100)
	if v, err := (Number{Text: huge}).BigInt(); err != nil || v.String() != huge {
		t.Errorf("BigInt(huge) = %v, %v", v, err)
	}
	if _, err := (Number{Text: "1.0"}).BigInt(); !errors.Is(err, ErrNotInteger) {
		t.Errorf("err = %v", err)
	}
	if _, err := (Number{Text: "z"}).BigInt(); !errors.Is(err, ErrNumberSyntax) {
		t.Errorf("err = %v", err)
	}
}
