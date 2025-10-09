package streams

// This file contain character stream pattern matching functions.
//
// There are several types:
// - ƒ: Stream -> ([]rune, bool)
// - ƒ: ...any -> (Stream -> ([]rune, bool))
// - ƒ: Pattern -> Pattern
// - ƒ: []Pattern -> Pattern
//
// The pattern matcher function extracts an array of runes from a character stream based on its internal logic.
//   It returns true if a match is found, regardless of whether any runes are produced. The input Stream instance
//   is mutated to reflect the consumption of character stream data.
//   Signature:
//   ƒ: Stream -> ([]rune, bool)
//
// The constructor function is used to create a pattern matcher function from its input parameters.
//   ƒ: ...any -> (Stream -> ([]rune, bool))
//   This closure type function returns a new function:
//     ƒ: Stream -> ([]rune, bool)
//   Examples: CharMatcher, StringMatcher
//
// The constructor function creates a pattern matcher function using the provided input parameters.
//   ƒ: ...any -> (Stream -> ([]rune, bool))
//   Output: A closure function with the following signature:
//     ƒ: Stream -> ([]rune, bool)
//   Examples of such matchers include CharMatcher and StringMatcher.
//
// Higher-Order Functions (HOFs) are powerful tools for implementing advanced pattern matching constructs such as
//   `Optional`, `Sequence`, `OneOf`, and more.
//   Function signatures:
//     ƒ: Pattern -> Pattern
//     ƒ: ...Pattern -> Pattern
//   All Higher-Order Functions (HOFs) must return a pattern-matching function with the following signature:
//     ƒ: Stream -> ([]rune, bool)
//    The outer function allows them to be part of other HOFs, and the enclosed function is used to process the stream.
//
//  The minimal string matching looks something like this:
//    StringMatcher(`abc`)(stream)
//    The StringMatcher function returns a function with this signature:
//      func(stream Stream) ([]rune, bool)
//    This function is then executed by the (stream) appendix
//
// The following type function signatures have been defined
//   type Pattern func(stream Stream) ([]rune, bool)

// A Pattern is a type that defines a function. This function takes a stream as input and returns a tuple consisting of
// a rune array and a boolean. The boolean indicates whether there is a match, with true signifying a match and false
// indicating no match
type Pattern func(stream Stream) ([]rune, bool)

// The CharMatcher function produces a rune array and true if there is a character match, otherwise nil and false
func CharMatcher(char rune) Pattern {
	return func(stream Stream) ([]rune, bool) {
		if r, ok := stream.NextChar(); ok && r == char {
			return []rune{char}, true
		}
		return nil, false
	}
}

// The StringMatcher function produces a rune array and true if there is a string match, otherwise nil and false
func StringMatcher(literal string) Pattern {
	var rLiteral = []rune(literal)
	return func(stream Stream) ([]rune, bool) {
		var value []rune

		for _, ch := range rLiteral {
			if r, ok := stream.NextChar(); ok && r == ch {
				value = append(value, r)
				continue
			}
			break
		}

		if len(value) != len(rLiteral) {
			return nil, false
		}
		return value, true
	}
}

// The Sequence function sequentially applies the matchers in the specified order. All matchers must succeed for the
// operation to succeed; otherwise, it will result in a failure.
func Sequence(patterns ...Pattern) Pattern {
	return func(stream Stream) ([]rune, bool) {
		var value []rune
		for _, pattern := range patterns {
			ra, ok := pattern(stream)
			if !ok {
				return nil, false
			}
			value = append(value, ra...)
		}
		return value, true
	}
}

// The OneOf function applies the matchers in order and returns the first match
func OneOf(patterns ...Pattern) Pattern {
	return func(stream Stream) ([]rune, bool) {
		for _, pattern := range patterns {
			cs := stream.Clone() // enables backtracking
			ra, ok := pattern(cs)
			if ok {
				stream.Match(cs) // update the parent stream to match the matching pattern match
				return ra, true
			}
		}
		return nil, false
	}
}

// The Optional function applies the matchers in order and returns the first match
func Optional(pattern Pattern) Pattern {
	return func(stream Stream) ([]rune, bool) {
		cs := stream.Clone()
		ra, ok := pattern(cs)
		if ok {
			stream.Match(cs) // update the parent stream to match the matching pattern match
			return ra, true
		}
		return nil, true
	}
}
