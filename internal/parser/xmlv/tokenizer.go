package xmlv

import (
	"unicode"

	"github.com/shapestone/shape/internal/tokenizer"
)

// Token kinds for XMLV
const (
	TokenTagOpen       = "TagOpen"       // <tag>
	TokenTagClose      = "TagClose"      // </tag>
	TokenTagSelfClose  = "TagSelfClose"  // <tag/>
	TokenText          = "Text"          // text content
	TokenIdentifier    = "Identifier"    // Type name (UUID, Email)
	TokenFunction      = "Function"      // Function call (String(1, 100))
	TokenNumber        = "Number"        // Number literal
	TokenString        = "String"        // String literal
	TokenTrue          = "True"          // true
	TokenFalse         = "False"         // false
	TokenNull          = "Null"          // null
)

// tagMatcher matches XML tags: <tag>, </tag>, <tag/>
func tagMatcher(stream tokenizer.Stream) *tokenizer.Token {
	first, ok := stream.NextChar()
	if !ok || first != '<' {
		return nil
	}

	value := []rune{first}
	isClosing := false
	isSelfClosing := false

	// Check for closing tag </
	r, ok := stream.NextChar()
	if !ok {
		return nil
	}
	value = append(value, r)

	if r == '/' {
		isClosing = true
		r, ok = stream.NextChar()
		if !ok {
			return nil
		}
		value = append(value, r)
	}

	// Tag name must start with letter
	if !unicode.IsLetter(r) {
		return nil
	}

	// Continue reading tag name
	for {
		r, ok = stream.NextChar()
		if !ok {
			return nil
		}
		value = append(value, r)

		if r == '>' {
			// End of tag
			break
		} else if r == '/' {
			// Might be self-closing
			r, ok = stream.NextChar()
			if !ok {
				return nil
			}
			value = append(value, r)
			if r == '>' {
				isSelfClosing = true
				break
			}
		} else if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_' && r != ':' {
			return nil
		}
	}

	// Determine token kind
	kind := TokenTagOpen
	if isClosing {
		kind = TokenTagClose
	} else if isSelfClosing {
		kind = TokenTagSelfClose
	}

	return tokenizer.NewToken(kind, value)
}

// textMatcher matches text content between tags
func textMatcher(stream tokenizer.Stream) *tokenizer.Token {
	first, ok := stream.NextChar()
	if !ok || first == '<' {
		return nil
	}

	value := []rune{first}
	for {
		r, ok := stream.NextChar()
		if !ok || r == '<' {
			break
		}
		value = append(value, r)
	}

	return tokenizer.NewToken(TokenText, value)
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

// GetMatchers returns all matchers for XMLV tokenization
// Note: WhiteSpaceMatcher is automatically prepended by NewTokenizer()
func GetMatchers() []tokenizer.Matcher {
	return []tokenizer.Matcher{
		tagMatcher,     // Match tags first
		textMatcher,    // Everything else is text (to be parsed later)
	}
}
