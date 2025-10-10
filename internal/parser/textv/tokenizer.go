package textv

import (
	"unicode"

	"github.com/shapestone/shape/internal/tokenizer"
)

// Token kinds for TEXTV
const (
	TokenPropertyName = "PropertyName" // property name (e.g., "user.id")
	TokenColon        = "Colon"        // :
	TokenIdentifier   = "Identifier"   // Type name (UUID, Email)
	TokenFunction     = "Function"     // Function call (String(1, 100))
	TokenNumber       = "Number"       // Number literal
	TokenString       = "String"       // String literal
	TokenTrue         = "True"         // true
	TokenFalse        = "False"        // false
	TokenNull         = "Null"         // null
	TokenComment      = "Comment"      // # comment
	TokenNewline      = "Newline"      // \n
)

// propertyNameMatcher matches property names: user.id, user.profile.name, tags[]
func propertyNameMatcher(stream tokenizer.Stream) *tokenizer.Token {
	first, ok := stream.NextChar()
	if !ok || !unicode.IsLetter(first) {
		return nil
	}

	value := []rune{first}
	for {
		r, ok := stream.NextChar()
		if !ok {
			break
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '_' || r == '[' || r == ']' {
			value = append(value, r)
		} else {
			break
		}
	}

	return tokenizer.NewToken(TokenPropertyName, value)
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

// numberMatcher matches numeric literals: 42, 3.14
func numberMatcher(stream tokenizer.Stream) *tokenizer.Token {
	first, ok := stream.NextChar()
	if !ok || (!unicode.IsDigit(first) && first != '-') {
		return nil
	}

	value := []rune{first}
	hasDecimal := false

	for {
		r, ok := stream.NextChar()
		if !ok {
			break
		}
		if unicode.IsDigit(r) {
			value = append(value, r)
		} else if r == '.' && !hasDecimal {
			value = append(value, r)
			hasDecimal = true
		} else {
			break
		}
	}

	return tokenizer.NewToken(TokenNumber, value)
}

// commentMatcher matches comments: # to end of line
func commentMatcher(stream tokenizer.Stream) *tokenizer.Token {
	first, ok := stream.NextChar()
	if !ok || first != '#' {
		return nil
	}

	value := []rune{first}
	for {
		r, ok := stream.NextChar()
		if !ok || r == '\n' {
			break
		}
		value = append(value, r)
	}

	return tokenizer.NewToken(TokenComment, value)
}

// GetMatchers returns all matchers for TEXTV tokenization
func GetMatchers() []tokenizer.Matcher {
	return []tokenizer.Matcher{
		commentMatcher,
		tokenizer.StringMatcherFunc(TokenTrue, "true"),
		tokenizer.StringMatcherFunc(TokenFalse, "false"),
		tokenizer.StringMatcherFunc(TokenNull, "null"),
		tokenizer.CharMatcherFunc(TokenColon, ':'),
		tokenizer.CharMatcherFunc(TokenNewline, '\n'),
		functionMatcher,
		identifierMatcher,
		numberMatcher,
		propertyNameMatcher,
	}
}
