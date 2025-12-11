package grammar

// TestCase represents a single test case generated from a grammar.
type TestCase struct {
	// Name is a descriptive name for this test case
	Name string

	// Input is the test input string
	Input string

	// ShouldSucceed indicates whether this input should parse successfully
	ShouldSucceed bool

	// Description explains what this test case is testing
	Description string

	// RulePath is the path of grammar rules that produced this test
	RulePath []string
}

// TestOptions configures test case generation.
type TestOptions struct {
	// MaxDepth limits nesting depth (default: 5)
	MaxDepth int

	// CoverAllRules ensures every grammar rule is exercised at least once
	CoverAllRules bool

	// EdgeCases generates edge cases (empty, single, multiple elements)
	EdgeCases bool

	// InvalidCases generates invalid inputs that should fail to parse
	InvalidCases bool

	// MaxAlternatives limits how many alternatives to try per rule (default: all)
	MaxAlternatives int
}

// DefaultOptions returns reasonable default test generation options.
func DefaultOptions() TestOptions {
	return TestOptions{
		MaxDepth:        5,
		CoverAllRules:   true,
		EdgeCases:       true,
		InvalidCases:    true,
		MaxAlternatives: 0, // 0 means all
	}
}

// CoverageReport contains grammar coverage statistics.
type CoverageReport struct {
	// TotalRules is the total number of rules in the grammar
	TotalRules int

	// CoveredRules is the number of rules that were invoked
	CoveredRules int

	// Percentage is the coverage percentage (0-100)
	Percentage float64

	// UncoveredRules lists rules that were never invoked
	UncoveredRules []string

	// RuleInvocations maps rule names to invocation counts
	RuleInvocations map[string]int
}
