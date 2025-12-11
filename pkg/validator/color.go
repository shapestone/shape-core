package validator

import (
	"os"
)

// ANSI color codes for terminal output
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorBlue   = "\033[34m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorGreen  = "\033[32m"
)

// shouldUseColor returns true if colored output should be used.
// Respects the NO_COLOR environment variable (https://no-color.org/)
func shouldUseColor() bool {
	// Respect NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return true
}

// colorize wraps text with ANSI color codes if colors are enabled
func colorize(colorCode, text string) string {
	if !shouldUseColor() {
		return text
	}
	return colorCode + text + colorReset
}

// Helper functions for specific colors
func red(text string) string {
	return colorize(colorRed, text)
}

func blue(text string) string {
	return colorize(colorBlue, text)
}

func yellow(text string) string {
	return colorize(colorYellow, text)
}

func cyan(text string) string {
	return colorize(colorCyan, text)
}

func gray(text string) string {
	return colorize(colorGray, text)
}

func green(text string) string {
	return colorize(colorGreen, text)
}
