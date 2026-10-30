package data

import (
	"context"
	"testing"
)

func BenchmarkGuardNode(b *testing.B) {
	g := NewGuard(Limits{})
	b.ReportAllocs()
	for b.Loop() {
		_ = g.Node()
	}
}

func BenchmarkGuardNodeAllLimits(b *testing.B) {
	g := NewGuard(Limits{MaxNodes: 1 << 60, MaxMemory: 1 << 60})
	b.ReportAllocs()
	for b.Loop() {
		_ = g.Node()
	}
}

func BenchmarkGuardNodeContext(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g := NewGuard(Limits{Context: ctx})
	b.ReportAllocs()
	for b.Loop() {
		_ = g.Node()
	}
}

func BenchmarkGuardEnterLeave(b *testing.B) {
	g := NewGuard(Limits{})
	b.ReportAllocs()
	for b.Loop() {
		_ = g.Enter()
		g.Leave()
	}
}

func BenchmarkGuardString(b *testing.B) {
	g := NewGuard(Limits{MaxStringBytes: 1 << 20})
	b.ReportAllocs()
	for b.Loop() {
		_ = g.String(32)
	}
}

func BenchmarkMapAppendGet(b *testing.B) {
	keys := make([]string, 64)
	for i := range keys {
		keys[i] = string(rune('a'+i%26)) + string(rune('A'+i/26))
	}
	b.ReportAllocs()
	for b.Loop() {
		m := NewMap(len(keys))
		for _, k := range keys {
			m.Add(k, NullValue())
		}
		_, _ = m.Get("zz")
	}
}
