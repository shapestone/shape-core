package data

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"
)

func TestGuardRejectsNegativeAndOverflow(t *testing.T) {
	tests := []struct {
		name string
		l    Limits
		op   func(*Guard) error
	}{
		{"alias negative", Limits{}, func(g *Guard) error { return g.Alias(-1) }},
		{"memory negative", Limits{MaxMemory: 100}, func(g *Guard) error { return g.Memory(-1_000_000) }},
		{"bytes negative", Limits{MaxBytes: 10}, func(g *Guard) error { return g.Bytes(-1000) }},
		{"string negative", Limits{}, func(g *Guard) error { return g.String(-1) }},
		{"number negative", Limits{}, func(g *Guard) error { return g.Number(-1) }},
		{"key negative", Limits{}, func(g *Guard) error { return g.Key(-1) }},
		{"check len negative", Limits{}, func(g *Guard) error { return g.CheckStringLen(-1) }},
		{"mapsize negative", Limits{}, func(g *Guard) error { return g.MapSize(-1) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.op(NewGuard(tt.l))
			if !errors.Is(err, ErrNegativeAmount) || errors.Is(err, ErrLimit) {
				t.Errorf("err = %v, want ErrNegativeAmount", err)
			}
		})
	}

	t.Run("negative does not credit budget", func(t *testing.T) {
		g := NewGuard(Limits{MaxMemory: 100})
		_ = g.Memory(50)
		_ = g.Memory(-1_000_000)
		if err := g.Memory(60); !errors.Is(err, ErrLimit) {
			t.Errorf("err = %v, want limit error", err)
		}
	})
	t.Run("alias cannot wrap when unlimited", func(t *testing.T) {
		g := NewGuard(Limits{MaxAliasExpansion: NoLimit})
		_ = g.Alias(math.MaxInt64)
		if err := g.Alias(math.MaxInt64); !errors.Is(err, ErrLimit) {
			t.Errorf("err = %v, want limit error", err)
		}
		var le *LimitError
		if err := g.Alias(2); !errors.As(err, &le) || le.Got != math.MaxInt64 {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("memory cannot wrap", func(t *testing.T) {
		g := NewGuard(Limits{MaxMemory: 100})
		_ = g.Memory(50)
		if err := g.Memory(math.MaxInt64); !errors.Is(err, ErrLimit) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("huge string and map size", func(t *testing.T) {
		g := NewGuard(Limits{MaxMemory: 1000})
		if err := g.String(math.MaxInt); !errors.Is(err, ErrLimit) {
			t.Errorf("String: %v", err)
		}
		if err := g.MapSize(math.MaxInt); err != nil {
			t.Errorf("MapSize unlimited: %v", err)
		}
		g = NewGuard(Limits{MaxKeysPerMap: 3})
		if err := g.MapSize(math.MaxInt); !errors.Is(err, ErrLimit) {
			t.Errorf("MapSize: %v", err)
		}
	})
	t.Run("failed charge is not applied", func(t *testing.T) {
		g := NewGuard(Limits{MaxBytes: 10})
		_ = g.Bytes(8)
		_ = g.Bytes(8)
		if err := g.Bytes(2); err != nil {
			t.Errorf("Bytes(2) after failed Bytes(8): %v", err)
		}
	})
}

func TestStreamingStringCheck(t *testing.T) {
	g := NewGuard(Limits{MaxStringBytes: 10, MaxMemory: 1000})
	for _, n := range []int{1, 5, 10} {
		if err := g.CheckStringLen(n); err != nil {
			t.Fatalf("CheckStringLen(%d): %v", n, err)
		}
	}
	if g.MemoryUsed() != 0 {
		t.Errorf("CheckStringLen charged memory: %d", g.MemoryUsed())
	}
	var le *LimitError
	if err := g.CheckStringLen(11); !errors.As(err, &le) || le.Limit != LimitStringBytes || le.Got != 11 {
		t.Errorf("err = %v", err)
	}
	if err := g.Number(10); err != nil {
		t.Fatal(err)
	}
	if err := g.Number(11); !errors.Is(err, ErrLimit) {
		t.Errorf("Number(11): %v", err)
	}
	if g.MemoryUsed() != 10+EstStringHeaderBytes {
		t.Errorf("MemoryUsed = %d", g.MemoryUsed())
	}
}

func TestEnterContextErrorLeavesDepth(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	g := NewGuard(Limits{Context: ctx})
	var err error
	n := 0
	for err == nil && n < 1000 {
		err = g.Enter()
		n++
	}
	if err != context.Canceled {
		t.Fatalf("err = %v", err)
	}
	if g.Depth() != n-1 {
		t.Errorf("depth = %d after %d Enters, the failed one must not count", g.Depth(), n)
	}
}

func TestContextPollThroughAllMethods(t *testing.T) {
	ops := map[string]func(*Guard) error{
		"MapSize":        func(g *Guard) error { return g.MapSize(1) },
		"Memory":         func(g *Guard) error { return g.Memory(1) },
		"CheckStringLen": func(g *Guard) error { return g.CheckStringLen(1) },
		"Number":         func(g *Guard) error { return g.Number(1) },
	}
	for name, op := range ops {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			g := NewGuard(Limits{Context: ctx})
			for i := 0; i < pollMask+1; i++ {
				if err := op(g); err != nil {
					if err != context.Canceled {
						t.Fatalf("err = %v", err)
					}
					return
				}
			}
			t.Error("cancellation not noticed within one poll interval")
		})
	}
}

func TestNilMapIsEmpty(t *testing.T) {
	m := StringValue("x").Map()
	if m != nil {
		t.Fatal("Map() of a string should be nil")
	}
	if m.Len() != 0 || m.Entries() != nil || m.Has("a") || m.Index("a") != -1 {
		t.Error("nil map not empty")
	}
	if _, ok := m.Get("a"); ok {
		t.Error("nil map Get found key")
	}
	for range m.All() {
		t.Error("nil map yielded")
	}
	for _, v := range []Value{NullValue(), BoolValue(true), IntValue(1), ListValue(nil)} {
		if v.Map().Len() != 0 {
			t.Errorf("%v: Map().Len() != 0", v.Kind())
		}
	}
}

func TestNewMapCapacity(t *testing.T) {
	for _, c := range []int{-1, math.MinInt, 0, 5, 1 << 20, math.MaxInt} {
		t.Run(fmt.Sprint(c), func(t *testing.T) {
			m := NewMap(c)
			if cap(m.entries) > maxPrealloc {
				t.Errorf("cap = %d", cap(m.entries))
			}
			m.Append("a", NullValue())
			if m.Len() != 1 {
				t.Error("append failed")
			}
		})
	}
}

func TestMapDuplicatesAcrossIndexThreshold(t *testing.T) {
	for _, n := range []int{indexThreshold - 2, indexThreshold - 1, indexThreshold, indexThreshold + 5} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			m := &Map{}
			for i := 0; i < n; i++ {
				m.Append(fmt.Sprint(i%4), IntValue(int64(i)))
			}
			for k := 0; k < 4; k++ {
				key := fmt.Sprint(k)
				if m.Index(key) != k {
					t.Fatalf("Index(%s) = %d, want first at %d", key, m.Index(key), k)
				}
			}
			// Set replaces only the first of the duplicates.
			if !m.Set("1", StringValue("new")) {
				t.Fatal("Set did not replace")
			}
			if m.At(1).Value.Str() != "new" || m.At(5).Value.Str() != "" {
				t.Error("Set must change only the first duplicate")
			}
			if m.Len() != n {
				t.Errorf("Len = %d", m.Len())
			}
			if m.Add("2", NullValue()) {
				t.Error("Add added an existing key")
			}
		})
	}
}

func TestEntriesIsACopy(t *testing.T) {
	m := &Map{}
	for i := 0; i < indexThreshold+4; i++ {
		m.Append(fmt.Sprint(i), NullValue())
	}
	e := m.Entries()
	e[0].Key = "zzz"
	if m.Has("zzz") || !m.Has("0") || m.At(0).Key != "0" {
		t.Error("mutating Entries() copy changed the map")
	}
}

func TestBigDigitCap(t *testing.T) {
	long := strings.Repeat("9", DefaultMaxBigDigits+1)
	tests := []struct {
		name string
		text string
		max  int
		err  error
	}{
		{"at default cap", strings.Repeat("9", DefaultMaxBigDigits), DefaultMaxBigDigits, nil},
		{"past default cap", long, DefaultMaxBigDigits, ErrNumberRange},
		{"fraction counts", "0." + long, DefaultMaxBigDigits, ErrNumberRange},
		{"custom cap ok", "12345", 5, nil},
		{"custom cap over", "123456", 5, ErrNumberRange},
		{"million digits fast fail", strings.Repeat("9", 1_000_000), DefaultMaxBigDigits, ErrNumberRange},
		{"refused by math/big is range", "0." + strings.Repeat("9", 1_000_000) + "e-65536", 2_000_000, ErrNumberRange}, // slow: skipped in -short
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if testing.Short() && len(tt.text) > 100_000 && tt.max > DefaultMaxBigDigits {
				t.Skip("quadratic math/big parse")
			}
			_, err := Number{Text: tt.text}.BigWithMaxDigits(tt.max)
			if tt.err == nil && err != nil || tt.err != nil && !errors.Is(err, tt.err) {
				t.Errorf("err = %v, want %v", err, tt.err)
			}
			if errors.Is(err, ErrNumberSyntax) {
				t.Error("valid text must not report a syntax error")
			}
		})
	}

	if _, err := (Number{Text: long}).BigInt(); !errors.Is(err, ErrNumberRange) {
		t.Errorf("BigInt past cap: %v", err)
	}
	if _, err := (Number{Text: long}).Big(); !errors.Is(err, ErrNumberRange) {
		t.Errorf("Big past cap: %v", err)
	}
	if v, err := (Number{Text: "123456"}).BigIntWithMaxDigits(6); err != nil || v.String() != "123456" {
		t.Errorf("BigIntWithMaxDigits: %v, %v", v, err)
	}
	if _, err := (Number{Text: "1234567"}).BigIntWithMaxDigits(6); !errors.Is(err, ErrNumberRange) {
		t.Errorf("BigIntWithMaxDigits over: %v", err)
	}
}

func TestNumberErrorsAreBounded(t *testing.T) {
	huge := strings.Repeat("x", 1<<20)
	multi := strings.Repeat("é", 100) // cut must not split a rune
	for name, f := range map[string]func(string) error{
		"ParseNumber": func(s string) error { _, err := ParseNumber(s); return err },
		"Int64":       func(s string) error { _, err := Number{Text: s}.Int64(); return err },
		"Float64":     func(s string) error { _, err := Number{Text: s}.Float64(); return err },
		"Big":         func(s string) error { _, err := Number{Text: s}.Big(); return err },
		"BigInt":      func(s string) error { _, err := Number{Text: s}.BigInt(); return err },
	} {
		for _, in := range []string{huge, multi} {
			err := f(in)
			if err == nil || len(err.Error()) > 200 {
				t.Errorf("%s: err length %d", name, len(fmt.Sprint(err)))
			}
			if err != nil && !strings.Contains(err.Error(), "truncated") {
				t.Errorf("%s: no truncation marker", name)
			}
		}
	}
	long := strings.Repeat("9", 100)
	if err := (func() error { _, e := Number{Text: long}.Int64(); return e })(); !errors.Is(err, ErrNumberRange) || len(err.Error()) > 200 {
		t.Errorf("Int64 range error: %v", err)
	}
}

func TestValueEqualAndString(t *testing.T) {
	mk := func() Value {
		m := NewMap(0)
		m.Append("a", IntValue(1))
		m.Append("b", ListValue(List{NullValue(), BoolValue(true), StringValue(`q"`), UncheckedNumberValue("1.5")}))
		return MapValue(m)
	}
	a, b := mk(), mk()
	if !a.Equal(b) {
		t.Error("equal values differ")
	}
	if got, want := a.String(), `{"a":1,"b":[null,true,"q\"",1.5]}`; got != want {
		t.Errorf("String = %s, want %s", got, want)
	}
	if got := NullValue().String(); got != "null" {
		t.Errorf("String = %s", got)
	}

	other := NewMap(0)
	other.Append("b", NullValue())
	other.Append("a", NullValue())
	swapped := NewMap(0)
	swapped.Append("b", ListValue(nil))
	swapped.Append("a", IntValue(1))
	tests := []struct {
		name string
		x, y Value
		want bool
	}{
		{"null", NullValue(), NullValue(), true},
		{"bool", BoolValue(true), BoolValue(false), false},
		{"kind", IntValue(1), StringValue("1"), false},
		{"1 vs 1.0", IntValue(1), UncheckedNumberValue("1.0"), false},
		{"string", StringValue("a"), StringValue("a"), true},
		{"list len", ListValue(List{NullValue()}), ListValue(nil), false},
		{"list elem", ListValue(List{IntValue(1)}), ListValue(List{IntValue(2)}), false},
		{"map order", a, MapValue(swapped), false},
		{"map vs other", a, MapValue(other), false},
		{"nil vs empty map", MapValue(nil), MapValue(NewMap(0)), true},
	}
	for _, tt := range tests {
		if got := tt.x.Equal(tt.y); got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestZeroNumber(t *testing.T) {
	var n Number
	if n.Valid() {
		t.Error("zero Number is valid")
	}
	if _, err := n.Int64(); !errors.Is(err, ErrNumberSyntax) {
		t.Errorf("err = %v", err)
	}
	if got := IntValue(1).Number(); !got.Valid() {
		t.Error("IntValue number invalid")
	}
	if StringValue("x").Number().Valid() {
		t.Error("non-numeric Value.Number() must be invalid")
	}
}
