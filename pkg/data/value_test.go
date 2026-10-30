package data

import (
	"fmt"
	"slices"
	"testing"
)

func TestKindString(t *testing.T) {
	tests := []struct {
		k    Kind
		want string
	}{
		{KindNull, "null"}, {KindBool, "bool"}, {KindInteger, "integer"},
		{KindDecimal, "decimal"}, {KindString, "string"}, {KindList, "list"},
		{KindMap, "map"}, {Kind(99), "Kind(99)"},
	}
	for _, tt := range tests {
		if got := tt.k.String(); got != tt.want {
			t.Errorf("%d: got %q, want %q", tt.k, got, tt.want)
		}
	}
}

func TestValueKinds(t *testing.T) {
	m := NewMap(0)
	num := func(s string) Value {
		v, err := NumberValue(s)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	tests := []struct {
		name string
		v    Value
		kind Kind
	}{
		{"zero", Value{}, KindNull},
		{"null", NullValue(), KindNull},
		{"true", BoolValue(true), KindBool},
		{"int", IntValue(-5), KindInteger},
		{"integer text", num("-0"), KindInteger},
		{"decimal text", num("1.0"), KindDecimal},
		{"exponent", num("1e3"), KindDecimal},
		{"string", StringValue("x"), KindString},
		{"list", ListValue(List{NullValue()}), KindList},
		{"map", MapValue(m), KindMap},
		{"nil map", MapValue(nil), KindMap},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.v.Kind() != tt.kind {
				t.Errorf("kind = %v, want %v", tt.v.Kind(), tt.kind)
			}
			if tt.v.IsNull() != (tt.kind == KindNull) {
				t.Error("IsNull mismatch")
			}
		})
	}
}

func TestValueAccessors(t *testing.T) {
	if !BoolValue(true).Bool() || BoolValue(false).Bool() || StringValue("x").Bool() {
		t.Error("Bool")
	}
	if got := IntValue(-5).Number().Text; got != "-5" {
		t.Errorf("Number = %q", got)
	}
	if got := StringValue("x").Number(); got.Text != "" {
		t.Errorf("Number of string = %q", got.Text)
	}
	if StringValue("hi").Str() != "hi" || IntValue(1).Str() != "" {
		t.Error("Str")
	}
	l := ListValue(List{NullValue(), BoolValue(true)})
	if l.List().Len() != 2 || NullValue().List() != nil {
		t.Error("List")
	}
	m := NewMap(1)
	if MapValue(m).Map() != m || NullValue().Map() != nil {
		t.Error("Map")
	}
	if _, err := NumberValue("nope"); err == nil {
		t.Error("NumberValue accepted invalid text")
	}
	if v := UncheckedNumberValue("2.50"); v.Kind() != KindDecimal || v.Number().Text != "2.50" {
		t.Error("UncheckedNumberValue")
	}
}

func keys(m *Map) []string {
	var out []string
	for k := range m.All() {
		out = append(out, k)
	}
	return out
}

func TestMapOrder(t *testing.T) {
	for _, n := range []int{0, 1, 5, indexThreshold - 1, indexThreshold, indexThreshold + 1, 200} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			m := &Map{}
			var want []string
			// descending keys so insertion order differs from sorted order
			for i := n; i > 0; i-- {
				k := fmt.Sprintf("k%03d", i)
				want = append(want, k)
				m.Append(k, IntValue(int64(i)))
			}
			if m.Len() != n || !slices.Equal(keys(m), want) {
				t.Fatalf("order wrong: %v", keys(m))
			}
			for i, e := range m.Entries() {
				if e.Key != want[i] {
					t.Fatalf("entry %d = %s", i, e.Key)
				}
				if m.Index(e.Key) != i {
					t.Fatalf("Index(%s) = %d, want %d", e.Key, m.Index(e.Key), i)
				}
			}
			if m.Has("missing") || m.Index("missing") != -1 {
				t.Error("found missing key")
			}
		})
	}
}

func TestMapDuplicatePrimitives(t *testing.T) {
	for _, n := range []int{3, 40} { // linear and indexed
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			fill := func() *Map {
				m := NewMap(n)
				for i := range n {
					m.Append(fmt.Sprintf("k%d", i), IntValue(int64(i)))
				}
				return m
			}

			m := fill()
			if m.Add("k1", IntValue(99)) {
				t.Error("Add replaced existing key")
			}
			if v, _ := m.Get("k1"); v.Number().Text != "1" {
				t.Errorf("first-wins: got %v", v.Number().Text)
			}
			if !m.Add("new", IntValue(7)) || m.Len() != n+1 {
				t.Error("Add of new key failed")
			}

			m = fill()
			if !m.Set("k1", IntValue(99)) {
				t.Error("Set did not report replace")
			}
			if v, _ := m.Get("k1"); v.Number().Text != "99" || m.Index("k1") != 1 || m.Len() != n {
				t.Error("last-wins must replace in place")
			}
			if m.Set("fresh", NullValue()) || m.Index("fresh") != n {
				t.Error("Set of new key must append")
			}

			m = fill()
			m.Append("k0", IntValue(-1)) // keep-all policy
			if m.Len() != n+1 {
				t.Error("Append dropped duplicate")
			}
			if v, _ := m.Get("k0"); v.Number().Text != "0" || m.Index("k0") != 0 {
				t.Error("lookup must find first duplicate")
			}
			if _, ok := m.Get("absent"); ok {
				t.Error("Get found absent key")
			}
		})
	}
}

func TestMapAllEarlyStop(t *testing.T) {
	m := &Map{}
	for i := range 5 {
		m.Append(fmt.Sprint(i), NullValue())
	}
	count := 0
	for range m.All() {
		count++
		if count == 2 {
			break
		}
	}
	if count != 2 {
		t.Errorf("count = %d", count)
	}
}
