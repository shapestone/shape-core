package data

import (
	"context"
	"errors"
	"math"
	"testing"
)

// limitCase drives one limit: ok runs up to and including the boundary, over
// pushes one past it and must return a *LimitError for want.
type limitCase struct {
	name string
	l    Limits
	want string
	max  int64
	got  int64
	at   func(g *Guard) error // exactly at the limit: must pass
	over func(g *Guard) error // one past: must fail
}

func TestLimitBoundaries(t *testing.T) {
	repeat := func(n int, f func() error) error {
		for range n {
			if err := f(); err != nil {
				return err
			}
		}
		return nil
	}
	tests := []limitCase{
		{
			name: "depth", l: Limits{MaxDepth: 3}, want: LimitDepth, max: 3, got: 4,
			at:   func(g *Guard) error { return repeat(3, g.Enter) },
			over: func(g *Guard) error { return repeat(4, g.Enter) },
		},
		{
			name: "bytes", l: Limits{MaxBytes: 100}, want: LimitBytes, max: 100, got: 101,
			at:   func(g *Guard) error { return g.Bytes(100) },
			over: func(g *Guard) error { return errors.Join(g.Bytes(60), g.Bytes(41)) },
		},
		{
			name: "nodes", l: Limits{MaxNodes: 5}, want: LimitNodes, max: 5, got: 6,
			at:   func(g *Guard) error { return repeat(5, g.Node) },
			over: func(g *Guard) error { return repeat(6, g.Node) },
		},
		{
			name: "string", l: Limits{MaxStringBytes: 8}, want: LimitStringBytes, max: 8, got: 9,
			at:   func(g *Guard) error { return g.String(8) },
			over: func(g *Guard) error { return g.String(9) },
		},
		{
			name: "key", l: Limits{MaxStringBytes: 8}, want: LimitStringBytes, max: 8, got: 9,
			at:   func(g *Guard) error { return g.Key(8) },
			over: func(g *Guard) error { return g.Key(9) },
		},
		{
			name: "keys per map", l: Limits{MaxKeysPerMap: 4}, want: LimitKeysPerMap, max: 4, got: 5,
			at:   func(g *Guard) error { return g.MapSize(4) },
			over: func(g *Guard) error { return g.MapSize(5) },
		},
		{
			name: "alias", l: Limits{MaxAliasExpansion: 10}, want: LimitAliasExpansion, max: 10, got: 11,
			at:   func(g *Guard) error { return g.Alias(10) },
			over: func(g *Guard) error { return errors.Join(g.Alias(10), g.Alias(1)) },
		},
		{
			name: "memory", l: Limits{MaxMemory: 10 * EstNodeBytes}, want: LimitMemory,
			max: 10 * EstNodeBytes, got: 11 * EstNodeBytes,
			at:   func(g *Guard) error { return repeat(10, g.Node) },
			over: func(g *Guard) error { return repeat(11, g.Node) },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.at(NewGuard(tt.l)); err != nil {
				t.Fatalf("at limit: %v", err)
			}
			err := tt.over(NewGuard(tt.l))
			var le *LimitError
			if !errors.As(err, &le) {
				t.Fatalf("past limit: err = %v, want *LimitError", err)
			}
			if !errors.Is(err, ErrLimit) {
				t.Error("not ErrLimit")
			}
			if le.Limit != tt.want || le.Max != tt.max || le.Got != tt.got {
				t.Errorf("got %+v, want limit=%s max=%d got=%d", *le, tt.want, tt.max, tt.got)
			}
		})
	}
}

func TestDefaults(t *testing.T) {
	t.Run("depth on by default", func(t *testing.T) {
		g := NewGuard(Limits{})
		for range DefaultMaxDepth {
			if err := g.Enter(); err != nil {
				t.Fatal(err)
			}
		}
		var le *LimitError
		if err := g.Enter(); !errors.As(err, &le) || le.Max != DefaultMaxDepth {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("alias on by default", func(t *testing.T) {
		g := NewGuard(Limits{})
		if err := g.Alias(DefaultMaxAliasExpansion); err != nil {
			t.Fatal(err)
		}
		if err := g.Alias(1); !errors.Is(err, ErrLimit) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("others off by default", func(t *testing.T) {
		g := NewGuard(Limits{})
		for _, err := range []error{
			g.Bytes(math.MaxInt32), g.Bytes(math.MaxInt32), g.String(1 << 30), g.Key(1 << 30),
			g.MapSize(1 << 30), g.Memory(1 << 50), g.Node(),
		} {
			if err != nil {
				t.Errorf("unexpected: %v", err)
			}
		}
	})
}

func TestNoLimit(t *testing.T) {
	g := NewGuard(Limits{MaxDepth: NoLimit, MaxAliasExpansion: NoLimit, MaxNodes: NoLimit, MaxBytes: -5})
	for range DefaultMaxDepth + 100 {
		if err := g.Enter(); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.Alias(DefaultMaxAliasExpansion * 10); err != nil {
		t.Fatal(err)
	}
	if g.Depth() != DefaultMaxDepth+100 {
		t.Errorf("depth = %d", g.Depth())
	}
}

func TestEnterLeave(t *testing.T) {
	g := NewGuard(Limits{MaxDepth: 2})
	_ = g.Enter()
	_ = g.Enter()
	if err := g.Enter(); err == nil {
		t.Fatal("expected error")
	}
	if g.Depth() != 2 {
		t.Errorf("failed Enter changed depth to %d", g.Depth())
	}
	g.Leave()
	if err := g.Enter(); err != nil {
		t.Errorf("Enter after Leave: %v", err)
	}
	g.Leave()
	g.Leave()
	g.Leave() // extra Leave is harmless
	if g.Depth() != 0 {
		t.Errorf("depth = %d", g.Depth())
	}
}

func TestMemoryEstimate(t *testing.T) {
	g := NewGuard(Limits{})
	_ = g.Node()
	_ = g.String(10)
	_ = g.Key(5)
	want := int64(EstNodeBytes + (10 + EstStringHeaderBytes) + (5 + EstStringHeaderBytes + EstMapEntryBytes))
	if g.MemoryUsed() != want {
		t.Errorf("MemoryUsed = %d, want %d", g.MemoryUsed(), want)
	}
	if g.Nodes() != 1 {
		t.Errorf("Nodes = %d", g.Nodes())
	}

	// String and Key charge memory too.
	for name, f := range map[string]func(*Guard) error{
		"string": func(g *Guard) error { return g.String(100) },
		"key":    func(g *Guard) error { return g.Key(100) },
	} {
		err := f(NewGuard(Limits{MaxMemory: 50}))
		var le *LimitError
		if !errors.As(err, &le) || le.Limit != LimitMemory {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}

// TestCancelInLongLoop cancels part way through a loop and checks the guard
// notices within one poll interval and returns the context error unchanged.
func TestCancelInLongLoop(t *testing.T) {
	ops := map[string]func(*Guard) error{
		"Node":   (*Guard).Node,
		"Enter":  func(g *Guard) error { err := g.Enter(); g.Leave(); return err },
		"Bytes":  func(g *Guard) error { return g.Bytes(1) },
		"String": func(g *Guard) error { return g.String(1) },
		"Key":    func(g *Guard) error { return g.Key(1) },
		"Alias":  func(g *Guard) error { return g.Alias(0) },
	}
	for name, op := range ops {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			g := NewGuard(Limits{Context: ctx})
			const cancelAt, total = 1000, 100000
			var err error
			i := 0
			for ; i < total; i++ {
				if i == cancelAt {
					cancel()
				}
				if err = op(g); err != nil {
					break
				}
			}
			if err != context.Canceled {
				t.Fatalf("err = %v after %d ops, want context.Canceled", err, i)
			}
			if i < cancelAt || i > cancelAt+pollMask+1 {
				t.Errorf("noticed after %d ops; cancelled at %d, poll interval %d", i, cancelAt, pollMask+1)
			}
		})
	}
}

func TestNoContext(t *testing.T) {
	if err := NewGuard(Limits{}).Check(); err != nil {
		t.Error(err)
	}
	// Background never cancels, so the guard skips polling it.
	if g := NewGuard(Limits{Context: context.Background()}); g.ctx != nil {
		t.Error("guard kept a never-cancelled context")
	}
	g := NewGuard(Limits{})
	for range 1000 {
		if err := g.Node(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestResolve(t *testing.T) {
	tests := []struct {
		v, def, want int64
	}{
		{0, 7, 7}, {5, 7, 5}, {-1, 7, math.MaxInt64}, {NoLimit, 0, math.MaxInt64},
	}
	for _, tt := range tests {
		if got := resolve(tt.v, tt.def); got != tt.want {
			t.Errorf("resolve(%d,%d) = %d, want %d", tt.v, tt.def, got, tt.want)
		}
	}
}
