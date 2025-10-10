package parser

import (
	"fmt"
	"strings"
	"unicode"
)

// Format represents a validation schema format.
type Format int

const (
	// FormatUnknown indicates the format could not be determined
	FormatUnknown Format = iota
	// FormatJSONV is JSON with validation expressions
	FormatJSONV
	// FormatXMLV is XML with validation expressions
	FormatXMLV
	// FormatPropsV is Properties (key=value) with validation
	FormatPropsV
	// FormatCSVV is CSV with validation headers
	FormatCSVV
	// FormatYAMLV is YAML with validation expressions
	FormatYAMLV
	// FormatTEXTV is Text patterns with validation
	FormatTEXTV
)

// String returns the format name.
func (f Format) String() string {
	switch f {
	case FormatJSONV:
		return "JSONV"
	case FormatXMLV:
		return "XMLV"
	case FormatPropsV:
		return "PropsV"
	case FormatCSVV:
		return "CSVV"
	case FormatYAMLV:
		return "YAMLV"
	case FormatTEXTV:
		return "TEXTV"
	default:
		return "Unknown"
	}
}

// DetectFormat attempts to detect the format from input using heuristics.
// It examines the structure and syntax to determine the format.
//
// Detection order:
//  1. JSONV - starts with { or [
//  2. XMLV - starts with <
//  3. CSVV - comma-separated header row
//  4. PropsV - uses = for key-value pairs
//  5. YAMLV - uses : with YAML-style structure
//  6. TEXTV - uses : with dot notation (property.name: value)
func DetectFormat(input string) (Format, error) {
	// Trim leading whitespace
	trimmed := strings.TrimLeftFunc(input, unicode.IsSpace)

	if len(trimmed) == 0 {
		return FormatUnknown, fmt.Errorf("empty input")
	}

	// Get first non-whitespace character
	firstChar := rune(trimmed[0])

	// 1. JSONV: starts with { or [
	if firstChar == '{' || firstChar == '[' {
		return FormatJSONV, nil
	}

	// 2. XMLV: starts with <
	if firstChar == '<' {
		return FormatXMLV, nil
	}

	// Get first line for further analysis
	firstLine := getFirstLine(trimmed)

	// 3. CSVV: first line has commas but no = or : (header row)
	if strings.Contains(firstLine, ",") {
		// CSV if it has commas and doesn't look like other formats
		hasEquals := strings.Contains(firstLine, "=")
		hasColon := strings.Contains(firstLine, ":")
		if !hasEquals && !hasColon {
			return FormatCSVV, nil
		}
		// Could be CSVV with complex values, check if it's header-like
		// (no validation expressions in first line)
		if !hasValidationPattern(firstLine) {
			return FormatCSVV, nil
		}
	}

	// 4. PropsV: uses = for key-value pairs
	if strings.Contains(trimmed, "=") {
		return FormatPropsV, nil
	}

	// 5 & 6. YAMLV or TEXTV: both use : for key-value
	if strings.Contains(trimmed, ":") {
		// TEXTV uses dot notation (property.name: value)
		// YAMLV uses nested structure with indentation
		if hasDotNotation(trimmed) {
			return FormatTEXTV, nil
		}
		// YAMLV has indented nested structure or no dots
		return FormatYAMLV, nil
	}

	// Unable to detect format
	return FormatUnknown, fmt.Errorf("unable to detect format from input")
}

// getFirstLine returns the first non-empty line from input
func getFirstLine(input string) string {
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 0 && !strings.HasPrefix(trimmed, "#") {
			return trimmed
		}
	}
	return ""
}

// hasDotNotation checks if input uses dot notation (property.name: value)
// which is characteristic of TEXTV format
func hasDotNotation(input string) bool {
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) == 0 || strings.HasPrefix(trimmed, "#") {
			continue
		}
		// Check if line has pattern: word.word: value
		if colonIdx := strings.Index(trimmed, ":"); colonIdx > 0 {
			key := strings.TrimSpace(trimmed[:colonIdx])
			if strings.Contains(key, ".") {
				return true
			}
		}
	}
	return false
}

// hasValidationPattern checks if string contains validation expressions
// like UUID, String(1,100), etc.
func hasValidationPattern(s string) bool {
	// Check for type identifiers (uppercase start)
	if strings.Contains(s, "UUID") || strings.Contains(s, "String") ||
		strings.Contains(s, "Integer") || strings.Contains(s, "Email") {
		return true
	}
	// Check for function calls with parentheses
	if strings.Contains(s, "(") && strings.Contains(s, ")") {
		return true
	}
	return false
}
