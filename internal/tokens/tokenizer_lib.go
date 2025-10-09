package tokens

import (
	"github.com/shapestone/shape/internal/streams"
	"unicode"
)

// The WhiteSpaceMatcher consumes all whitespace and yields a Whitespace Token
func WhiteSpaceMatcher(stream streams.Stream) *Token {
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

// The CharMatcher produces a Token if there is a character match
func CharMatcher(tokenName string, char rune) Matcher {
	return func(stream streams.Stream) *Token {
		if r, ok := stream.NextChar(); ok && r != char {
			return nil
		}
		return NewToken(tokenName, []rune{char})
	}
}

// The StringMatcher produces a Token if there is a string match
func StringMatcher(tokenName string, literal string) Matcher {
	var rLiteral = []rune(literal)
	return func(stream streams.Stream) *Token {
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
