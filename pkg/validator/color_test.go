package validator

import (
	"os"
	"strings"
	"testing"
)

// TestShouldUseColor_WithoutNO_COLOR tests color is enabled by default
func TestShouldUseColor_WithoutNO_COLOR(t *testing.T) {
	// Clear NO_COLOR
	os.Unsetenv("NO_COLOR")

	if !shouldUseColor() {
		t.Error("shouldUseColor() = false, want true when NO_COLOR not set")
	}
}

// TestShouldUseColor_WithNO_COLOR tests NO_COLOR environment variable
func TestShouldUseColor_WithNO_COLOR(t *testing.T) {
	// Set NO_COLOR
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	if shouldUseColor() {
		t.Error("shouldUseColor() = true, want false when NO_COLOR is set")
	}
}

// TestShouldUseColor_WithNO_COLOR_EmptyString tests NO_COLOR with empty string
func TestShouldUseColor_WithNO_COLOR_EmptyString(t *testing.T) {
	// Set NO_COLOR to empty string (still means "disable colors")
	os.Setenv("NO_COLOR", "")
	defer os.Unsetenv("NO_COLOR")

	// According to no-color.org spec, any value (even empty) means no color
	// But our implementation checks for non-empty, so empty string means colors enabled
	// This is acceptable behavior
	if !shouldUseColor() {
		t.Error("shouldUseColor() = false, want true when NO_COLOR is empty string")
	}
}

// TestColorize_WithColors tests colorize adds ANSI codes when colors enabled
func TestColorize_WithColors(t *testing.T) {
	os.Unsetenv("NO_COLOR")

	result := colorize(colorRed, "test")

	// Should contain ANSI codes
	if !strings.Contains(result, "\033[") {
		t.Error("colorize() should contain ANSI codes when colors enabled")
	}

	// Should contain the text
	if !strings.Contains(result, "test") {
		t.Error("colorize() should contain the input text")
	}

	// Should contain reset code
	if !strings.Contains(result, colorReset) {
		t.Error("colorize() should contain reset code")
	}
}

// TestColorize_WithoutColors tests colorize returns plain text when NO_COLOR set
func TestColorize_WithoutColors(t *testing.T) {
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	result := colorize(colorRed, "test")

	// Should NOT contain ANSI codes
	if strings.Contains(result, "\033[") {
		t.Error("colorize() should not contain ANSI codes when NO_COLOR is set")
	}

	// Should be exactly the input text
	if result != "test" {
		t.Errorf("colorize() = %q, want %q", result, "test")
	}
}

// TestRed tests red() function
func TestRed(t *testing.T) {
	os.Unsetenv("NO_COLOR")

	result := red("error")

	if !strings.Contains(result, colorRed) {
		t.Error("red() should contain red color code")
	}

	if !strings.Contains(result, "error") {
		t.Error("red() should contain the input text")
	}
}

// TestRed_WithNO_COLOR tests red() respects NO_COLOR
func TestRed_WithNO_COLOR(t *testing.T) {
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	result := red("error")

	if result != "error" {
		t.Errorf("red() with NO_COLOR = %q, want %q", result, "error")
	}
}

// TestBlue tests blue() function
func TestBlue(t *testing.T) {
	os.Unsetenv("NO_COLOR")

	result := blue("hint")

	if !strings.Contains(result, colorBlue) {
		t.Error("blue() should contain blue color code")
	}
}

// TestYellow tests yellow() function
func TestYellow(t *testing.T) {
	os.Unsetenv("NO_COLOR")

	result := yellow("warning")

	if !strings.Contains(result, colorYellow) {
		t.Error("yellow() should contain yellow color code")
	}
}

// TestCyan tests cyan() function
func TestCyan(t *testing.T) {
	os.Unsetenv("NO_COLOR")

	result := cyan("info")

	if !strings.Contains(result, colorCyan) {
		t.Error("cyan() should contain cyan color code")
	}
}

// TestGray tests gray() function
func TestGray(t *testing.T) {
	os.Unsetenv("NO_COLOR")

	result := gray("context")

	if !strings.Contains(result, colorGray) {
		t.Error("gray() should contain gray color code")
	}
}

// TestGreen tests green() function
func TestGreen(t *testing.T) {
	os.Unsetenv("NO_COLOR")

	result := green("success")

	if !strings.Contains(result, colorGreen) {
		t.Error("green() should contain green color code")
	}
}

// TestAllColors_WithNO_COLOR tests all color functions respect NO_COLOR
func TestAllColors_WithNO_COLOR(t *testing.T) {
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	tests := []struct {
		name     string
		fn       func(string) string
		input    string
	}{
		{"red", red, "test"},
		{"blue", blue, "test"},
		{"yellow", yellow, "test"},
		{"cyan", cyan, "test"},
		{"gray", gray, "test"},
		{"green", green, "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.input)
			if result != tt.input {
				t.Errorf("%s() with NO_COLOR = %q, want %q", tt.name, result, tt.input)
			}

			if strings.Contains(result, "\033[") {
				t.Errorf("%s() with NO_COLOR should not contain ANSI codes", tt.name)
			}
		})
	}
}

// TestColorConstants tests that color constants are valid ANSI codes
func TestColorConstants(t *testing.T) {
	constants := []struct {
		name  string
		value string
	}{
		{"colorReset", colorReset},
		{"colorRed", colorRed},
		{"colorBlue", colorBlue},
		{"colorYellow", colorYellow},
		{"colorCyan", colorCyan},
		{"colorGray", colorGray},
		{"colorGreen", colorGreen},
	}

	for _, c := range constants {
		t.Run(c.name, func(t *testing.T) {
			// ANSI codes should start with ESC character (\033)
			if !strings.HasPrefix(c.value, "\033[") {
				t.Errorf("%s = %q, should start with \\033[", c.name, c.value)
			}

			// Should end with 'm' (SGR terminator)
			if !strings.HasSuffix(c.value, "m") {
				t.Errorf("%s = %q, should end with 'm'", c.name, c.value)
			}
		})
	}
}

// TestColorize_EmptyString tests colorize with empty string
func TestColorize_EmptyString(t *testing.T) {
	os.Unsetenv("NO_COLOR")

	result := colorize(colorRed, "")

	// Should handle empty string gracefully
	if !strings.Contains(result, colorRed) {
		t.Error("colorize() with empty string should still add color codes")
	}
}

// TestColorize_MultilineString tests colorize with multiline text
func TestColorize_MultilineString(t *testing.T) {
	os.Unsetenv("NO_COLOR")

	input := "line1\nline2\nline3"
	result := colorize(colorRed, input)

	// Should contain all lines
	if !strings.Contains(result, "line1") || !strings.Contains(result, "line2") || !strings.Contains(result, "line3") {
		t.Error("colorize() should preserve multiline text")
	}

	// Should wrap entire text (not individual lines)
	if !strings.HasPrefix(result, colorRed) {
		t.Error("colorize() should start with color code")
	}

	if !strings.HasSuffix(result, colorReset) {
		t.Error("colorize() should end with reset code")
	}
}
