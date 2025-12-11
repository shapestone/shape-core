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
