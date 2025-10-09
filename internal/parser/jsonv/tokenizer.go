package jsonv

import (
	"unicode"

	"github.com/shapestone/shape/internal/tokenizer"
)

// Token kinds for JSONV
const (
	TokenObjectStart = "ObjectStart"
	TokenObjectEnd   = "ObjectEnd"
	TokenArrayStart  = "ArrayStart"
	TokenArrayEnd    = "ArrayEnd"
	TokenColon       = "Colon"
	TokenComma       = "Comma"
	TokenString      = "String"
	TokenNumber      = "Number"
	TokenTrue        = "True"
	TokenFalse       = "False"
	TokenNull        = "Null"
	TokenIdentifier  = "Identifier"
	TokenFunction    = "Function"
	TokenWhitespace  = "Whitespace"
)

// identifierMatcher matches type identifiers: UUID, Email, ISO-8601
// Pattern: [A-Z][A-Za-z0-9-]*
func identifierMatcher(stream tokenizer.Stream) *tokenizer.Token {
	// Must start with uppercase letter
	first, ok := stream.NextChar()
	if !ok || !unicode.IsUpper(first) {
		return nil
	}

	value := []rune{first}

	// Continue with letters, digits, or hyphens
	for {
		r, ok := stream.NextChar()
		if !ok {
			break
		}

		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			value = append(value, r)
		} else {
			break
		}
	}

	return tokenizer.NewToken(TokenIdentifier, value)
}

// functionMatcher matches function calls: Integer(1, 100)
// Pattern: Identifier '(' arguments ')'
func functionMatcher(stream tokenizer.Stream) *tokenizer.Token {
	// Must start with uppercase letter (function name)
	first, ok := stream.NextChar()
	if !ok || !unicode.IsUpper(first) {
		return nil
	}

	value := []rune{first}

	// Continue with letters/digits/hyphens until we hit '('
	foundParen := false
	for {
		r, ok := stream.NextChar()
		if !ok {
			return nil
		}

		if r == '(' {
			value = append(value, r)
			foundParen = true
			break
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			value = append(value, r)
		} else {
			return nil
		}
	}

	if !foundParen {
		return nil
	}

	// Now consume everything until we find the matching ')'
	parenDepth := 1
	for {
		r, ok := stream.NextChar()
		if !ok {
			return nil
		}

		value = append(value, r)

		if r == '(' {
			parenDepth++
		} else if r == ')' {
			parenDepth--
			if parenDepth == 0 {
				break
			}
		}
	}

	return tokenizer.NewToken(TokenFunction, value)
}

// stringMatcher matches JSON string literals: "text"
func stringMatcher(stream tokenizer.Stream) *tokenizer.Token {
	// Must start with double quote
	first, ok := stream.NextChar()
	if !ok || first != '"' {
		return nil
	}

	value := []rune{first}
	escaped := false

	for {
		r, ok := stream.NextChar()
		if !ok {
			// Unterminated string
			return nil
		}

		value = append(value, r)

		if escaped {
			escaped = false
			continue
		}

		if r == '\\' {
			escaped = true
			continue
		}

		if r == '"' {
			// End of string
			break
		}
	}

	return tokenizer.NewToken(TokenString, value)
}

// numberMatcher matches JSON numbers: 42, -3.14, 1.5e10
func numberMatcher(stream tokenizer.Stream) *tokenizer.Token {
	value := []rune{}

	// Optional minus sign
	r, ok := stream.NextChar()
	if !ok {
		return nil
	}

	if r == '-' {
		value = append(value, r)
		r, ok = stream.NextChar()
		if !ok {
			return nil
		}
	}

	// Must have at least one digit
	if !unicode.IsDigit(r) {
		return nil
	}

	// Integer part
	if r == '0' {
		value = append(value, r)
	} else {
		value = append(value, r)
		for {
			r, ok = stream.NextChar()
			if !ok {
				break
			}
			if unicode.IsDigit(r) {
				value = append(value, r)
			} else {
				break
			}
		}
	}

	// Optional decimal part
	if ok && r == '.' {
		value = append(value, r)
		foundDigit := false
		for {
			r, ok = stream.NextChar()
			if !ok {
				break
			}
			if unicode.IsDigit(r) {
				value = append(value, r)
				foundDigit = true
			} else {
				break
			}
		}
		if !foundDigit {
			return nil
		}
	}

	// Optional exponent part
	if ok && (r == 'e' || r == 'E') {
		value = append(value, r)
		r, ok = stream.NextChar()
		if !ok {
			return nil
		}

		// Optional sign
		if r == '+' || r == '-' {
			value = append(value, r)
			r, ok = stream.NextChar()
			if !ok {
				return nil
			}
		}

		// Must have at least one digit
		if !unicode.IsDigit(r) {
			return nil
		}

		value = append(value, r)
		for {
			r, ok = stream.NextChar()
			if !ok {
				break
			}
			if unicode.IsDigit(r) {
				value = append(value, r)
			} else {
				break
			}
		}
	}

	return tokenizer.NewToken(TokenNumber, value)
}

// GetMatchers returns all matchers for JSONV tokenization in priority order.
func GetMatchers() []tokenizer.Matcher {
	return []tokenizer.Matcher{
		// Whitespace (built-in, skip)
		tokenizer.WhiteSpaceMatcher,

		// Delimiters (single characters)
		tokenizer.CharMatcherFunc(TokenObjectStart, '{'),
		tokenizer.CharMatcherFunc(TokenObjectEnd, '}'),
		tokenizer.CharMatcherFunc(TokenArrayStart, '['),
		tokenizer.CharMatcherFunc(TokenArrayEnd, ']'),
		tokenizer.CharMatcherFunc(TokenColon, ':'),
		tokenizer.CharMatcherFunc(TokenComma, ','),

		// Keywords (must come before identifier to avoid conflicts)
		tokenizer.StringMatcherFunc(TokenTrue, "true"),
		tokenizer.StringMatcherFunc(TokenFalse, "false"),
		tokenizer.StringMatcherFunc(TokenNull, "null"),

		// String literals
		stringMatcher,

		// Numbers
		numberMatcher,

		// Functions (must come before identifier)
		functionMatcher,

		// Identifiers (type names)
		identifierMatcher,
	}
}
