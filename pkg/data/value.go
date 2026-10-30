package data

import (
	"slices"
	"strconv"
	"strings"
)

// Kind identifies which type a Value holds.
type Kind uint8

// The kinds of Value. The zero Kind is KindNull, so the zero Value is null.
const (
	KindNull Kind = iota
	KindBool
	KindInteger
	KindDecimal
	KindString
	KindList
	KindMap
)

var kindNames = [...]string{"null", "bool", "integer", "decimal", "string", "list", "map"}

// String returns the lower-case name of the kind.
func (k Kind) String() string {
	if int(k) < len(kindNames) {
		return kindNames[k]
	}
	return "Kind(" + strconv.Itoa(int(k)) + ")"
}

// List is an ordered sequence of values.
type List []Value

// Len returns the number of elements.
func (l List) Len() int { return len(l) }

// Value holds exactly one of the kinds. The zero Value is null. Accessors for
// a kind other than the value's own return the zero result.
type Value struct {
	kind Kind
	b    bool
	text string // string content, or exact number text
	list List
	m    *Map
}

// NullValue returns a null value.
func NullValue() Value { return Value{} }

// BoolValue returns a bool value.
func BoolValue(b bool) Value { return Value{kind: KindBool, b: b} }

// IntValue returns an Integer value written in decimal.
func IntValue(i int64) Value {
	return Value{kind: KindInteger, text: strconv.FormatInt(i, 10)}
}

// NumberValue validates text against the JSON number grammar and returns an
// Integer or Decimal value holding the exact text.
func NumberValue(text string) (Value, error) {
	n, err := ParseNumber(text)
	if err != nil {
		return Value{}, err
	}
	return UncheckedNumberValue(n.Text), nil
}

// UncheckedNumberValue is NumberValue for text the caller has already
// validated, such as a parser that tokenised it with the JSON grammar. It
// still classifies the text as Integer or Decimal.
func UncheckedNumberValue(text string) Value {
	return Value{kind: Number{Text: text}.Kind(), text: text}
}

// StringValue returns a string value.
func StringValue(s string) Value { return Value{kind: KindString, text: s} }

// ListValue returns a list value.
func ListValue(l List) Value { return Value{kind: KindList, list: l} }

// MapValue returns a map value. A nil map is treated as empty.
func MapValue(m *Map) Value {
	if m == nil {
		m = NewMap(0)
	}
	return Value{kind: KindMap, m: m}
}

// Kind returns the kind of the value.
func (v Value) Kind() Kind { return v.kind }

// IsNull reports whether the value is null.
func (v Value) IsNull() bool { return v.kind == KindNull }

// Bool returns the bool, or false for other kinds.
func (v Value) Bool() bool { return v.kind == KindBool && v.b }

// Number returns the number, or the zero Number for non-numeric kinds. The
// zero Number is not valid (see Number): check Kind first.
func (v Value) Number() Number {
	if v.kind == KindInteger || v.kind == KindDecimal {
		return Number{Text: v.text}
	}
	return Number{}
}

// Str returns the string content, or "" for other kinds.
func (v Value) Str() string {
	if v.kind == KindString {
		return v.text
	}
	return ""
}

// List returns the list, or nil for other kinds.
func (v Value) List() List { return v.list }

// Map returns the map, or nil for other kinds. A nil *Map is safe to read
// and behaves as empty.
func (v Value) Map() *Map { return v.m }

// Equal reports whether two values have the same kind and content. Numbers
// compare by exact text, so 1 and 1.0 differ. Maps compare entries in order.
func (v Value) Equal(o Value) bool {
	if v.kind != o.kind {
		return false
	}
	switch v.kind {
	case KindNull:
		return true
	case KindBool:
		return v.b == o.b
	case KindList:
		return slices.EqualFunc(v.list, o.list, Value.Equal)
	case KindMap:
		return slices.EqualFunc(v.m.Entries(), o.m.Entries(), func(a, b Entry) bool {
			return a.Key == b.Key && a.Value.Equal(b.Value)
		})
	}
	return v.text == o.text // string and numbers
}

// String returns a compact JSON-like rendering for debugging and test
// failures. It is not a serializer: use the format packages to write data.
// Use Str for the content of a string value.
func (v Value) String() string {
	var sb strings.Builder
	v.debug(&sb)
	return sb.String()
}

func (v Value) debug(sb *strings.Builder) {
	switch v.kind {
	case KindNull:
		sb.WriteString("null")
	case KindBool:
		sb.WriteString(strconv.FormatBool(v.b))
	case KindInteger, KindDecimal:
		sb.WriteString(v.text)
	case KindString:
		sb.WriteString(strconv.Quote(v.text))
	case KindList:
		sb.WriteByte('[')
		for i, e := range v.list {
			if i > 0 {
				sb.WriteByte(',')
			}
			e.debug(sb)
		}
		sb.WriteByte(']')
	case KindMap:
		sb.WriteByte('{')
		for i := 0; i < v.m.Len(); i++ {
			if i > 0 {
				sb.WriteByte(',')
			}
			e := v.m.At(i)
			sb.WriteString(strconv.Quote(e.Key))
			sb.WriteByte(':')
			e.Value.debug(sb)
		}
		sb.WriteByte('}')
	}
}
