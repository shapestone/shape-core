package tokenizer

import (
	"unicode"
)

//
// Built-in Matchers - Common token matchers for various use cases
//

// WhiteSpaceMatcher consumes all consecutive whitespace and yields a Whitespace token.
// Returns nil if no whitespace is found.
func WhiteSpaceMatcher(stream Stream) *Token {
	var value []rune
	for {
		if r, ok := stream.NextChar(); ok {
			if unicode.IsSpace(r) {
				value = append(value, r)
				continue
			}
		}
		break
	}
	if len(value) == 0 {
		return nil
	}
	return NewToken(`Whitespace`, value)
}

// CharMatcherFunc creates a matcher that matches a single character and returns a token.
// The tokenName parameter specifies the token kind.
func CharMatcherFunc(tokenName string, char rune) Matcher {
	return func(stream Stream) *Token {
		if r, ok := stream.NextChar(); ok && r == char {
			return NewToken(tokenName, []rune{char})
		}
		return nil
	}
}

// StringMatcherFunc creates a matcher that matches a literal string and returns a token.
// The tokenName parameter specifies the token kind.
func StringMatcherFunc(tokenName string, literal string) Matcher {
	var rLiteral = []rune(literal)
	return func(stream Stream) *Token {
		var value []rune

		for _, ch := range rLiteral {
			if r, ok := stream.NextChar(); ok && r == ch {
				value = append(value, r)
				continue
			}
			break
		}

		if len(value) != len(rLiteral) {
			return nil
		}
		return NewToken(tokenName, value)
	}
}
