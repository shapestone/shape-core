package data

import (
	"context"
	"fmt"
	"math"
)

// Per-node memory estimates used for the MaxMemory budget. They model the
// in-memory size of the result on a 64-bit platform. They are a heuristic,
// not exact allocation: slice-growth slack (up to 2x on append) and Go map
// overhead are ignored, and tiny maps are over-counted.
const (
	// EstNodeBytes is charged for every value: the size of a Value.
	EstNodeBytes = 64

	// EstStringHeaderBytes is charged for every string and key on top of
	// its content length: the string header and allocator slack.
	EstStringHeaderBytes = 16

	// EstMapEntryBytes is charged for every map key on top of the key's
	// string cost: the Entry key header plus amortised index overhead.
	EstMapEntryBytes = 48
)

// pollMask makes the guard poll the context once every 256 operations.
const pollMask = 255

// Guard enforces Limits while a parser runs. Create one per parse with
// NewGuard; it is not safe for concurrent use.
//
// Every check returns nil, a *LimitError, ErrNegativeAmount (wrapped) for a
// negative argument, or the context's error unchanged. Once a check fails the
// parse should stop; the guard is not reusable. A failed check does not add
// to its counter. Counters saturate rather than wrap, so no argument can
// overflow a limit.
//
// Memory estimate: Node charges EstNodeBytes, String and Key charge their
// byte length plus EstStringHeaderBytes, and Key also charges
// EstMapEntryBytes. Parsers that allocate more can add to the budget with
// Memory.
//
// Context polling: Enter, Node, Bytes, String, Key, CheckStringLen,
// MapSize, Alias and Memory count as one operation each, and every 256th
// polls the context. Leave does not. A parser that calls the guard rarely
// per unit of work (for example CSV rows with large fields) should call
// Check at coarse boundaries.
//
// Streaming strings: while accumulating a long string, call CheckStringLen
// with the length so far to fail before allocating, then call String (or Key
// or Number) once with the final length to charge memory.
//
// Unlimited limits are stored as math.MaxInt64, so the hot path is one
// subtract and one compare per check with no branch on "enabled".
type Guard struct {
	ctx   context.Context
	ticks uint32

	depth    int64
	maxDepth int64

	nodes    int64
	maxNodes int64

	bytes    int64
	maxBytes int64

	maxString int64
	maxKeys   int64

	alias    int64
	maxAlias int64

	mem    int64
	maxMem int64
}

// NewGuard returns a guard for one parse, applying the defaults of Limits.
func NewGuard(l Limits) *Guard {
	g := &Guard{
		maxDepth:  resolve(l.MaxDepth, DefaultMaxDepth),
		maxNodes:  resolve(l.MaxNodes, math.MaxInt64),
		maxBytes:  resolve(l.MaxBytes, math.MaxInt64),
		maxString: resolve(l.MaxStringBytes, math.MaxInt64),
		maxKeys:   resolve(l.MaxKeysPerMap, math.MaxInt64),
		maxAlias:  resolve(l.MaxAliasExpansion, DefaultMaxAliasExpansion),
		maxMem:    resolve(l.MaxMemory, math.MaxInt64),
	}
	if l.Context != nil && l.Context.Done() != nil {
		g.ctx = l.Context // a never-cancelled context needs no polling
	}
	return g
}

// Check polls the context immediately. Call it once before parsing to
// reject an already-cancelled context, and at coarse boundaries (see Guard).
func (g *Guard) Check() error {
	if g.ctx != nil {
		return g.ctx.Err()
	}
	return nil
}

// tick counts one operation and polls the context every 256th.
func (g *Guard) tick() error {
	g.ticks++
	if g.ticks&pollMask == 0 && g.ctx != nil {
		return g.ctx.Err()
	}
	return nil
}

func exceeded(limit string, max, got int64) error {
	return &LimitError{Limit: limit, Max: max, Got: got}
}

func negative(limit string, n int64) error {
	return fmt.Errorf("data: %s: %w (%d)", limit, ErrNegativeAmount, n)
}

// charge adds n to *cur if the result stays within max. It rejects negative n
// and cannot overflow: the comparison is against max-*cur, and *cur never
// exceeds max because a failed charge is not applied.
func charge(cur *int64, n, max int64, limit string) error {
	if n < 0 {
		return negative(limit, n)
	}
	if n > max-*cur {
		got := *cur + n
		if got < *cur { // overflowed
			got = math.MaxInt64
		}
		return exceeded(limit, max, got)
	}
	*cur += n
	return nil
}

// Enter descends one nesting level. Depth is counted by the guard, not by
// the Go stack. On error the depth is unchanged.
func (g *Guard) Enter() error {
	if g.depth >= g.maxDepth {
		return exceeded(LimitDepth, g.maxDepth, g.depth+1)
	}
	if err := g.tick(); err != nil {
		return err
	}
	g.depth++
	return nil
}

// Leave ascends one nesting level. It never fails, and an unmatched Leave
// is clamped at depth 0 rather than panicking.
func (g *Guard) Leave() {
	if g.depth > 0 {
		g.depth--
	}
}

// Depth returns the current nesting depth.
func (g *Guard) Depth() int { return int(g.depth) }

// Node counts one created value and charges EstNodeBytes.
func (g *Guard) Node() error {
	if g.nodes >= g.maxNodes {
		return exceeded(LimitNodes, g.maxNodes, g.nodes+1)
	}
	if err := charge(&g.mem, EstNodeBytes, g.maxMem, LimitMemory); err != nil {
		return err
	}
	g.nodes++
	return g.tick()
}

// Nodes returns the number of values counted so far.
func (g *Guard) Nodes() int64 { return g.nodes }

// Bytes records n more input bytes consumed. Call it per chunk, or once with
// the total when the input size is known up front. Negative n is an error.
func (g *Guard) Bytes(n int) error {
	if err := charge(&g.bytes, int64(n), g.maxBytes, LimitBytes); err != nil {
		return err
	}
	return g.tick()
}

// CheckStringLen checks a string or key length of n bytes against
// MaxStringBytes without charging memory. Use it while accumulating a long
// string, before allocating; see Guard for the pattern.
func (g *Guard) CheckStringLen(n int) error {
	if n < 0 {
		return negative(LimitStringBytes, int64(n))
	}
	if int64(n) > g.maxString {
		return exceeded(LimitStringBytes, g.maxString, int64(n))
	}
	return g.tick()
}

// String checks a finished string of n bytes against MaxStringBytes and
// charges its estimated memory. Call it once per string with the final
// length.
func (g *Guard) String(n int) error {
	return g.text(n, 0)
}

// Number checks the text of a number of n bytes like String. Parsers should
// call it for every number so MaxStringBytes covers number text.
func (g *Guard) Number(n int) error {
	return g.text(n, 0)
}

// Key checks a finished map key of n bytes like String, and also charges
// EstMapEntryBytes.
func (g *Guard) Key(n int) error {
	return g.text(n, EstMapEntryBytes)
}

func (g *Guard) text(n int, extra int64) error {
	if n < 0 {
		return negative(LimitStringBytes, int64(n))
	}
	if int64(n) > g.maxString {
		return exceeded(LimitStringBytes, g.maxString, int64(n))
	}
	if err := charge(&g.mem, int64(n), g.maxMem, LimitMemory); err != nil {
		return err
	}
	if err := charge(&g.mem, EstStringHeaderBytes+extra, g.maxMem, LimitMemory); err != nil {
		return err
	}
	return g.tick()
}

// MapSize checks the number of keys a map holds, counting the key just
// added, against MaxKeysPerMap.
func (g *Guard) MapSize(keys int) error {
	if keys < 0 {
		return negative(LimitKeysPerMap, int64(keys))
	}
	if int64(keys) > g.maxKeys {
		return exceeded(LimitKeysPerMap, g.maxKeys, int64(keys))
	}
	return g.tick()
}

// Alias records n nodes produced by alias expansion (YAML). The expanded
// nodes still need Node calls to count against the other budgets.
func (g *Guard) Alias(n int64) error {
	if err := charge(&g.alias, n, g.maxAlias, LimitAliasExpansion); err != nil {
		return err
	}
	return g.tick()
}

// Memory charges n bytes against the estimated-memory budget. Negative n is
// an error; the budget cannot be credited back.
func (g *Guard) Memory(n int64) error {
	if err := charge(&g.mem, n, g.maxMem, LimitMemory); err != nil {
		return err
	}
	return g.tick()
}

// MemoryUsed returns the estimated memory charged so far.
func (g *Guard) MemoryUsed() int64 { return g.mem }
