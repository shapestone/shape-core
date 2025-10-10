package yamlv

import (
	"unicode"

	"github.com/shapestone/shape/internal/tokenizer"
)

// Token kinds for YAMLV
const (
	TokenKey        = "Key"        // property key before :
	TokenColon      = "Colon"      // :
	TokenDash       = "Dash"       // - (array item)
	TokenIdentifier = "Identifier" // Type name (UUID, Email)
	TokenFunction   = "Function"   // Function call (String(1, 100))
	TokenNumber     = "Number"     // Number literal
	TokenString     = "String"     // String literal (quoted)
	TokenTrue       = "True"       // true
	TokenFalse      = "False"      // false
	TokenNull       = "Null"       // null
	TokenComment    = "Comment"    // # comment
	TokenNewline    = "Newline"    // \n
	TokenIndent     = "Indent"     // indentation increase
	TokenDedent     = "Dedent"     // indentation decrease
)

// keyMatcher matches property keys (before the colon)
// Matches: letters, digits, underscore, hyphen
func keyMatcher(stream tokenizer.Stream) *tokenizer.Token {
	first, ok := stream.NextChar()
	if !ok || (!unicode.IsLetter(first) && first != '_') {
		return nil
	}

	value := []rune{first}
	for {
		r, ok := stream.PeekChar()
		if !ok {
			break
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
			stream.NextChar()
			value = append(value, r)
		} else {
			break
		}
	}

	return tokenizer.NewToken(TokenKey, value)
}

// functionMatcher matches function calls: Integer(1,100)
func functionMatcher(stream tokenizer.Stream) *tokenizer.Token {
	first, ok := stream.NextChar()
	if !ok || !unicode.IsUpper(first) {
		return nil
	}

	value := []rune{first}
	foundParen := false

	for {
		r, ok := stream.PeekChar()
		if !ok {
			return nil
		}
		if r == '(' {
			stream.NextChar()
			value = append(value, r)
			foundParen = true
			break
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			stream.NextChar()
			value = append(value, r)
		} else {
			return nil
		}
	}

	if !foundParen {
		return nil
	}

	// Consume everything until matching )
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

// identifierMatcher matches type identifiers: UUID, Email, ISO-8601
func identifierMatcher(stream tokenizer.Stream) *tokenizer.Token {
	first, ok := stream.NextChar()
	if !ok || !unicode.IsUpper(first) {
		return nil
	}

	value := []rune{first}
	for {
		r, ok := stream.PeekChar()
		if !ok {
			break
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			stream.NextChar()
			value = append(value, r)
		} else {
			break
		}
	}

	return tokenizer.NewToken(TokenIdentifier, value)
}

// numberMatcher matches numeric literals: 42, 3.14, -5
func numberMatcher(stream tokenizer.Stream) *tokenizer.Token {
	first, ok := stream.NextChar()
	if !ok || (!unicode.IsDigit(first) && first != '-') {
		return nil
	}

	value := []rune{first}
	hasDecimal := false

	for {
		r, ok := stream.PeekChar()
		if !ok {
			break
		}
		if unicode.IsDigit(r) {
			stream.NextChar()
			value = append(value, r)
		} else if r == '.' && !hasDecimal {
			stream.NextChar()
			value = append(value, r)
			hasDecimal = true
		} else {
			break
		}
	}

	return tokenizer.NewToken(TokenNumber, value)
}

// stringMatcher matches quoted strings: "hello world"
func stringMatcher(stream tokenizer.Stream) *tokenizer.Token {
	first, ok := stream.NextChar()
	if !ok || first != '"' {
		return nil
	}

	value := []rune{first}
	escaped := false

	for {
		r, ok := stream.NextChar()
		if !ok {
			return nil // Unterminated string
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
			break
		}
	}

	return tokenizer.NewToken(TokenString, value)
}

// commentMatcher matches comments: # to end of line
func commentMatcher(stream tokenizer.Stream) *tokenizer.Token {
	first, ok := stream.NextChar()
	if !ok || first != '#' {
		return nil
	}

	value := []rune{first}
	for {
		r, ok := stream.PeekChar()
		if !ok || r == '\n' {
			break
		}
		stream.NextChar()
		value = append(value, r)
	}

	return tokenizer.NewToken(TokenComment, value)
}

// GetMatchers returns all matchers for YAMLV tokenization
// Order matters: try most specific matchers first
func GetMatchers() []tokenizer.Matcher {
	return []tokenizer.Matcher{
		commentMatcher,
		tokenizer.StringMatcherFunc(TokenTrue, "true"),
		tokenizer.StringMatcherFunc(TokenFalse, "false"),
		tokenizer.StringMatcherFunc(TokenNull, "null"),
		tokenizer.CharMatcherFunc(TokenColon, ':'),
		tokenizer.CharMatcherFunc(TokenDash, '-'),
		tokenizer.CharMatcherFunc(TokenNewline, '\n'),
		stringMatcher,
		functionMatcher,
		identifierMatcher,
		numberMatcher,
		keyMatcher,
	}
}
