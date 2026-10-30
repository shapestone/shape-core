package data

import (
	"iter"
	"slices"
)

// indexThreshold is the entry count at which a Map builds its lookup index.
// Below it a linear scan is faster and allocates nothing.
const indexThreshold = 16

// Entry is one key/value pair of a Map.
type Entry struct {
	Key   string
	Value Value
}

// maxPrealloc caps the capacity NewMap reserves, so a hostile declared
// length cannot force a large uncharged allocation.
const maxPrealloc = 1024

// Map is an ordered map from string keys to values. Entries keep insertion
// order. Duplicate-key policy belongs to each format, so Map offers
// primitives to apply one: Has and Index to check, Append to keep
// duplicates, Add for first-wins and Set for last-wins.
//
// When duplicates were appended, lookups (Get, Index, Has) and Set act on
// the first entry with the key. The read methods (Len, Entries, At, All,
// Index, Has, Get) are safe on a nil *Map and see it as empty, so
// Value.Map() can be used on any value. A Map is not safe for concurrent
// use. The zero Map is empty and ready to use.
type Map struct {
	entries []Entry
	index   map[string]int // key -> position of its first entry; nil until large
}

// NewMap returns an empty map with room for capacity entries. Negative
// capacity is treated as 0, and the reservation is silently capped at 1024
// entries because the size may come from untrusted input; the map still
// grows as needed. Use NewMapTrusted when the size is known to be sound.
// The reservation is not charged to a Guard.
func NewMap(capacity int) *Map {
	return NewMapTrusted(min(capacity, maxPrealloc))
}

// NewMapTrusted is NewMap without the 1024-entry cap, for builders that know
// the real size (for example from an already-parsed row count). Negative
// capacity is treated as 0. Never pass a size read from untrusted input
// without checking it against a Guard first.
func NewMapTrusted(capacity int) *Map {
	return &Map{entries: make([]Entry, 0, max(0, capacity))}
}

// Len returns the number of entries, counting appended duplicates.
func (m *Map) Len() int {
	if m == nil {
		return 0
	}
	return len(m.entries)
}

// Entries returns a copy of the entries in order. Because it is a copy,
// changing it cannot corrupt the key index, but every call allocates: do not
// call it inside a loop. Use Len with At, or All, to read without copying.
func (m *Map) Entries() []Entry {
	if m == nil {
		return nil
	}
	return slices.Clone(m.entries)
}

// At returns the i-th entry in insertion order. It panics if i is out of
// range, like a slice index.
func (m *Map) At(i int) Entry { return m.entries[i] }

// All iterates over the entries in order.
func (m *Map) All() iter.Seq2[string, Value] {
	return func(yield func(string, Value) bool) {
		if m == nil {
			return
		}
		for _, e := range m.entries {
			if !yield(e.Key, e.Value) {
				return
			}
		}
	}
}

// Index returns the position of the first entry with key, or -1.
func (m *Map) Index(key string) int {
	if m == nil {
		return -1
	}
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

// Set replaces the value of the existing entry for key in place, keeping its
// position, or appends the entry (last value wins for keys added with Set
// or Add). If duplicates were appended, only the first entry is replaced.
// It reports whether an existing entry was replaced.
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
