package data

import (
	"context"
	"math"
)

// Per-node memory estimates used for the MaxMemory budget. They model the
// in-memory size of the result on a 64-bit platform, not exact allocation.
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
// Every check returns nil, a *LimitError, or the context's error unchanged.
// Once a check fails the parse should stop; the guard is not reusable.
//
// Memory estimate: Node charges EstNodeBytes, String and Key charge their
// byte length plus EstStringHeaderBytes, and Key also charges
// EstMapEntryBytes. Parsers that allocate more can add to the budget with
// Memory.
//
// Unlimited limits are stored as math.MaxInt64, so the hot path is one add
// and one compare per check with no branch on "enabled".
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
		maxDepth:  resolve(int64(l.MaxDepth), DefaultMaxDepth),
		maxNodes:  resolve(l.MaxNodes, math.MaxInt64),
		maxBytes:  resolve(l.MaxBytes, math.MaxInt64),
		maxString: resolve(int64(l.MaxStringBytes), math.MaxInt64),
		maxKeys:   resolve(int64(l.MaxKeysPerMap), math.MaxInt64),
		maxAlias:  resolve(l.MaxAliasExpansion, DefaultMaxAliasExpansion),
		maxMem:    resolve(l.MaxMemory, math.MaxInt64),
	}
	if l.Context != nil && l.Context.Done() != nil {
		g.ctx = l.Context // a never-cancelled context needs no polling
	}
	return g
}

// Check polls the context immediately. Call it once before parsing to
// reject an already-cancelled context.
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

// Enter descends one nesting level. Depth is counted by the guard, not by
// the Go stack. On error the depth is unchanged.
func (g *Guard) Enter() error {
	if g.depth+1 > g.maxDepth {
		return exceeded(LimitDepth, g.maxDepth, g.depth+1)
	}
	g.depth++
	return g.tick()
}

// Leave ascends one nesting level. It never fails.
func (g *Guard) Leave() {
	if g.depth > 0 {
		g.depth--
	}
}

// Depth returns the current nesting depth.
func (g *Guard) Depth() int { return int(g.depth) }

// Node counts one created value and charges EstNodeBytes.
func (g *Guard) Node() error {
	g.nodes++
	if g.nodes > g.maxNodes {
		return exceeded(LimitNodes, g.maxNodes, g.nodes)
	}
	if err := g.Memory(EstNodeBytes); err != nil {
		return err
	}
	return g.tick()
}

// Nodes returns the number of values counted so far.
func (g *Guard) Nodes() int64 { return g.nodes }

// Bytes records n more input bytes consumed. Call it per chunk, or once with
// the total when the input size is known up front.
func (g *Guard) Bytes(n int) error {
	g.bytes += int64(n)
	if g.bytes > g.maxBytes {
		return exceeded(LimitBytes, g.maxBytes, g.bytes)
	}
	return g.tick()
}

// String checks a string of n bytes against MaxStringBytes and charges its
// estimated memory.
func (g *Guard) String(n int) error {
	if int64(n) > g.maxString {
		return exceeded(LimitStringBytes, g.maxString, int64(n))
	}
	if err := g.Memory(int64(n) + EstStringHeaderBytes); err != nil {
		return err
	}
	return g.tick()
}

// Key checks a map key of n bytes like String, and also charges
// EstMapEntryBytes.
func (g *Guard) Key(n int) error {
	if int64(n) > g.maxString {
		return exceeded(LimitStringBytes, g.maxString, int64(n))
	}
	if err := g.Memory(int64(n) + EstStringHeaderBytes + EstMapEntryBytes); err != nil {
		return err
	}
	return g.tick()
}

// MapSize checks the number of keys a map holds, counting the key just
// added, against MaxKeysPerMap.
func (g *Guard) MapSize(keys int) error {
	if int64(keys) > g.maxKeys {
		return exceeded(LimitKeysPerMap, g.maxKeys, int64(keys))
	}
	return nil
}

// Alias records n nodes produced by alias expansion (YAML). The expanded
// nodes still need Node calls to count against the other budgets.
func (g *Guard) Alias(n int64) error {
	g.alias += n
	if g.alias > g.maxAlias {
		return exceeded(LimitAliasExpansion, g.maxAlias, g.alias)
	}
	return g.tick()
}

// Memory charges n bytes against the estimated-memory budget.
func (g *Guard) Memory(n int64) error {
	g.mem += n
	if g.mem > g.maxMem {
		return exceeded(LimitMemory, g.maxMem, g.mem)
	}
	return nil
}

// MemoryUsed returns the estimated memory charged so far.
func (g *Guard) MemoryUsed() int64 { return g.mem }
