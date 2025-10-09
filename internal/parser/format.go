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
// It examines the first non-whitespace character(s) to determine the format.
func DetectFormat(input string) (Format, error) {
	// Trim leading whitespace
	trimmed := strings.TrimLeftFunc(input, unicode.IsSpace)

	if len(trimmed) == 0 {
		return FormatUnknown, fmt.Errorf("empty input")
	}

	// Check first character
	firstChar := rune(trimmed[0])

	switch firstChar {
	case '{', '[':
		// JSON-like structure
		return FormatJSONV, nil
	case '<':
		// XML-like structure
		return FormatXMLV, nil
	default:
		// Check for key=value pattern (PropsV)
		if strings.Contains(trimmed, "=") && !strings.Contains(trimmed, ",") {
			return FormatPropsV, nil
		}

		// Check for CSV (comma-separated values)
		if strings.Contains(trimmed, ",") {
			return FormatCSVV, nil
		}

		// Default to unknown
		return FormatUnknown, fmt.Errorf("unable to detect format from input")
	}
}
