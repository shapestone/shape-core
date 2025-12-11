package validator

import (
	"fmt"
	"sort"
	"sync"

	"github.com/shapestone/shape/pkg/ast"
)

// FunctionDescriptor describes a validation function.
type FunctionDescriptor struct {
	Name         string
	Description  string
	MinArgs      int                       // Minimum number of arguments
	MaxArgs      int                       // Maximum number of arguments (-1 = unlimited)
	ValidateArgs func([]interface{}) error // Optional custom argument validation
}

// FunctionRegistry is a thread-safe registry of validation functions.
// It uses a RWMutex for efficient concurrent reads.
type FunctionRegistry struct {
	functions map[string]*FunctionDescriptor
	mu        sync.RWMutex
}

// NewFunctionRegistry creates a new function registry pre-populated with 7 built-in functions.
func NewFunctionRegistry() *FunctionRegistry {
	registry := &FunctionRegistry{
		functions: make(map[string]*FunctionDescriptor),
	}

	// Pre-populate with 7 built-in functions
	builtInFunctions := []FunctionDescriptor{
		{
			Name:         "String",
			Description:  "String length validation (min, max)",
			MinArgs:      1,
			MaxArgs:      2,
			ValidateArgs: validateRangeArgs,
		},
		{
			Name:         "Integer",
			Description:  "Integer range validation (min, max)",
			MinArgs:      1,
			MaxArgs:      2,
			ValidateArgs: validateRangeArgs,
		},
		{
			Name:         "Float",
			Description:  "Float range validation (min, max)",
			MinArgs:      1,
			MaxArgs:      2,
			ValidateArgs: validateRangeArgs,
		},
		{
			Name:        "Enum",
			Description: "Enumeration of allowed values",
			MinArgs:     1,
			MaxArgs:     -1, // Unlimited
		},
		{
			Name:        "Pattern",
			Description: "Regular expression pattern matching",
			MinArgs:     1,
			MaxArgs:     1,
			ValidateArgs: func(args []interface{}) error {
				if _, ok := args[0].(string); !ok {
					return fmt.Errorf("argument must be a string pattern")
				}
				return nil
			},
		},
		{
			Name:         "Length",
			Description:  "Length validation (min, max)",
			MinArgs:      1,
			MaxArgs:      2,
			ValidateArgs: validateRangeArgs,
		},
		{
			Name:         "Range",
			Description:  "Range validation (min, max)",
			MinArgs:      1,
			MaxArgs:      2,
			ValidateArgs: validateRangeArgs,
		},
	}

	for _, desc := range builtInFunctions {
		// Use string interning for memory efficiency
		name := ast.InternString(desc.Name)
		registry.functions[name] = &FunctionDescriptor{
			Name:         name,
			Description:  desc.Description,
			MinArgs:      desc.MinArgs,
			MaxArgs:      desc.MaxArgs,
			ValidateArgs: desc.ValidateArgs,
		}
	}

	return registry
}

// Register registers a new function or replaces an existing one.
// Returns nil (for compatibility with error-returning patterns).
func (r *FunctionRegistry) Register(name string, desc FunctionDescriptor) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Use string interning for memory efficiency
	name = ast.InternString(name)
	desc.Name = name

	r.functions[name] = &desc
	return nil
}

// Lookup looks up a function by name.
// Returns the descriptor and true if found, nil and false otherwise.
func (r *FunctionRegistry) Lookup(name string) (*FunctionDescriptor, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	desc, ok := r.functions[name]
	if !ok {
		return nil, false
	}

	// Return a copy to prevent external modification (except ValidateArgs which is a function pointer)
	return &FunctionDescriptor{
		Name:         desc.Name,
		Description:  desc.Description,
		MinArgs:      desc.MinArgs,
		MaxArgs:      desc.MaxArgs,
		ValidateArgs: desc.ValidateArgs,
	}, true
}

// Has checks if a function is registered.
func (r *FunctionRegistry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.functions[name]
	return ok
}

// List returns a sorted list of all registered function names.
func (r *FunctionRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.functions))
	for name := range r.functions {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

// Size returns the number of registered functions.
func (r *FunctionRegistry) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.functions)
}

// Unregister removes a function from the registry.
func (r *FunctionRegistry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.functions, name)
}

// Clear removes all functions from the registry.
func (r *FunctionRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.functions = make(map[string]*FunctionDescriptor)
}

// validateRangeArgs validates range arguments (min, max) or (min+).
// Reuses the logic from validator.go for consistency.
func validateRangeArgs(args []interface{}) error {
	if len(args) == 1 {
		// Single argument must be a number
		switch args[0].(type) {
		case int64, float64:
			return nil
		default:
			return fmt.Errorf("argument must be a number")
		}
	}

	if len(args) == 2 {
		// Two arguments: (min, max) or (min, "+")
		var min, max int64
		var unbounded bool

		// First argument must be a number
		switch v := args[0].(type) {
		case int64:
			min = v
		case float64:
			min = int64(v)
		default:
			return fmt.Errorf("first argument must be a number")
		}

		// Second argument: number or "+"
		switch v := args[1].(type) {
		case int64:
			max = v
		case float64:
			max = int64(v)
		case string:
			if v != "+" {
				return fmt.Errorf("second argument must be a number or '+'")
			}
			unbounded = true
		default:
			return fmt.Errorf("second argument must be a number or '+'")
		}

		// If bounded, check min <= max
		if !unbounded && min > max {
			return fmt.Errorf("min (%d) must be less than or equal to max (%d)", min, max)
		}
	}

	return nil
}
