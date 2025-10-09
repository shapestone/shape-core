package tokenizer

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

//
// Text Utilities - Functions for text manipulation and comparison
//

// Regex patterns for text processing
var (
	stripMarginGroup  = regexp.MustCompile(`(?m)^[ \t]*\|(.*)(?:\n|$)`)
	stripColumnGroup  = regexp.MustCompile(`(?m)^[ \t]*\|(.*)(?:\|[ \t]*\n|\|[ \t]*$)`)
	reLastSpace       = regexp.MustCompile(` $`)
	reLastTab         = regexp.MustCompile(`\t$`)
	reTrailingWS      = regexp.MustCompile(`[\n\r\t ]*$`)
	reSpaces          = regexp.MustCompile(`[\t ]$`)
	reTab             = regexp.MustCompile(`\t`)
	reSpace           = regexp.MustCompile(` `)
	reNewline         = regexp.MustCompile(`\n`)
	reReturn          = regexp.MustCompile(`\r`)
	reFormFeed        = regexp.MustCompile(`\f`)
)

// StripMargin lets you define multiline strings where each line is prepended
// with optional whitespace and a pipeline symbol.
//
// Example:
//
//	StripMargin(`
//	  |line 1
//	  |line 2
//	`)
func StripMargin(s string) string {
	ms := stripMarginGroup.FindAllStringSubmatch(s, -1)
	if ms == nil {
		return ``
	}

	lines := ``
	for idx, m := range ms {
		if idx > 0 {
			lines += "\n"
		}
		lines += m[1]
	}

	return lines
}

// StripColumn lets you define multiline strings where each line is enclosed
// by pipeline symbols with optional whitespace.
//
// Example:
//
//	StripColumn(`
//	  |line 1|
//	  |line 2|
//	`)
func StripColumn(s string) string {
	ms := stripColumnGroup.FindAllStringSubmatch(s, -1)
	if ms == nil {
		return ``
	}

	lines := ``
	for idx, m := range ms {
		if idx > 0 {
			lines += "\n"
		}
		lines += m[1]
	}

	return lines
}

// Diff compares two strings and outputs a diff format.
// Returns the diff string and true if the strings matched.
func Diff(expected string, actual string) (string, bool) {
	expectedArr := showTabsArray(strings.Split(expected, "\n"))
	expectedWidth := len("Expected")
	for _, s := range expectedArr {
		if expectedWidth < len(s) {
			expectedWidth = len(s)
		}
	}

	actualArr := showTabsArray(strings.Split(actual, "\n"))
	actualWidth := len("Actual")
	for _, s := range actualArr {
		if actualWidth < len(s) {
			actualWidth = len(s)
		}
	}
	width := maxInt(expectedWidth, actualWidth)

	minVal := minInt(len(expectedArr), len(actualArr))
	var sb strings.Builder
	status := true
	sb.WriteString(Pad("Expected", width) + ` | ` + Pad("Actual", width) + "\n")
	sb.WriteString(strings.Repeat(`-`, width) + ` | ` + strings.Repeat(`-`, width) + "\n")
	for i := 0; i < minVal; i++ {
		if expectedArr[i] == actualArr[i] {
			sb.WriteString(Pad(expectedArr[i], width) + ` | ` + Pad(actualArr[i], width) + "\n")
		} else if expectedArr[i] != actualArr[i] {
			expected, actual := SpaceDiff(expectedArr[i], actualArr[i])
			sb.WriteString(Pad(expected, width) + " ≠ " + Pad(actual, width) + "\n")
			sd := StringDiff(expectedArr[i], actualArr[i])
			sb.WriteString(Pad(sd, width) + `   ` + Pad(sd, width) + "\n")
			return sb.String(), false
		}
	}
	if len(expectedArr) > len(actualArr) {
		expectedStr := expectedArr[minVal]
		actualStr := ``
		if utf8.RuneCountInString(expectedStr) == 0 {
			expectedStr = `␤`
		}
		sb.WriteString(Pad(expectedStr, width) + " ← " + Pad(actualStr, width) + "\n")
		status = false
	} else if len(expectedArr) < len(actualArr) {
		expectedStr := ``
		actualStr := actualArr[minVal]
		if utf8.RuneCountInString(actualStr) == 0 {
			actualStr = `␤`
		}
		sb.WriteString(Pad(expectedStr, width) + " → " + Pad(actualStr, width) + "\n")
		status = false
	}
	return sb.String(), status
}

// Pad performs right space padding on a string to reach the specified length.
func Pad(str string, length int) string {
	rc := length - utf8.RuneCountInString(str)
	return str + strings.Repeat(` `, rc)
}

// SpaceDiff compares two strings and highlights invisible whitespace differences.
func SpaceDiff(a, b string) (string, string) {
	ar := []rune(a)
	br := []rune(b)
	sl := minInt(len(ar), len(br))
	i := 0
	for ; i < sl; i++ {
		if ar[i] != br[i] {
			break
		}
	}
	as := showTabsAndSpaces(ar, i)
	bs := showTabsAndSpaces(br, i)
	return string(as), string(bs)
}

func showTabsAndSpaces(ra []rune, i int) []rune {
	for j := i - 1; j >= 0; j-- {
		if !isInvisible(ra[j]) {
			break
		}
		ra[j] = showInvisible(ra[j])
	}
	for j := i; j < len(ra); j++ {
		if !isInvisible(ra[j]) {
			break
		}
		ra[j] = showInvisible(ra[j])
	}
	return ra
}

func isInvisible(r rune) bool {
	switch r {
	case ' ':
		return true
	case '\t':
		return true
	default:
		return false
	}
}

func showInvisible(r rune) rune {
	switch r {
	case ' ':
		return '␣'
	case '\t':
		return '␉'
	default:
		return r
	}
}

// StringDiff returns a visual indicator showing where two strings differ.
func StringDiff(a, b string) string {
	ar := []rune(a)
	br := []rune(b)
	ml := minInt(len(ar), len(br))
	i := 0
	for i < ml {
		if ar[i] != br[i] {
			break
		}
		i++
	}
	return strings.Repeat(" ", i) + "△"
}

func showTabsArray(orig []string) []string {
	arr := make([]string, len(orig))

	for i, str := range orig {
		arr[i] = showTabs(str)
	}

	return arr
}

func showTabs(str string) string {
	str = reTab.ReplaceAllString(str, "␉")
	return str
}

// Flatten converts whitespace characters to their escaped string representations.
func Flatten(str string) string {
	str = strings.Replace(str, `%`, `%%`, -1)
	str = reNewline.ReplaceAllString(str, `\n`)
	str = reReturn.ReplaceAllString(str, `\r`)
	str = reTab.ReplaceAllString(str, `\t`)
	str = reFormFeed.ReplaceAllString(str, `\f`)
	return str
}

// FitString will shorten or pad the input string to the defined length.
// - If the input string is shorter than the defined length, it returns a padded string.
// - If the input string length equals the defined length, it returns the original string.
// - If the input string is longer, it returns "left ... right" format.
func FitString(str string, length int) string {
	str = strings.Replace(str, `%`, `%%`, -1)
	runeStr := []rune(str)
	strLen := len(runeStr)
	if strLen == length {
		return str
	} else if strLen <= length {
		return Pad(str, length)
	} else if length <= 7 {
		return string(runeStr[0:length])
	}
	remaining := length - 5
	right := remaining / 2
	left := remaining - right
	firstPart := string(runeStr[0:left])
	lastPart := string(runeStr[strLen-right : strLen])
	if len(lastPart) > 0 {
		return firstPart + ` ... ` + lastPart
	}
	return firstPart
}

//
// Rune Utilities - Functions for comparing rune slices
//

// RunesMatch returns true if two rune slices are equal.
func RunesMatch(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// RunesNoMatch returns true if two rune slices are not equal.
func RunesNoMatch(a, b []rune) bool {
	return !RunesMatch(a, b)
}

//
// Helper functions
//

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
