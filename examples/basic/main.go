package main

import (
	"fmt"
	"log"

	"github.com/shapestone/shape/internal/parser"
	"github.com/shapestone/shape/pkg/shape"
)

func main() {
	// Example 1: Parse with explicit format
	fmt.Println("=== Example 1: Parse with explicit format ===")
	schema1 := `{
		"id": UUID,
		"name": String(1, 100),
		"email": Email,
		"age": Integer(18, 120)
	}`

	node1, err := shape.Parse(parser.FormatJSONV, schema1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Parsed successfully! Node type: %s\n", node1.Type())
	fmt.Printf("AST: %s\n\n", node1.String())

	// Example 2: Auto-detect format
	fmt.Println("=== Example 2: Auto-detect format ===")
	schema2 := `{
		"user": {
			"profile": {
				"firstName": String(1, 50),
				"lastName": String(1, 50)
			},
			"roles": [String(1, 30)]
		}
	}`

	node2, format, err := shape.ParseAuto(schema2)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Detected format: %s\n", format)
	fmt.Printf("Node type: %s\n", node2.Type())
	fmt.Printf("AST: %s\n\n", node2.String())

	// Example 3: Parse array
	fmt.Println("=== Example 3: Parse array ===")
	schema3 := `[String(1, 30)]`

	node3, err := shape.Parse(parser.FormatJSONV, schema3)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Array schema parsed! Node type: %s\n", node3.Type())
	fmt.Printf("AST: %s\n\n", node3.String())

	// Example 4: Using MustParse
	fmt.Println("=== Example 4: Using MustParse ===")
	schema4 := `{"tags": [String(1, 30)]}`

	node4 := shape.MustParse(parser.FormatJSONV, schema4)
	fmt.Printf("Parsed with MustParse! AST: %s\n", node4.String())
}
