package main

import (
	"fmt"

	"github.com/shapestone/shape/pkg/tokenizer"
)

func main() {
	input := `# Sample configuration
server {
    host: "localhost"
    port: 8080
    enabled: true
}

database {
    connection: "postgresql://localhost/mydb"
    pool_size: 10
}
`

	fmt.Println("=== Custom DSL Tokenization Example ===\n")
	fmt.Println("Input:")
	fmt.Println(input)
	fmt.Println("\nTokens:")
	fmt.Println("-------")

	// Create tokenizer with custom matchers
	tok := NewConfigTokenizer()
	tok.Initialize(input)

	// Tokenize
	tokens, isEOS := tok.Tokenize()

	// Display tokens
	for _, token := range tokens {
		// Skip whitespace and comments for cleaner output
		if token.Kind() == "Whitespace" || token.Kind() == "Comment" {
			continue
		}

		fmt.Printf("%-12s %q at line %d, col %d\n",
			token.Kind(),
			token.ValueString(),
			token.Row(),
			token.Column(),
		)
	}

	if !isEOS {
		fmt.Println("\nError: Unexpected end of tokenization (stream not fully consumed)")
		return
	}

	fmt.Println("\n✓ Tokenization successful!")
}

// NewConfigTokenizer creates a tokenizer for our configuration DSL
func NewConfigTokenizer() tokenizer.Tokenizer {
	return tokenizer.NewTokenizer(
		// Comments
		CommentMatcher,

		// Keywords (must come before identifiers)
		KeywordMatcher,

		// Literals
		StringLiteralMatcher,
		NumberMatcher,
		BooleanMatcher,

		// Identifiers (catch-all for names)
		IdentifierMatcher,

		// Punctuation
		tokenizer.CharMatcherFunc("LBrace", '{'),
		tokenizer.CharMatcherFunc("RBrace", '}'),
		tokenizer.CharMatcherFunc("Colon", ':'),
	)
	// Note: WhiteSpaceMatcher is automatically prepended by the tokenizer
}
