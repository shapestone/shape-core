package tokenizer

import "math"

//
// Number Utilities - Functions for numeric operations
//

// NearlyEqual compares two float64 numbers and returns true if they are nearly equal
// within the specified epsilon tolerance.
func NearlyEqual(a float64, b float64, epsilon float64) bool {
	// already equal?
	if a == b {
		return true
	}

	diff := math.Abs(a - b)
	if a == 0.0 || b == 0.0 || diff < math.SmallestNonzeroFloat64 {
		return diff < epsilon*math.SmallestNonzeroFloat64
	}

	return diff/(math.Abs(a)+math.Abs(b)) < epsilon
}

// Difference returns the absolute difference between two float64 numbers.
func Difference(a float64, b float64) float64 {
	return math.Abs(a - b)
}

// MaxInt returns the maximum of two integers.
func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// MinInt returns the minimum of two integers.
func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// PowInt calculates m to the n-th power for integers.
func PowInt(m, n int64) int64 {
	if n == 0 {
		return 1
	}

	if n == 1 {
		return m
	}

	result := m
	for i := int64(2); i <= n; i++ {
		result *= m
	}
	return result
}
