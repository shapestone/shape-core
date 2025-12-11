package tokenizer

import (
	"testing"
)

func TestNearlyEqualShouldPass(t *testing.T) {
	// Given
	a := 1.0001
	b := 1.0002
	e := 0.001

	// When
	r := NearlyEqual(a, b, e)

	// Then
	if !r {
		t.Fatalf("%f is not nearly equal (%v) to %f for epsilon %f", a, r, b, e)
	}
}

func TestNearlyEqualShouldFail(t *testing.T) {
	// Given
	a := 1.0001
	b := 1.0002
	e := 0.00001

	// When
	r := NearlyEqual(a, b, e)

	// Then
	if r {
		t.Fatalf("%f is not nearly equal (%v) to %f for epsilon %f", a, r, b, e)
	}
}

func TestDifference(t *testing.T) {
	// Given
	a := 1.0001
	b := 1.0002
	e := 0.0001

	// When
	d := Difference(a, b)

	// Then
	if d == e {
		t.Fatalf("The difference %f between %f and %f is different than %f", d, a, b, e)
	}
}

func TestMinShouldYieldLeftValue(t *testing.T) {
	// Given
	left := 5
	right := 9

	// When
	res := MinInt(left, right)

	// Then
	if res != left {
		t.Fatalf("Min of %v and %v is expected to yield %v", left, right, left)
	}
}

func TestMinShouldYieldRightValue(t *testing.T) {
	// Given
	left := 11
	right := 7

	// When
	res := MinInt(left, right)

	// Then
	if res != right {
		t.Fatalf("Min of %v and %v is expected to yield %v", left, right, right)
	}
}

func TestMaxShouldYieldLeftValue(t *testing.T) {
	// Given
	left := 27
	right := 11

	// When
	res := MaxInt(left, right)

	// Then
	if res != left {
		t.Fatalf("Max of %v and %v is expected to yield %v", left, right, left)
	}
}

func TestMaxShouldYieldRightValue(t *testing.T) {
	// Given
	left := 13
	right := 42

	// When
	res := MaxInt(left, right)

	// Then
	if res != right {
		t.Fatalf("Max of %v and %v is expected to yield %v", left, right, left)
	}
}

func TestTwoToPowerOfZeroShouldYieldOne(t *testing.T) {
	// Given
	var m int64 = 2
	var n int64 = 0

	// When
	res := PowInt(m, n)

	// Then
	if res != 1 {
		t.Fatalf("Power of %v and %v is expected to yield %v", m, n, 1)
	}
}

func TestTwoToPowerOfOneShouldYieldTwo(t *testing.T) {
	// Given
	var m int64 = 2
	var n int64 = 1

	// When
	res := PowInt(m, n)

	// Then
	if res != m {
		t.Fatalf("Power of %v and %v is expected to yield %v", m, n, m)
	}
}

func TestTwoToPowerOfThreeShouldYieldEight(t *testing.T) {
	// Given
	var m int64 = 2
	var n int64 = 3

	// When
	res := PowInt(m, n)

	// Then
	if res != m*m*m {
		t.Fatalf("Power of %v and %v is expected to yield %v", m, n, m*m*m)
	}
}
