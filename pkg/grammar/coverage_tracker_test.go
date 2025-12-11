package grammar

import "testing"

func TestCoverageTracker_Basic(t *testing.T) {
	input := `
		Value = Type | Literal ;
		Type = Identifier ;
		Literal = String ;
	`

	grammar, err := ParseEBNF(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	tracker := NewCoverageTracker(grammar)

	// Record some rules
	tracker.RecordRule("Value")
	tracker.RecordRule("Type")

	// Generate report
	report := tracker.Report()

	if report.TotalRules != 3 {
		t.Errorf("expected 3 total rules, got %d", report.TotalRules)
	}

	if report.CoveredRules != 2 {
		t.Errorf("expected 2 covered rules, got %d", report.CoveredRules)
	}

	// Expected: 2/3 * 100 = 66.7%
	if report.Percentage < 66.0 || report.Percentage > 67.0 {
		t.Errorf("expected ~66.7%% coverage, got %.1f%%", report.Percentage)
	}

	if len(report.UncoveredRules) != 1 {
		t.Errorf("expected 1 uncovered rule, got %d", len(report.UncoveredRules))
	}

	if report.UncoveredRules[0] != "Literal" {
		t.Errorf("expected 'Literal' uncovered, got %s", report.UncoveredRules[0])
	}
}

func TestCoverageTracker_FullCoverage(t *testing.T) {
	input := `Value = "test" ;`

	grammar, err := ParseEBNF(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	tracker := NewCoverageTracker(grammar)
	tracker.RecordRule("Value")

	report := tracker.Report()

	if report.Percentage != 100.0 {
		t.Errorf("expected 100%% coverage, got %.1f%%", report.Percentage)
	}

	if len(report.UncoveredRules) != 0 {
		t.Errorf("expected no uncovered rules, got %v", report.UncoveredRules)
	}
}

func TestCoverageTracker_Reset(t *testing.T) {
	input := `Value = "test" ;`

	grammar, err := ParseEBNF(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	tracker := NewCoverageTracker(grammar)
	tracker.RecordRule("Value")

	// Reset
	tracker.Reset()

	report := tracker.Report()

	if report.Percentage != 0.0 {
		t.Errorf("expected 0%% coverage after reset, got %.1f%%", report.Percentage)
	}
}

func TestCoverageTracker_GetInvocationCount(t *testing.T) {
	input := `Value = "test" ;`

	grammar, err := ParseEBNF(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	tracker := NewCoverageTracker(grammar)

	// Record rule multiple times
	tracker.RecordRule("Value")
	tracker.RecordRule("Value")
	tracker.RecordRule("Value")

	count := tracker.GetInvocationCount("Value")
	if count != 3 {
		t.Errorf("expected invocation count 3, got %d", count)
	}

	// Test non-existent rule
	count = tracker.GetInvocationCount("NonExistent")
	if count != 0 {
		t.Errorf("expected invocation count 0 for non-existent rule, got %d", count)
	}
}

func TestCoverageTracker_IsCovered(t *testing.T) {
	input := `
		Value = Type ;
		Type = "test" ;
	`

	grammar, err := ParseEBNF(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	tracker := NewCoverageTracker(grammar)
	tracker.RecordRule("Value")

	if !tracker.IsCovered("Value") {
		t.Error("expected Value to be covered")
	}

	if tracker.IsCovered("Type") {
		t.Error("expected Type to not be covered")
	}

	if tracker.IsCovered("NonExistent") {
		t.Error("expected NonExistent to not be covered")
	}
}
