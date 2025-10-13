package validator

import (
	"sort"
	"sync"
	"testing"
)

// TestFunctionRegistry_Register tests registering a new function
func TestFunctionRegistry_Register(t *testing.T) {
	registry := NewFunctionRegistry()

	descriptor := FunctionDescriptor{
		Name:        "Custom",
		Description: "Custom validation function",
		MinArgs:     1,
		MaxArgs:     2,
	}

	err := registry.Register("Custom", descriptor)
	if err != nil {
		t.Errorf("Register() error = %v, want nil", err)
	}

	if !registry.Has("Custom") {
		t.Error("Has('Custom') = false, want true after registration")
	}
}

// TestFunctionRegistry_Register_Duplicate tests replacing existing function
func TestFunctionRegistry_Register_Duplicate(t *testing.T) {
	registry := NewFunctionRegistry()

	descriptor1 := FunctionDescriptor{
		Name:        "Validate",
		Description: "Original description",
		MinArgs:     1,
		MaxArgs:     1,
	}

	descriptor2 := FunctionDescriptor{
		Name:        "Validate",
		Description: "Updated description",
		MinArgs:     1,
		MaxArgs:     2,
	}

	registry.Register("Validate", descriptor1)
	registry.Register("Validate", descriptor2)

	desc, found := registry.Lookup("Validate")
	if !found {
		t.Fatal("Lookup('Validate') not found")
	}

	if desc.Description != "Updated description" {
		t.Errorf("Description = %q, want %q", desc.Description, "Updated description")
	}

	if desc.MaxArgs != 2 {
		t.Errorf("MaxArgs = %d, want 2", desc.MaxArgs)
	}
}

// TestFunctionRegistry_Lookup tests lookup operations
func TestFunctionRegistry_Lookup(t *testing.T) {
	registry := NewFunctionRegistry()

	descriptor := FunctionDescriptor{
		Name:        "String",
		Description: "String length validation",
		MinArgs:     1,
		MaxArgs:     2,
	}

	registry.Register("String", descriptor)

	// Lookup existing
	desc, found := registry.Lookup("String")
	if !found {
		t.Error("Lookup('String') found = false, want true")
	}

	if desc.Name != "String" {
		t.Errorf("Lookup returned Name = %q, want %q", desc.Name, "String")
	}

	// Lookup non-existent
	_, found = registry.Lookup("NonExistent")
	if found {
		t.Error("Lookup('NonExistent') found = true, want false")
	}
}

// TestFunctionRegistry_Has tests Has operation
func TestFunctionRegistry_Has(t *testing.T) {
	registry := NewFunctionRegistry()

	registry.Register("Integer", FunctionDescriptor{Name: "Integer"})

	if !registry.Has("Integer") {
		t.Error("Has('Integer') = false, want true")
	}

	if registry.Has("NonExistent") {
		t.Error("Has('NonExistent') = true, want false")
	}
}

// TestFunctionRegistry_List tests listing all functions
func TestFunctionRegistry_List(t *testing.T) {
	registry := NewFunctionRegistry()

	// Registry is pre-populated with 7 built-in functions
	list := registry.List()
	if len(list) < 7 {
		t.Errorf("List() returned %d functions, want at least 7 (built-ins)", len(list))
	}

	// Register additional custom functions
	customFunctions := []string{"CustomFunc1", "CustomFunc2", "CustomFunc3"}
	for _, name := range customFunctions {
		registry.Register(name, FunctionDescriptor{Name: name, MinArgs: 1, MaxArgs: 2})
	}

	// List should now include built-ins + custom functions
	list = registry.List()
	if len(list) < 7+len(customFunctions) {
		t.Errorf("List() returned %d functions, want at least %d", len(list), 7+len(customFunctions))
	}

	// Verify custom functions are in the list
	listMap := make(map[string]bool)
	for _, name := range list {
		listMap[name] = true
	}

	for _, name := range customFunctions {
		if !listMap[name] {
			t.Errorf("List() missing custom function %q", name)
		}
	}
}

// TestFunctionRegistry_List_Sorted tests that list is sorted
func TestFunctionRegistry_List_Sorted(t *testing.T) {
	registry := NewFunctionRegistry()

	// Register custom functions in random order
	registry.Register("Zebra", FunctionDescriptor{Name: "Zebra", MinArgs: 1, MaxArgs: 1})
	registry.Register("Apple", FunctionDescriptor{Name: "Apple", MinArgs: 1, MaxArgs: 1})
	registry.Register("Mango", FunctionDescriptor{Name: "Mango", MinArgs: 1, MaxArgs: 1})

	list := registry.List()
	// Verify list is sorted (includes built-ins + custom functions)
	if !sort.StringsAreSorted(list) {
		t.Errorf("List() is not sorted: %v", list)
	}

	// Verify our custom functions are present
	customFuncs := []string{"Apple", "Mango", "Zebra"}
	listMap := make(map[string]bool)
	for _, name := range list {
		listMap[name] = true
	}

	for _, name := range customFuncs {
		if !listMap[name] {
			t.Errorf("List() missing custom function %q", name)
		}
	}
}

// TestFunctionRegistry_List_AfterClear tests listing after clearing registry
func TestFunctionRegistry_List_AfterClear(t *testing.T) {
	registry := NewFunctionRegistry()

	// Clear the built-in functions
	registry.Clear()

	list := registry.List()
	if len(list) != 0 {
		t.Errorf("List() after Clear() returned %d items, want 0", len(list))
	}
}

// TestFunctionRegistry_Concurrent_Reads tests concurrent reads
func TestFunctionRegistry_Concurrent_Reads(t *testing.T) {
	registry := NewFunctionRegistry()

	// Pre-populate
	for i := 0; i < 10; i++ {
		name := "Func" + string(rune('A'+i))
		registry.Register(name, FunctionDescriptor{Name: name})
	}

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			name := "Func" + string(rune('A'+(n%10)))
			registry.Has(name)
			registry.Lookup(name)
			registry.List()
		}(i)
	}

	wg.Wait()
	// Test passes if no race detected
}

// TestFunctionRegistry_Concurrent_Writes tests concurrent writes
func TestFunctionRegistry_Concurrent_Writes(t *testing.T) {
	registry := NewFunctionRegistry()

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			name := "Func" + string(rune('A'+(n%26)))
			registry.Register(name, FunctionDescriptor{Name: name})
		}(i)
	}

	wg.Wait()

	list := registry.List()
	if len(list) == 0 {
		t.Error("After concurrent writes, registry should not be empty")
	}
}

// TestFunctionRegistry_Concurrent_ReadWrite tests concurrent reads and writes
func TestFunctionRegistry_Concurrent_ReadWrite(t *testing.T) {
	registry := NewFunctionRegistry()

	// Pre-populate
	for i := 0; i < 5; i++ {
		name := "Func" + string(rune('A'+i))
		registry.Register(name, FunctionDescriptor{Name: name})
	}

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		if i%2 == 0 {
			// Reader
			go func(n int) {
				defer wg.Done()
				name := "Func" + string(rune('A'+(n%10)))
				registry.Has(name)
				registry.Lookup(name)
				registry.List()
			}(i)
		} else {
			// Writer
			go func(n int) {
				defer wg.Done()
				name := "NewFunc" + string(rune('A'+(n%26)))
				registry.Register(name, FunctionDescriptor{Name: name})
			}(i)
		}
	}

	wg.Wait()
}

// TestFunctionRegistry_BuiltInFunctions tests built-in functions
func TestFunctionRegistry_BuiltInFunctions(t *testing.T) {
	registry := NewFunctionRegistry()

	expectedBuiltIns := []string{
		"String", "Integer", "Float", "Enum",
		"Pattern", "Length", "Range",
	}

	list := registry.List()
	if len(list) != len(expectedBuiltIns) {
		t.Errorf("NewFunctionRegistry() has %d built-in functions, want %d", len(list), len(expectedBuiltIns))
	}

	for _, name := range expectedBuiltIns {
		if !registry.Has(name) {
			t.Errorf("Built-in function %q not found", name)
		}

		desc, found := registry.Lookup(name)
		if !found {
			t.Errorf("Built-in function %q lookup failed", name)
		}

		if desc.Name != name {
			t.Errorf("Built-in function %q has wrong name: %q", name, desc.Name)
		}

		if desc.Description == "" {
			t.Errorf("Built-in function %q has empty description", name)
		}
	}
}

// TestFunctionRegistry_BuiltInFunctions_ArgCounts tests built-in function argument counts
func TestFunctionRegistry_BuiltInFunctions_ArgCounts(t *testing.T) {
	registry := NewFunctionRegistry()

	tests := []struct {
		name    string
		minArgs int
		maxArgs int
	}{
		{"String", 1, 2},
		{"Integer", 1, 2},
		{"Float", 1, 2},
		{"Enum", 1, -1}, // Unlimited
		{"Pattern", 1, 1},
		{"Length", 1, 2},
		{"Range", 1, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc, found := registry.Lookup(tt.name)
			if !found {
				t.Errorf("Built-in function %q not found", tt.name)
				return
			}

			if desc.MinArgs != tt.minArgs {
				t.Errorf("%s MinArgs = %d, want %d", tt.name, desc.MinArgs, tt.minArgs)
			}

			if desc.MaxArgs != tt.maxArgs {
				t.Errorf("%s MaxArgs = %d, want %d", tt.name, desc.MaxArgs, tt.maxArgs)
			}
		})
	}
}

// TestFunctionRegistry_BuiltInFunctions_ValidateArgs tests ValidateArgs for built-in functions
func TestFunctionRegistry_BuiltInFunctions_ValidateArgs(t *testing.T) {
	registry := NewFunctionRegistry()

	tests := []struct {
		name    string
		args    []interface{}
		wantErr bool
	}{
		// String function
		{"String", []interface{}{int64(1), int64(100)}, false},
		{"String", []interface{}{int64(1), "+"}, false},
		{"String", []interface{}{int64(100), int64(1)}, true}, // min > max
		{"String", []interface{}{"not a number"}, true},

		// Integer function
		{"Integer", []interface{}{int64(0), int64(120)}, false},
		{"Integer", []interface{}{int64(18), "+"}, false},
		{"Integer", []interface{}{int64(120), int64(18)}, true}, // min > max

		// Pattern function
		{"Pattern", []interface{}{"^[a-z]+$"}, false},
		{"Pattern", []interface{}{int64(123)}, true}, // not a string

		// Enum function (any values allowed)
		{"Enum", []interface{}{"a", "b", "c"}, false},
		{"Enum", []interface{}{int64(1), int64(2), int64(3)}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc, found := registry.Lookup(tt.name)
			if !found {
				t.Fatalf("Function %q not found", tt.name)
			}

			if desc.ValidateArgs == nil {
				// No custom validation
				return
			}

			err := desc.ValidateArgs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestFunctionDescriptor_MinMaxArgs tests MinArgs/MaxArgs validation
func TestFunctionDescriptor_MinMaxArgs(t *testing.T) {
	tests := []struct {
		name    string
		desc    FunctionDescriptor
		argLen  int
		wantErr bool
	}{
		{
			name:    "exact match",
			desc:    FunctionDescriptor{MinArgs: 2, MaxArgs: 2},
			argLen:  2,
			wantErr: false,
		},
		{
			name:    "within range",
			desc:    FunctionDescriptor{MinArgs: 1, MaxArgs: 3},
			argLen:  2,
			wantErr: false,
		},
		{
			name:    "too few args",
			desc:    FunctionDescriptor{MinArgs: 2, MaxArgs: 3},
			argLen:  1,
			wantErr: true,
		},
		{
			name:    "too many args",
			desc:    FunctionDescriptor{MinArgs: 1, MaxArgs: 2},
			argLen:  3,
			wantErr: true,
		},
		{
			name:    "unlimited args",
			desc:    FunctionDescriptor{MinArgs: 1, MaxArgs: -1},
			argLen:  100,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate argument count validation
			hasError := tt.argLen < tt.desc.MinArgs ||
				(tt.desc.MaxArgs >= 0 && tt.argLen > tt.desc.MaxArgs)

			if hasError != tt.wantErr {
				t.Errorf("Arg count validation error = %v, wantErr %v", hasError, tt.wantErr)
			}
		})
	}
}

// TestFunctionRegistry_Unregister tests unregistering a function
func TestFunctionRegistry_Unregister(t *testing.T) {
	registry := NewFunctionRegistry()

	registry.Register("Custom", FunctionDescriptor{Name: "Custom"})
	if !registry.Has("Custom") {
		t.Fatal("Custom function should be registered")
	}

	registry.Unregister("Custom")
	if registry.Has("Custom") {
		t.Error("Custom function should be unregistered")
	}
}

// TestFunctionRegistry_Unregister_NonExistent tests unregistering non-existent function
func TestFunctionRegistry_Unregister_NonExistent(t *testing.T) {
	registry := NewFunctionRegistry()

	// Should not panic
	registry.Unregister("NonExistent")
}

// TestFunctionRegistry_Clear tests clearing all functions
func TestFunctionRegistry_Clear(t *testing.T) {
	registry := NewFunctionRegistry()

	registry.Register("Func1", FunctionDescriptor{Name: "Func1"})
	registry.Register("Func2", FunctionDescriptor{Name: "Func2"})

	if len(registry.List()) == 0 {
		t.Fatal("Registry should not be empty before Clear()")
	}

	registry.Clear()

	list := registry.List()
	if len(list) != 0 {
		t.Errorf("After Clear(), registry should be empty, got %d functions", len(list))
	}
}
