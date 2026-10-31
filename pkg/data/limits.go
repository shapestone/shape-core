package data

import (
	"context"
	"math"
)

// NoLimit turns a limit off, as does any negative value. It is the only way to
// disable the limits that are on by default (MaxDepth, MaxAliasExpansion).
const NoLimit = -1

// Default values applied when a limit field is zero.
const (
	// DefaultMaxDepth matches encoding/json, encoding/xml and shape-yaml.
	DefaultMaxDepth = 10000

	// DefaultMaxAliasExpansion is the default budget of nodes produced by
	// expanding YAML aliases. Legitimate documents expand to at most a few
	// thousand nodes; one million nodes costs roughly 64 MiB under the
	// estimate in Guard, which is cheap enough to allow and small enough to
	// stop a "billion laughs" document long before it exhausts memory.
	// SnakeYAML instead caps the alias count at 50; go-yaml v3 uses a
	// ratio of aliased to decoded nodes. A plain node budget is simpler
	// to explain and independent of document size.
	DefaultMaxAliasExpansion = 1_000_000
)

// Limits are the resource ceilings for one parse. The zero value means the
// defaults, so Limits{} is safe to use as is.
//
// For every integer field:
//   - zero selects the default: DefaultMaxDepth for MaxDepth,
//     DefaultMaxAliasExpansion for MaxAliasExpansion, and "unlimited" for
//     the other fields;
//   - a positive value is the maximum allowed;
//   - NoLimit (or any negative value) means unlimited.
//
// A value equal to the limit is allowed; one more is an error.
type Limits struct {
	// MaxDepth is the maximum nesting depth of lists and maps. On by default.
	// Turning it off removes the only protection against Go stack
	// exhaustion in recursive parsers, which is a fatal, unrecoverable
	// crash rather than an error. Only disable it for iterative parsers.
	MaxDepth int64

	// MaxBytes is the maximum number of input bytes consumed.
	MaxBytes int64

	// MaxNodes is the maximum number of values created.
	MaxNodes int64

	// MaxStringBytes is the maximum length in bytes of any single string,
	// map key or number text. Parsers report each through Guard.String,
	// Guard.Key or Guard.Number.
	MaxStringBytes int64

	// MaxKeysPerMap is the maximum number of keys in one map.
	MaxKeysPerMap int64

	// MaxAliasExpansion is the maximum number of nodes produced by YAML
	// alias expansion. On by default.
	MaxAliasExpansion int64

	// MaxMemory is the budget for the estimated memory of the result, in
	// bytes. The estimate of the finished result is approximate: it uses
	// fixed per-node and per-string constants (a small integer is charged
	// 81 bytes where the result really holds about 56) and ignores Go map
	// overhead. The estimate is not the main risk; the peak is. See Guard.
	//
	// The peak heap during a parse is higher than the estimate: an append
	// that grows a list keeps the old and the new slice alive together, and
	// the garbage left behind is not collected at once. Without Guard.Grow
	// the peak reached 1.8 to 2.1 times MaxMemory on a parse of small
	// integers and of one large object (256 MiB budget; up to 2.5 times seen
	// at 1 GiB). A parser that calls Grow before each reallocating append
	// peaked at 0.9 and 1.1 times MaxMemory. A client with a hard memory
	// ceiling should set MaxMemory to at most half of it for parsers that
	// use Grow, and a third of it for those that do not.
	MaxMemory int64

	// Context, when non-nil, cancels the parse. The guard polls it
	// periodically and returns its error unchanged.
	Context context.Context
}

// resolve maps a configured value to an effective ceiling: zero becomes def
// and negative becomes unlimited.
func resolve(v, def int64) int64 {
	switch {
	case v == 0:
		return def
	case v < 0:
		return math.MaxInt64
	}
	return v
}
