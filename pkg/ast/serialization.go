package ast

import (
	"encoding/json"
	"fmt"
)

// SerializableNode is a helper struct for JSON serialization.
type SerializableNode struct {
	Type       string                 `json:"type"`
	Value      interface{}            `json:"value,omitempty"`
	TypeName   string                 `json:"typeName,omitempty"`
	Name       string                 `json:"name,omitempty"`
	Arguments  []interface{}          `json:"arguments,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Element    interface{}            `json:"element,omitempty"`
	Elements   []interface{}          `json:"elements,omitempty"`
	Position   *Position              `json:"position,omitempty"`
}

// MarshalJSON implements json.Marshaler for LiteralNode.
func (n *LiteralNode) MarshalJSON() ([]byte, error) {
	return json.Marshal(&SerializableNode{
		Type:     "literal",
		Value:    n.value,
		Position: &n.position,
	})
}

// MarshalJSON implements json.Marshaler for TypeNode.
func (n *TypeNode) MarshalJSON() ([]byte, error) {
	return json.Marshal(&SerializableNode{
		Type:     "type",
		TypeName: n.typeName,
		Position: &n.position,
	})
}

// MarshalJSON implements json.Marshaler for FunctionNode.
func (n *FunctionNode) MarshalJSON() ([]byte, error) {
	return json.Marshal(&SerializableNode{
		Type:      "function",
		Name:      n.name,
		Arguments: n.arguments,
		Position:  &n.position,
	})
}

// MarshalJSON implements json.Marshaler for ObjectNode.
func (n *ObjectNode) MarshalJSON() ([]byte, error) {
	props := make(map[string]interface{})
	for k, v := range n.properties {
		props[k] = v
	}

	return json.Marshal(&SerializableNode{
		Type:       "object",
		Properties: props,
		Position:   &n.position,
	})
}

// MarshalJSON implements json.Marshaler for ArrayNode.
func (n *ArrayNode) MarshalJSON() ([]byte, error) {
	return json.Marshal(&SerializableNode{
		Type:     "array",
		Element:  n.elementSchema,
		Position: &n.position,
	})
}

// MarshalJSON implements json.Marshaler for ArrayDataNode.
func (n *ArrayDataNode) MarshalJSON() ([]byte, error) {
	elements := make([]interface{}, len(n.elements))
	for i, elem := range n.elements {
		elements[i] = elem
	}

	return json.Marshal(&SerializableNode{
		Type:     "arraydata",
		Elements: elements,
		Position: &n.position,
	})
}

// UnmarshalSchemaNode unmarshals JSON into a SchemaNode.
func UnmarshalSchemaNode(data []byte) (SchemaNode, error) {
	var sn SerializableNode
	if err := json.Unmarshal(data, &sn); err != nil {
		return nil, err
	}

	pos := ZeroPosition()
	if sn.Position != nil {
		pos = *sn.Position
	}

	switch sn.Type {
	case "literal":
		return NewLiteralNode(sn.Value, pos), nil

	case "type":
		return NewTypeNode(sn.TypeName, pos), nil

	case "function":
		return NewFunctionNode(sn.Name, sn.Arguments, pos), nil

	case "object":
		props := make(map[string]SchemaNode)
		for k, v := range sn.Properties {
			vBytes, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal property %q: %w", k, err)
			}

			node, err := UnmarshalSchemaNode(vBytes)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal property %q: %w", k, err)
			}

			props[k] = node
		}
		return NewObjectNode(props, pos), nil

	case "array":
		elemBytes, err := json.Marshal(sn.Element)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal array element: %w", err)
		}

		elemNode, err := UnmarshalSchemaNode(elemBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal array element: %w", err)
		}

		return NewArrayNode(elemNode, pos), nil

	case "arraydata":
		elements := make([]SchemaNode, len(sn.Elements))
		for i, elem := range sn.Elements {
			elemBytes, err := json.Marshal(elem)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal array element %d: %w", i, err)
			}

			elemNode, err := UnmarshalSchemaNode(elemBytes)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal array element %d: %w", i, err)
			}

			elements[i] = elemNode
		}
		return NewArrayDataNode(elements, pos), nil

	default:
		return nil, fmt.Errorf("unknown node type: %q", sn.Type)
	}
}
