package main

import (
	"unicode"

	"github.com/shapestone/shape-core/pkg/tokenizer"
)

// KeywordMatcher matches reserved words in our DSL
func KeywordMatcher(stream tokenizer.Stream) *tokenizer.Token {
	keywords := []string{
		"server", "database", "host", "port",
		"enabled", "connection", "pool_size",
	}

	for _, kw := range keywords {
		matcher := tokenizer.StringMatcherFunc("Keyword", kw)
		if match := matcher(stream); match != nil {
			// Verify it's not part of a larger identifier
			// Peek ahead to ensure keyword is followed by whitespace or punctuation
			ch, ok := stream.PeekChar()
			if ok && (unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_') {
				// It's part of a larger identifier, not a keyword
				return nil
			}
			return match
		}
	}
	return nil
}

// IdentifierMatcher matches variable names and non-keyword identifiers
func IdentifierMatcher(stream tokenizer.Stream) *tokenizer.Token {
	var value []rune

	// First character must be letter or underscore
	ch, ok := stream.PeekChar()
	if !ok || !(unicode.IsLetter(ch) || ch == '_') {
		return nil
	}

	stream.NextChar()
	value = append(value, ch)

	// Subsequent characters: letter, digit, or underscore
	for {
		ch, ok := stream.PeekChar()
		if !ok {
			break
		}

		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' {
			stream.NextChar()
			value = append(value, ch)
		} else {
			break
		}
	}

	return tokenizer.NewToken("Identifier", value)
}

// StringLiteralMatcher matches quoted strings
func StringLiteralMatcher(stream tokenizer.Stream) *tokenizer.Token {
	// Must start with quote
	first, ok := stream.NextChar()
	if !ok || first != '"' {
		return nil
	}

	// Include opening quote in value (following Shape's pattern)
	value := []rune{first}
	escaped := false

	for {
		ch, ok := stream.NextChar()
		if !ok {
			return nil // Unclosed string (error)
		}

		value = append(value, ch)

		if escaped {
			escaped = false
			continue
		}

		if ch == '\\' {
			escaped = true
			continue
		}

		if ch == '"' {
			// Closing quote found (already added to value)
			return tokenizer.NewToken("String", value)
		}
	}
}

// NumberMatcher matches integers and floats
func NumberMatcher(stream tokenizer.Stream) *tokenizer.Token {
	var value []rune
	hasDecimal := false

	for {
		ch, ok := stream.PeekChar()
		if !ok {
			break
		}

		if unicode.IsDigit(ch) {
			stream.NextChar()
			value = append(value, ch)
		} else if ch == '.' && !hasDecimal {
			// Allow decimal point (simplified - doesn't check if followed by digit)
			stream.NextChar()
			value = append(value, ch)
			hasDecimal = true
		} else {
			break
		}
	}

	if len(value) == 0 {
		return nil
	}

	kind := "Integer"
	if hasDecimal {
		kind = "Float"
	}

	return tokenizer.NewToken(kind, value)
}

// CommentMatcher matches # comments
func CommentMatcher(stream tokenizer.Stream) *tokenizer.Token {
	ch, ok := stream.PeekChar()
	if !ok || ch != '#' {
		return nil
	}

	var value []rune
	for {
		ch, ok := stream.NextChar()
		if !ok || ch == '\n' {
			break
		}
		value = append(value, ch)
	}

	return tokenizer.NewToken("Comment", value)
}

// BooleanMatcher matches true/false literals
func BooleanMatcher(stream tokenizer.Stream) *tokenizer.Token {
	trueMatcher := tokenizer.StringMatcherFunc("Boolean", "true")
	if match := trueMatcher(stream); match != nil {
		// Verify it's not part of a larger identifier
		ch, ok := stream.PeekChar()
		if ok && (unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_') {
			return nil
		}
		return match
	}

	falseMatcher := tokenizer.StringMatcherFunc("Boolean", "false")
	if match := falseMatcher(stream); match != nil {
		// Verify it's not part of a larger identifier
		ch, ok := stream.PeekChar()
		if ok && (unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_') {
			return nil
		}
		return match
	}

	return nil
}
