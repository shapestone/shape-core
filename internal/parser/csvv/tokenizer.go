package csvv

import (
	"github.com/shapestone/shape/internal/tokenizer"
)

// Token kinds for CSVV
const (
	TokenComment = "Comment" // # comment
	TokenCell    = "Cell"    // CSV cell value
	TokenComma   = "Comma"   // ,
	TokenNewline = "Newline" // \n
)

// cellMatcher matches a CSV cell (unquoted or quoted)
func cellMatcher(stream tokenizer.Stream) *tokenizer.Token {
	first, ok := stream.NextChar()
	if !ok {
		return nil
	}

	// Check for quoted cell
	if first == '"' {
		return quotedCellMatcher(stream, first)
	}

	// Check for comma or newline (cell boundary) - not part of cell
	if first == ',' || first == '\n' {
		return nil
	}

	// Unquoted cell - read until comma, newline, or EOF
	// But track parentheses depth to handle function calls like String(1,100)
	value := []rune{first}
	parenDepth := 0
	if first == '(' {
		parenDepth = 1
	}

	for {
		r, ok := stream.NextChar()
		if !ok {
			break
		}

		// Track parentheses depth
		if r == '(' {
			parenDepth++
		} else if r == ')' {
			parenDepth--
		}

		// Only treat comma/newline as boundary if not inside parentheses
		if parenDepth == 0 && (r == ',' || r == '\n') {
			break
		}

		value = append(value, r)
	}

	return tokenizer.NewToken(TokenCell, value)
}

// quotedCellMatcher matches a quoted CSV cell: "value"
func quotedCellMatcher(stream tokenizer.Stream, firstQuote rune) *tokenizer.Token {
	value := []rune{firstQuote}
	escaped := false

	for {
		r, ok := stream.NextChar()
		if !ok {
			// Unterminated quoted cell
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
			// Check for escaped quote (two quotes in a row)
			next, ok := stream.NextChar()
			if ok && next == '"' {
				// Double quote = escaped quote, consume it
				value = append(value, next)
				continue
			}
			// End of quoted cell
			break
		}
	}

	return tokenizer.NewToken(TokenCell, value)
}

// commentMatcher matches CSV comments: # to end of line
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

// GetMatchers returns all matchers for CSVV tokenization
// Note: WhiteSpaceMatcher is automatically prepended by NewTokenizer()
func GetMatchers() []tokenizer.Matcher {
	return []tokenizer.Matcher{
		commentMatcher,
		tokenizer.CharMatcherFunc(TokenComma, ','),
		tokenizer.CharMatcherFunc(TokenNewline, '\n'),
		cellMatcher, // Must come after comma and newline
	}
}
