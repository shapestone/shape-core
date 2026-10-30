package data

import "iter"

// indexThreshold is the entry count at which a Map builds its lookup index.
// Below it a linear scan is faster and allocates nothing.
const indexThreshold = 16

// Entry is one key/value pair of a Map.
type Entry struct {
	Key   string
	Value Value
}

// Map is an ordered map from string keys to values. Entries keep insertion
// order. Duplicate-key policy belongs to each format, so Map offers
// primitives to apply one: Has and Index to check, Append to keep
// duplicates, Add for first-wins and Set for last-wins.
//
// When duplicates were appended, lookups (Get, Index, Has) find the first
// entry with the key. A Map is not safe for concurrent use. The zero Map is
// empty and ready to use.
type Map struct {
	entries []Entry
	index   map[string]int // key -> position of its first entry; nil until large
}

// NewMap returns an empty map with room for capacity entries.
func NewMap(capacity int) *Map {
	return &Map{entries: make([]Entry, 0, capacity)}
}

// Len returns the number of entries, counting appended duplicates.
func (m *Map) Len() int { return len(m.entries) }

// Entries returns the entries in order. The slice is shared with the map:
// callers may change values but must not change keys or length.
func (m *Map) Entries() []Entry { return m.entries }

// All iterates over the entries in order.
func (m *Map) All() iter.Seq2[string, Value] {
	return func(yield func(string, Value) bool) {
		for _, e := range m.entries {
			if !yield(e.Key, e.Value) {
				return
			}
		}
	}
}

// Index returns the position of the first entry with key, or -1.
func (m *Map) Index(key string) int {
	if m.index != nil {
		if i, ok := m.index[key]; ok {
			return i
		}
		return -1
	}
	for i := range m.entries {
		if m.entries[i].Key == key {
			return i
		}
	}
	return -1
}

// Has reports whether an entry with key exists.
func (m *Map) Has(key string) bool { return m.Index(key) >= 0 }

// Get returns the value of the first entry with key.
func (m *Map) Get(key string) (Value, bool) {
	if i := m.Index(key); i >= 0 {
		return m.entries[i].Value, true
	}
	return Value{}, false
}

// Append adds an entry at the end without checking for duplicates.
func (m *Map) Append(key string, v Value) {
	m.entries = append(m.entries, Entry{Key: key, Value: v})
	switch {
	case m.index != nil:
		if _, ok := m.index[key]; !ok {
			m.index[key] = len(m.entries) - 1
		}
	case len(m.entries) >= indexThreshold:
		m.buildIndex()
	}
}

// Add appends the entry only if key is absent (first wins) and reports
// whether it was added.
func (m *Map) Add(key string, v Value) bool {
	if m.Has(key) {
		return false
	}
	m.Append(key, v)
	return true
}

// Set replaces the value of an existing key in place, keeping its original
// position (last wins), or appends the entry. It reports whether an existing
// entry was replaced.
func (m *Map) Set(key string, v Value) bool {
	if i := m.Index(key); i >= 0 {
		m.entries[i].Value = v
		return true
	}
	m.Append(key, v)
	return false
}

func (m *Map) buildIndex() {
	m.index = make(map[string]int, len(m.entries)*2)
	for i := range m.entries {
		if _, ok := m.index[m.entries[i].Key]; !ok {
			m.index[m.entries[i].Key] = i
		}
	}
}
