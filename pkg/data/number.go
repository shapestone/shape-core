package data

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Errors returned by Number accessors.
var (
	// ErrNumberSyntax means the text is not a valid JSON number.
	ErrNumberSyntax = errors.New("invalid number syntax")

	// ErrNumberRange means the value does not fit the requested type.
	ErrNumberRange = errors.New("number out of range")

	// ErrNotInteger means an integer was requested from a Decimal number.
	ErrNotInteger = errors.New("number is not an integer")
)

// DefaultMaxBigDigits is the most mantissa digits Big and BigInt accept.
// math/big parsing is quadratic in the digit count (one million digits takes
// about a second), so the default keeps a hostile number cheap. Use
// BigWithMaxDigits or BigIntWithMaxDigits to choose another bound.
const DefaultMaxBigDigits = 4096

// maxErrText is the most bytes of number text quoted in an error message.
const maxErrText = 32

// maxBigExponent bounds the exponent accepted by Big, so that text such as
// "1e999999999" cannot make math/big allocate gigabytes.
const maxBigExponent = 1 << 16

// Number is a numeric value that keeps its exact source text. Writers must
// emit Text unchanged. The accepted grammar is JSON's:
//
//	-? (0 | [1-9][0-9]*) (. [0-9]+)? ([eE] [+-]? [0-9]+)?
//
// A Number is an Integer when its text has no '.' and no exponent, and a
// Decimal otherwise, so "1.0", "1e3" and "-0.5" are all Decimal.
//
// The zero Number has empty text, which is not valid: Valid reports false,
// every accessor returns ErrNumberSyntax, but Kind and IsInteger look only
// for '.', 'e' and 'E' and so report Integer. Check Valid, or the kind of the
// enclosing Value, before trusting them on a Number of unknown origin.
type Number struct {
	Text string
}

// ParseNumber validates text against the JSON number grammar.
func ParseNumber(text string) (Number, error) {
	if !validNumber(text) {
		return Number{}, numErr(ErrNumberSyntax, text)
	}
	return Number{Text: text}, nil
}

// String returns the exact source text.
func (n Number) String() string { return n.Text }

// Valid reports whether Text follows the JSON number grammar.
func (n Number) Valid() bool { return validNumber(n.Text) }

// Kind returns KindInteger or KindDecimal, decided by the written form.
// The text is not validated.
func (n Number) Kind() Kind {
	if n.IsInteger() {
		return KindInteger
	}
	return KindDecimal
}

// IsInteger reports whether the text has no '.' and no exponent.
func (n Number) IsInteger() bool {
	return !strings.ContainsAny(n.Text, ".eE")
}

// Int64 returns the value as an int64. It returns an error wrapping
// ErrNumberRange on overflow (it never clamps), ErrNotInteger for Decimal
// numbers and ErrNumberSyntax for invalid text. "-0" is 0.
func (n Number) Int64() (int64, error) {
	if !n.Valid() {
		return 0, numErr(ErrNumberSyntax, n.Text)
	}
	if !n.IsInteger() {
		return 0, numErr(ErrNotInteger, n.Text)
	}
	v, err := strconv.ParseInt(n.Text, 10, 64)
	if err != nil {
		return 0, numErr(ErrNumberRange, n.Text)
	}
	return v, nil
}

// Float64 returns the nearest float64. When the magnitude is too large it
// returns ±Inf and an error wrapping ErrNumberRange. Precision loss is not
// an error; use Big for exact values.
func (n Number) Float64() (float64, error) {
	if !n.Valid() {
		return 0, numErr(ErrNumberSyntax, n.Text)
	}
	v, err := strconv.ParseFloat(n.Text, 64)
	if err != nil {
		return v, numErr(ErrNumberRange, n.Text)
	}
	return v, nil
}

// Big returns the exact value as a rational number. It is
// BigWithMaxDigits(DefaultMaxBigDigits).
func (n Number) Big() (*big.Rat, error) {
	return n.BigWithMaxDigits(DefaultMaxBigDigits)
}

// BigWithMaxDigits returns the exact value as a rational number. Text with
// more than maxDigits mantissa digits, or an exponent with a magnitude above
// 65536, returns an error wrapping ErrNumberRange; so does any text math/big
// refuses. The bounds limit CPU and memory: even "1e65536" becomes a
// roughly 27 KB Rat from 7 bytes of text.
func (n Number) BigWithMaxDigits(maxDigits int) (*big.Rat, error) {
	if !n.Valid() {
		return nil, numErr(ErrNumberSyntax, n.Text)
	}
	if mantissaDigits(n.Text) > maxDigits {
		return nil, numErr(ErrNumberRange, n.Text)
	}
	if exp := exponent(n.Text); exp > maxBigExponent || exp < -maxBigExponent {
		return nil, numErr(ErrNumberRange, n.Text)
	}
	r, ok := new(big.Rat).SetString(n.Text)
	if !ok { // text is valid, so math/big refused it for size
		return nil, numErr(ErrNumberRange, n.Text)
	}
	return r, nil
}

// BigInt returns the exact value of an Integer number. It is
// BigIntWithMaxDigits(DefaultMaxBigDigits).
func (n Number) BigInt() (*big.Int, error) {
	return n.BigIntWithMaxDigits(DefaultMaxBigDigits)
}

// BigIntWithMaxDigits returns the exact value of an Integer number, or an
// error wrapping ErrNumberRange if it has more than maxDigits digits.
func (n Number) BigIntWithMaxDigits(maxDigits int) (*big.Int, error) {
	if !n.Valid() {
		return nil, numErr(ErrNumberSyntax, n.Text)
	}
	if !n.IsInteger() {
		return nil, numErr(ErrNotInteger, n.Text)
	}
	if mantissaDigits(n.Text) > maxDigits {
		return nil, numErr(ErrNumberRange, n.Text)
	}
	v, ok := new(big.Int).SetString(n.Text, 10)
	if !ok {
		return nil, numErr(ErrNumberSyntax, n.Text)
	}
	return v, nil
}

// numErr wraps sentinel with the number text, truncated so that a huge
// invalid number cannot make a huge error message.
func numErr(sentinel error, text string) error {
	if len(text) > maxErrText {
		cut := maxErrText
		for cut > 0 && !utf8.RuneStart(text[cut]) {
			cut--
		}
		return fmt.Errorf("%w: %q (truncated)", sentinel, text[:cut])
	}
	return fmt.Errorf("%w: %q", sentinel, text)
}

// mantissaDigits counts the digits before any exponent.
func mantissaDigits(s string) int {
	d := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == 'e' || c == 'E' {
			break
		}
		if c >= '0' && c <= '9' {
			d++
		}
	}
	return d
}

// exponent returns the exponent of valid number text, clamped to ±MaxInt32
// so very long exponent digits cannot overflow.
func exponent(s string) int {
	i := strings.IndexAny(s, "eE")
	if i < 0 {
		return 0
	}
	i++
	neg := false
	if s[i] == '+' || s[i] == '-' {
		neg = s[i] == '-'
		i++
	}
	e := 0
	for ; i < len(s); i++ {
		e = e*10 + int(s[i]-'0')
		if e > math.MaxInt32 {
			e = math.MaxInt32
		}
	}
	if neg {
		return -e
	}
	return e
}

func validNumber(s string) bool {
	i, n := 0, len(s)
	if i < n && s[i] == '-' {
		i++
	}
	if i >= n {
		return false
	}
	switch {
	case s[i] == '0':
		i++
	case s[i] >= '1' && s[i] <= '9':
		i = skipDigits(s, i)
	default:
		return false
	}
	if i < n && s[i] == '.' {
		j := skipDigits(s, i+1)
		if j == i+1 {
			return false
		}
		i = j
	}
	if i < n && (s[i] == 'e' || s[i] == 'E') {
		i++
		if i < n && (s[i] == '+' || s[i] == '-') {
			i++
		}
		j := skipDigits(s, i)
		if j == i {
			return false
		}
		i = j
	}
	return i == n
}

func skipDigits(s string, i int) int {
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	return i
}
