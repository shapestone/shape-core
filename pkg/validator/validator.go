package validator

import (
	"fmt"

	"github.com/shapestone/shape/pkg/ast"
)

// Validator validates schema ASTs for correctness.
type Validator struct {
	knownTypes     map[string]bool
	knownFunctions map[string]FunctionRule
}

// FunctionRule defines validation rules for a function.
type FunctionRule struct {
	MinArgs      int
	MaxArgs      int  // -1 means unlimited
	ValidateArgs func(args []interface{}) error
}

// NewValidator creates a new schema validator with default rules.
func NewValidator() *Validator {
	return &Validator{
		knownTypes:     defaultTypes(),
		knownFunctions: defaultFunctions(),
	}
}

// Validate validates a schema AST.
func (v *Validator) Validate(node ast.SchemaNode) error {
	return node.Accept(v)
}

// VisitLiteral validates a literal node.
func (v *Validator) VisitLiteral(node *ast.LiteralNode) error {
	// Literals are always valid
	return nil
}

// VisitType validates a type node.
func (v *Validator) VisitType(node *ast.TypeNode) error {
	typeName := node.TypeName()
	if !v.knownTypes[typeName] {
		return &ValidationError{
			Position: node.Position(),
			Message:  fmt.Sprintf("unknown type: %s", typeName),
		}
	}
	return nil
}

// VisitFunction validates a function node.
func (v *Validator) VisitFunction(node *ast.FunctionNode) error {
	name := node.Name()
	args := node.Arguments()

	// Check if function is known
	rule, ok := v.knownFunctions[name]
	if !ok {
		return &ValidationError{
			Position: node.Position(),
			Message:  fmt.Sprintf("unknown function: %s", name),
		}
	}

	// Check argument count
	argCount := len(args)
	if argCount < rule.MinArgs {
		return &ValidationError{
			Position: node.Position(),
			Message:  fmt.Sprintf("%s requires at least %d arguments, got %d", name, rule.MinArgs, argCount),
		}
	}
	if rule.MaxArgs >= 0 && argCount > rule.MaxArgs {
		return &ValidationError{
			Position: node.Position(),
			Message:  fmt.Sprintf("%s accepts at most %d arguments, got %d", name, rule.MaxArgs, argCount),
		}
	}

	// Custom argument validation
	if rule.ValidateArgs != nil {
		if err := rule.ValidateArgs(args); err != nil {
			return &ValidationError{
				Position: node.Position(),
				Message:  fmt.Sprintf("%s: %v", name, err),
			}
		}
	}

	return nil
}

// VisitObject validates an object node.
func (v *Validator) VisitObject(node *ast.ObjectNode) error {
	for key, prop := range node.Properties() {
		if err := prop.Accept(v); err != nil {
			return &ValidationError{
				Position: node.Position(),
				Message:  fmt.Sprintf("property %q: %v", key, err),
			}
		}
	}
	return nil
}

// VisitArray validates an array node.
func (v *Validator) VisitArray(node *ast.ArrayNode) error {
	return node.ElementSchema().Accept(v)
}

// ValidationError represents a schema validation error.
type ValidationError struct {
	Position ast.Position
	Message  string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error at line %d, column %d: %s",
		e.Position.Line, e.Position.Column, e.Message)
}

// defaultTypes returns the default set of known types.
func defaultTypes() map[string]bool {
	return map[string]bool{
		"UUID":     true,
		"Email":    true,
		"String":   true,
		"Integer":  true,
		"Float":    true,
		"Boolean":  true,
		"ISO-8601": true,
		"URL":      true,
		"IPv4":     true,
		"IPv6":     true,
		"Date":     true,
		"Time":     true,
		"DateTime": true,
		"JSON":     true,
		"Base64":   true,
	}
}

// defaultFunctions returns the default set of known functions with their rules.
func defaultFunctions() map[string]FunctionRule {
	return map[string]FunctionRule{
		"String": {
			MinArgs:      1,
			MaxArgs:      2,
			ValidateArgs: validateRangeArgs,
		},
		"Integer": {
			MinArgs:      1,
			MaxArgs:      2,
			ValidateArgs: validateRangeArgs,
		},
		"Float": {
			MinArgs:      1,
			MaxArgs:      2,
			ValidateArgs: validateRangeArgs,
		},
		"Enum": {
			MinArgs: 1,
			MaxArgs: -1, // Unlimited
		},
		"Pattern": {
			MinArgs: 1,
			MaxArgs: 1,
			ValidateArgs: func(args []interface{}) error {
				if _, ok := args[0].(string); !ok {
					return fmt.Errorf("argument must be a string pattern")
				}
				return nil
			},
		},
		"Length": {
			MinArgs:      1,
			MaxArgs:      2,
			ValidateArgs: validateRangeArgs,
		},
		"Range": {
			MinArgs:      1,
			MaxArgs:      2,
			ValidateArgs: validateRangeArgs,
		},
	}
}

// validateRangeArgs validates range arguments (min, max) or (min+).
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
