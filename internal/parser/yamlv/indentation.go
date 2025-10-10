package yamlv

import (
	"strings"
)

// YAMLLine represents a parsed line with its indentation and content.
type YAMLLine struct {
	Indent  int    // Number of leading spaces
	Content string // Content after stripping indent and comments
	LineNum int    // Line number (1-indexed)
}

// ParseLines splits input into lines and tracks indentation.
func ParseLines(input string) []YAMLLine {
	lines := strings.Split(input, "\n")
	result := make([]YAMLLine, 0)

	for i, line := range lines {
		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Calculate indentation
		indent := 0
		for _, r := range line {
			if r == ' ' {
				indent++
			} else if r == '\t' {
				indent += 2 // Treat tab as 2 spaces
			} else {
				break
			}
		}

		// Get content (trimmed)
		content := strings.TrimSpace(line)

		// Skip comment-only lines
		if strings.HasPrefix(content, "#") {
			continue
		}

		// Strip inline comments
		if idx := strings.Index(content, "#"); idx >= 0 {
			content = strings.TrimSpace(content[:idx])
		}

		result = append(result, YAMLLine{
			Indent:  indent,
			Content: content,
			LineNum: i + 1,
		})
	}

	return result
}
