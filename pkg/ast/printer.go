package ast

import (
	"fmt"
	"sort"
	"strings"
)

// PrettyPrint returns a human-readable, indented representation of the AST.
func PrettyPrint(node SchemaNode) string {
	return prettyPrint(node, 0)
}

func prettyPrint(node SchemaNode, indent int) string {
	prefix := strings.Repeat("  ", indent)

	switch n := node.(type) {
	case *LiteralNode:
		return fmt.Sprintf("%sLiteral: %s", prefix, n.String())

	case *TypeNode:
		return fmt.Sprintf("%sType: %s", prefix, n.typeName)

	case *FunctionNode:
		args := make([]string, len(n.arguments))
		for i, arg := range n.arguments {
			switch v := arg.(type) {
			case string:
				if v == "+" {
					args[i] = v
				} else {
					args[i] = fmt.Sprintf("%q", v)
				}
			case nil:
				args[i] = "null"
			default:
				args[i] = fmt.Sprintf("%v", v)
			}
		}
		return fmt.Sprintf("%sFunction: %s(%s)", prefix, n.name, strings.Join(args, ", "))

	case *ObjectNode:
		if len(n.properties) == 0 {
			return fmt.Sprintf("%sObject: {}", prefix)
		}

		// Sort keys for deterministic output
		keys := make([]string, 0, len(n.properties))
		for k := range n.properties {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		lines := []string{fmt.Sprintf("%sObject:", prefix)}
		for _, k := range keys {
			childStr := prettyPrint(n.properties[k], indent+1)
			lines = append(lines, fmt.Sprintf("%s  %q:", prefix, k))
			lines = append(lines, childStr)
		}
		return strings.Join(lines, "\n")

	case *ArrayNode:
		lines := []string{fmt.Sprintf("%sArray:", prefix)}
		lines = append(lines, fmt.Sprintf("%s  element:", prefix))
		lines = append(lines, prettyPrint(n.elementSchema, indent+2))
		return strings.Join(lines, "\n")

	default:
		return fmt.Sprintf("%sUnknown node type", prefix)
	}
}

// TreePrint returns a tree-style representation of the AST.
func TreePrint(node SchemaNode) string {
	return treePrint(node, "", true)
}

func treePrint(node SchemaNode, prefix string, isLast bool) string {
	connector := "└── "
	if !isLast {
		connector = "├── "
	}

	var result strings.Builder
	result.WriteString(prefix)
	result.WriteString(connector)

	switch n := node.(type) {
	case *LiteralNode:
		result.WriteString(fmt.Sprintf("Literal: %s\n", n.String()))

	case *TypeNode:
		result.WriteString(fmt.Sprintf("Type: %s\n", n.typeName))

	case *FunctionNode:
		result.WriteString(fmt.Sprintf("Function: %s\n", n.String()))

	case *ObjectNode:
		result.WriteString("Object\n")

		// Sort keys
		keys := make([]string, 0, len(n.properties))
		for k := range n.properties {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		childPrefix := prefix
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += "│   "
		}

		for i, k := range keys {
			result.WriteString(childPrefix)
			if i == len(keys)-1 {
				result.WriteString("└── ")
			} else {
				result.WriteString("├── ")
			}
			result.WriteString(fmt.Sprintf("%q:\n", k))

			grandChildPrefix := childPrefix
			if i == len(keys)-1 {
				grandChildPrefix += "    "
			} else {
				grandChildPrefix += "│   "
			}

			result.WriteString(treePrint(n.properties[k], grandChildPrefix, true))
		}

	case *ArrayNode:
		result.WriteString("Array\n")

		childPrefix := prefix
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += "│   "
		}

		result.WriteString(treePrint(n.elementSchema, childPrefix, true))

	default:
		result.WriteString("Unknown\n")
	}

	return result.String()
}
