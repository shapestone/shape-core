package validator

import (
	"sort"
	"sync"
	"testing"
)

// TestTypeRegistry_Register tests registering a new type
func TestTypeRegistry_Register(t *testing.T) {
	registry := NewTypeRegistry()

	// Register a new type
	descriptor := TypeDescriptor{
		Name:        "SSN",
		Description: "Social Security Number",
	}

	err := registry.Register("SSN", descriptor)
	if err != nil {
		t.Errorf("Register() error = %v, want nil", err)
	}

	// Verify it was registered
	if !registry.Has("SSN") {
		t.Error("Has('SSN') = false, want true after registration")
	}
}

// TestTypeRegistry_Register_Duplicate tests registering duplicate type (should replace)
func TestTypeRegistry_Register_Duplicate(t *testing.T) {
	registry := NewTypeRegistry()

	descriptor1 := TypeDescriptor{
		Name:        "Custom",
		Description: "Original description",
	}

	descriptor2 := TypeDescriptor{
		Name:        "Custom",
		Description: "Updated description",
	}

	// Register first time
	err := registry.Register("Custom", descriptor1)
	if err != nil {
		t.Errorf("First Register() error = %v, want nil", err)
	}

	// Register again (should replace)
	err = registry.Register("Custom", descriptor2)
	if err != nil {
		t.Errorf("Second Register() error = %v, want nil", err)
	}

	// Verify it has the updated description
	desc, found := registry.Lookup("Custom")
	if !found {
		t.Fatal("Lookup('Custom') not found")
	}

	if desc.Description != "Updated description" {
		t.Errorf("Description = %q, want %q", desc.Description, "Updated description")
	}
}

// TestTypeRegistry_Lookup_Existing tests lookup of existing type
func TestTypeRegistry_Lookup_Existing(t *testing.T) {
	registry := NewTypeRegistry()

	descriptor := TypeDescriptor{
		Name:        "Email",
		Description: "Email address format",
	}

	registry.Register("Email", descriptor)

	// Lookup existing type
	desc, found := registry.Lookup("Email")
	if !found {
		t.Error("Lookup('Email') found = false, want true")
	}

	if desc.Name != "Email" {
		t.Errorf("Lookup returned Name = %q, want %q", desc.Name, "Email")
	}
}

// TestTypeRegistry_Lookup_NonExistent tests lookup of non-existent type
func TestTypeRegistry_Lookup_NonExistent(t *testing.T) {
	registry := NewTypeRegistry()

	// Lookup non-existent type
	_, found := registry.Lookup("UnknownType")
	if found {
		t.Error("Lookup('UnknownType') found = true, want false")
	}
}

// TestTypeRegistry_Has_Existing tests Has for existing type
func TestTypeRegistry_Has_Existing(t *testing.T) {
	registry := NewTypeRegistry()
	registry.Register("UUID", TypeDescriptor{Name: "UUID"})

	if !registry.Has("UUID") {
		t.Error("Has('UUID') = false, want true")
	}
}

// TestTypeRegistry_Has_NonExistent tests Has for non-existent type
func TestTypeRegistry_Has_NonExistent(t *testing.T) {
	registry := NewTypeRegistry()

	if registry.Has("NonExistent") {
		t.Error("Has('NonExistent') = true, want false")
	}
}

// TestTypeRegistry_List tests listing all types
func TestTypeRegistry_List(t *testing.T) {
	registry := NewTypeRegistry()

	// Registry is pre-populated with 15 built-in types
	list := registry.List()
	if len(list) < 15 {
		t.Errorf("List() returned %d types, want at least 15 (built-ins)", len(list))
	}

	// Register additional custom types
	customTypes := []string{"CustomType1", "CustomType2", "CustomType3"}
	for _, typeName := range customTypes {
		registry.Register(typeName, TypeDescriptor{Name: typeName})
	}

	// List should now include built-ins + custom types
	list = registry.List()
	if len(list) < 15+len(customTypes) {
		t.Errorf("List() returned %d types, want at least %d", len(list), 15+len(customTypes))
	}

	// Verify custom types are in the list
	listMap := make(map[string]bool)
	for _, name := range list {
		listMap[name] = true
	}

	for _, typeName := range customTypes {
		if !listMap[typeName] {
			t.Errorf("List() missing custom type %q", typeName)
		}
	}
}

// TestTypeRegistry_List_Sorted tests that list is sorted alphabetically
func TestTypeRegistry_List_Sorted(t *testing.T) {
	registry := NewTypeRegistry()

	// Register types in random order
	registry.Register("Zebra", TypeDescriptor{Name: "Zebra"})
	registry.Register("Apple", TypeDescriptor{Name: "Apple"})
	registry.Register("Mango", TypeDescriptor{Name: "Mango"})
	registry.Register("Banana", TypeDescriptor{Name: "Banana"})

	list := registry.List()

	// Verify list is sorted (includes built-ins + custom types)
	if !sort.StringsAreSorted(list) {
		t.Errorf("List() is not sorted: %v", list)
	}

	// Verify our custom types are present and sorted
	customTypes := []string{"Apple", "Banana", "Mango", "Zebra"}
	listMap := make(map[string]bool)
	for _, name := range list {
		listMap[name] = true
	}

	for _, typeName := range customTypes {
		if !listMap[typeName] {
			t.Errorf("List() missing custom type %q", typeName)
		}
	}
}

// TestTypeRegistry_List_AfterClear tests listing after clearing registry
func TestTypeRegistry_List_AfterClear(t *testing.T) {
	registry := NewTypeRegistry()

	// Clear the built-in types
	registry.Clear()

	list := registry.List()
	if len(list) != 0 {
		t.Errorf("List() after Clear() returned %d types, want 0", len(list))
	}
}

// TestTypeRegistry_Concurrent_Reads tests concurrent reads (should have no races)
func TestTypeRegistry_Concurrent_Reads(t *testing.T) {
	registry := NewTypeRegistry()

	// Pre-populate registry
	for i := 0; i < 10; i++ {
		registry.Register("Type"+string(rune('A'+i)), TypeDescriptor{Name: "Type" + string(rune('A'+i))})
	}

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			typeName := "Type" + string(rune('A'+(n%10)))
			registry.Has(typeName)
			registry.Lookup(typeName)
			registry.List()
		}(i)
	}

	wg.Wait()
	// Test passes if no race detected (run with go test -race)
}

// TestTypeRegistry_Concurrent_Writes tests concurrent writes (should have no races)
func TestTypeRegistry_Concurrent_Writes(t *testing.T) {
	registry := NewTypeRegistry()

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			typeName := "Type" + string(rune('A'+(n%26)))
			registry.Register(typeName, TypeDescriptor{Name: typeName})
		}(i)
	}

	wg.Wait()

	// Verify at least some types were registered
	list := registry.List()
	if len(list) == 0 {
		t.Error("After concurrent writes, registry should not be empty")
	}
	// Test passes if no race detected (run with go test -race)
}

// TestTypeRegistry_Concurrent_ReadWrite tests concurrent reads and writes
func TestTypeRegistry_Concurrent_ReadWrite(t *testing.T) {
	registry := NewTypeRegistry()

	// Pre-populate
	for i := 0; i < 5; i++ {
		registry.Register("Type"+string(rune('A'+i)), TypeDescriptor{Name: "Type" + string(rune('A'+i))})
	}

	var wg sync.WaitGroup
	numGoroutines := 100

	// Half readers, half writers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		if i%2 == 0 {
			// Reader
			go func(n int) {
				defer wg.Done()
				typeName := "Type" + string(rune('A'+(n%10)))
				registry.Has(typeName)
				registry.Lookup(typeName)
				registry.List()
			}(i)
		} else {
			// Writer
			go func(n int) {
				defer wg.Done()
				typeName := "NewType" + string(rune('A'+(n%26)))
				registry.Register(typeName, TypeDescriptor{Name: typeName})
			}(i)
		}
	}

	wg.Wait()
	// Test passes if no race detected (run with go test -race)
}

// TestTypeRegistry_BuiltInTypes tests that NewTypeRegistry has 15 built-in types
func TestTypeRegistry_BuiltInTypes(t *testing.T) {
	registry := NewTypeRegistry()

	expectedBuiltIns := []string{
		"UUID", "Email", "String", "Integer", "Float",
		"Boolean", "ISO-8601", "Date", "Time", "DateTime",
		"IPv4", "IPv6", "JSON", "Base64", "URL",
	}

	list := registry.List()
	if len(list) != len(expectedBuiltIns) {
		t.Errorf("NewTypeRegistry() has %d built-in types, want %d", len(list), len(expectedBuiltIns))
	}

	// Verify all expected built-ins exist
	for _, typeName := range expectedBuiltIns {
		if !registry.Has(typeName) {
			t.Errorf("Built-in type %q not found in registry", typeName)
		}

		desc, found := registry.Lookup(typeName)
		if !found {
			t.Errorf("Built-in type %q lookup failed", typeName)
		}

		if desc.Name != typeName {
			t.Errorf("Built-in type %q descriptor has wrong name: %q", typeName, desc.Name)
		}

		if desc.Description == "" {
			t.Errorf("Built-in type %q has empty description", typeName)
		}
	}
}

// TestTypeRegistry_BuiltInTypes_Details tests built-in type details
func TestTypeRegistry_BuiltInTypes_Details(t *testing.T) {
	registry := NewTypeRegistry()

	tests := []struct {
		typeName string
		wantDesc bool // Should have description
	}{
		{"UUID", true},
		{"Email", true},
		{"String", true},
		{"Integer", true},
		{"Float", true},
		{"Boolean", true},
		{"ISO-8601", true},
		{"Date", true},
		{"Time", true},
		{"DateTime", true},
		{"IPv4", true},
		{"IPv6", true},
		{"JSON", true},
		{"Base64", true},
		{"URL", true},
	}

	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			desc, found := registry.Lookup(tt.typeName)
			if !found {
				t.Errorf("Built-in type %q not found", tt.typeName)
				return
			}

			if tt.wantDesc && desc.Description == "" {
				t.Errorf("Built-in type %q should have description", tt.typeName)
			}
		})
	}
}

// TestTypeRegistry_Unregister tests unregistering a type
func TestTypeRegistry_Unregister(t *testing.T) {
	registry := NewTypeRegistry()

	// Register and unregister
	registry.Register("Custom", TypeDescriptor{Name: "Custom"})
	if !registry.Has("Custom") {
		t.Fatal("Custom type should be registered")
	}

	registry.Unregister("Custom")
	if registry.Has("Custom") {
		t.Error("Custom type should be unregistered")
	}
}

// TestTypeRegistry_Unregister_NonExistent tests unregistering non-existent type (should not panic)
func TestTypeRegistry_Unregister_NonExistent(t *testing.T) {
	registry := NewTypeRegistry()

	// Should not panic
	registry.Unregister("NonExistent")
}

// TestTypeRegistry_Clear tests clearing all types
func TestTypeRegistry_Clear(t *testing.T) {
	registry := NewTypeRegistry()

	// Register some types
	registry.Register("Type1", TypeDescriptor{Name: "Type1"})
	registry.Register("Type2", TypeDescriptor{Name: "Type2"})

	if len(registry.List()) == 0 {
		t.Fatal("Registry should not be empty before Clear()")
	}

	registry.Clear()

	list := registry.List()
	if len(list) != 0 {
		t.Errorf("After Clear(), registry should be empty, got %d types", len(list))
	}
}
