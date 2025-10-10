package ast

import "sync"

// intern provides string interning for commonly used type names and function names.
// This reduces memory allocations by reusing the same string instances.
var intern = newStringInterner()

// stringInterner manages interned strings for type and function names.
type stringInterner struct {
	mu      sync.RWMutex
	strings map[string]string
}

// newStringInterner creates a new string interner with common type and function names pre-populated.
func newStringInterner() *stringInterner {
	si := &stringInterner{
		strings: make(map[string]string, 32),
	}

	// Pre-populate with common type names
	commonTypes := []string{
		"UUID", "Email", "URL",
		"String", "Integer", "Float", "Boolean",
		"ISO-8601", "Date", "Time", "DateTime",
		"IPv4", "IPv6",
		"JSON", "Base64",
	}

	// Pre-populate with common function names
	commonFunctions := []string{
		"String", "Integer", "Float", "Boolean",
		"Enum", "Pattern", "Length", "Range",
	}

	// Intern all common names
	for _, s := range commonTypes {
		si.strings[s] = s
	}
	for _, s := range commonFunctions {
		si.strings[s] = s
	}

	return si
}

// Get returns an interned version of the string.
// If the string is already interned, returns the existing instance.
// Otherwise, interns the new string and returns it.
func (si *stringInterner) Get(s string) string {
	// Fast path: read lock for existing strings
	si.mu.RLock()
	if interned, ok := si.strings[s]; ok {
		si.mu.RUnlock()
		return interned
	}
	si.mu.RUnlock()

	// Slow path: write lock to add new string
	si.mu.Lock()
	defer si.mu.Unlock()

	// Double-check in case another goroutine added it
	if interned, ok := si.strings[s]; ok {
		return interned
	}

	// Intern the new string
	si.strings[s] = s
	return s
}

// InternString returns an interned version of the string.
// For type and function names, this reduces memory usage by reusing string instances.
func InternString(s string) string {
	return intern.Get(s)
}
