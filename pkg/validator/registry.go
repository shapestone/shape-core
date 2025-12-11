package validator

import (
	"sort"
	"sync"

	"github.com/shapestone/shape/pkg/ast"
)

// TypeDescriptor describes a validation type.
type TypeDescriptor struct {
	Name        string
	Description string
}

// TypeRegistry is a thread-safe registry of validation types.
// It uses a RWMutex for efficient concurrent reads.
type TypeRegistry struct {
	types map[string]*TypeDescriptor
	mu    sync.RWMutex
}

// NewTypeRegistry creates a new type registry pre-populated with 15 built-in types.
func NewTypeRegistry() *TypeRegistry {
	registry := &TypeRegistry{
		types: make(map[string]*TypeDescriptor),
	}

	// Pre-populate with 15 built-in types
	builtInTypes := []TypeDescriptor{
		{Name: "UUID", Description: "Universally Unique Identifier"},
		{Name: "Email", Description: "Email address format"},
		{Name: "String", Description: "String data type"},
		{Name: "Integer", Description: "Integer number"},
		{Name: "Float", Description: "Floating-point number"},
		{Name: "Boolean", Description: "Boolean true/false value"},
		{Name: "ISO-8601", Description: "ISO-8601 date/time format"},
		{Name: "Date", Description: "Date without time"},
		{Name: "Time", Description: "Time without date"},
		{Name: "DateTime", Description: "Combined date and time"},
		{Name: "IPv4", Description: "IPv4 address"},
		{Name: "IPv6", Description: "IPv6 address"},
		{Name: "JSON", Description: "JSON data structure"},
		{Name: "Base64", Description: "Base64 encoded data"},
		{Name: "URL", Description: "URL/URI format"},
	}

	for _, desc := range builtInTypes {
		// Use string interning for memory efficiency
		name := ast.InternString(desc.Name)
		registry.types[name] = &TypeDescriptor{
			Name:        name,
			Description: desc.Description,
		}
	}

	return registry
}

// Register registers a new type or replaces an existing one.
// Returns nil (for compatibility with error-returning patterns).
func (r *TypeRegistry) Register(name string, desc TypeDescriptor) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Use string interning for memory efficiency
	name = ast.InternString(name)
	desc.Name = name

	r.types[name] = &desc
	return nil
}

// Lookup looks up a type by name.
// Returns the descriptor and true if found, nil and false otherwise.
func (r *TypeRegistry) Lookup(name string) (*TypeDescriptor, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	desc, ok := r.types[name]
	if !ok {
		return nil, false
	}

	// Return a copy to prevent external modification
	return &TypeDescriptor{
		Name:        desc.Name,
		Description: desc.Description,
	}, true
}

// Has checks if a type is registered.
func (r *TypeRegistry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.types[name]
	return ok
}

// List returns a sorted list of all registered type names.
func (r *TypeRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.types))
	for name := range r.types {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

// Size returns the number of registered types.
func (r *TypeRegistry) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.types)
}

// Unregister removes a type from the registry.
func (r *TypeRegistry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.types, name)
}

// Clear removes all types from the registry.
func (r *TypeRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.types = make(map[string]*TypeDescriptor)
}
