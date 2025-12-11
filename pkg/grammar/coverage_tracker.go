package grammar

import "fmt"

// CoverageTracker tracks which grammar rules are exercised during testing.
//
// Example usage:
//
//	grammar, _ := ParseEBNF(grammarSource)
//	tracker := NewCoverageTracker(grammar)
//
//	// During parsing, record rule invocations
//	tracker.RecordRule("ObjectNode")
//	tracker.RecordRule("Property")
//
//	// Generate report
//	report := tracker.Report()
//	fmt.Printf("Coverage: %.1f%%\n", report.Percentage)
type CoverageTracker struct {
	grammar     *Grammar
	invocations map[string]int
}

// NewCoverageTracker creates a new coverage tracker for a grammar.
func NewCoverageTracker(grammar *Grammar) *CoverageTracker {
	return &CoverageTracker{
		grammar:     grammar,
		invocations: make(map[string]int),
	}
}

// RecordRule records that a grammar rule was invoked.
//
// This should be called by parser implementations when entering
// a parse function that corresponds to a grammar rule.
//
// Example:
//
//	func (p *Parser) parseObject() (*ast.ObjectNode, error) {
//	    if p.tracker != nil {
//	        p.tracker.RecordRule("ObjectNode")
//	    }
//	    // ... parsing logic
//	}
func (t *CoverageTracker) RecordRule(ruleName string) {
	t.invocations[ruleName]++
}

// Reset clears all recorded invocations.
func (t *CoverageTracker) Reset() {
	t.invocations = make(map[string]int)
}

// Report generates a coverage report showing which rules were exercised.
func (t *CoverageTracker) Report() CoverageReport {
	totalRules := len(t.grammar.Rules)
	coveredRules := 0
	uncovered := []string{}

	// Count covered rules and find uncovered ones
	for _, rule := range t.grammar.Rules {
		if t.invocations[rule.Name] > 0 {
			coveredRules++
		} else {
			uncovered = append(uncovered, rule.Name)
		}
	}

	percentage := 0.0
	if totalRules > 0 {
		percentage = float64(coveredRules) / float64(totalRules) * 100.0
	}

	return CoverageReport{
		TotalRules:      totalRules,
		CoveredRules:    coveredRules,
		Percentage:      percentage,
		UncoveredRules:  uncovered,
		RuleInvocations: t.copyInvocations(),
	}
}

// GetInvocationCount returns the number of times a rule was invoked.
func (t *CoverageTracker) GetInvocationCount(ruleName string) int {
	return t.invocations[ruleName]
}

// IsCovered returns true if a rule has been invoked at least once.
func (t *CoverageTracker) IsCovered(ruleName string) bool {
	return t.invocations[ruleName] > 0
}

func (t *CoverageTracker) copyInvocations() map[string]int {
	copy := make(map[string]int, len(t.invocations))
	for k, v := range t.invocations {
		copy[k] = v
	}
	return copy
}

// FormatReport returns a human-readable coverage report.
func (r *CoverageReport) FormatReport() string {
	report := fmt.Sprintf("Grammar Coverage Report:\n")
	report += fmt.Sprintf("  Total Rules: %d\n", r.TotalRules)
	report += fmt.Sprintf("  Covered: %d\n", r.CoveredRules)
	report += fmt.Sprintf("  Coverage: %.1f%%\n\n", r.Percentage)

	if len(r.UncoveredRules) > 0 {
		report += fmt.Sprintf("Uncovered Rules (%d):\n", len(r.UncoveredRules))
		for _, rule := range r.UncoveredRules {
			report += fmt.Sprintf("  - %s\n", rule)
		}
	} else {
		report += "All rules covered!\n"
	}

	return report
}
