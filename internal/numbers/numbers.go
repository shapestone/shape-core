package numbers

import "math"

// The NearlyEqual function compares two float64 numbers and returns true if they are nearly equal
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

// The Difference function returns the differences between two float64 numbers
func Difference(a float64, b float64) float64 {
	return math.Abs(a - b)
}

// The MaxInt function returns the max of two values
func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// The MinInt function returns the min of two values
func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// The PowInt function calculates m to the n-th power
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
