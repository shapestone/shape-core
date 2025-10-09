package text

import (
	"github.com/shapestone/shape/internal/numbers"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Regex support functions
var stripMarginGroup = regexp.MustCompile(`(?m)^[ \t]*\|(.*)(?:\n|$)`)
var stripColumnGroup = regexp.MustCompile(`(?m)^[ \t]*\|(.*)(?:\|[ \t]*\n|\|[ \t]*$)`)

// The StripMargin function lets you define multiline strings where each line is prepended with optional whitespace
// and a pipeline symbol
//
// Code example:
//
// text.StripMargin(`
//
//	|<content line 1>
//	|<content line 2>
//
// `)
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

// The StripColumn function lets you define multiline strings where each line is prepended with optional whitespace
// and pipeline symbols
//
// Code example:
//
// text.StripColumn(`
//
//	|<content line 1>|
//	|<content line 2>|
//
// `)
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

// The Diff function compares two strings and outputs a diff format and a boolean value to indicate if the two strings
// matched
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
	width := numbers.MaxInt(expectedWidth, actualWidth)

	minVal := numbers.MinInt(len(expectedArr), len(actualArr))
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

// The Pad function is a right space padding function
func Pad(str string, length int) string {
	rc := length - utf8.RuneCountInString(str)
	return str + strings.Repeat(` `, rc)
}

var reLastSpace = regexp.MustCompile(` $`)
var reLastTab = regexp.MustCompile(`\t$`)

// Todo: deprecate?
func showLastSpace(str string) string {
	str = reLastSpace.ReplaceAllString(str, "␣")
	str = reLastTab.ReplaceAllString(str, "␉")
	return str
}

// The SpaceDiff function compares two strings and replace invisible whitespaces with symbols that represents them
func SpaceDiff(a, b string) (string, string) {
	ar := []rune(a)
	br := []rune(b)
	sl := numbers.MinInt(len(ar), len(br))
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

func StringDiff(a, b string) string {
	ar := []rune(a)
	br := []rune(b)
	ml := numbers.MinInt(len(ar), len(br))
	i := 0
	for i < ml {
		if ar[i] != br[i] {
			break
		}
		i++
	}
	return strings.Repeat(" ", i) + "△"
}

var reTrailingWS = regexp.MustCompile(`[\n\r\t ]*$`)

var reSpaces = regexp.MustCompile(`[\t ]$`)
var reTab = regexp.MustCompile(`\t`)
var reSpace = regexp.MustCompile(` `)

// Todo: deprecate?
func showAllSpaces(str string) string {
	if !reSpaces.MatchString(str) {
		return str
	}
	str = reTab.ReplaceAllString(str, "␉")
	str = reSpace.ReplaceAllString(str, "␣")
	return str
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

// Todo clean up
// Shard REs with above function
var reNewline = regexp.MustCompile(`\n`)
var reReturn = regexp.MustCompile(`\r`)

// var reTab = regexp.MustCompile(`\t`)
var reFormFeed = regexp.MustCompile(`\f`)

func Flatten(str string) string {
	str = strings.Replace(str, `%`, `%%`, -1)
	str = reNewline.ReplaceAllString(str, `\n`)
	str = reReturn.ReplaceAllString(str, `\r`)
	str = reTab.ReplaceAllString(str, `\t`)
	str = reFormFeed.ReplaceAllString(str, `\f`)
	return str
}

// FitString will shorten the input string to the defined string length
// If the input string is shorter than the defined length then return a padded string
// If the input string length equals the defined length then return the original string
// if the input string is shorter than 7 then return the first x number of characters
// Else return a string "left ... right" containing the beginning of the string plus " ... " plus the end of the string
// TODO: need tests
// TODO: Strip "String"
// TODO: need variations SquashString, no padding only truncation
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
