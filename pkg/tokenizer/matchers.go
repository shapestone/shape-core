package tokenizer

import (
	"unicode"
)

//
// Built-in Matchers - Common token matchers for various use cases
//

// WhiteSpaceMatcher consumes all consecutive whitespace and yields a Whitespace token.
// Returns nil if no whitespace is found.
// Uses SWAR acceleration when ByteStream is available for 2-8x speedup on common whitespace (space, tab, LF, CR).
func WhiteSpaceMatcher(stream Stream) *Token {
	// Fast path: Use SWAR SkipWhitespace if ByteStream available
	// Note: SWAR only handles space, tab, LF, CR (JSON whitespace)
	// Falls through to byte-by-byte for rare whitespace (\v, \f, etc.)
	if byteStream, ok := stream.(ByteStream); ok {
		startPos := byteStream.BytePosition()

		// Use SWAR to skip common whitespace (processes 8 bytes at once)
		remaining := byteStream.RemainingBytes()
		skipped := SkipWhitespace(remaining)

		if skipped > 0 {
			// Advance stream by SWAR-skipped bytes
			for i := 0; i < skipped; i++ {
				byteStream.NextByte()
			}
		}

		// Check for additional rare whitespace (unicode.IsSpace but not in SWAR set)
		for {
			b, ok := byteStream.PeekByte()
			if !ok {
				break
			}
			// Check for rare whitespace: \v (0x0B), \f (0x0C), and Unicode spaces
			r := rune(b)
			if unicode.IsSpace(r) && b != ' ' && b != '\t' && b != '\n' && b != '\r' {
				byteStream.NextByte()
				continue
			}
			break
		}

		endPos := byteStream.BytePosition()
		if endPos == startPos {
			return nil // No whitespace found
		}

		// Extract the whitespace as a token
		value := byteStream.SliceFrom(startPos)
		return NewToken(`Whitespace`, []rune(string(value)))
	}

	// Fallback: Rune-based implementation for non-ByteStream
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
