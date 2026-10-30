// Package data provides the shared data model, resource limits, parse guard
// and typed errors used by Shape format parsers (JSON, YAML, CSV, XML).
//
// The package is purely additive: it does not depend on, or change, the
// ast package.
//
// # Value model
//
// A [Value] holds exactly one of seven kinds: Null, Bool, Integer, Decimal,
// String, List or Map. Maps are ordered ([Map]) and keep keys in insertion
// order. Numbers keep their exact source text ([Number]) so nothing is lost
// to float rounding and writers can emit the text unchanged.
//
// # Limits and guard
//
// [Limits] describes the resource ceilings for one parse. The zero value
// means the secure defaults: nesting depth 10,000 and YAML alias expansion
// [DefaultMaxAliasExpansion]; everything else is unlimited. A [Guard] enforces
// the limits while a parser runs. A parser creates one guard per parse and
// calls its methods as it reads input and builds nodes.
//
// # Errors
//
// Syntax problems are reported as *[SyntaxError] and limit violations as
// *[LimitError]. Both work with errors.Is (against [ErrSyntax] and
// [ErrLimit]) and errors.As. Context errors pass through unwrapped.
package data
