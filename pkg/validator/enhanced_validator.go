package validator

import (
	"fmt"
	"strings"

	"github.com/shapestone/shape-core/pkg/ast"
)

// SchemaValidator validates schema ASTs for semantic correctness.
// It collects ALL errors found during validation (not just the first one).
// It tracks JSONPath for better error messages.
type SchemaValidator struct {
	typeRegistry     *TypeRegistry
	functionRegistry *FunctionRegistry
	currentPath      []string // Track JSONPath during traversal
	result           *ValidationResult
	sourceText       string // Original schema text for source context display
}

// NewSchemaValidator creates a new schema validator with built-in types and functions.
func NewSchemaValidator() *SchemaValidator {
	return &SchemaValidator{
		typeRegistry:     NewTypeRegistry(),
		functionRegistry: NewFunctionRegistry(),
		currentPath:      []string{},
		result:           NewValidationResult(),
	}
}

// ValidateAll validates a schema AST and returns all errors found.
// Unlike the old Validate() method, this collects ALL errors, not just the first one.
// The sourceText parameter is optional but recommended for better error messages with source context.
func (v *SchemaValidator) ValidateAll(node ast.SchemaNode, sourceText ...string) *ValidationResult {
	// Reset state for new validation
	v.result = NewValidationResult()
	v.currentPath = []string{}
	v.sourceText = ""

	// Store source text if provided (optional parameter)
	if len(sourceText) > 0 {
		v.sourceText = sourceText[0]
	}

	// Traverse the AST and collect all errors
	// nolint:errcheck // Error is intentionally ignored as errors are collected in v.result
	node.Accept(v)

	// Add source context to errors if source text was provided
	if v.sourceText != "" {
		v.addSourceContext()
	}

	return v.result
}

// addSourceContext adds source lines to each error for context display.
func (v *SchemaValidator) addSourceContext() {
	lines := strings.Split(v.sourceText, "\n")

	for i := range v.result.Errors {
		err := &v.result.Errors[i]
		if err.Position.Line > 0 && err.Position.Line <= len(lines) {
			// Extract 2 lines before, the error line, and 2 lines after
			startLine := maxInt(1, err.Position.Line-2)
			endLine := minInt(len(lines), err.Position.Line+2)

			err.SourceLines = make([]string, 0, endLine-startLine+1)
			for line := startLine; line <= endLine; line++ {
				err.SourceLines = append(err.SourceLines, lines[line-1]) // lines are 0-indexed
			}
			err.Source = v.sourceText
		}
	}
}

// RegisterType registers a custom type.
// Returns the validator for method chaining.
func (v *SchemaValidator) RegisterType(name string, desc TypeDescriptor) *SchemaValidator {
	// nolint:errcheck // Registry.Register always returns nil
	v.typeRegistry.Register(name, desc)
	return v
}

// RegisterFunction registers a custom function.
// Returns the validator for method chaining.
func (v *SchemaValidator) RegisterFunction(name string, desc FunctionDescriptor) *SchemaValidator {
	// nolint:errcheck // Registry.Register always returns nil
	v.functionRegistry.Register(name, desc)
	return v
}

// VisitLiteral validates a literal node (always valid).
func (v *SchemaValidator) VisitLiteral(node *ast.LiteralNode) error {
	// Literals are always valid
	return nil
}

// VisitType validates a type node.
func (v *SchemaValidator) VisitType(node *ast.TypeNode) error {
	typeName := node.TypeName()

	if !v.typeRegistry.Has(typeName) {
		v.result.AddError(ValidationError{
			Position: node.Position(),
			Path:     v.currentJSONPath(),
			Code:     ErrCodeUnknownType,
			Message:  fmt.Sprintf("unknown type: %s", typeName),
			Hint:     v.generateTypeHint(typeName),
		})
	}

	// Don't return error - continue collecting all errors
	return nil
}

// VisitFunction validates a function node.
func (v *SchemaValidator) VisitFunction(node *ast.FunctionNode) error {
	name := node.Name()
	args := node.Arguments()

	// Check if function is known
	desc, ok := v.functionRegistry.Lookup(name)
	if !ok {
		v.result.AddError(ValidationError{
			Position: node.Position(),
			Path:     v.currentJSONPath(),
			Code:     ErrCodeUnknownFunction,
			Message:  fmt.Sprintf("unknown function: %s", name),
			Hint:     v.generateFunctionHint(name),
		})
		// Don't return - continue collecting errors
		return nil
	}

	// Check argument count
	argCount := len(args)
	if argCount < desc.MinArgs {
		v.result.AddError(ValidationError{
			Position: node.Position(),
			Path:     v.currentJSONPath(),
			Code:     ErrCodeInvalidArgCount,
			Message:  fmt.Sprintf("%s requires at least %d arguments, got %d", name, desc.MinArgs, argCount),
			Hint:     fmt.Sprintf("Expected %s to have between %d and %s arguments", name, desc.MinArgs, v.formatMaxArgs(desc.MaxArgs)),
		})
		return nil
	}

	if desc.MaxArgs >= 0 && argCount > desc.MaxArgs {
		v.result.AddError(ValidationError{
			Position: node.Position(),
			Path:     v.currentJSONPath(),
			Code:     ErrCodeInvalidArgCount,
			Message:  fmt.Sprintf("%s accepts at most %d arguments, got %d", name, desc.MaxArgs, argCount),
			Hint:     fmt.Sprintf("Expected %s to have between %d and %d arguments", name, desc.MinArgs, desc.MaxArgs),
		})
		return nil
	}

	// Custom argument validation
	if desc.ValidateArgs != nil {
		if err := desc.ValidateArgs(args); err != nil {
			v.result.AddError(ValidationError{
				Position: node.Position(),
				Path:     v.currentJSONPath(),
				Code:     ErrCodeInvalidArgValue,
				Message:  fmt.Sprintf("%s: %v", name, err),
				Hint:     "Check the argument types and values",
			})
		}
	}

	return nil
}

// VisitObject validates an object node.
func (v *SchemaValidator) VisitObject(node *ast.ObjectNode) error {
	for key, prop := range node.Properties() {
		// Push property name onto path
		v.currentPath = append(v.currentPath, key)

		// Validate the property (errors are collected, not returned)
		// nolint:errcheck // Error is intentionally ignored as errors are collected in v.result
		prop.Accept(v)

		// Pop property name from path
		v.currentPath = v.currentPath[:len(v.currentPath)-1]
	}

	return nil
}

// VisitArray validates an array node.
func (v *SchemaValidator) VisitArray(node *ast.ArrayNode) error {
	// Push array indicator onto path
	v.currentPath = append(v.currentPath, "[]")

	// Validate the element schema (errors are collected, not returned)
	// nolint:errcheck // Error is intentionally ignored as errors are collected in v.result
	node.ElementSchema().Accept(v)

	// Pop array indicator from path
	v.currentPath = v.currentPath[:len(v.currentPath)-1]

	return nil
}

// currentJSONPath returns the current JSONPath as a string.
func (v *SchemaValidator) currentJSONPath() string {
	if len(v.currentPath) == 0 {
		return "$"
	}

	path := "$"
	for _, segment := range v.currentPath {
		if segment == "[]" {
			path += "[]"
		} else {
			path += "." + segment
		}
	}
	return path
}

// generateTypeHint generates a helpful hint for unknown types.
func (v *SchemaValidator) generateTypeHint(typeName string) string {
	availableTypes := v.typeRegistry.List()

	// Find closest match using simple string similarity
	closest := findClosestMatch(typeName, availableTypes, 3)

	if closest != "" {
		return fmt.Sprintf("Did you mean '%s'?", closest)
	}

	// Show available types
	if len(availableTypes) <= 10 {
		return fmt.Sprintf("Available types: %s", strings.Join(availableTypes, ", "))
	}

	return fmt.Sprintf("Available types include: %s, ... (%d total)", strings.Join(availableTypes[:10], ", "), len(availableTypes))
}

// generateFunctionHint generates a helpful hint for unknown functions.
func (v *SchemaValidator) generateFunctionHint(funcName string) string {
	availableFunctions := v.functionRegistry.List()

	// Find closest match
	closest := findClosestMatch(funcName, availableFunctions, 3)

	if closest != "" {
		return fmt.Sprintf("Did you mean '%s'?", closest)
	}

	// Show available functions
	if len(availableFunctions) <= 10 {
		return fmt.Sprintf("Available functions: %s", strings.Join(availableFunctions, ", "))
	}

	return fmt.Sprintf("Available functions include: %s, ... (%d total)", strings.Join(availableFunctions[:10], ", "), len(availableFunctions))
}

// formatMaxArgs formats the maximum arguments for hints.
func (v *SchemaValidator) formatMaxArgs(maxArgs int) string {
	if maxArgs < 0 {
		return "unlimited"
	}
	return fmt.Sprintf("%d", maxArgs)
}

// findClosestMatch finds the closest matching string using Levenshtein distance.
// Returns empty string if no match is close enough (within maxDistance).
func findClosestMatch(target string, candidates []string, maxDistance int) string {
	if len(candidates) == 0 {
		return ""
	}

	target = strings.ToLower(target)
	closestMatch := ""
	closestDistance := maxDistance + 1

	for _, candidate := range candidates {
		distance := levenshteinDistance(target, strings.ToLower(candidate))
		if distance < closestDistance {
			closestDistance = distance
			closestMatch = candidate
		}
	}

	if closestDistance <= maxDistance {
		return closestMatch
	}

	return ""
}

// levenshteinDistance calculates the Levenshtein distance between two strings.
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create a 2D matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	// Initialize first row and column
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	// Fill the matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

// min returns the minimum of three integers.
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// maxInt returns the maximum of two integers.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// minInt returns the minimum of two integers.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
