package propsv

import (
	"unicode"

	"github.com/shapestone/shape/internal/tokenizer"
)

// Token kinds for PropsV
const (
	TokenPropertyName = "PropertyName"
	TokenEquals       = "Equals"
	TokenComment      = "Comment"
	TokenIdentifier   = "Identifier"
	TokenFunction     = "Function"
	TokenString       = "String"
	TokenNumber       = "Number"
	TokenTrue         = "True"
	TokenFalse        = "False"
	TokenNull         = "Null"
)

//propertyNameMatcher matches property names: user.profile.name or tags[]
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

// identifierMatcher matches type identifiers: UUID, Email
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

// functionMatcher matches function calls: Integer(1, 100)
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

// numberMatcher matches numeric literals: 42, 3.14
func numberMatcher(stream tokenizer.Stream) *tokenizer.Token {
	first, ok := stream.NextChar()
	if !ok || !unicode.IsDigit(first) {
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

// GetMatchers returns all matchers for PropsV tokenization
// Note: WhiteSpaceMatcher is automatically prepended by NewTokenizer()
func GetMatchers() []tokenizer.Matcher {
	return []tokenizer.Matcher{
		commentMatcher,
		tokenizer.CharMatcherFunc(TokenEquals, '='),
		tokenizer.StringMatcherFunc(TokenTrue, "true"),
		tokenizer.StringMatcherFunc(TokenFalse, "false"),
		tokenizer.StringMatcherFunc(TokenNull, "null"),
		functionMatcher,
		identifierMatcher,
		numberMatcher,
		propertyNameMatcher,
	}
}
